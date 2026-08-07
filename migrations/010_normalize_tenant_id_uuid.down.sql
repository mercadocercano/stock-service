-- Migration: 010_normalize_tenant_id_uuid (down)
-- Description: Revert la normalización — vuelve las 3 tablas a VARCHAR(255) y la policy
-- tenant_isolation a la comparación texto-texto que 009 instaló.
--
-- Orden: DROP policy (forma uuid) → ALTER columna de vuelta a varchar → CREATE policy
-- (forma texto-texto). No se puede recrear la policy texto-texto mientras la columna es
-- uuid: `uuid = text` no tiene operador y el CREATE POLICY fallaría.

-- locations
DROP POLICY IF EXISTS tenant_isolation ON locations;
ALTER TABLE locations ALTER COLUMN tenant_id TYPE character varying(255) USING tenant_id::text;
CREATE POLICY tenant_isolation ON locations
  USING      (tenant_id = current_setting('app.tenant_id'))
  WITH CHECK (tenant_id = current_setting('app.tenant_id'));

-- warehouses
DROP POLICY IF EXISTS tenant_isolation ON warehouses;
ALTER TABLE warehouses ALTER COLUMN tenant_id TYPE character varying(255) USING tenant_id::text;
CREATE POLICY tenant_isolation ON warehouses
  USING      (tenant_id = current_setting('app.tenant_id'))
  WITH CHECK (tenant_id = current_setting('app.tenant_id'));

-- stock_locations
DROP POLICY IF EXISTS tenant_isolation ON stock_locations;
ALTER TABLE stock_locations ALTER COLUMN tenant_id TYPE character varying(255) USING tenant_id::text;
CREATE POLICY tenant_isolation ON stock_locations
  USING      (tenant_id = current_setting('app.tenant_id'))
  WITH CHECK (tenant_id = current_setting('app.tenant_id'));