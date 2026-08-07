-- Migration: 009_rls_stock.sql
-- Description: RLS fail-closed en locations, warehouses, stock_locations, stock_entries y
-- stock_availability (enable + force + policy tenant_isolation con USING y WITH CHECK, sin
-- caso global, sin break_glass).
-- PLAT-E30 T-STK2 — RLS retrofit RULE-09/RULE-10. El GRANT del rol de aplicación vive en el
-- DDL de roles (migrations/roles/), NO acá: las migraciones numeradas quedan role-agnostic,
-- igual que PLAT-E24..E29.
--
-- Nota de tipos (stock es el primer servicio de la serie con tenant_id MIXTO): a diferencia de
-- los hermanos E24-E29 (todas UUID), stock lleva tenant_id VARCHAR(255) en
-- locations/warehouses/stock_locations y UUID en stock_entries/stock_availability. La policy
-- compara contra current_setting('app.tenant_id'), que el app setea como texto (go-shared
-- postgres.SetRLSContextTx / WithRLSInTransaction): cast ::uuid SÓLO en las columnas uuid;
-- comparación texto-texto en las varchar (no introduce cast-failure modes sobre datos legacy).
--
-- Precondición verificada 2026-08-07 contra lab-postgres: 0 filas con tenant_id NULL en las 5
-- (las 5 son NOT NULL por DDL 001/004). schema_migrations queda fuera (bookkeeping, sin
-- tenant_id). Las 5 tablas pasan de relrowsecurity=f/relforcerowsecurity=f a t/t.

ALTER TABLE locations ENABLE ROW LEVEL SECURITY;
ALTER TABLE locations FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON locations;
CREATE POLICY tenant_isolation ON locations
  USING      (tenant_id = current_setting('app.tenant_id'))
  WITH CHECK (tenant_id = current_setting('app.tenant_id'));

ALTER TABLE warehouses ENABLE ROW LEVEL SECURITY;
ALTER TABLE warehouses FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON warehouses;
CREATE POLICY tenant_isolation ON warehouses
  USING      (tenant_id = current_setting('app.tenant_id'))
  WITH CHECK (tenant_id = current_setting('app.tenant_id'));

ALTER TABLE stock_locations ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_locations FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON stock_locations;
CREATE POLICY tenant_isolation ON stock_locations
  USING      (tenant_id = current_setting('app.tenant_id'))
  WITH CHECK (tenant_id = current_setting('app.tenant_id'));

ALTER TABLE stock_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_entries FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON stock_entries;
CREATE POLICY tenant_isolation ON stock_entries
  USING      (tenant_id = current_setting('app.tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

ALTER TABLE stock_availability ENABLE ROW LEVEL SECURITY;
ALTER TABLE stock_availability FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON stock_availability;
CREATE POLICY tenant_isolation ON stock_availability
  USING      (tenant_id = current_setting('app.tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);