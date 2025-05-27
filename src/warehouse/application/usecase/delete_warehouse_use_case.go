package usecase

import (
	"context"

	"stock/src/warehouse/domain/service"
)

// DeleteWarehouseUseCase define el caso de uso para eliminar un almacén
type DeleteWarehouseUseCase struct {
	warehouseService *service.WarehouseService
}

// NewDeleteWarehouseUseCase crea una nueva instancia del caso de uso
func NewDeleteWarehouseUseCase(warehouseService *service.WarehouseService) *DeleteWarehouseUseCase {
	return &DeleteWarehouseUseCase{
		warehouseService: warehouseService,
	}
}

// Execute ejecuta el caso de uso para eliminar un almacén
func (uc *DeleteWarehouseUseCase) Execute(ctx context.Context, tenantID string, warehouseID string) error {
	// Delegar la eliminación al servicio de dominio
	return uc.warehouseService.DeleteWarehouse(ctx, warehouseID, tenantID)
}
