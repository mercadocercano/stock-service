-- Create the schema for the stock service

-- Locations table
CREATE TABLE IF NOT EXISTS locations (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(code, tenant_id)
);

-- Warehouses table
CREATE TABLE IF NOT EXISTS warehouses (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    location_id UUID NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(code, tenant_id),
    FOREIGN KEY (location_id) REFERENCES locations(id)
);

-- Stock Locations table
CREATE TABLE IF NOT EXISTS stock_locations (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    warehouse_id UUID NOT NULL,
    parent_id UUID,
    path VARCHAR(4000) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(code, warehouse_id, tenant_id),
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id),
    FOREIGN KEY (parent_id) REFERENCES stock_locations(id)
);

-- Indexes
CREATE INDEX idx_locations_tenant_id ON locations(tenant_id);
CREATE INDEX idx_warehouses_tenant_id ON warehouses(tenant_id);
CREATE INDEX idx_warehouses_location_id ON warehouses(location_id);
CREATE INDEX idx_stock_locations_tenant_id ON stock_locations(tenant_id);
CREATE INDEX idx_stock_locations_warehouse_id ON stock_locations(warehouse_id);
CREATE INDEX idx_stock_locations_parent_id ON stock_locations(parent_id);
CREATE INDEX idx_stock_locations_path ON stock_locations(path); 