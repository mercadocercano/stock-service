-- Migration: 004_create_stock_entries.sql
-- Description: Crear tablas para entradas de stock y disponibilidad
-- Date: 2026-02-01
-- HITO 2: Stock Entry module

-- =====================================================
-- Tabla: stock_entries
-- Propósito: Registrar entradas/movimientos de stock
-- =====================================================
CREATE TABLE IF NOT EXISTS stock_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    
    -- Referencia al producto (SKU)
    product_sku VARCHAR(255) NOT NULL,
    product_id UUID,  -- Opcional: ID del producto en PIM
    product_name VARCHAR(500),
    
    -- Referencia a location (warehouse, store, etc.)
    location_id UUID,
    
    -- Tipo de movimiento
    entry_type VARCHAR(50) NOT NULL DEFAULT 'initial_stock',  -- initial_stock, purchase, adjustment, transfer_in, transfer_out, sale
    
    -- Cantidades
    quantity DECIMAL(15, 3) NOT NULL CHECK (quantity != 0),
    unit_of_measure VARCHAR(50) DEFAULT 'unit',  -- unit, kg, liter, meter, etc.
    
    -- Costo/Precio de referencia
    unit_cost DECIMAL(15, 2),
    total_cost DECIMAL(15, 2),
    
    -- Metadata
    reference_number VARCHAR(100),  -- Número de orden, factura, etc.
    notes TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Estado
    status VARCHAR(50) DEFAULT 'confirmed',  -- pending, confirmed, cancelled
    is_active BOOLEAN DEFAULT true,
    
    -- Auditoría
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT stock_entries_entry_type_check CHECK (
        entry_type IN ('initial_stock', 'purchase', 'adjustment', 'transfer_in', 'transfer_out', 'sale', 'return')
    ),
    CONSTRAINT stock_entries_status_check CHECK (
        status IN ('pending', 'confirmed', 'cancelled')
    )
);

-- =====================================================
-- Tabla: stock_availability
-- Propósito: Vista consolidada de disponibilidad actual por producto
-- =====================================================
CREATE TABLE IF NOT EXISTS stock_availability (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    
    -- Referencia al producto
    product_sku VARCHAR(255) NOT NULL,
    product_id UUID,
    product_name VARCHAR(500),
    
    -- Location (opcional, NULL = disponibilidad total)
    location_id UUID,
    
    -- Cantidades
    available_quantity DECIMAL(15, 3) NOT NULL DEFAULT 0,
    reserved_quantity DECIMAL(15, 3) NOT NULL DEFAULT 0,
    total_quantity DECIMAL(15, 3) NOT NULL DEFAULT 0,
    unit_of_measure VARCHAR(50) DEFAULT 'unit',
    
    -- Valores agregados
    avg_unit_cost DECIMAL(15, 2),
    total_value DECIMAL(15, 2),
    
    -- Alertas
    min_stock_level DECIMAL(15, 3) DEFAULT 0,
    max_stock_level DECIMAL(15, 3),
    is_low_stock BOOLEAN DEFAULT false,
    is_out_of_stock BOOLEAN DEFAULT false,
    
    -- Metadata
    last_entry_at TIMESTAMP WITH TIME ZONE,
    last_movement_type VARCHAR(50),
    
    -- Auditoría
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT stock_availability_tenant_sku_location_unique UNIQUE (tenant_id, product_sku, location_id)
);

-- =====================================================
-- Índices para optimización
-- =====================================================

