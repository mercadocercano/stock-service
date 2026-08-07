package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/hornosg/go-shared/criteria"
	"github.com/hornosg/go-shared/infrastructure/postgres"

	"stock/src/warehouse/domain/entity"
	"stock/src/warehouse/domain/exception"
	"stock/src/warehouse/domain/port"
)

// PostgresWarehouseRepository implementa la interfaz WarehouseRepository con PostgreSQL.
//
// RLS (PLAT-E30, RULE-09/RULE-10): `warehouses` tiene ROW LEVEL SECURITY forzado con la
// policy `tenant_isolation` (migración 009). Cada operación corre dentro de
// postgres.WithRLSInTransaction, que fija `app.tenant_id` con SET LOCAL — sin él, cualquier
// query erra (fail-closed) bajo el rol NOBYPASSRLS `stock_app`. El filtro manual
// `WHERE tenant_id = $` se mantiene como defensa en profundidad.
type PostgresWarehouseRepository struct {
	db *sql.DB
}

// NewPostgresWarehouseRepository crea una nueva instancia del repositorio de almacenes
func NewPostgresWarehouseRepository(db *sql.DB) port.WarehouseRepository {
	return &PostgresWarehouseRepository{
		db: db,
	}
}

// Save guarda un almacén en la base de datos. Tenant: warehouse.TenantID (path del WITH CHECK).
func (r *PostgresWarehouseRepository) Save(ctx context.Context, warehouse *entity.Warehouse) error {
	query := `
		INSERT INTO warehouses (
			id, tenant_id, location_id, name, code, type, description,
			priority, active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	rc := postgres.RLSContext{TenantID: warehouse.TenantID}
	return postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(
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
	})
}

// FindByID busca un almacén por su ID. Tenant: tenantID param.
func (r *PostgresWarehouseRepository) FindByID(ctx context.Context, id string, tenantID string) (*entity.Warehouse, error) {
	query := `
		SELECT
			id, tenant_id, location_id, name, code, type, description,
			priority, active, created_at, updated_at
		FROM warehouses
		WHERE id = $1 AND tenant_id = $2
	`

	rc := postgres.RLSContext{TenantID: tenantID}
	var warehouse entity.Warehouse
	var warehouseType string
	var found bool
	err := postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		scanErr := tx.QueryRowContext(ctx, query, id, tenantID).Scan(
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
		if scanErr == sql.ErrNoRows {
			return nil
		}
		if scanErr != nil {
			return scanErr
		}
		found = true
		return nil
	})

	// Convertir el tipo de string a WarehouseType
	warehouse.Type = entity.WarehouseType(warehouseType)

	if err != nil {
		return nil, err
	}
	if !found {
		return nil, exception.NewWarehouseNotFound(id, tenantID)
	}

	return &warehouse, nil
}

// Update actualiza un almacén existente. Tenant: warehouse.TenantID (path del WITH CHECK).
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

	rc := postgres.RLSContext{TenantID: warehouse.TenantID}
	var rowsAffected int64
	err := postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		result, execErr := tx.ExecContext(
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
		if execErr != nil {
			return execErr
		}
		ra, raErr := result.RowsAffected()
		if raErr != nil {
			return raErr
		}
		rowsAffected = ra
		return nil
	})

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return exception.NewWarehouseNotFound(warehouse.ID, warehouse.TenantID)
	}

	return nil
}

// Delete elimina un almacén por su ID. Tenant: tenantID param.
func (r *PostgresWarehouseRepository) Delete(ctx context.Context, id string, tenantID string) error {
	query := "DELETE FROM warehouses WHERE id = $1 AND tenant_id = $2"

	rc := postgres.RLSContext{TenantID: tenantID}
	var rowsAffected int64
	err := postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		result, execErr := tx.ExecContext(ctx, query, id, tenantID)
		if execErr != nil {
			return execErr
		}
		ra, raErr := result.RowsAffected()
		if raErr != nil {
			return raErr
		}
		rowsAffected = ra
		return nil
	})

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return exception.NewWarehouseNotFound(id, tenantID)
	}

	return nil
}

// FindByCriteria busca almacenes según criterios específicos. Tenant: tenantID param.
func (r *PostgresWarehouseRepository) FindByCriteria(ctx context.Context, tenantID string, crit criteria.Criteria) ([]*entity.Warehouse, int, error) {
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
	countQuery := "SELECT COUNT(*) FROM warehouses WHERE tenant_id = $1"
	if len(where) > 1 { // Si hay más condiciones además de tenant_id
		countQuery += " AND " + strings.Join(where[1:], " AND ")
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
	if order := crit.PrimaryOrder(); !order.IsEmpty() {
		selectQuery += fmt.Sprintf(" ORDER BY %s %s", order.Field, order.Direction)
	}

	// Agregar LIMIT y OFFSET si están especificados
	if !crit.Pagination.IsEmpty() {
		selectQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", crit.Pagination.Limit, crit.Pagination.Offset)
	}

	rc := postgres.RLSContext{TenantID: tenantID}
	var totalCount int
	warehouses := make([]*entity.Warehouse, 0)
	err := postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		if scanErr := tx.QueryRowContext(ctx, countQuery, params...).Scan(&totalCount); scanErr != nil {
			return scanErr
		}

		rows, queryErr := tx.QueryContext(ctx, selectQuery, params...)
		if queryErr != nil {
			return queryErr
		}
		defer rows.Close()

		for rows.Next() {
			var warehouse entity.Warehouse
			var wt string

			if scanErr := rows.Scan(
				&warehouse.ID,
				&warehouse.TenantID,
				&warehouse.LocationID,
				&warehouse.Name,
				&warehouse.Code,
				&wt,
				&warehouse.Description,
				&warehouse.Priority,
				&warehouse.Active,
				&warehouse.CreatedAt,
				&warehouse.UpdatedAt,
			); scanErr != nil {
				return scanErr
			}

			// Convertir el tipo de string a WarehouseType
			warehouse.Type = entity.WarehouseType(wt)
			warehouses = append(warehouses, &warehouse)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, 0, err
	}

	return warehouses, totalCount, nil
}

// FindByLocationID busca almacenes por el ID de su ubicación
func (r *PostgresWarehouseRepository) FindByLocationID(ctx context.Context, locationID string, tenantID string, crit criteria.Criteria) ([]*entity.Warehouse, int, error) {
	locationFilter := criteria.NewFilters()
	for _, f := range crit.Filters.Items {
		locationFilter.Add(f)
	}
	locationFilter.Add(criteria.NewFilter("location_id", criteria.OpEqual, locationID))
	locationCriteria := criteria.NewCriteria(locationFilter, crit.Orders, crit.Pagination)
	return r.FindByCriteria(ctx, tenantID, locationCriteria)
}