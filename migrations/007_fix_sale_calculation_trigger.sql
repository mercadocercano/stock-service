-- Migration 007: Fix sale calculation in stock_availability trigger
-- PROBLEMA: El trigger suma las ventas en lugar de restarlas
-- SOLUCIÓN: Cambiar THEN quantity por THEN -quantity para sale y transfer_out

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
    -- CORREGIDO: sale y transfer_out deben RESTAR (quantity es positivo en tabla)
    SELECT 
        COALESCE(SUM(
            CASE 
                WHEN entry_type IN ('initial_stock', 'purchase', 'transfer_in', 'return') THEN quantity
                WHEN entry_type IN ('transfer_out', 'sale') THEN -quantity  -- NEGATIVO
                WHEN entry_type = 'adjustment' THEN quantity  -- Puede ser +/-
                ELSE 0
            END
        ), 0),
        COALESCE(AVG(unit_cost), 0)
    INTO v_total_quantity, v_avg_cost
    FROM stock_entries
    WHERE tenant_id = NEW.tenant_id
      AND variant_sku = NEW.variant_sku
      AND status = 'confirmed'
      AND (NEW.location_id IS NULL OR location_id = NEW.location_id OR location_id IS NULL);

    -- Obtener reserved_quantity existente para preservarlo
    SELECT COALESCE(reserved_quantity, 0)
    INTO v_existing_reserved
    FROM stock_availability
    WHERE tenant_id = NEW.tenant_id
      AND variant_sku = NEW.variant_sku
      AND (location_id = NEW.location_id OR (location_id IS NULL AND NEW.location_id IS NULL))
    LIMIT 1;

    -- Insertar o actualizar stock_availability (simplificado - solo sin location)
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
        NEW.variant_sku,
        NEW.product_id,
        NEW.product_name,
        NULL,
        v_total_quantity - v_existing_reserved,
        v_existing_reserved,
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
        available_quantity = v_total_quantity - v_existing_reserved,
        reserved_quantity = v_existing_reserved,
        total_quantity = v_total_quantity,
        unit_of_measure = NEW.unit_of_measure,
        avg_unit_cost = v_avg_cost,
        total_value = v_total_quantity * v_avg_cost,
        is_low_stock = v_total_quantity < 10,
        is_out_of_stock = v_total_quantity <= 0,
        last_entry_at = NEW.created_at,
        last_movement_type = NEW.entry_type,
        updated_at = NOW();

    RETURN NEW;
END;
$function$;

-- Recrear trigger
DROP TRIGGER IF EXISTS trigger_update_stock_availability ON stock_entries;

CREATE TRIGGER trigger_update_stock_availability
AFTER INSERT OR UPDATE ON stock_entries
FOR EACH ROW
WHEN (NEW.status = 'confirmed' AND NEW.is_active = true)
EXECUTE FUNCTION update_stock_availability_v2();

COMMENT ON FUNCTION update_stock_availability_v2() IS 'HITO A - Trigger corregido: sale y transfer_out restan stock';
