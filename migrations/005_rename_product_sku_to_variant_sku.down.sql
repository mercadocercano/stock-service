-- Down: irreversible (rename de columna + swap de unique constraint + backfill de datos).
-- ADR-001 Param 1: las migraciones de datos no se revierten automáticamente.
DO $$ BEGIN RAISE EXCEPTION 'Migration 005 is irreversible (data/constraint migration). Restore from backup if a rollback is required.'; END $$;
