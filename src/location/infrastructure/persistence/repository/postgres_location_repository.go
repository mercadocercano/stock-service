package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"stock/src/location/domain/entity"
	"stock/src/location/domain/exception"
	"stock/src/location/domain/port"
	"stock/src/shared/domain/criteria"
)

// PostgresLocationRepository implementa la interfaz LocationRepository con PostgreSQL
type PostgresLocationRepository struct {
	db *sql.DB
}

// NewPostgresLocationRepository crea una nueva instancia del repositorio de ubicaciones
func NewPostgresLocationRepository(db *sql.DB) port.LocationRepository {
	return &PostgresLocationRepository{
		db: db,
	}
}

// Save guarda una ubicación en la base de datos
func (r *PostgresLocationRepository) Save(ctx context.Context, location *entity.Location) error {
	query := `
		INSERT INTO locations (
			id, tenant_id, name, type, address, city, state, country, 
			postal_code, phone, email, active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err := r.db.ExecContext(
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
}

// FindByID busca una ubicación por su ID
func (r *PostgresLocationRepository) FindByID(ctx context.Context, id string, tenantID string) (*entity.Location, error) {
	query := `
		SELECT 
			id, tenant_id, name, type, address, city, state, country, 
			postal_code, phone, email, active, created_at, updated_at
		FROM locations
		WHERE id = $1 AND tenant_id = $2
	`

	var location entity.Location
	err := r.db.QueryRowContext(ctx, query, id, tenantID).Scan(
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

	if err == sql.ErrNoRows {
		return nil, exception.NewLocationNotFound(id, tenantID)
	}

	if err != nil {
		return nil, err
	}

	return &location, nil
}

// Update actualiza una ubicación existente
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

	result, err := r.db.ExecContext(
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

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return exception.NewLocationNotFound(location.ID, location.TenantID)
	}

	return nil
}

// Delete elimina una ubicación por su ID
func (r *PostgresLocationRepository) Delete(ctx context.Context, id string, tenantID string) error {
	query := "DELETE FROM locations WHERE id = $1 AND tenant_id = $2"

	result, err := r.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return exception.NewLocationNotFound(id, tenantID)
	}

	return nil
}

// FindByCriteria busca ubicaciones según criterios específicos
func (r *PostgresLocationRepository) FindByCriteria(ctx context.Context, tenantID string, crit criteria.Criteria) ([]*entity.Location, int, error) {
	// Construir la consulta base
	baseQuery := `
		FROM locations
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
			id, tenant_id, name, type, address, city, state, country, 
			postal_code, phone, email, active, created_at, updated_at
		FROM locations
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

	var locations []*entity.Location
	for rows.Next() {
		var location entity.Location
		err := rows.Scan(
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

		if err != nil {
			return nil, 0, err
		}

		locations = append(locations, &location)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return locations, totalCount, nil
}

// FindStores busca solo ubicaciones de tipo tienda
func (r *PostgresLocationRepository) FindStores(ctx context.Context, tenantID string, crit criteria.Criteria) ([]*entity.Location, int, error) {
	// Añadir filtro de tipo = 'store'
	storeFilter := crit.Filters
	storeFilter.Add(criteria.NewFilter("type", "=", "store"))

	storeCriteria := criteria.NewCriteria(storeFilter, crit.Order, crit.Limit, crit.Offset)

	return r.FindByCriteria(ctx, tenantID, storeCriteria)
}

// FindDistributionCenters busca solo ubicaciones de tipo centro de distribución
func (r *PostgresLocationRepository) FindDistributionCenters(ctx context.Context, tenantID string, crit criteria.Criteria) ([]*entity.Location, int, error) {
	// Añadir filtro de tipo = 'distribution_center'
	dcFilter := crit.Filters
	dcFilter.Add(criteria.NewFilter("type", "=", "distribution_center"))

	dcCriteria := criteria.NewCriteria(dcFilter, crit.Order, crit.Limit, crit.Offset)

	return r.FindByCriteria(ctx, tenantID, dcCriteria)
}
