-- Rollback de migrations/roles/stock_app.sql — PLAT-E30 T-STK3 (best-effort, dev local).
-- Revoca los grants y elimina el rol. DROP ROLE falla si el rol es dueño de objetos o tiene
-- privilegios pendientes en otras DBs — best-effort, mismo patrón que tenant_app/E29.
-- Aplicar: psql ... -d stock_db -f stock_app.down.sql (una sola DB, sin \connect).

REVOKE ALL ON schema_migrations      FROM stock_app;
REVOKE ALL ON stock_availability     FROM stock_app;
REVOKE ALL ON stock_entries          FROM stock_app;
REVOKE ALL ON stock_locations        FROM stock_app;
REVOKE ALL ON warehouses             FROM stock_app;
REVOKE ALL ON locations              FROM stock_app;
REVOKE USAGE ON SCHEMA public       FROM stock_app;
REVOKE CONNECT ON DATABASE stock_db FROM stock_app;
DROP ROLE IF EXISTS stock_app;