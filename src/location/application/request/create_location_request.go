package request

import (
	"errors"
	"strings"
)

// CreateLocationRequest representa la solicitud para crear una ubicación
type CreateLocationRequest struct {
	TenantID   string `json:"tenant_id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Address    string `json:"address"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
}

// Validate valida que la solicitud sea válida
func (r *CreateLocationRequest) Validate() error {
	if r.TenantID == "" {
		return errors.New("tenant_id is required")
	}

	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}

	if r.Type != "store" && r.Type != "distribution_center" {
		return errors.New("type must be 'store' or 'distribution_center'")
	}

	if strings.TrimSpace(r.Address) == "" {
		return errors.New("address is required")
	}

	if strings.TrimSpace(r.City) == "" {
		return errors.New("city is required")
	}

	if strings.TrimSpace(r.Country) == "" {
		return errors.New("country is required")
	}

	return nil
}
