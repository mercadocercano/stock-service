package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"stock/src/stock_entry/domain/entity"
	"stock/src/stock_entry/domain/exception"
	"stock/src/stock_entry/domain/port"
)

// dbExecer abstrae *sql.DB y *sql.Tx para reusar queries
type dbExecer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

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

	if err := r.recalcAvailability(ctx, r.db, entry.TenantID, entry.VariantSKU, entry); err != nil {
		return fmt.Errorf("error updating availability after save: %w", err)
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
	
	skuSeen := make(map[string]*entity.StockEntry)
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
		skuSeen[entry.VariantSKU] = entry
	}
	stmt.Close()

	for sku, entry := range skuSeen {
		if err := r.recalcAvailability(ctx, tx, entry.TenantID, sku, entry); err != nil {
			return fmt.Errorf("error updating availability for SKU %s: %w", sku, err)
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

// ProcessSaleAtomic valida y descuenta stock en una sola transacción atómica
// HITO D: Operación atómica con SELECT FOR UPDATE para eliminar race conditions
func (r *PostgresStockEntryRepository) ProcessSaleAtomic(
	ctx context.Context,
	tenantID uuid.UUID,
	variantSKU string,
	quantity float64,
	reference string,
) (*entity.StockEntry, error) {
	// Iniciar transacción
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. SELECT FOR UPDATE - Lock row y obtener disponibilidad actual
	// Falla con sql.ErrNoRows si el producto nunca tuvo movimientos (correcto)
	var availableQty float64
	lockQuery := `
		SELECT available_quantity 
		FROM stock_availability 
		WHERE tenant_id = $1 
		  AND variant_sku = $2 
		  AND location_id IS NULL
		FOR UPDATE
	`
	
	err = tx.QueryRowContext(ctx, lockQuery, tenantID, variantSKU).Scan(&availableQty)
	if err == sql.ErrNoRows {
		// Producto sin stock inicializado → no vendible
		return nil, exception.ErrStockNotInitialized
	}
	if err != nil {
		return nil, fmt.Errorf("failed to lock stock availability: %w", err)
	}

	// 2. Validación en Go (no en DB, no en trigger)
	if availableQty < quantity {
		return nil, fmt.Errorf("%w: available=%.2f, requested=%.2f", 
			exception.ErrInsufficientStock, availableQty, quantity)
	}

	// 3. Crear entidad de dominio
	stockEntry, err := entity.NewStockEntry(
		tenantID,
		variantSKU,
		entity.EntryTypeSale,
		quantity,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stock entry entity: %w", err)
	}
	
	stockEntry.SetReference(reference)
	if err := stockEntry.Confirm(); err != nil {
		return nil, fmt.Errorf("failed to confirm stock entry: %w", err)
	}

	// 4. Persistir movimiento de venta
	insertQuery := `
		INSERT INTO stock_entries (
			id, tenant_id, variant_sku, product_sku, 
			entry_type, quantity, reference_number, 
			status, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, NOW(), NOW())
	`
	
	_, err = tx.ExecContext(ctx, insertQuery,
		stockEntry.ID,
		tenantID,
		variantSKU,
		variantSKU, // Copiar a product_sku por compatibilidad
		entity.EntryTypeSale,
		quantity,
		reference,
		"confirmed",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert stock entry: %w", err)
	}

	// 5. Recalcular stock_availability dentro de la misma TX
	if err := r.recalcAvailability(ctx, tx, tenantID, variantSKU, stockEntry); err != nil {
		return nil, fmt.Errorf("failed to update availability: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return stockEntry, nil
}

// CompensateSale revierte una venta creando un movimiento inverso
// HITO D: Compensación para rollback cuando falla persistencia de sale/order
func (r *PostgresStockEntryRepository) CompensateSale(
	ctx context.Context,
	tenantID uuid.UUID,
	stockEntryID uuid.UUID,
	reason string,
) error {
	// 1. Buscar el stock_entry original
	original, err := r.FindByID(ctx, stockEntryID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to find original stock entry: %w", err)
	}

	// 2. Validar que sea tipo 'sale'
	if original.EntryType != entity.EntryTypeSale {
		return fmt.Errorf("can only compensate sale entries, got: %s", original.EntryType)
	}

	// 3. Crear movimiento inverso (return suma stock)
	compensationEntry, err := entity.NewStockEntry(
		tenantID,
		original.VariantSKU,
		entity.EntryTypeReturn, // Tipo 'return' suma stock
		original.Quantity,      // Misma cantidad positiva
	)
	if err != nil {
		return fmt.Errorf("failed to create compensation entry: %w", err)
	}

	compensationEntry.SetReference(fmt.Sprintf("COMPENSATION-%s", stockEntryID.String()[:8]))
	compensationEntry.SetNotes(fmt.Sprintf("Compensation for sale %s. Reason: %s", stockEntryID, reason))
	if err := compensationEntry.Confirm(); err != nil {
		return fmt.Errorf("failed to confirm compensation entry: %w", err)
	}

	// 4. Guardar compensación (Save recalcula stock_availability)
	if err := r.Save(ctx, compensationEntry); err != nil {
		return fmt.Errorf("failed to save compensation entry: %w", err)
	}

	return nil
}

// recalcAvailability recalcula stock_availability desde stock_entries.
// Reemplaza la lógica que antes vivía en el trigger update_stock_availability_v2.
// Acepta dbExecer para funcionar tanto dentro de una TX como con r.db.
func (r *PostgresStockEntryRepository) recalcAvailability(ctx context.Context, ex dbExecer, tenantID uuid.UUID, variantSKU string, entry *entity.StockEntry) error {
	var totalQty float64
	var avgCost sql.NullFloat64
	sumQuery := `
		SELECT
			COALESCE(SUM(
				CASE
					WHEN entry_type IN ('initial_stock','purchase','transfer_in','return') THEN quantity
					WHEN entry_type IN ('sale','transfer_out') THEN -quantity
					WHEN entry_type = 'adjustment' THEN quantity
					ELSE 0
				END
			), 0),
			COALESCE(AVG(unit_cost), 0)
		FROM stock_entries
		WHERE tenant_id = $1
		  AND variant_sku = $2
		  AND status = 'confirmed'
	`
	if err := ex.QueryRowContext(ctx, sumQuery, tenantID, variantSKU).Scan(&totalQty, &avgCost); err != nil {
		return fmt.Errorf("recalcAvailability sum: %w", err)
	}

	var reservedQty float64
	resQuery := `
		SELECT COALESCE(reserved_quantity, 0)
		FROM stock_availability
		WHERE tenant_id = $1 AND variant_sku = $2 AND location_id IS NULL
	`
	err := ex.QueryRowContext(ctx, resQuery, tenantID, variantSKU).Scan(&reservedQty)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("recalcAvailability reserved: %w", err)
	}

	availableQty := totalQty - reservedQty
	cost := avgCost.Float64
	totalValue := totalQty * cost
	isLow := totalQty > 0 && totalQty < 10
	isOut := totalQty <= 0

	var productName, unitOfMeasure, entryType *string
	if entry != nil {
		if entry.ProductName != "" {
			productName = &entry.ProductName
		}
		if entry.UnitOfMeasure != "" {
			unitOfMeasure = &entry.UnitOfMeasure
		}
		et := string(entry.EntryType)
		entryType = &et
	}

	upsertQuery := `
		INSERT INTO stock_availability (
			tenant_id, variant_sku, product_sku, product_name, location_id,
			available_quantity, reserved_quantity, total_quantity,
			unit_of_measure, avg_unit_cost, total_value,
			is_low_stock, is_out_of_stock,
			last_entry_at, last_movement_type, updated_at
		) VALUES (
			$1, $2, $2, $3, NULL,
			$4, $5, $6,
			COALESCE($7, 'unit'), $8, $9,
			$10, $11,
			NOW(), $12, NOW()
		)
		ON CONFLICT (tenant_id, variant_sku) WHERE location_id IS NULL
		DO UPDATE SET
			available_quantity  = EXCLUDED.available_quantity,
			reserved_quantity   = EXCLUDED.reserved_quantity,
			total_quantity      = EXCLUDED.total_quantity,
			unit_of_measure     = COALESCE(EXCLUDED.unit_of_measure, stock_availability.unit_of_measure),
			avg_unit_cost       = EXCLUDED.avg_unit_cost,
			total_value         = EXCLUDED.total_value,
			is_low_stock        = EXCLUDED.is_low_stock,
			is_out_of_stock     = EXCLUDED.is_out_of_stock,
			last_entry_at       = EXCLUDED.last_entry_at,
			last_movement_type  = EXCLUDED.last_movement_type,
			updated_at          = NOW()
	`
	_, err = ex.ExecContext(ctx, upsertQuery,
		tenantID, variantSKU, productName,
		availableQty, reservedQty, totalQty,
		unitOfMeasure, cost, totalValue,
		isLow, isOut,
		entryType,
	)
	if err != nil {
		return fmt.Errorf("recalcAvailability upsert: %w", err)
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

// CountByTenant cuenta el total de registros de disponibilidad para un tenant
func (r *PostgresStockAvailabilityRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM stock_availability WHERE tenant_id = $1`
	var count int
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
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

