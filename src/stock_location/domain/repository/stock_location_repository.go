package repository

import (
	"context"

	"stock/src/stock_location/domain/entity"
)

// StockLocationRepository define la interfaz del repositorio de ubicaciones de stock
type StockLocationRepository interface {
	Create(ctx context.Context, stockLocation *entity.StockLocation) error
	Update(ctx context.Context, stockLocation *entity.StockLocation) error
	Delete(ctx context.Context, id string, tenantID string) error
	GetByID(ctx context.Context, id string, tenantID string) (*entity.StockLocation, error)
	FindByCriteria(ctx context.Context, criteria StockLocationCriteria) ([]*entity.StockLocation, error)
}

// StockLocationCriteria define los criterios de búsqueda para ubicaciones de stock
type StockLocationCriteria struct {
	TenantID    string  `json:"tenant_id"`
	WarehouseID string  `json:"warehouse_id"`
	ParentID    *string `json:"parent_id,omitempty"`
	Active      *bool   `json:"active,omitempty"`
	Name        *string `json:"name,omitempty"`
	Code        *string `json:"code,omitempty"`
}
