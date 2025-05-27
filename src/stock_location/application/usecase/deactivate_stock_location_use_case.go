package usecase

import (
	"context"

	"stock/src/stock_location/application/response"
	"stock/src/stock_location/domain/service"
)

// DeactivateStockLocationUseCase define el caso de uso para desactivar una ubicación de stock
type DeactivateStockLocationUseCase struct {
	stockLocationService *service.StockLocationService
}

// NewDeactivateStockLocationUseCase crea una nueva instancia del caso de uso
func NewDeactivateStockLocationUseCase(stockLocationService *service.StockLocationService) *DeactivateStockLocationUseCase {
	return &DeactivateStockLocationUseCase{
		stockLocationService: stockLocationService,
	}
}

// Execute ejecuta el caso de uso para desactivar una ubicación de stock
func (uc *DeactivateStockLocationUseCase) Execute(ctx context.Context, tenantID string, stockLocationID string) (*response.StockLocationResponse, error) {
	// Obtener la ubicación de stock existente
	stockLocation, err := uc.stockLocationService.GetStockLocationByID(ctx, stockLocationID, tenantID)
	if err != nil {
		return nil, err
	}

	// Desactivar la ubicación de stock
	stockLocation.Deactivate()

	// Guardar los cambios
	err = uc.stockLocationService.UpdateStockLocationEntity(ctx, stockLocation)
	if err != nil {
		return nil, err
	}

	// Transformar entidad de dominio en DTO de respuesta
	return response.NewStockLocationResponse(stockLocation), nil
}
