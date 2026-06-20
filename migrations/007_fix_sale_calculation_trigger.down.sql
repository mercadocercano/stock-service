-- Down: irreversible (corrección de lógica de cálculo en trigger).
-- ADR-001 Param 1.
DO $$ BEGIN RAISE EXCEPTION 'Migration 007 is irreversible (trigger logic fix). Restore from backup if a rollback is required.'; END $$;
