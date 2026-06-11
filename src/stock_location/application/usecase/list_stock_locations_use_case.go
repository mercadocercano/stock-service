package usecase

import (
	"context"

	"github.com/hornosg/go-shared/criteria"
	"stock/src/stock_location/application/response"
	"stock/src/stock_location/domain/service"
)

// ListStockLocationsUseCase define el caso de uso para listar ubicaciones de stock
type ListStockLocationsUseCase struct {
	stockLocationService *service.StockLocationService
}

// NewListStockLocationsUseCase crea una nueva instancia del caso de uso
func NewListStockLocationsUseCase(stockLocationService *service.StockLocationService) *ListStockLocationsUseCase {
	return &ListStockLocationsUseCase{
		stockLocationService: stockLocationService,
	}
}

// Execute ejecuta el caso de uso para listar ubicaciones de stock
func (uc *ListStockLocationsUseCase) Execute(ctx context.Context, tenantID string, crit criteria.Criteria) (*response.StockLocationListResponse, error) {
	// Obtener las ubicaciones de stock a través del servicio de dominio
	stockLocations, total, err := uc.stockLocationService.FindStockLocationsByCriteria(ctx, tenantID, crit)
	if err != nil {
		return nil, err
	}

	// Transformar las entidades de dominio en DTO de respuesta
	return response.NewStockLocationListResponse(stockLocations, total), nil
}

// ExecuteByWarehouseID ejecuta el caso de uso para listar ubicaciones de stock por almacén
func (uc *ListStockLocationsUseCase) ExecuteByWarehouseID(ctx context.Context, warehouseID string, tenantID string, crit criteria.Criteria) (*response.StockLocationListResponse, error) {
	// Obtener las ubicaciones de stock por almacén a través del servicio de dominio
	stockLocations, total, err := uc.stockLocationService.FindStockLocationsByWarehouseID(ctx, warehouseID, tenantID, crit)
	if err != nil {
		return nil, err
	}

	// Transformar las entidades de dominio en DTO de respuesta
	return response.NewStockLocationListResponse(stockLocations, total), nil
}

// ExecuteChildren ejecuta el caso de uso para listar ubicaciones de stock hijas
func (uc *ListStockLocationsUseCase) ExecuteChildren(ctx context.Context, parentID string, tenantID string, crit criteria.Criteria) (*response.StockLocationListResponse, error) {
	// Obtener las ubicaciones de stock hijas a través del servicio de dominio
	stockLocations, total, err := uc.stockLocationService.FindChildrenStockLocations(ctx, parentID, tenantID, crit)
	if err != nil {
		return nil, err
	}

	// Transformar las entidades de dominio en DTO de respuesta
	return response.NewStockLocationListResponse(stockLocations, total), nil
}

// ExecuteRoots ejecuta el caso de uso para listar ubicaciones de stock raíz por almacén
func (uc *ListStockLocationsUseCase) ExecuteRoots(ctx context.Context, warehouseID string, tenantID string, crit criteria.Criteria) (*response.StockLocationListResponse, error) {
	// Obtener las ubicaciones de stock raíz a través del servicio de dominio
	stockLocations, total, err := uc.stockLocationService.FindRootStockLocations(ctx, warehouseID, tenantID, crit)
	if err != nil {
		return nil, err
	}

	// Transformar las entidades de dominio en DTO de respuesta
	return response.NewStockLocationListResponse(stockLocations, total), nil
}
