package persistence

import (
	"context"
	"database/sql"
	"fmt"
	
	"github.com/google/uuid"
	// "github.com/lib/pq" // No usado actualmente
	
	"stock/src/stock_entry/domain/entity"
	"stock/src/stock_entry/domain/exception"
	"stock/src/stock_entry/domain/port"
)

// PostgresStockEntryRepository implementación PostgreSQL del repositorio
type PostgresStockEntryRepository struct {
	db *sql.DB
}

// NewPostgresStockEntryRepository crea una nueva instancia
func NewPostgresStockEntryRepository(db *sql.DB) port.StockEntryRepository {
	return &PostgresStockEntryRepository{db: db}
}

// Save guarda una entrada de stock (HITO 2.1 - con variant_sku)
func (r *PostgresStockEntryRepository) Save(ctx context.Context, entry *entity.StockEntry) error {
	query := `
		INSERT INTO stock_entries (
			id, tenant_id, variant_sku, product_sku, product_id, product_name, location_id,
			entry_type, quantity, unit_of_measure, unit_cost, total_cost,
			reference_number, notes, status, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		entry.ID,
		entry.TenantID,
		entry.VariantSKU,
		entry.VariantSKU,  // Copiar a product_sku por compatibilidad
		entry.ProductID,
		entry.ProductName,
		entry.LocationID,
		entry.EntryType,
		entry.Quantity,
		entry.UnitOfMeasure,
		entry.UnitCost,
		entry.TotalCost,
		entry.ReferenceNumber,
		entry.Notes,
		entry.Status,
		entry.IsActive,
		entry.CreatedAt,
		entry.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("error saving stock entry: %w", err)
	}
	
	return nil
}

// SaveBulk guarda múltiples entradas (HITO 2.1 - con variant_sku)
func (r *PostgresStockEntryRepository) SaveBulk(ctx context.Context, entries []*entity.StockEntry) error {
	if len(entries) == 0 {
		return nil
	}
	
	// Usar transacción para bulk insert
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO stock_entries (
			id, tenant_id, variant_sku, product_sku, product_id, product_name, location_id,
			entry_type, quantity, unit_of_measure, unit_cost, total_cost,
			reference_number, notes, status, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	
	for _, entry := range entries {
		_, err = stmt.ExecContext(ctx,
			entry.ID,
			entry.TenantID,
			entry.VariantSKU,
			entry.VariantSKU,  // Copiar a product_sku
			entry.ProductID,
			entry.ProductName,
			entry.LocationID,
			entry.EntryType,
			entry.Quantity,
			entry.UnitOfMeasure,
			entry.UnitCost,
			entry.TotalCost,
			entry.ReferenceNumber,
			entry.Notes,
			entry.Status,
			entry.IsActive,
			entry.CreatedAt,
			entry.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("error saving entry for variant SKU %s: %w", entry.VariantSKU, err)
		}
	}
	
	return tx.Commit()
}

// FindByID busca una entrada por ID
func (r *PostgresStockEntryRepository) FindByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*entity.StockEntry, error) {
	query := `
		SELECT id, tenant_id, variant_sku, product_id, product_name, location_id,
			   entry_type, quantity, unit_of_measure, unit_cost, total_cost,
			   reference_number, notes, status, is_active, created_at, updated_at
		FROM stock_entries
		WHERE id = $1 AND tenant_id = $2
	`
	
	entry := &entity.StockEntry{}
	err := r.db.QueryRowContext(ctx, query, id, tenantID).Scan(
		&entry.ID,
		&entry.TenantID,
		&entry.VariantSKU,
		&entry.ProductID,
		&entry.ProductName,
		&entry.LocationID,
		&entry.EntryType,
		&entry.Quantity,
		&entry.UnitOfMeasure,
		&entry.UnitCost,
		&entry.TotalCost,
		&entry.ReferenceNumber,
		&entry.Notes,
		&entry.Status,
		&entry.IsActive,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)
	
	// Mantener product_sku sincronizado
	entry.ProductSKU = entry.VariantSKU
	
	if err == sql.ErrNoRows {
		return nil, exception.ErrStockEntryNotFound
	}
	if err != nil {
		return nil, err
	}
	
	return entry, nil
}

// FindByTenantAndSKU busca entradas por tenant y variant SKU
func (r *PostgresStockEntryRepository) FindByTenantAndSKU(ctx context.Context, tenantID uuid.UUID, variantSKU string) ([]*entity.StockEntry, error) {
	query := `
		SELECT id, tenant_id, variant_sku, product_id, product_name, location_id,
			   entry_type, quantity, unit_of_measure, unit_cost, total_cost,
			   reference_number, notes, status, is_active, created_at, updated_at
		FROM stock_entries
		WHERE tenant_id = $1 AND (variant_sku = $2 OR product_sku = $2) AND is_active = true
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, tenantID, variantSKU)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	entries := make([]*entity.StockEntry, 0)
	for rows.Next() {
		entry := &entity.StockEntry{}
		err := rows.Scan(
			&entry.ID,
			&entry.TenantID,
			&entry.VariantSKU,
			&entry.ProductID,
			&entry.ProductName,
			&entry.LocationID,
			&entry.EntryType,
			&entry.Quantity,
			&entry.UnitOfMeasure,
			&entry.UnitCost,
			&entry.TotalCost,
			&entry.ReferenceNumber,
			&entry.Notes,
			&entry.Status,
			&entry.IsActive,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		entry.ProductSKU = entry.VariantSKU  // Mantener sincronizado
		entries = append(entries, entry)
	}
	
	return entries, nil
}

