package usecase

import (
	"context"

	"stock/src/stock_location/application/request"
	"stock/src/stock_location/application/response"
	"stock/src/stock_location/domain/service"
)

// UpdateStockLocationUseCase define el caso de uso para actualizar una ubicación de stock
type UpdateStockLocationUseCase struct {
	stockLocationService *service.StockLocationService
}

// NewUpdateStockLocationUseCase crea una nueva instancia del caso de uso
func NewUpdateStockLocationUseCase(stockLocationService *service.StockLocationService) *UpdateStockLocationUseCase {
	return &UpdateStockLocationUseCase{
		stockLocationService: stockLocationService,
	}
}

// Execute ejecuta el caso de uso para actualizar una ubicación de stock
func (uc *UpdateStockLocationUseCase) Execute(ctx context.Context, tenantID string, stockLocationID string, req request.UpdateStockLocationRequest) (*response.StockLocationResponse, error) {
	// Validar la solicitud
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Obtener la ubicación de stock existente
	stockLocation, err := uc.stockLocationService.GetStockLocationByID(ctx, stockLocationID, tenantID)
	if err != nil {
		return nil, err
	}

	// Actualizar los datos
	stockLocation.Update(req.Name, req.Code, req.Description)

	// Guardar los cambios
	err = uc.stockLocationService.UpdateStockLocationEntity(ctx, stockLocation)
	if err != nil {
		return nil, err
	}

	// Transformar la entidad de dominio en DTO de respuesta
	return response.NewStockLocationResponse(stockLocation), nil
}
