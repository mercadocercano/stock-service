package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/hornosg/go-shared/criteria"
	"github.com/hornosg/go-shared/infrastructure/postgres"

	"stock/src/location/domain/entity"
	"stock/src/location/domain/exception"
	"stock/src/location/domain/port"
)

// PostgresLocationRepository implementa la interfaz LocationRepository con PostgreSQL.
//
// RLS (PLAT-E30, RULE-09/RULE-10): `locations` tiene ROW LEVEL SECURITY forzado con la
// policy `tenant_isolation` (migración 009). Cada operación corre dentro de
// postgres.WithRLSInTransaction, que fija `app.tenant_id` con SET LOCAL — sin él, cualquier
// query erra (fail-closed) bajo el rol NOBYPASSRLS `stock_app`. El filtro manual
// `WHERE tenant_id = $` se mantiene como defensa en profundidad.
type PostgresLocationRepository struct {
	db *sql.DB
}

// NewPostgresLocationRepository crea una nueva instancia del repositorio de ubicaciones
func NewPostgresLocationRepository(db *sql.DB) port.LocationRepository {
	return &PostgresLocationRepository{
		db: db,
	}
}

// Save guarda una ubicación en la base de datos. Tenant: location.TenantID (path del WITH CHECK).
func (r *PostgresLocationRepository) Save(ctx context.Context, location *entity.Location) error {
	query := `
		INSERT INTO locations (
			id, tenant_id, name, type, address, city, state, country,
			postal_code, phone, email, active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	rc := postgres.RLSContext{TenantID: location.TenantID}
	return postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			query,
			location.ID,
			location.TenantID,
			location.Name,
			location.Type,
			location.Address,
			location.City,
			location.State,
			location.Country,
			location.PostalCode,
			location.Phone,
			location.Email,
			location.Active,
			location.CreatedAt,
			location.UpdatedAt,
		)
		return err
	})
}

// FindByID busca una ubicación por su ID. Tenant: tenantID param.
func (r *PostgresLocationRepository) FindByID(ctx context.Context, id string, tenantID string) (*entity.Location, error) {
	query := `
		SELECT
			id, tenant_id, name, type, address, city, state, country,
			postal_code, phone, email, active, created_at, updated_at
		FROM locations
		WHERE id = $1 AND tenant_id = $2
	`

	rc := postgres.RLSContext{TenantID: tenantID}
	var location entity.Location
	var found bool
	err := postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		scanErr := tx.QueryRowContext(ctx, query, id, tenantID).Scan(
			&location.ID,
			&location.TenantID,
			&location.Name,
			&location.Type,
			&location.Address,
			&location.City,
			&location.State,
			&location.Country,
			&location.PostalCode,
			&location.Phone,
			&location.Email,
			&location.Active,
			&location.CreatedAt,
			&location.UpdatedAt,
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

	if err != nil {
		return nil, err
	}
	if !found {
		return nil, exception.NewLocationNotFound(id, tenantID)
	}

	return &location, nil
}

// Update actualiza una ubicación existente. Tenant: location.TenantID (path del WITH CHECK).
func (r *PostgresLocationRepository) Update(ctx context.Context, location *entity.Location) error {
	query := `
		UPDATE locations SET
			name = $1,
			address = $2,
			city = $3,
			state = $4,
			country = $5,
			postal_code = $6,
			phone = $7,
			email = $8,
			active = $9,
			updated_at = $10
		WHERE id = $11 AND tenant_id = $12
	`

	rc := postgres.RLSContext{TenantID: location.TenantID}
	var rowsAffected int64
	err := postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		result, execErr := tx.ExecContext(
			ctx,
			query,
			location.Name,
			location.Address,
			location.City,
			location.State,
			location.Country,
			location.PostalCode,
			location.Phone,
			location.Email,
			location.Active,
			location.UpdatedAt,
			location.ID,
			location.TenantID,
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
		return exception.NewLocationNotFound(location.ID, location.TenantID)
	}

	return nil
}

// Delete elimina una ubicación por su ID. Tenant: tenantID param.
func (r *PostgresLocationRepository) Delete(ctx context.Context, id string, tenantID string) error {
	query := "DELETE FROM locations WHERE id = $1 AND tenant_id = $2"

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
		return exception.NewLocationNotFound(id, tenantID)
	}

	return nil
}

// FindByCriteria busca ubicaciones según criterios específicos. Tenant: tenantID param.
func (r *PostgresLocationRepository) FindByCriteria(ctx context.Context, tenantID string, crit criteria.Criteria) ([]*entity.Location, int, error) {
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
		case "type":
			where = append(where, fmt.Sprintf("type = $%d", paramCount))
			params = append(params, filter.Value)
			paramCount++
		case "city":
			where = append(where, fmt.Sprintf("city ILIKE $%d", paramCount))
			params = append(params, "%"+filter.Value.(string)+"%")
			paramCount++
		case "active":
			where = append(where, fmt.Sprintf("active = $%d", paramCount))
			params = append(params, filter.Value)
			paramCount++
		}
	}

	// Consulta para contar el total
	countQuery := "SELECT COUNT(*) FROM locations WHERE tenant_id = $1"
	if len(where) > 1 { // Si hay más condiciones además de tenant_id
		countQuery += " AND " + strings.Join(where[1:], " AND ")
	}

	// Construir la consulta de selección con ORDER BY, LIMIT y OFFSET
	selectQuery := `
		SELECT
			id, tenant_id, name, type, address, city, state, country,
			postal_code, phone, email, active, created_at, updated_at
		FROM locations
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
	locations := make([]*entity.Location, 0)
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
			var location entity.Location
			if scanErr := rows.Scan(
				&location.ID,
				&location.TenantID,
				&location.Name,
				&location.Type,
				&location.Address,
				&location.City,
				&location.State,
				&location.Country,
				&location.PostalCode,
				&location.Phone,
				&location.Email,
				&location.Active,
				&location.CreatedAt,
				&location.UpdatedAt,
			); scanErr != nil {
				return scanErr
			}

			locations = append(locations, &location)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, 0, err
	}

	return locations, totalCount, nil
}

// FindStores busca solo ubicaciones de tipo tienda
func (r *PostgresLocationRepository) FindStores(ctx context.Context, tenantID string, crit criteria.Criteria) ([]*entity.Location, int, error) {
	storeFilter := criteria.NewFilters()
	for _, f := range crit.Filters.Items {
		storeFilter.Add(f)
	}
	storeFilter.Add(criteria.NewFilter("type", criteria.OpEqual, "store"))
	storeCriteria := criteria.NewCriteria(storeFilter, crit.Orders, crit.Pagination)
	return r.FindByCriteria(ctx, tenantID, storeCriteria)
}

// FindDistributionCenters busca solo ubicaciones de tipo centro de distribución
func (r *PostgresLocationRepository) FindDistributionCenters(ctx context.Context, tenantID string, crit criteria.Criteria) ([]*entity.Location, int, error) {
	dcFilter := criteria.NewFilters()
	for _, f := range crit.Filters.Items {
		dcFilter.Add(f)
	}
	dcFilter.Add(criteria.NewFilter("type", criteria.OpEqual, "distribution_center"))
	dcCriteria := criteria.NewCriteria(dcFilter, crit.Orders, crit.Pagination)
	return r.FindByCriteria(ctx, tenantID, dcCriteria)
}