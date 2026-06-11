package port

import (
	"context"

	"github.com/hornosg/go-shared/criteria"
	"stock/src/warehouse/domain/entity"
)

// WarehouseRepository define la interfaz para el repositorio de almacenes
type WarehouseRepository interface {
	// Save guarda un almacén en la base de datos
	Save(ctx context.Context, warehouse *entity.Warehouse) error

	// FindByID busca un almacén por su ID
	FindByID(ctx context.Context, id string, tenantID string) (*entity.Warehouse, error)

	// Update actualiza un almacén existente
	Update(ctx context.Context, warehouse *entity.Warehouse) error

	// Delete elimina un almacén por su ID
	Delete(ctx context.Context, id string, tenantID string) error

	// FindByCriteria busca almacenes según criterios específicos
	FindByCriteria(ctx context.Context, tenantID string, criteria criteria.Criteria) ([]*entity.Warehouse, int, error)

	// FindByLocationID busca almacenes por el ID de su ubicación
	FindByLocationID(ctx context.Context, locationID string, tenantID string, criteria criteria.Criteria) ([]*entity.Warehouse, int, error)
}
