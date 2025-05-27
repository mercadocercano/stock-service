package usecase

import (
	"context"

	"stock/src/warehouse/application/response"
	"stock/src/warehouse/domain/service"
)

// GetWarehouseUseCase define el caso de uso para obtener un almacén por su ID
type GetWarehouseUseCase struct {
	warehouseService *service.WarehouseService
}

// NewGetWarehouseUseCase crea una nueva instancia del caso de uso
func NewGetWarehouseUseCase(warehouseService *service.WarehouseService) *GetWarehouseUseCase {
	return &GetWarehouseUseCase{
		warehouseService: warehouseService,
	}
}

// Execute ejecuta el caso de uso para obtener un almacén por su ID
func (uc *GetWarehouseUseCase) Execute(ctx context.Context, tenantID string, warehouseID string) (*response.WarehouseResponse, error) {
	// Obtener almacén del servicio de dominio
	warehouse, err := uc.warehouseService.GetWarehouseByID(ctx, warehouseID, tenantID)
	if err != nil {
		return nil, err
	}

	// Transformar entidad de dominio en DTO de respuesta
	return response.NewWarehouseResponse(warehouse), nil
}
