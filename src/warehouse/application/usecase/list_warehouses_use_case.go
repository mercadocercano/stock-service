package usecase

import (
	"context"

	"github.com/mercadocercano/criteria"
	"stock/src/warehouse/application/response"
	"stock/src/warehouse/domain/service"
)

// ListWarehousesUseCase define el caso de uso para listar almacenes
type ListWarehousesUseCase struct {
	warehouseService *service.WarehouseService
}

// NewListWarehousesUseCase crea una nueva instancia del caso de uso
func NewListWarehousesUseCase(warehouseService *service.WarehouseService) *ListWarehousesUseCase {
	return &ListWarehousesUseCase{
		warehouseService: warehouseService,
	}
}

// Execute ejecuta el caso de uso para listar almacenes
func (uc *ListWarehousesUseCase) Execute(ctx context.Context, tenantID string, crit criteria.Criteria) (*response.WarehouseListResponse, error) {
	// Obtener almacenes del servicio de dominio
	warehouses, total, err := uc.warehouseService.FindWarehousesByCriteria(ctx, tenantID, crit)
	if err != nil {
		return nil, err
	}

	// Transformar entidades de dominio en DTO de respuesta
	warehouseDTOs := make([]response.WarehouseDTO, 0, len(warehouses))
	for _, warehouse := range warehouses {
		warehouseDTOs = append(warehouseDTOs, response.WarehouseDTO{
			ID:          warehouse.ID,
			TenantID:    warehouse.TenantID,
			LocationID:  warehouse.LocationID,
			Name:        warehouse.Name,
			Code:        warehouse.Code,
			Type:        string(warehouse.Type),
			Description: warehouse.Description,
			Priority:    warehouse.Priority,
			Active:      warehouse.Active,
			CreatedAt:   warehouse.CreatedAt,
			UpdatedAt:   warehouse.UpdatedAt,
		})
	}

	// Construir respuesta
	resp := &response.WarehouseListResponse{
		Total:      total,
		Count:      len(warehouseDTOs),
		Warehouses: warehouseDTOs,
	}

	return resp, nil
}

// ExecuteByLocationID ejecuta el caso de uso para listar almacenes por ubicación
func (uc *ListWarehousesUseCase) ExecuteByLocationID(ctx context.Context, locationID string, tenantID string, crit criteria.Criteria) (*response.WarehouseListResponse, error) {
	// Obtener almacenes del servicio de dominio filtrados por locationID
	warehouses, total, err := uc.warehouseService.FindWarehousesByLocationID(ctx, locationID, tenantID, crit)
	if err != nil {
		return nil, err
	}

	// Transformar entidades de dominio en DTO de respuesta
	warehouseDTOs := make([]response.WarehouseDTO, 0, len(warehouses))
	for _, warehouse := range warehouses {
		warehouseDTOs = append(warehouseDTOs, response.WarehouseDTO{
			ID:          warehouse.ID,
			TenantID:    warehouse.TenantID,
			LocationID:  warehouse.LocationID,
			Name:        warehouse.Name,
			Code:        warehouse.Code,
			Type:        string(warehouse.Type),
			Description: warehouse.Description,
			Priority:    warehouse.Priority,
			Active:      warehouse.Active,
			CreatedAt:   warehouse.CreatedAt,
			UpdatedAt:   warehouse.UpdatedAt,
		})
	}

	// Construir respuesta
	resp := &response.WarehouseListResponse{
		Total:      total,
		Count:      len(warehouseDTOs),
		Warehouses: warehouseDTOs,
	}

	return resp, nil
}
