-- Migration: 010_normalize_tenant_id_uuid.sql
-- Description: Normaliza tenant_id de VARCHAR(255) a UUID en locations, warehouses y
-- stock_locations (las 3 tablas que 009 dejó con comparación texto-texto). Unifica la
-- policy tenant_isolation a la forma `tenant_id = current_setting('app.tenant_id')::uuid`
-- usada por stock_entries/stock_availability (patrón E24-E29). Elimina la comparación
-- texto-texto y el cast (tenant_id)::text que 009 instaló sobre las varchar.
-- PLAT-E30 T-STK2b — decisión del owner 2026-08-07.
--
-- Precondición verificada contra lab-postgres 2026-08-07: las 3 tablas están vacías (0
-- filas), así que ALTER TYPE uuid USING tenant_id::uuid no convierte datos ni tiene
-- riesgo de pérdida. ENABLE+FORCE RLS ya aplicados por 009; esta migración NO los toca
-- (no re-ENABLE, no re-FORCE): sólo cambia el tipo de columna y la forma de la policy.
--
-- Orden dentro de cada tabla: DROP policy (depende de la columna) → ALTER TYPE uuid →
-- CREATE policy (forma ::uuid). Postgres rechaza ALTER TYPE de una columna referenciada
-- por una policy, por eso el DROP va primero. El ALTER TYPE reconstruye automáticamente
-- los índices y UNIQUE constraints que dependen de tenant_id:
--   idx_locations_tenant_id, idx_warehouses_tenant_id, idx_stock_locations_tenant_id,
--   UNIQUE(code, tenant_id) en locations/warehouses,
--   UNIQUE(code, warehouse_id, tenant_id) en stock_locations.

-- locations
DROP POLICY IF EXISTS tenant_isolation ON locations;
ALTER TABLE locations ALTER COLUMN tenant_id TYPE uuid USING tenant_id::uuid;
CREATE POLICY tenant_isolation ON locations
  USING      (tenant_id = current_setting('app.tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

-- warehouses
DROP POLICY IF EXISTS tenant_isolation ON warehouses;
ALTER TABLE warehouses ALTER COLUMN tenant_id TYPE uuid USING tenant_id::uuid;
CREATE POLICY tenant_isolation ON warehouses
  USING      (tenant_id = current_setting('app.tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

-- stock_locations
DROP POLICY IF EXISTS tenant_isolation ON stock_locations;
ALTER TABLE stock_locations ALTER COLUMN tenant_id TYPE uuid USING tenant_id::uuid;
CREATE POLICY tenant_isolation ON stock_locations
  USING      (tenant_id = current_setting('app.tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);