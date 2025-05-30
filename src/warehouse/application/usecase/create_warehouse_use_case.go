package usecase

import (
	"context"

	"stock/src/warehouse/application/request"
	"stock/src/warehouse/application/response"
	"stock/src/warehouse/domain/entity"
	"stock/src/warehouse/domain/service"
)

// CreateWarehouseUseCase define el caso de uso para crear un almacén
type CreateWarehouseUseCase struct {
	warehouseService service.WarehouseServiceInterface
}

// NewCreateWarehouseUseCase crea una nueva instancia del caso de uso
func NewCreateWarehouseUseCase(warehouseService service.WarehouseServiceInterface) *CreateWarehouseUseCase {
	return &CreateWarehouseUseCase{
		warehouseService: warehouseService,
	}
}

// Execute ejecuta el caso de uso para crear un almacén
func (uc *CreateWarehouseUseCase) Execute(ctx context.Context, req request.CreateWarehouseRequest) (*response.WarehouseResponse, error) {
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

	// Crear el almacén a través del servicio de dominio
	warehouse, err := uc.warehouseService.CreateWarehouse(
		ctx,
		req.TenantID,
		req.LocationID,
		req.Name,
		req.Code,
		warehouseType,
		req.Description,
		req.Priority,
	)

	if err != nil {
		return nil, err
	}

	// Transformar la entidad de dominio en DTO de respuesta
	return response.NewWarehouseResponse(warehouse), nil
}
