package service

import (
	"context"

	"github.com/mercadocercano/criteria"
	"stock/src/stock_location/domain/entity"
)

// StockLocationServiceInterface define la interfaz del servicio de ubicaciones de stock
type StockLocationServiceInterface interface {
	CreateStockLocation(ctx context.Context, tenantID, warehouseID string, parentID *string, name, code, description string) (*entity.StockLocation, error)
	GetStockLocationByID(ctx context.Context, id, tenantID string) (*entity.StockLocation, error)
	UpdateStockLocationEntity(ctx context.Context, stockLocation *entity.StockLocation) error
	DeleteStockLocation(ctx context.Context, id, tenantID string) error
	FindStockLocationsByCriteria(ctx context.Context, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error)
	FindStockLocationsByWarehouseID(ctx context.Context, warehouseID, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error)
	FindChildrenStockLocations(ctx context.Context, parentID, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error)
	FindRootStockLocations(ctx context.Context, warehouseID, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error)
}