// FindByTenant busca entradas por tenant con paginación
func (r *PostgresStockEntryRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.StockEntry, error) {
	query := `
		SELECT id, tenant_id, variant_sku, product_id, product_name, location_id,
			   entry_type, quantity, unit_of_measure, unit_cost, total_cost,
			   reference_number, notes, status, is_active, created_at, updated_at
		FROM stock_entries
		WHERE tenant_id = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	entries := make([]*entity.StockEntry, 0)
	for rows.Next() {
		entry := &entity.StockEntry{}
		err := rows.Scan(
			&entry.ID,
			&entry.TenantID,
			&entry.VariantSKU,
			&entry.ProductID,
			&entry.ProductName,
			&entry.LocationID,
			&entry.EntryType,
			&entry.Quantity,
			&entry.UnitOfMeasure,
			&entry.UnitCost,
			&entry.TotalCost,
			&entry.ReferenceNumber,
			&entry.Notes,
			&entry.Status,
			&entry.IsActive,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		entry.ProductSKU = entry.VariantSKU  // Mantener sincronizado
		entries = append(entries, entry)
	}
	
	return entries, nil
}

// Delete soft delete de una entrada
func (r *PostgresStockEntryRepository) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	query := `
		UPDATE stock_entries
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`
	
	result, err := r.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		return err
	}
	
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return exception.ErrStockEntryNotFound
	}
	
	return nil
}

// PostgresStockAvailabilityRepository implementación PostgreSQL
type PostgresStockAvailabilityRepository struct {
	db *sql.DB
}

// NewPostgresStockAvailabilityRepository crea una nueva instancia
func NewPostgresStockAvailabilityRepository(db *sql.DB) port.StockAvailabilityRepository {
	return &PostgresStockAvailabilityRepository{db: db}
}

// FindByTenantAndSKU busca disponibilidad por tenant y variant SKU (HITO 2.1)
func (r *PostgresStockAvailabilityRepository) FindByTenantAndSKU(ctx context.Context, tenantID uuid.UUID, variantSKU string) (*entity.StockAvailability, error) {
	query := `
		SELECT id, tenant_id, variant_sku, product_id, product_name, location_id,
			   available_quantity, reserved_quantity, total_quantity, unit_of_measure,
			   avg_unit_cost, total_value, min_stock_level, max_stock_level,
			   is_low_stock, is_out_of_stock, last_entry_at, updated_at
		FROM stock_availability
		WHERE tenant_id = $1 AND (variant_sku = $2 OR product_sku = $2)
		ORDER BY updated_at DESC
		LIMIT 1
	`
	
	availability := &entity.StockAvailability{}
	
	err := r.db.QueryRowContext(ctx, query, tenantID, variantSKU).Scan(
		&availability.ID,
		&availability.TenantID,
		&availability.VariantSKU,
		&availability.ProductID,
		&availability.ProductName,
		&availability.LocationID,
		&availability.AvailableQuantity,
		&availability.ReservedQuantity,
		&availability.TotalQuantity,
		&availability.UnitOfMeasure,
		&availability.AvgUnitCost,
		&availability.TotalValue,
		&availability.MinStockLevel,
		&availability.MaxStockLevel,
		&availability.IsLowStock,
		&availability.IsOutOfStock,
		&availability.LastEntryAt,
		&availability.UpdatedAt,
	)
	
	// Mantener product_sku sincronizado
	availability.ProductSKU = availability.VariantSKU
	
	if err == sql.ErrNoRows {
		return nil, exception.ErrStockAvailabilityNotFound
	}
	if err != nil {
		return nil, err
	}
	
	return availability, nil
}

// FindByTenant busca disponibilidad por tenant
func (r *PostgresStockAvailabilityRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.StockAvailability, error) {
	query := `
		SELECT id, tenant_id, product_sku, product_id, product_name, location_id,
			   available_quantity, reserved_quantity, total_quantity, unit_of_measure,
			   avg_unit_cost, total_value, min_stock_level, max_stock_level,
			   is_low_stock, is_out_of_stock, last_entry_at, updated_at
		FROM stock_availability
		WHERE tenant_id = $1
		ORDER BY product_sku
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	availabilities := make([]*entity.StockAvailability, 0)
	for rows.Next() {
		a := &entity.StockAvailability{}
		err := rows.Scan(
			&a.ID, &a.TenantID, &a.ProductSKU, &a.ProductID, &a.ProductName, &a.LocationID,
			&a.AvailableQuantity, &a.ReservedQuantity, &a.TotalQuantity, &a.UnitOfMeasure,
			&a.AvgUnitCost, &a.TotalValue, &a.MinStockLevel, &a.MaxStockLevel,
			&a.IsLowStock, &a.IsOutOfStock, &a.LastEntryAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		availabilities = append(availabilities, a)
	}
	
	return availabilities, nil
}

// FindLowStock busca productos con bajo stock
func (r *PostgresStockAvailabilityRepository) FindLowStock(ctx context.Context, tenantID uuid.UUID) ([]*entity.StockAvailability, error) {
	query := `
		SELECT id, tenant_id, product_sku, product_id, product_name, location_id,
			   available_quantity, reserved_quantity, total_quantity, unit_of_measure,
			   avg_unit_cost, total_value, min_stock_level, max_stock_level,
			   is_low_stock, is_out_of_stock, last_entry_at, updated_at
		FROM stock_availability
		WHERE tenant_id = $1 AND is_low_stock = true
		ORDER BY available_quantity ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	availabilities := make([]*entity.StockAvailability, 0)
	for rows.Next() {
		a := &entity.StockAvailability{}
		err := rows.Scan(
			&a.ID, &a.TenantID, &a.ProductSKU, &a.ProductID, &a.ProductName, &a.LocationID,
			&a.AvailableQuantity, &a.ReservedQuantity, &a.TotalQuantity, &a.UnitOfMeasure,
			&a.AvgUnitCost, &a.TotalValue, &a.MinStockLevel, &a.MaxStockLevel,
			&a.IsLowStock, &a.IsOutOfStock, &a.LastEntryAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		availabilities = append(availabilities, a)
	}
	
	return availabilities, nil
}

// FindOutOfStock busca productos sin stock
func (r *PostgresStockAvailabilityRepository) FindOutOfStock(ctx context.Context, tenantID uuid.UUID) ([]*entity.StockAvailability, error) {
	query := `
		SELECT id, tenant_id, product_sku, product_id, product_name, location_id,
			   available_quantity, reserved_quantity, total_quantity, unit_of_measure,
			   avg_unit_cost, total_value, min_stock_level, max_stock_level,
			   is_low_stock, is_out_of_stock, last_entry_at, updated_at
		FROM stock_availability
		WHERE tenant_id = $1 AND is_out_of_stock = true
		ORDER BY product_sku
	`
	
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	availabilities := make([]*entity.StockAvailability, 0)
	for rows.Next() {
		a := &entity.StockAvailability{}
		err := rows.Scan(
			&a.ID, &a.TenantID, &a.ProductSKU, &a.ProductID, &a.ProductName, &a.LocationID,
			&a.AvailableQuantity, &a.ReservedQuantity, &a.TotalQuantity, &a.UnitOfMeasure,
			&a.AvgUnitCost, &a.TotalValue, &a.MinStockLevel, &a.MaxStockLevel,
			&a.IsLowStock, &a.IsOutOfStock, &a.LastEntryAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		availabilities = append(availabilities, a)
	}
	
	return availabilities, nil
}

// Save guarda o actualiza disponibilidad
func (r *PostgresStockAvailabilityRepository) Save(ctx context.Context, availability *entity.StockAvailability) error {
	query := `
		INSERT INTO stock_availability (
			id, tenant_id, product_sku, product_id, product_name, location_id,
			available_quantity, reserved_quantity, total_quantity, unit_of_measure,
			avg_unit_cost, total_value, min_stock_level, max_stock_level,
			is_low_stock, is_out_of_stock, last_entry_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT (tenant_id, product_sku, location_id)
		DO UPDATE SET
			available_quantity = EXCLUDED.available_quantity,
			reserved_quantity = EXCLUDED.reserved_quantity,
			total_quantity = EXCLUDED.total_quantity,
			avg_unit_cost = EXCLUDED.avg_unit_cost,
			total_value = EXCLUDED.total_value,
			is_low_stock = EXCLUDED.is_low_stock,
			is_out_of_stock = EXCLUDED.is_out_of_stock,
			updated_at = EXCLUDED.updated_at
	`
	
	_, err := r.db.ExecContext(ctx, query,
		availability.ID, availability.TenantID, availability.ProductSKU,
		availability.ProductID, availability.ProductName, availability.LocationID,
		availability.AvailableQuantity, availability.ReservedQuantity, availability.TotalQuantity,
		availability.UnitOfMeasure, availability.AvgUnitCost, availability.TotalValue,
		availability.MinStockLevel, availability.MaxStockLevel,
		availability.IsLowStock, availability.IsOutOfStock,
		availability.LastEntryAt, availability.UpdatedAt,
	)
	
	return err
}

// Update actualiza disponibilidad existente
func (r *PostgresStockAvailabilityRepository) Update(ctx context.Context, availability *entity.StockAvailability) error {
	query := `
		UPDATE stock_availability SET
			available_quantity = $1,
			reserved_quantity = $2,
			total_quantity = $3,
			avg_unit_cost = $4,
			total_value = $5,
			is_low_stock = $6,
			is_out_of_stock = $7,
			updated_at = $8
		WHERE id = $9 AND tenant_id = $10
	`
	
	result, err := r.db.ExecContext(ctx, query,
		availability.AvailableQuantity, availability.ReservedQuantity, availability.TotalQuantity,
		availability.AvgUnitCost, availability.TotalValue,
		availability.IsLowStock, availability.IsOutOfStock, availability.UpdatedAt,
		availability.ID, availability.TenantID,
	)
	if err != nil {
		return err
	}
	
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return exception.ErrStockAvailabilityNotFound
	}
	
	return nil
}

