package usecase

import (
	"context"

	"stock/src/warehouse/application/response"
	"stock/src/warehouse/domain/service"
)

// DeactivateWarehouseUseCase define el caso de uso para desactivar un almacén
type DeactivateWarehouseUseCase struct {
	warehouseService *service.WarehouseService
}

// NewDeactivateWarehouseUseCase crea una nueva instancia del caso de uso
func NewDeactivateWarehouseUseCase(warehouseService *service.WarehouseService) *DeactivateWarehouseUseCase {
	return &DeactivateWarehouseUseCase{
		warehouseService: warehouseService,
	}
}

// Execute ejecuta el caso de uso para desactivar un almacén
func (uc *DeactivateWarehouseUseCase) Execute(ctx context.Context, tenantID string, warehouseID string) (*response.WarehouseResponse, error) {
	// Obtener el almacén existente
	warehouse, err := uc.warehouseService.GetWarehouseByID(ctx, warehouseID, tenantID)
	if err != nil {
		return nil, err
	}

	// Desactivar el almacén
	warehouse.Deactivate()

	// Guardar los cambios
	err = uc.warehouseService.UpdateWarehouseEntity(ctx, warehouse)
	if err != nil {
		return nil, err
	}

	// Transformar entidad de dominio en DTO de respuesta
	return response.NewWarehouseResponse(warehouse), nil
}
