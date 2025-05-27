package port

import (
	"context"

	"stock/src/shared/domain/criteria"
	"stock/src/stock_location/domain/entity"
)

// StockLocationRepository define la interfaz para el repositorio de ubicaciones de stock
type StockLocationRepository interface {
	// Save guarda una nueva ubicación de stock
	Save(ctx context.Context, stockLocation *entity.StockLocation) error

	// FindByID busca una ubicación de stock por su ID
	FindByID(ctx context.Context, id string, tenantID string) (*entity.StockLocation, error)

	// Update actualiza una ubicación de stock
	Update(ctx context.Context, stockLocation *entity.StockLocation) error

	// Delete elimina una ubicación de stock
	Delete(ctx context.Context, id string, tenantID string) error

	// FindByCriteria busca ubicaciones de stock según criterios
	FindByCriteria(ctx context.Context, tenantID string, criteria criteria.Criteria) ([]*entity.StockLocation, int, error)

	// FindByWarehouseID busca ubicaciones de stock por el ID del almacén
	FindByWarehouseID(ctx context.Context, warehouseID string, tenantID string, criteria criteria.Criteria) ([]*entity.StockLocation, int, error)

	// FindChildren busca ubicaciones de stock hijas de una ubicación padre
	FindChildren(ctx context.Context, parentID string, tenantID string, criteria criteria.Criteria) ([]*entity.StockLocation, int, error)

	// FindRoots busca ubicaciones de stock de nivel raíz en un almacén
	FindRoots(ctx context.Context, warehouseID string, tenantID string, criteria criteria.Criteria) ([]*entity.StockLocation, int, error)
}
