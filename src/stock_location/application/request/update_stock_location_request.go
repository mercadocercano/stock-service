package request

import (
	"errors"
	"strings"
)

// UpdateStockLocationRequest representa la solicitud para actualizar una ubicación de stock
type UpdateStockLocationRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

// Validate valida que la solicitud sea válida
func (r *UpdateStockLocationRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}

	if strings.TrimSpace(r.Code) == "" {
		return errors.New("code is required")
	}

	return nil
}
