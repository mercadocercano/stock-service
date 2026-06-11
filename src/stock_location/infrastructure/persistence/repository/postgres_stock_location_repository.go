package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/hornosg/go-shared/criteria"
	"stock/src/stock_location/domain/entity"
	"stock/src/stock_location/domain/exception"
	"stock/src/stock_location/domain/port"
)

// PostgresStockLocationRepository implementación PostgreSQL del repositorio de ubicaciones de stock
type PostgresStockLocationRepository struct {
	db *sql.DB
}

// NewPostgresStockLocationRepository crea una nueva instancia del repositorio
func NewPostgresStockLocationRepository(db *sql.DB) port.StockLocationRepository {
	return &PostgresStockLocationRepository{
		db: db,
	}
}

// Save guarda una nueva ubicación de stock
func (r *PostgresStockLocationRepository) Save(ctx context.Context, stockLocation *entity.StockLocation) error {
	query := `
		INSERT INTO stock_locations (
			id, tenant_id, warehouse_id, parent_id, name, code, path, level, 
			description, active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		stockLocation.ID,
		stockLocation.TenantID,
		stockLocation.WarehouseID,
		stockLocation.ParentID,
		stockLocation.Name,
		stockLocation.Code,
		stockLocation.Path,
		stockLocation.Level,
		stockLocation.Description,
		stockLocation.Active,
		stockLocation.CreatedAt,
		stockLocation.UpdatedAt,
	)

	return err
}

// FindByID busca una ubicación de stock por su ID
func (r *PostgresStockLocationRepository) FindByID(ctx context.Context, id string, tenantID string) (*entity.StockLocation, error) {
	query := `
		SELECT 
			id, tenant_id, warehouse_id, parent_id, name, code, path, level,
			description, active, created_at, updated_at
		FROM stock_locations
		WHERE id = $1 AND tenant_id = $2
	`

	var stockLocation entity.StockLocation
	var parentID sql.NullString
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, id, tenantID).Scan(
		&stockLocation.ID,
		&stockLocation.TenantID,
		&stockLocation.WarehouseID,
		&parentID,
		&stockLocation.Name,
		&stockLocation.Code,
		&stockLocation.Path,
		&stockLocation.Level,
		&stockLocation.Description,
		&stockLocation.Active,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &exception.StockLocationNotFound{
				ID:       id,
				TenantID: tenantID,
			}
		}
		return nil, err
	}

	stockLocation.CreatedAt = createdAt
	stockLocation.UpdatedAt = updatedAt

	if parentID.Valid {
		stockLocation.ParentID = &parentID.String
	}

	return &stockLocation, nil
}

// Update actualiza una ubicación de stock
func (r *PostgresStockLocationRepository) Update(ctx context.Context, stockLocation *entity.StockLocation) error {
	query := `
		UPDATE stock_locations
		SET name = $1, code = $2, description = $3, active = $4, updated_at = $5
		WHERE id = $6 AND tenant_id = $7
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		stockLocation.Name,
		stockLocation.Code,
		stockLocation.Description,
		stockLocation.Active,
		time.Now(),
		stockLocation.ID,
		stockLocation.TenantID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return &exception.StockLocationNotFound{
			ID:       stockLocation.ID,
			TenantID: stockLocation.TenantID,
		}
	}

	return nil
}

// Delete elimina una ubicación de stock
func (r *PostgresStockLocationRepository) Delete(ctx context.Context, id string, tenantID string) error {
	query := `DELETE FROM stock_locations WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return &exception.StockLocationNotFound{
			ID:       id,
			TenantID: tenantID,
		}
	}

	return nil
}

// FindByCriteria busca ubicaciones de stock según criterios
func (r *PostgresStockLocationRepository) FindByCriteria(ctx context.Context, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error) {
	whereClause, args := r.buildWhereClause(crit.Filters, []interface{}{tenantID})
	orderClause := r.buildOrderClause(crit.PrimaryOrder())
	limitClause, offsetClause := "", ""

	if !crit.Pagination.IsEmpty() {
		limitClause = fmt.Sprintf("LIMIT %d", crit.Pagination.Limit)
		offsetClause = fmt.Sprintf("OFFSET %d", crit.Pagination.Offset)
	}

	query := fmt.Sprintf(`
		SELECT 
			id, tenant_id, warehouse_id, parent_id, name, code, path, level,
			description, active, created_at, updated_at
		FROM stock_locations
		WHERE tenant_id = $1 %s
		%s
		%s
		%s
	`, whereClause, orderClause, limitClause, offsetClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	stockLocations := make([]*entity.StockLocation, 0)
	for rows.Next() {
		var stockLocation entity.StockLocation
		var parentID sql.NullString
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&stockLocation.ID,
			&stockLocation.TenantID,
			&stockLocation.WarehouseID,
			&parentID,
			&stockLocation.Name,
			&stockLocation.Code,
			&stockLocation.Path,
			&stockLocation.Level,
			&stockLocation.Description,
			&stockLocation.Active,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, 0, err
		}

		stockLocation.CreatedAt = createdAt
		stockLocation.UpdatedAt = updatedAt

		if parentID.Valid {
			stockLocation.ParentID = &parentID.String
		}

		stockLocations = append(stockLocations, &stockLocation)
	}

	// Obtener el total de registros
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM stock_locations
		WHERE tenant_id = $1 %s
	`, whereClause)

	var total int
	err = r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return stockLocations, total, nil
}

