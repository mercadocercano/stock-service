package service

import (
	"context"

	"github.com/hornosg/go-shared/criteria"
	"stock/src/warehouse/domain/entity"
)

// WarehouseServiceInterface define la interfaz del servicio de almacenes
type WarehouseServiceInterface interface {
	CreateWarehouse(ctx context.Context, tenantID, locationID, name, code string, warehouseType entity.WarehouseType, description string, priority int) (*entity.Warehouse, error)
	GetWarehouseByID(ctx context.Context, id, tenantID string) (*entity.Warehouse, error)
	UpdateWarehouseEntity(ctx context.Context, warehouse *entity.Warehouse) error
	DeleteWarehouse(ctx context.Context, id, tenantID string) error
	ActivateWarehouse(ctx context.Context, id, tenantID string) (*entity.Warehouse, error)
	DeactivateWarehouse(ctx context.Context, id, tenantID string) (*entity.Warehouse, error)
	FindWarehousesByCriteria(ctx context.Context, tenantID string, crit criteria.Criteria) ([]*entity.Warehouse, int, error)
	FindWarehousesByLocationID(ctx context.Context, locationID, tenantID string, crit criteria.Criteria) ([]*entity.Warehouse, int, error)
}