-- Stock Entries
CREATE INDEX IF NOT EXISTS idx_stock_entries_tenant_id ON stock_entries(tenant_id);
CREATE INDEX IF NOT EXISTS idx_stock_entries_product_sku ON stock_entries(product_sku);
CREATE INDEX IF NOT EXISTS idx_stock_entries_product_id ON stock_entries(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_entries_tenant_sku ON stock_entries(tenant_id, product_sku);
CREATE INDEX IF NOT EXISTS idx_stock_entries_tenant_location ON stock_entries(tenant_id, location_id);
CREATE INDEX IF NOT EXISTS idx_stock_entries_entry_type ON stock_entries(entry_type);
CREATE INDEX IF NOT EXISTS idx_stock_entries_status ON stock_entries(status);
CREATE INDEX IF NOT EXISTS idx_stock_entries_created_at ON stock_entries(created_at DESC);

-- Stock Availability
CREATE INDEX IF NOT EXISTS idx_stock_availability_tenant_id ON stock_availability(tenant_id);
CREATE INDEX IF NOT EXISTS idx_stock_availability_product_sku ON stock_availability(product_sku);
CREATE INDEX IF NOT EXISTS idx_stock_availability_product_id ON stock_availability(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_availability_location_id ON stock_availability(location_id);
CREATE INDEX IF NOT EXISTS idx_stock_availability_tenant_sku ON stock_availability(tenant_id, product_sku);
CREATE INDEX IF NOT EXISTS idx_stock_availability_low_stock ON stock_availability(tenant_id, is_low_stock) WHERE is_low_stock = true;
CREATE INDEX IF NOT EXISTS idx_stock_availability_out_of_stock ON stock_availability(tenant_id, is_out_of_stock) WHERE is_out_of_stock = true;

-- =====================================================
-- Función: Actualizar stock_availability automáticamente
-- =====================================================
CREATE OR REPLACE FUNCTION update_stock_availability()
RETURNS TRIGGER AS $$
DECLARE
    v_total_quantity DECIMAL(15, 3);
    v_avg_cost DECIMAL(15, 2);
BEGIN
    -- Calcular total de stock para el producto
    SELECT 
        COALESCE(SUM(
            CASE 
                WHEN entry_type IN ('initial_stock', 'purchase', 'adjustment', 'transfer_in', 'return') THEN quantity
                WHEN entry_type IN ('transfer_out', 'sale') THEN -quantity
                ELSE 0
            END
        ), 0),
        COALESCE(AVG(unit_cost), 0)
    INTO v_total_quantity, v_avg_cost
    FROM stock_entries
    WHERE tenant_id = NEW.tenant_id
      AND product_sku = NEW.product_sku
      AND status = 'confirmed'
      AND is_active = true
      AND (location_id = NEW.location_id OR (location_id IS NULL AND NEW.location_id IS NULL));

    -- Insertar o actualizar stock_availability
    INSERT INTO stock_availability (
        tenant_id,
        product_sku,
        product_id,
        product_name,
        location_id,
        available_quantity,
        total_quantity,
        unit_of_measure,
        avg_unit_cost,
        total_value,
        is_low_stock,
        is_out_of_stock,
        last_entry_at,
        last_movement_type,
        updated_at
    ) VALUES (
        NEW.tenant_id,
        NEW.product_sku,
        NEW.product_id,
        NEW.product_name,
        NEW.location_id,
        v_total_quantity,
        v_total_quantity,
        NEW.unit_of_measure,
        v_avg_cost,
        v_total_quantity * v_avg_cost,
        v_total_quantity < 10,  -- Simplificado: low stock si < 10
        v_total_quantity <= 0,
        NEW.created_at,
        NEW.entry_type,
        NOW()
    )
    ON CONFLICT (tenant_id, product_sku, location_id)
    DO UPDATE SET
        product_id = EXCLUDED.product_id,
        product_name = EXCLUDED.product_name,
        available_quantity = EXCLUDED.available_quantity,
        total_quantity = EXCLUDED.total_quantity,
        unit_of_measure = EXCLUDED.unit_of_measure,
        avg_unit_cost = EXCLUDED.avg_unit_cost,
        total_value = EXCLUDED.total_value,
        is_low_stock = EXCLUDED.is_low_stock,
        is_out_of_stock = EXCLUDED.is_out_of_stock,
        last_entry_at = EXCLUDED.last_entry_at,
        last_movement_type = EXCLUDED.last_movement_type,
        updated_at = NOW();

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger para actualizar automáticamente stock_availability
CREATE TRIGGER trigger_update_stock_availability
AFTER INSERT OR UPDATE ON stock_entries
FOR EACH ROW
WHEN (NEW.status = 'confirmed' AND NEW.is_active = true)
EXECUTE FUNCTION update_stock_availability();

-- =====================================================
-- Comentarios para documentación
-- =====================================================
COMMENT ON TABLE stock_entries IS 'Registro de todos los movimientos de entrada/salida de stock';
COMMENT ON TABLE stock_availability IS 'Vista consolidada de disponibilidad actual de stock por producto y location';
COMMENT ON COLUMN stock_entries.entry_type IS 'Tipo de movimiento: initial_stock, purchase, adjustment, transfer_in, transfer_out, sale, return';
COMMENT ON COLUMN stock_availability.available_quantity IS 'Cantidad disponible para venta (total - reserved)';
COMMENT ON COLUMN stock_availability.reserved_quantity IS 'Cantidad reservada en órdenes pendientes';
COMMENT ON COLUMN stock_availability.total_quantity IS 'Cantidad total en stock';

