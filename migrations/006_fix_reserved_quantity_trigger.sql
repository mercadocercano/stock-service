-- Migration 006: Fix Reserved Quantity Trigger
-- Fecha: 02/02/2026
-- Descripción: Corrige el trigger update_stock_availability_v2 para que preserve reserved_quantity
--              y no lo sobrescriba cuando se insertan stock_entries

-- Crear índices UNIQUE para evitar duplicados
CREATE UNIQUE INDEX IF NOT EXISTS idx_stock_availability_tenant_variant_location
ON stock_availability (tenant_id, variant_sku, location_id)
WHERE location_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_stock_availability_tenant_variant_no_location
ON stock_availability (tenant_id, variant_sku)
WHERE location_id IS NULL;

-- Eliminar duplicados existentes (quedarse con el más reciente)
DELETE FROM stock_availability sa1
WHERE EXISTS (
    SELECT 1 FROM stock_availability sa2
    WHERE sa1.tenant_id = sa2.tenant_id
      AND sa1.variant_sku = sa2.variant_sku
      AND (sa1.location_id = sa2.location_id OR (sa1.location_id IS NULL AND sa2.location_id IS NULL))
      AND sa1.updated_at < sa2.updated_at
);

-- Recrear función del trigger con lógica corregida
CREATE OR REPLACE FUNCTION public.update_stock_availability_v2()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
DECLARE
    v_total_quantity DECIMAL(15, 3);
    v_avg_cost DECIMAL(15, 2);
    v_existing_reserved DECIMAL(15, 3);
BEGIN
    -- Calcular total de stock para la variante
    -- IMPORTANTE: adjustment puede ser positivo o negativo, usar quantity directamente
    SELECT 
        COALESCE(SUM(
            CASE 
                WHEN entry_type IN ('initial_stock', 'purchase', 'transfer_in', 'return') THEN quantity
                WHEN entry_type IN ('transfer_out', 'sale') THEN quantity  -- Ya es negativo
                WHEN entry_type = 'adjustment' THEN quantity  -- Puede ser positivo o negativo
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

    -- Obtener reserved_quantity existente para NO sobrescribirlo
    SELECT COALESCE(reserved_quantity, 0)
    INTO v_existing_reserved
    FROM stock_availability
    WHERE tenant_id = NEW.tenant_id
      AND variant_sku = NEW.variant_sku
      AND (location_id = NEW.location_id OR (location_id IS NULL AND NEW.location_id IS NULL));

    -- Si no existe, inicializar en 0
    IF v_existing_reserved IS NULL THEN
        v_existing_reserved := 0;
    END IF;

    -- Insertar o actualizar stock_availability
    INSERT INTO stock_availability (
        tenant_id,
        variant_sku,
        product_sku,
        product_id,
        product_name,
        location_id,
        available_quantity,
        reserved_quantity,
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
        v_total_quantity - v_existing_reserved,  -- available = total - reserved
        v_existing_reserved,  -- mantener reserved
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
    ON CONFLICT (tenant_id, variant_sku) 
    WHERE location_id IS NULL
    DO UPDATE SET
        product_sku = EXCLUDED.variant_sku,  -- Mantener sincronizado
        product_id = EXCLUDED.product_id,
        product_name = EXCLUDED.product_name,
        available_quantity = v_total_quantity - stock_availability.reserved_quantity,  -- Preservar reserved
        reserved_quantity = stock_availability.reserved_quantity,  -- NO sobrescribir
        total_quantity = v_total_quantity,
        unit_of_measure = EXCLUDED.unit_of_measure,
        avg_unit_cost = EXCLUDED.avg_unit_cost,
        total_value = v_total_quantity * EXCLUDED.avg_unit_cost,
        is_low_stock = v_total_quantity < 10,
        is_out_of_stock = v_total_quantity <= 0,
        last_entry_at = EXCLUDED.last_entry_at,
        last_movement_type = EXCLUDED.last_movement_type,
        updated_at = NOW();

    RETURN NEW;
END;
$function$;
