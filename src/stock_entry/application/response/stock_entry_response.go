package response

import (
	"time"
	
	"stock/src/stock_entry/domain/entity"
)

// StockEntryResponse representa la respuesta de una entrada de stock
type StockEntryResponse struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	ProductSKU      string                 `json:"product_sku"`
	ProductID       *string                `json:"product_id,omitempty"`
	ProductName     string                 `json:"product_name,omitempty"`
	LocationID      *string                `json:"location_id,omitempty"`
	EntryType       string                 `json:"entry_type"`
	Quantity        float64                `json:"quantity"`
	UnitOfMeasure   string                 `json:"unit_of_measure"`
	UnitCost        *float64               `json:"unit_cost,omitempty"`
	TotalCost       *float64               `json:"total_cost,omitempty"`
	ReferenceNumber *string                `json:"reference_number,omitempty"`
	Notes           *string                `json:"notes,omitempty"`
	Status          string                 `json:"status"`
	IsActive        bool                   `json:"is_active"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// BulkCreateResponse respuesta de creación masiva
type BulkCreateResponse struct {
	Success       bool                  `json:"success"`
	TotalEntries  int                   `json:"total_entries"`
	EntriesCreated int                  `json:"entries_created"`
	EntriesFailed int                   `json:"entries_failed"`
	Errors        []string              `json:"errors,omitempty"`
	CreatedEntries []StockEntryResponse `json:"created_entries,omitempty"`
}

// StockAvailabilityResponse respuesta de disponibilidad
type StockAvailabilityResponse struct {
	ProductSKU        string     `json:"product_sku"`
	ProductID         *string    `json:"product_id,omitempty"`
	ProductName       string     `json:"product_name,omitempty"`
	LocationID        *string    `json:"location_id,omitempty"`
	AvailableQuantity float64    `json:"available_quantity"`
	ReservedQuantity  float64    `json:"reserved_quantity"`
	TotalQuantity     float64    `json:"total_quantity"`
	UnitOfMeasure     string     `json:"unit_of_measure"`
	AvgUnitCost       *float64   `json:"avg_unit_cost,omitempty"`
	TotalValue        *float64   `json:"total_value,omitempty"`
	IsLowStock        bool       `json:"is_low_stock"`
	IsOutOfStock      bool       `json:"is_out_of_stock"`
	LastEntryAt       *time.Time `json:"last_entry_at,omitempty"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// FromStockEntry convierte una entidad a response
func FromStockEntry(entry *entity.StockEntry) StockEntryResponse {
	resp := StockEntryResponse{
		ID:            entry.ID.String(),
		TenantID:      entry.TenantID.String(),
		ProductSKU:    entry.ProductSKU,
		ProductName:   entry.ProductName,
		EntryType:     string(entry.EntryType),
		Quantity:      entry.Quantity,
		UnitOfMeasure: entry.UnitOfMeasure,
		Status:        string(entry.Status),
		IsActive:      entry.IsActive,
		CreatedAt:     entry.CreatedAt,
		UpdatedAt:     entry.UpdatedAt,
	}
	
	if entry.ProductID != nil {
		productID := entry.ProductID.String()
		resp.ProductID = &productID
	}
	
	if entry.LocationID != nil {
		locationID := entry.LocationID.String()
		resp.LocationID = &locationID
	}
	
	resp.UnitCost = entry.UnitCost
	resp.TotalCost = entry.TotalCost
	resp.ReferenceNumber = entry.ReferenceNumber
	resp.Notes = entry.Notes
	
	return resp
}

// FromStockAvailability convierte una entidad a response
func FromStockAvailability(availability *entity.StockAvailability) StockAvailabilityResponse {
	resp := StockAvailabilityResponse{
		ProductSKU:        availability.ProductSKU,
		ProductName:       availability.ProductName,
		AvailableQuantity: availability.AvailableQuantity,
		ReservedQuantity:  availability.ReservedQuantity,
		TotalQuantity:     availability.TotalQuantity,
		UnitOfMeasure:     availability.UnitOfMeasure,
		IsLowStock:        availability.IsLowStock,
		IsOutOfStock:      availability.IsOutOfStock,
		UpdatedAt:         availability.UpdatedAt,
	}
	
	if availability.ProductID != nil {
		productID := availability.ProductID.String()
		resp.ProductID = &productID
	}
	
	if availability.LocationID != nil {
		locationID := availability.LocationID.String()
		resp.LocationID = &locationID
	}
	
	resp.AvgUnitCost = availability.AvgUnitCost
	resp.TotalValue = availability.TotalValue
	resp.LastEntryAt = availability.LastEntryAt
	
	return resp
}