// FindByWarehouseID busca ubicaciones de stock por el ID del almacén
func (r *PostgresStockLocationRepository) FindByWarehouseID(ctx context.Context, warehouseID string, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error) {
	filters := criteria.NewFilters()
	for _, f := range crit.Filters.Items {
		filters.Add(f)
	}
	filters.Add(criteria.NewFilter("warehouse_id", criteria.OpEqual, warehouseID))
	newCriteria := criteria.NewCriteria(filters, crit.Orders, crit.Pagination)
	return r.FindByCriteria(ctx, tenantID, newCriteria)
}

// FindChildren busca ubicaciones de stock hijas de una ubicación padre
func (r *PostgresStockLocationRepository) FindChildren(ctx context.Context, parentID string, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error) {
	filters := criteria.NewFilters()
	for _, f := range crit.Filters.Items {
		filters.Add(f)
	}
	filters.Add(criteria.NewFilter("parent_id", criteria.OpEqual, parentID))
	newCriteria := criteria.NewCriteria(filters, crit.Orders, crit.Pagination)
	return r.FindByCriteria(ctx, tenantID, newCriteria)
}

// FindRoots busca ubicaciones de stock de nivel raíz en un almacén
func (r *PostgresStockLocationRepository) FindRoots(ctx context.Context, warehouseID string, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error) {
	filters := criteria.NewFilters()
	for _, f := range crit.Filters.Items {
		filters.Add(f)
	}
	filters.Add(criteria.NewFilter("warehouse_id", criteria.OpEqual, warehouseID))
	filters.Add(criteria.NewFilter("parent_id", criteria.OpIsNull, nil))
	newCriteria := criteria.NewCriteria(filters, crit.Orders, crit.Pagination)
	return r.FindByCriteria(ctx, tenantID, newCriteria)
}

// buildWhereClause construye la cláusula WHERE para la consulta SQL
func (r *PostgresStockLocationRepository) buildWhereClause(filters criteria.Filters, args []interface{}) (string, []interface{}) {
	if filters.IsEmpty() {
		return "", args
	}

	conditions := make([]string, 0)
	for _, filter := range filters.Items {
		paramIndex := len(args) + 1

		switch filter.Operator {
		case criteria.OpEqual, criteria.OpNotEqual, criteria.OpGreaterThan, criteria.OpGreaterThanOrEqual, criteria.OpLessThan, criteria.OpLessThanOrEqual:
			conditions = append(conditions, fmt.Sprintf("%s %s $%d", filter.Field, filter.Operator, paramIndex))
			args = append(args, filter.Value)
		case criteria.OpLike:
			conditions = append(conditions, fmt.Sprintf("%s ILIKE $%d", filter.Field, paramIndex))
			args = append(args, filter.Value)
		case criteria.OpIsNull:
			conditions = append(conditions, fmt.Sprintf("%s IS NULL", filter.Field))
		case criteria.OpIsNotNull:
			conditions = append(conditions, fmt.Sprintf("%s IS NOT NULL", filter.Field))
		default:
			continue
		}
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "AND " + strings.Join(conditions, " AND "), args
}

// buildOrderClause construye la cláusula ORDER BY para la consulta SQL
func (r *PostgresStockLocationRepository) buildOrderClause(order criteria.Order) string {
	if order.IsEmpty() {
		return "ORDER BY created_at DESC"
	}
	return fmt.Sprintf("ORDER BY %s %s", order.Field, order.Direction)
}
