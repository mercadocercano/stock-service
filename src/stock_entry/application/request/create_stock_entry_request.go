package request

import "github.com/google/uuid"

// CreateStockEntryRequest representa una petición para crear una entrada de stock
type CreateStockEntryRequest struct {
	TenantID        string  `json:"tenant_id"`
	ProductSKU      string  `json:"product_sku" binding:"required"`
	ProductID       string  `json:"product_id,omitempty"`
	ProductName     string  `json:"product_name,omitempty"`
	LocationID      string  `json:"location_id,omitempty"`
	EntryType       string  `json:"entry_type" binding:"required"`
	Quantity        float64 `json:"quantity" binding:"required"`
	UnitOfMeasure   string  `json:"unit_of_measure"`
	UnitCost        float64 `json:"unit_cost,omitempty"`
	ReferenceNumber string  `json:"reference_number,omitempty"`
	Notes           string  `json:"notes,omitempty"`
}

// BulkCreateStockEntriesRequest representa una petición para crear múltiples entradas
type BulkCreateStockEntriesRequest struct {
	TenantID string                     `json:"tenant_id"`
	Entries  []CreateStockEntryRequest  `json:"entries" binding:"required,min=1"`
}

// Validate valida la petición
func (r *CreateStockEntryRequest) Validate() error {
	if r.ProductSKU == "" {
		return ErrProductSKURequired
	}
	
	if r.Quantity == 0 {
		return ErrInvalidQuantity
	}
	
	if r.EntryType == "" {
		return ErrEntryTypeRequired
	}
	
	return nil
}

// ParseProductID parsea el product ID si existe
func (r *CreateStockEntryRequest) ParseProductID() (*uuid.UUID, error) {
	if r.ProductID == "" {
		return nil, nil
	}
	
	id, err := uuid.Parse(r.ProductID)
	if err != nil {
		return nil, err
	}
	
	return &id, nil
}

// ParseLocationID parsea el location ID si existe
func (r *CreateStockEntryRequest) ParseLocationID() (*uuid.UUID, error) {
	if r.LocationID == "" {
		return nil, nil
	}
	
	id, err := uuid.Parse(r.LocationID)
	if err != nil {
		return nil, err
	}
	
	return &id, nil
}

// Errors comunes
var (
	ErrProductSKURequired = &ValidationError{Message: "product_sku is required"}
	ErrInvalidQuantity    = &ValidationError{Message: "quantity must be non-zero"}
	ErrEntryTypeRequired  = &ValidationError{Message: "entry_type is required"}
)

// ValidationError error de validación
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

