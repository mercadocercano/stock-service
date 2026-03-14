package service

import (
	"context"

	"stock/src/location/domain/entity"
	"github.com/mercadocercano/criteria"
)

// LocationServiceInterface define la interfaz del servicio de ubicaciones
type LocationServiceInterface interface {
	CreateLocation(ctx context.Context, tenantID, name string, locationType entity.LocationType,
		address, city, state, country, postalCode, phone, email string) (*entity.Location, error)
	GetLocationByID(ctx context.Context, id, tenantID string) (*entity.Location, error)
	UpdateLocationEntity(ctx context.Context, location *entity.Location) error
	DeleteLocation(ctx context.Context, id, tenantID string) error
	FindLocationsByCriteria(ctx context.Context, tenantID string, criteria criteria.Criteria) ([]*entity.Location, int, error)
	FindStores(ctx context.Context, tenantID string, criteria criteria.Criteria) ([]*entity.Location, int, error)
	FindDistributionCenters(ctx context.Context, tenantID string, criteria criteria.Criteria) ([]*entity.Location, int, error)
}
