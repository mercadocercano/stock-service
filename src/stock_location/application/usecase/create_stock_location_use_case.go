package usecase

import (
	"context"

	"stock/src/stock_location/application/request"
	"stock/src/stock_location/application/response"
	"stock/src/stock_location/domain/service"
)

// CreateStockLocationUseCase define el caso de uso para crear una ubicación de stock
type CreateStockLocationUseCase struct {
	stockLocationService service.StockLocationServiceInterface
}

// NewCreateStockLocationUseCase crea una nueva instancia del caso de uso
func NewCreateStockLocationUseCase(stockLocationService service.StockLocationServiceInterface) *CreateStockLocationUseCase {
	return &CreateStockLocationUseCase{
		stockLocationService: stockLocationService,
	}
}

// Execute ejecuta el caso de uso para crear una ubicación de stock
func (uc *CreateStockLocationUseCase) Execute(ctx context.Context, req request.CreateStockLocationRequest) (*response.StockLocationResponse, error) {
	// Validar la solicitud
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Crear la ubicación de stock a través del servicio de dominio
	stockLocation, err := uc.stockLocationService.CreateStockLocation(
		ctx,
		req.TenantID,
		req.WarehouseID,
		req.ParentID,
		req.Name,
		req.Code,
		req.Description,
	)

	if err != nil {
		return nil, err
	}

	// Transformar la entidad de dominio en DTO de respuesta
	return response.NewStockLocationResponse(stockLocation), nil
}
