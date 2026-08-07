-- migrations/roles/stock_app.sql — PLAT-E30 T-STK3 (RULE-09/RULE-10: rol de aplicación least-privilege, NOBYPASSRLS)
--
-- stock-service corre hoy en runtime como `postgres` (superuser → BYPASSRLS: FORCE ROW LEVEL
-- SECURITY no aplica a superusers, así que la RLS de la migración 009 quedaría "activa" pero nunca
-- ejercida). Este rol least-privilege lo reemplaza en runtime. El flip del compose
-- (DB_USER: ${STOCK_DB_USER:-stock_app}) y el fail-fast anti-superuser en main.go se DIFIEREN a
-- T-STK6 — flipear antes de T-STK4/T-STK5 rompe toda lectura: los 19 sitios aún leen el header
-- crudo (D1) y los repos aún no fijan el GUC app.tenant_id (WithRLSInTransaction, T-STK5).
--
-- Permisos (least-privilege, mapeado del DML real de los repositorios Go):
--   • SELECT/INSERT/UPDATE/DELETE sobre locations, warehouses, stock_locations — CRUD completo
--     (postgres_{location,warehouse,stock_location}_repository.go: Create/Get/Update/Delete).
--   • SELECT/INSERT/UPDATE sobre stock_entries — SIN DELETE (postgres_stock_entry_repository.go:
--     inserts, lecturas, updates; el compensatorio no borra filas, ajusta cantidades).
--   • SELECT/INSERT/UPDATE sobre stock_availability — SIN DELETE (recalcAvailability Go: upsert
--     ON CONFLICT + updates; el runtime nunca DELETEa availability).
--   • SELECT sobre schema_migrations — el runtime solo chequea la versión al arrancar
--     (sharedmigrate.RunMigrations); no aplica DDL nuevo bajo este rol (NOCREATEROLE → las
--     migraciones numeradas se aplican out-of-band por un admin, igual que E24..E29).
--   • NOBYPASSRLS: la policy tenant_isolation de la migración 009 se ejerce de verdad.
--
-- Una sola conexión (stock_db) — a diferencia de tenant-service (E29), stock NO publica eventos:
-- 0 referencias a event_bus/event_consumers/EventStore/EventWorker en el código (verificado
-- 2026-08-07). Sin segunda conexión, sin grants sobre eventbus.
--
-- Sin secuencias en stock_db (IDs son UUID: app-side en locations/warehouses/stock_locations,
-- gen_random_uuid() en stock_entries/stock_availability — verificado 2026-08-07): sin
-- GRANT USAGE ON SEQUENCES.
--
-- Triggers: NO hay triggers ni funciones trigger vivas en stock_db (verificado 2026-08-07). La
-- migración 008 dropeó trigger_update_stock_availability y moveió el recálculo de availability
-- al código Go (recalcAvailability). El enunciado de C2 (T-STK7) los nombra como "lógica de
-- negocio viva" — hoy esa lógica vive en el repositorio Go, no en triggers. El rol concede los
-- DML que ese repositorio ejecuta directamente; sin triggers no hay writes cross-tabla implícitos
-- que requieran grants adicionales.
--
-- NO es una migración numerada del boot: vive en migrations/roles/ (fuera del alcance de
-- `//go:embed migrations/*.sql` en migrations_embed.go, glob NO recursivo → este archivo no entra
-- al FS embebido) para que NO se auto-aplique en el arranque bajo stock_app (NOCREATEROLE no podría
-- crear roles). Se aplica UNA sola vez, out-of-band, contra el cluster por un admin (postgres).
-- Idempotente (DO $$ IF NOT EXISTS). Una sola DB → sin \connect, puede ir con --single-transaction.
--
-- Seguridad (mismo patrón que tenant_app/E29, customer_app/E27, onboarding_app/E28): este script
-- NO fija password —un literal quedaría en el git history para siempre—. El rol se crea con LOGIN
-- pero SIN password (login efectivo deshabilitado hasta setearla); la password real se fija
-- out-of-band vía `ALTER ROLE stock_app PASSWORD '...'` corrido a mano contra lab-postgres, nunca
-- versionada (dev local: stock_app123, ver .env.example / docker-compose STOCK_DB_PASSWORD).
--
-- Aplicar: psql ... -d stock_db -f stock_app.sql (ON_ERROR_STOP=1; --single-transaction OK).

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'stock_app') THEN
    CREATE ROLE stock_app LOGIN
      NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS;
  END IF;
END
$$;

GRANT CONNECT ON DATABASE stock_db TO stock_app;
GRANT USAGE ON SCHEMA public TO stock_app;

-- Hardening least-privilege (serie E27/E29): revocar privilegios de tabla que PUBLIC pudiera
-- conceder implícitamente; el rol solo obtiene lo que se le concede explícito abajo.
-- (Baseline 2026-08-07: PUBLIC sin grants en stock_db — defensivo de todas formas.)
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;

-- CRUD completo (Create/Get/Update/Delete en los repositorios).
GRANT SELECT, INSERT, UPDATE, DELETE ON locations      TO stock_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON warehouses      TO stock_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON stock_locations  TO stock_app;

-- Sin DELETE (los repos nunca borran filas de estas dos tablas).
GRANT SELECT, INSERT, UPDATE ON stock_entries       TO stock_app;
GRANT SELECT, INSERT, UPDATE ON stock_availability  TO stock_app;

-- El runtime solo chequea la versión al arrancar (no aplica DDL nuevo → SELECT alcanza).
GRANT SELECT ON schema_migrations TO stock_app;