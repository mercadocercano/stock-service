package usecase

import (
	"context"

	"stock/src/stock_location/domain/service"
)

// DeleteStockLocationUseCase define el caso de uso para eliminar una ubicación de stock
type DeleteStockLocationUseCase struct {
	stockLocationService *service.StockLocationService
}

// NewDeleteStockLocationUseCase crea una nueva instancia del caso de uso
func NewDeleteStockLocationUseCase(stockLocationService *service.StockLocationService) *DeleteStockLocationUseCase {
	return &DeleteStockLocationUseCase{
		stockLocationService: stockLocationService,
	}
}

// Execute ejecuta el caso de uso para eliminar una ubicación de stock
func (uc *DeleteStockLocationUseCase) Execute(ctx context.Context, tenantID string, stockLocationID string) error {
	// Delegar la eliminación al servicio de dominio
	return uc.stockLocationService.DeleteStockLocation(ctx, stockLocationID, tenantID)
}
