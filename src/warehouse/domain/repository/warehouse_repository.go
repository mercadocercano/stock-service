package repository

import (
	"context"

	"stock/src/warehouse/domain/entity"
)

// WarehouseRepository define la interfaz del repositorio de almacenes
type WarehouseRepository interface {
	Create(ctx context.Context, warehouse *entity.Warehouse) error
	Update(ctx context.Context, warehouse *entity.Warehouse) error
	Delete(ctx context.Context, id string, tenantID string) error
	GetByID(ctx context.Context, id string, tenantID string) (*entity.Warehouse, error)
	FindByCriteria(ctx context.Context, criteria WarehouseCriteria) ([]*entity.Warehouse, error)
}

// WarehouseCriteria define los criterios de búsqueda para almacenes
type WarehouseCriteria struct {
	TenantID   string  `json:"tenant_id"`
	LocationID *string `json:"location_id,omitempty"`
	Name       *string `json:"name,omitempty"`
	Code       *string `json:"code,omitempty"`
	Active     *bool   `json:"active,omitempty"`
}
