package port

import (
	"context"

	"stock/src/location/domain/entity"
	"stock/src/shared/domain/criteria"
)

// LocationRepository define las operaciones disponibles para el repositorio de ubicaciones
type LocationRepository interface {
	// Save guarda una ubicación en el repositorio
	Save(ctx context.Context, location *entity.Location) error

	// FindByID busca una ubicación por su ID
	FindByID(ctx context.Context, id string, tenantID string) (*entity.Location, error)

	// Update actualiza una ubicación existente
	Update(ctx context.Context, location *entity.Location) error

	// Delete elimina una ubicación por su ID
	Delete(ctx context.Context, id string, tenantID string) error

	// FindByCriteria busca ubicaciones según criterios específicos
	FindByCriteria(ctx context.Context, tenantID string, criteria criteria.Criteria) ([]*entity.Location, int, error)

	// FindStores busca solo ubicaciones de tipo tienda
	FindStores(ctx context.Context, tenantID string, criteria criteria.Criteria) ([]*entity.Location, int, error)

	// FindDistributionCenters busca solo ubicaciones de tipo centro de distribución
	FindDistributionCenters(ctx context.Context, tenantID string, criteria criteria.Criteria) ([]*entity.Location, int, error)
}
