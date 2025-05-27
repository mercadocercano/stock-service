package response

import (
	"time"

	"stock/src/warehouse/domain/entity"
)

// WarehouseDTO representa un almacén en el formato de respuesta
type WarehouseDTO struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	LocationID  string    `json:"location_id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WarehouseListResponse representa la respuesta para listar almacenes
type WarehouseListResponse struct {
	Total      int            `json:"total"`
	Count      int            `json:"count"`
	Warehouses []WarehouseDTO `json:"warehouses"`
}

// WarehouseResponse representa la respuesta para un único almacén
type WarehouseResponse struct {
	Warehouse WarehouseDTO `json:"warehouse"`
}

// NewWarehouseResponse crea una nueva respuesta de almacén a partir de una entidad
func NewWarehouseResponse(warehouse *entity.Warehouse) *WarehouseResponse {
	return &WarehouseResponse{
		Warehouse: WarehouseDTO{
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
		},
	}
}

// NewWarehouseResponses convierte una lista de entidades en una lista de respuestas
func NewWarehouseResponses(warehouses []*entity.Warehouse) []*WarehouseResponse {
	responses := make([]*WarehouseResponse, len(warehouses))
	for i, warehouse := range warehouses {
		responses[i] = NewWarehouseResponse(warehouse)
	}
	return responses
}
