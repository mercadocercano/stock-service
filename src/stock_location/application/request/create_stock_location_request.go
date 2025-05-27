package request

import (
	"errors"
	"strings"
)

// CreateStockLocationRequest representa la solicitud para crear una ubicación de stock
type CreateStockLocationRequest struct {
	TenantID    string  `json:"tenant_id"`
	WarehouseID string  `json:"warehouse_id"`
	ParentID    *string `json:"parent_id"`
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Description string  `json:"description"`
}

// Validate valida que la solicitud sea válida
func (r *CreateStockLocationRequest) Validate() error {
	if r.TenantID == "" {
		return errors.New("tenant_id is required")
	}

	if r.WarehouseID == "" {
		return errors.New("warehouse_id is required")
	}

	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}

	if strings.TrimSpace(r.Code) == "" {
		return errors.New("code is required")
	}

	return nil
}
