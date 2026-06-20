-- Down: revierte 001_initial_schema. DROP en cascada elimina índices y constraints.
DROP TABLE IF EXISTS stock_locations CASCADE;
DROP TABLE IF EXISTS warehouses CASCADE;
DROP TABLE IF EXISTS locations CASCADE;
