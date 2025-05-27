package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"stock/src/shared/domain/criteria"
	"stock/src/warehouse/domain/entity"
	"stock/src/warehouse/domain/exception"
	"stock/src/warehouse/domain/port"
)

// PostgresWarehouseRepository implementa la interfaz WarehouseRepository con PostgreSQL
type PostgresWarehouseRepository struct {
	db *sql.DB
}

// NewPostgresWarehouseRepository crea una nueva instancia del repositorio de almacenes
func NewPostgresWarehouseRepository(db *sql.DB) port.WarehouseRepository {
	return &PostgresWarehouseRepository{
		db: db,
	}
}

// Save guarda un almacén en la base de datos
func (r *PostgresWarehouseRepository) Save(ctx context.Context, warehouse *entity.Warehouse) error {
	query := `
		INSERT INTO warehouses (
			id, tenant_id, location_id, name, code, type, description, 
			priority, active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		warehouse.ID,
		warehouse.TenantID,
		warehouse.LocationID,
		warehouse.Name,
		warehouse.Code,
		warehouse.Type,
		warehouse.Description,
		warehouse.Priority,
		warehouse.Active,
		warehouse.CreatedAt,
		warehouse.UpdatedAt,
	)

	return err
}

// FindByID busca un almacén por su ID
func (r *PostgresWarehouseRepository) FindByID(ctx context.Context, id string, tenantID string) (*entity.Warehouse, error) {
	query := `
		SELECT 
			id, tenant_id, location_id, name, code, type, description,
			priority, active, created_at, updated_at
		FROM warehouses
		WHERE id = $1 AND tenant_id = $2
	`

	var warehouse entity.Warehouse
	var warehouseType string

	err := r.db.QueryRowContext(ctx, query, id, tenantID).Scan(
		&warehouse.ID,
		&warehouse.TenantID,
		&warehouse.LocationID,
		&warehouse.Name,
		&warehouse.Code,
		&warehouseType,
		&warehouse.Description,
		&warehouse.Priority,
		&warehouse.Active,
		&warehouse.CreatedAt,
		&warehouse.UpdatedAt,
	)

	// Convertir el tipo de string a WarehouseType
	warehouse.Type = entity.WarehouseType(warehouseType)

	if err == sql.ErrNoRows {
		return nil, exception.NewWarehouseNotFound(id, tenantID)
	}

	if err != nil {
		return nil, err
	}

	return &warehouse, nil
}

// Update actualiza un almacén existente
func (r *PostgresWarehouseRepository) Update(ctx context.Context, warehouse *entity.Warehouse) error {
	query := `
		UPDATE warehouses SET
			name = $1,
			code = $2,
			type = $3,
			description = $4,
			priority = $5,
			active = $6,
			updated_at = $7
		WHERE id = $8 AND tenant_id = $9
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		warehouse.Name,
		warehouse.Code,
		warehouse.Type,
		warehouse.Description,
		warehouse.Priority,
		warehouse.Active,
		warehouse.UpdatedAt,
		warehouse.ID,
		warehouse.TenantID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return exception.NewWarehouseNotFound(warehouse.ID, warehouse.TenantID)
	}

	return nil
}

// Delete elimina un almacén por su ID
func (r *PostgresWarehouseRepository) Delete(ctx context.Context, id string, tenantID string) error {
	query := "DELETE FROM warehouses WHERE id = $1 AND tenant_id = $2"

	result, err := r.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return exception.NewWarehouseNotFound(id, tenantID)
	}

	return nil
}

// FindByCriteria busca almacenes según criterios específicos
func (r *PostgresWarehouseRepository) FindByCriteria(ctx context.Context, tenantID string, crit criteria.Criteria) ([]*entity.Warehouse, int, error) {
	// Construir la consulta base
	baseQuery := `
		FROM warehouses
		WHERE tenant_id = $1
	`

	// Construir la parte WHERE con los filtros
	where := []string{"tenant_id = $1"}
	params := []interface{}{tenantID}
	paramCount := 2 // Comenzamos desde 2 porque ya usamos $1 para tenant_id

	for _, filter := range crit.Filters.Items {
		switch filter.Field {
		case "name":
			where = append(where, fmt.Sprintf("name ILIKE $%d", paramCount))
			params = append(params, "%"+filter.Value.(string)+"%")
			paramCount++
		case "code":
			where = append(where, fmt.Sprintf("code ILIKE $%d", paramCount))
			params = append(params, "%"+filter.Value.(string)+"%")
			paramCount++
		case "type":
			where = append(where, fmt.Sprintf("type = $%d", paramCount))
			params = append(params, filter.Value)
			paramCount++
		case "location_id":
			where = append(where, fmt.Sprintf("location_id = $%d", paramCount))
			params = append(params, filter.Value)
			paramCount++
		case "active":
			where = append(where, fmt.Sprintf("active = $%d", paramCount))
			params = append(params, filter.Value)
			paramCount++
		}
	}

	// Consulta para contar el total
	countQuery := "SELECT COUNT(*) " + baseQuery
	if len(where) > 1 { // Si hay más condiciones además de tenant_id
		countQuery += " AND " + strings.Join(where[1:], " AND ")
	}

	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, params...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Construir la consulta de selección con ORDER BY, LIMIT y OFFSET
	selectQuery := `
		SELECT 
			id, tenant_id, location_id, name, code, type, description,
			priority, active, created_at, updated_at
		FROM warehouses
		WHERE tenant_id = $1
	`

	if len(where) > 1 { // Si hay más condiciones además de tenant_id
		selectQuery += " AND " + strings.Join(where[1:], " AND ")
	}

	// Agregar ORDER BY si está especificado
	if !crit.Order.IsEmpty() {
		selectQuery += fmt.Sprintf(" ORDER BY %s %s", crit.Order.Field, crit.Order.OrderType)
	}

	// Agregar LIMIT y OFFSET si están especificados
	if crit.Limit != nil {
		selectQuery += fmt.Sprintf(" LIMIT %d", *crit.Limit)
	}

	if crit.Offset != nil {
		selectQuery += fmt.Sprintf(" OFFSET %d", *crit.Offset)
	}

	rows, err := r.db.QueryContext(ctx, selectQuery, params...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var warehouses []*entity.Warehouse
	for rows.Next() {
		var warehouse entity.Warehouse
		var warehouseType string

		err := rows.Scan(
			&warehouse.ID,
			&warehouse.TenantID,
			&warehouse.LocationID,
			&warehouse.Name,
			&warehouse.Code,
			&warehouseType,
			&warehouse.Description,
			&warehouse.Priority,
			&warehouse.Active,
			&warehouse.CreatedAt,
			&warehouse.UpdatedAt,
		)

		// Convertir el tipo de string a WarehouseType
		warehouse.Type = entity.WarehouseType(warehouseType)

		if err != nil {
			return nil, 0, err
		}

		warehouses = append(warehouses, &warehouse)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return warehouses, totalCount, nil
}

// FindByLocationID busca almacenes por el ID de su ubicación
func (r *PostgresWarehouseRepository) FindByLocationID(ctx context.Context, locationID string, tenantID string, crit criteria.Criteria) ([]*entity.Warehouse, int, error) {
	// Crear nuevos criterios con el filtro de location_id
	locationFilter := crit.Filters
	locationFilter.Add(criteria.NewFilter("location_id", "=", locationID))

	locationCriteria := criteria.NewCriteria(locationFilter, crit.Order, crit.Limit, crit.Offset)

	return r.FindByCriteria(ctx, tenantID, locationCriteria)
}
