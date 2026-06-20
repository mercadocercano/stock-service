-- Down: irreversible (fix de trigger + creación de índices únicos sobre datos).
-- ADR-001 Param 1.
DO $$ BEGIN RAISE EXCEPTION 'Migration 006 is irreversible (trigger/index fix). Restore from backup if a rollback is required.'; END $$;
