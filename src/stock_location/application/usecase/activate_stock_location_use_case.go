package usecase

import (
	"context"

	"stock/src/stock_location/application/response"
	"stock/src/stock_location/domain/service"
)

// ActivateStockLocationUseCase define el caso de uso para activar una ubicación de stock
type ActivateStockLocationUseCase struct {
	stockLocationService *service.StockLocationService
}

// NewActivateStockLocationUseCase crea una nueva instancia del caso de uso
func NewActivateStockLocationUseCase(stockLocationService *service.StockLocationService) *ActivateStockLocationUseCase {
	return &ActivateStockLocationUseCase{
		stockLocationService: stockLocationService,
	}
}

// Execute ejecuta el caso de uso para activar una ubicación de stock
func (uc *ActivateStockLocationUseCase) Execute(ctx context.Context, tenantID string, stockLocationID string) (*response.StockLocationResponse, error) {
	// Obtener la ubicación de stock existente
	stockLocation, err := uc.stockLocationService.GetStockLocationByID(ctx, stockLocationID, tenantID)
	if err != nil {
		return nil, err
	}

	// Activar la ubicación de stock
	stockLocation.Activate()

	// Guardar los cambios
	err = uc.stockLocationService.UpdateStockLocationEntity(ctx, stockLocation)
	if err != nil {
		return nil, err
	}

	// Transformar entidad de dominio en DTO de respuesta
	return response.NewStockLocationResponse(stockLocation), nil
}
