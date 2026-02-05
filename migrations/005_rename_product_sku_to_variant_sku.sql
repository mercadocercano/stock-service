-- Migration: 005_rename_product_sku_to_variant_sku.sql
-- Description: Renombrar product_sku a variant_sku para claridad del modelo (HITO 2.1)
-- Date: 2026-02-01
-- Mantiene compatibilidad con código existente usando alias

-- =====================================================
-- PASO 1: Agregar nueva columna variant_sku
-- =====================================================

ALTER TABLE stock_entries 
ADD COLUMN IF NOT EXISTS variant_sku VARCHAR(255);

ALTER TABLE stock_availability 
ADD COLUMN IF NOT EXISTS variant_sku VARCHAR(255);

-- =====================================================
-- PASO 2: Migrar datos existentes
-- =====================================================

-- Copiar product_sku a variant_sku en registros existentes
UPDATE stock_entries 
SET variant_sku = product_sku 
WHERE variant_sku IS NULL;

UPDATE stock_availability 
SET variant_sku = product_sku 
WHERE variant_sku IS NULL;

-- =====================================================
-- PASO 3: Crear índices para variant_sku
-- =====================================================

CREATE INDEX IF NOT EXISTS idx_stock_entries_variant_sku ON stock_entries(variant_sku);
CREATE INDEX IF NOT EXISTS idx_stock_entries_tenant_variant ON stock_entries(tenant_id, variant_sku);

CREATE INDEX IF NOT EXISTS idx_stock_availability_variant_sku ON stock_availability(variant_sku);
CREATE INDEX IF NOT EXISTS idx_stock_availability_tenant_variant ON stock_availability(tenant_id, variant_sku);

-- =====================================================
-- PASO 4: Actualizar unique constraint en stock_availability
-- =====================================================

-- Eliminar constraint anterior
ALTER TABLE stock_availability 
DROP CONSTRAINT IF EXISTS stock_availability_tenant_sku_location_unique;

-- Agregar nuevo constraint con variant_sku
ALTER TABLE stock_availability 
ADD CONSTRAINT stock_availability_tenant_variant_location_unique 
UNIQUE (tenant_id, variant_sku, location_id);

-- =====================================================
-- PASO 5: Actualizar trigger para usar variant_sku
-- =====================================================

CREATE OR REPLACE FUNCTION update_stock_availability_v2()
RETURNS TRIGGER AS $$
DECLARE
    v_total_quantity DECIMAL(15, 3);
    v_avg_cost DECIMAL(15, 2);
BEGIN
    -- Calcular total de stock para la variante
    SELECT 
        COALESCE(SUM(
            CASE 
                WHEN entry_type IN ('initial_stock', 'purchase', 'adjustment', 'transfer_in', 'return') THEN quantity
                WHEN entry_type IN ('transfer_out', 'sale') THEN -quantity
                ELSE 0
            END
        ), 0),
        COALESCE(AVG(unit_cost), 0)
    INTO v_total_quantity, v_avg_cost
    FROM stock_entries
    WHERE tenant_id = NEW.tenant_id
      AND variant_sku = NEW.variant_sku
      AND status = 'confirmed'
      AND is_active = true
      AND (location_id = NEW.location_id OR (location_id IS NULL AND NEW.location_id IS NULL));

    -- Insertar o actualizar stock_availability
    INSERT INTO stock_availability (
        tenant_id,
        variant_sku,
        product_sku,  -- Mantener por compatibilidad
        product_id,
        product_name,
        location_id,
        available_quantity,
        total_quantity,
        unit_of_measure,
        avg_unit_cost,
        total_value,
        is_low_stock,
        is_out_of_stock,
        last_entry_at,
        last_movement_type,
        updated_at
    ) VALUES (
        NEW.tenant_id,
        NEW.variant_sku,
        NEW.variant_sku,  -- Copiar a product_sku por compatibilidad
        NEW.product_id,
        NEW.product_name,
        NEW.location_id,
        v_total_quantity,
        v_total_quantity,
        NEW.unit_of_measure,
        v_avg_cost,
        v_total_quantity * v_avg_cost,
        v_total_quantity < 10,
        v_total_quantity <= 0,
        NEW.created_at,
        NEW.entry_type,
        NOW()
    )
    ON CONFLICT (tenant_id, variant_sku, location_id)
    DO UPDATE SET
        product_sku = EXCLUDED.variant_sku,  -- Mantener sincronizado
        product_id = EXCLUDED.product_id,
        product_name = EXCLUDED.product_name,
        available_quantity = EXCLUDED.available_quantity,
        total_quantity = EXCLUDED.total_quantity,
        unit_of_measure = EXCLUDED.unit_of_measure,
        avg_unit_cost = EXCLUDED.avg_unit_cost,
        total_value = EXCLUDED.total_value,
        is_low_stock = EXCLUDED.is_low_stock,
        is_out_of_stock = EXCLUDED.is_out_of_stock,
        last_entry_at = EXCLUDED.last_entry_at,
        last_movement_type = EXCLUDED.last_movement_type,
        updated_at = NOW();

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Reemplazar trigger existente con la nueva función
DROP TRIGGER IF EXISTS trigger_update_stock_availability ON stock_entries;

CREATE TRIGGER trigger_update_stock_availability
AFTER INSERT OR UPDATE ON stock_entries
FOR EACH ROW
WHEN (NEW.status = 'confirmed' AND NEW.is_active = true)
EXECUTE FUNCTION update_stock_availability_v2();

-- =====================================================
-- COMENTARIOS
-- =====================================================

COMMENT ON COLUMN stock_entries.variant_sku IS 'SKU de la variante del producto (campo principal desde HITO 2.1)';
COMMENT ON COLUMN stock_entries.product_sku IS 'Alias de variant_sku para compatibilidad hacia atrás';
COMMENT ON COLUMN stock_availability.variant_sku IS 'SKU de la variante del producto (campo principal desde HITO 2.1)';
COMMENT ON COLUMN stock_availability.product_sku IS 'Alias de variant_sku para compatibilidad hacia atrás';

