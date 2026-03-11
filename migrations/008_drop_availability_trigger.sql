-- ============================================================================
-- Migration 008: Eliminar trigger de stock_availability
-- La lógica de recálculo se mueve al código Go (repositorio).
-- ============================================================================

DROP TRIGGER IF EXISTS trigger_update_stock_availability ON stock_entries;
DROP FUNCTION IF EXISTS update_stock_availability_v2();
DROP FUNCTION IF EXISTS update_stock_availability();

-- Índice parcial único para ON CONFLICT en recalcAvailability (Go).
-- El constraint existente UNIQUE(tenant_id, variant_sku, location_id) no detecta
-- conflictos cuando location_id IS NULL (NULLs son distintos en PG).
CREATE UNIQUE INDEX IF NOT EXISTS idx_stock_availability_tenant_variant_no_loc
    ON stock_availability (tenant_id, variant_sku)
    WHERE location_id IS NULL;
