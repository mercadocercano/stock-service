-- Migration: 009_rls_stock (down)
-- Description: Revert RLS + policies (orden inverso al up). Deja las 5 tablas como antes del
-- retrofit: relrowsecurity=f / relforcerowsecurity=f, sin policy tenant_isolation.

DROP POLICY IF EXISTS tenant_isolation ON stock_availability;
ALTER TABLE stock_availability NO FORCE ROW LEVEL SECURITY;
ALTER TABLE stock_availability DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON stock_entries;
ALTER TABLE stock_entries NO FORCE ROW LEVEL SECURITY;
ALTER TABLE stock_entries DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON stock_locations;
ALTER TABLE stock_locations NO FORCE ROW LEVEL SECURITY;
ALTER TABLE stock_locations DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON warehouses;
ALTER TABLE warehouses NO FORCE ROW LEVEL SECURITY;
ALTER TABLE warehouses DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON locations;
ALTER TABLE locations NO FORCE ROW LEVEL SECURITY;
ALTER TABLE locations DISABLE ROW LEVEL SECURITY;