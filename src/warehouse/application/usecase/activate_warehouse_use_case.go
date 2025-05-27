package usecase

import (
	"context"

	"stock/src/warehouse/application/response"
	"stock/src/warehouse/domain/service"
)

// ActivateWarehouseUseCase define el caso de uso para activar un almacén
type ActivateWarehouseUseCase struct {
	warehouseService *service.WarehouseService
}

// NewActivateWarehouseUseCase crea una nueva instancia del caso de uso
func NewActivateWarehouseUseCase(warehouseService *service.WarehouseService) *ActivateWarehouseUseCase {
	return &ActivateWarehouseUseCase{
		warehouseService: warehouseService,
	}
}

// Execute ejecuta el caso de uso para activar un almacén
func (uc *ActivateWarehouseUseCase) Execute(ctx context.Context, tenantID string, warehouseID string) (*response.WarehouseResponse, error) {
	// Obtener el almacén existente
	warehouse, err := uc.warehouseService.GetWarehouseByID(ctx, warehouseID, tenantID)
	if err != nil {
		return nil, err
	}

	// Activar el almacén
	warehouse.Activate()

	// Guardar los cambios
	err = uc.warehouseService.UpdateWarehouseEntity(ctx, warehouse)
	if err != nil {
		return nil, err
	}

	// Transformar entidad de dominio en DTO de respuesta
	return response.NewWarehouseResponse(warehouse), nil
}
