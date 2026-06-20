-- Down: irreversible (drop de trigger + índice único; la lógica se movió a Go).
-- ADR-001 Param 1.
DO $$ BEGIN RAISE EXCEPTION 'Migration 008 is irreversible (trigger removed, logic moved to Go). Restore from backup if a rollback is required.'; END $$;
