package response

import (
	"time"

	"stock/src/stock_location/domain/entity"
)

// StockLocationResponse representa la respuesta con los datos de una ubicación de stock
type StockLocationResponse struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	WarehouseID string    `json:"warehouse_id"`
	ParentID    *string   `json:"parent_id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Path        string    `json:"path"`
	Level       int       `json:"level"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewStockLocationResponse crea una nueva respuesta a partir de una entidad de ubicación de stock
func NewStockLocationResponse(stockLocation *entity.StockLocation) *StockLocationResponse {
	return &StockLocationResponse{
		ID:          stockLocation.ID,
		TenantID:    stockLocation.TenantID,
		WarehouseID: stockLocation.WarehouseID,
		ParentID:    stockLocation.ParentID,
		Name:        stockLocation.Name,
		Code:        stockLocation.Code,
		Path:        stockLocation.Path,
		Level:       stockLocation.Level,
		Description: stockLocation.Description,
		Active:      stockLocation.Active,
		CreatedAt:   stockLocation.CreatedAt,
		UpdatedAt:   stockLocation.UpdatedAt,
	}
}

// StockLocationListResponse representa la respuesta con una lista paginada de ubicaciones de stock
type StockLocationListResponse struct {
	Items      []*StockLocationResponse `json:"items"`
	TotalItems int                      `json:"total_items"`
}

// NewStockLocationListResponse crea una nueva respuesta de lista a partir de entidades y total
func NewStockLocationListResponse(stockLocations []*entity.StockLocation, total int) *StockLocationListResponse {
	items := make([]*StockLocationResponse, len(stockLocations))
	for i, stockLocation := range stockLocations {
		items[i] = NewStockLocationResponse(stockLocation)
	}

	return &StockLocationListResponse{
		Items:      items,
		TotalItems: total,
	}
}
