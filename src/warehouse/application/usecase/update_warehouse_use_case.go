package usecase

import (
	"context"

	"stock/src/warehouse/application/request"
	"stock/src/warehouse/application/response"
	"stock/src/warehouse/domain/entity"
	"stock/src/warehouse/domain/service"
)

// UpdateWarehouseUseCase define el caso de uso para actualizar un almacén
type UpdateWarehouseUseCase struct {
	warehouseService *service.WarehouseService
}

// NewUpdateWarehouseUseCase crea una nueva instancia del caso de uso
func NewUpdateWarehouseUseCase(warehouseService *service.WarehouseService) *UpdateWarehouseUseCase {
	return &UpdateWarehouseUseCase{
		warehouseService: warehouseService,
	}
}

// Execute ejecuta el caso de uso para actualizar un almacén
func (uc *UpdateWarehouseUseCase) Execute(ctx context.Context, tenantID string, warehouseID string, req request.UpdateWarehouseRequest) (*response.WarehouseResponse, error) {
	// Convertir el tipo de almacén desde string a WarehouseType
	var warehouseType entity.WarehouseType
	switch req.Type {
	case "regular":
		warehouseType = entity.RegularWarehouseType
	case "special":
		warehouseType = entity.SpecialWarehouseType
	case "virtual":
		warehouseType = entity.VirtualWarehouseType
	default:
		warehouseType = entity.RegularWarehouseType
	}

	// Obtener el almacén existente
	warehouse, err := uc.warehouseService.GetWarehouseByID(ctx, warehouseID, tenantID)
	if err != nil {
		return nil, err
	}

	// Actualizar los datos
	warehouse.Update(req.Name, req.Code, warehouseType, req.Description, req.Priority)

	// Guardar los cambios
	err = uc.warehouseService.UpdateWarehouseEntity(ctx, warehouse)
	if err != nil {
		return nil, err
	}

	// Transformar la entidad de dominio en DTO de respuesta
	return response.NewWarehouseResponse(warehouse), nil
}
