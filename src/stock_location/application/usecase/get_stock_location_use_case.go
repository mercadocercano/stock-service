package usecase

import (
	"context"

	"stock/src/stock_location/application/response"
	"stock/src/stock_location/domain/exception"
	"stock/src/stock_location/domain/service"
)

// GetStockLocationUseCase define el caso de uso para obtener una ubicación de stock por su ID
type GetStockLocationUseCase struct {
	stockLocationService *service.StockLocationService
}

// NewGetStockLocationUseCase crea una nueva instancia del caso de uso
func NewGetStockLocationUseCase(stockLocationService *service.StockLocationService) *GetStockLocationUseCase {
	return &GetStockLocationUseCase{
		stockLocationService: stockLocationService,
	}
}

// Execute ejecuta el caso de uso para obtener una ubicación de stock por su ID
func (uc *GetStockLocationUseCase) Execute(ctx context.Context, tenantID string, stockLocationID string) (*response.StockLocationResponse, error) {
	// Obtener la ubicación de stock a través del servicio de dominio
	stockLocation, err := uc.stockLocationService.GetStockLocationByID(ctx, stockLocationID, tenantID)
	if err != nil {
		return nil, err
	}

	// Verificar si se encontró la ubicación de stock
	if stockLocation == nil {
		return nil, &exception.StockLocationNotFound{
			ID:       stockLocationID,
			TenantID: tenantID,
		}
	}

	// Transformar la entidad de dominio en DTO de respuesta
	return response.NewStockLocationResponse(stockLocation), nil
}
