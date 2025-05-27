package request

import (
	"errors"
	"strings"
)

// UpdateLocationRequest representa la solicitud para actualizar una ubicación
type UpdateLocationRequest struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
}

// Validate valida que la solicitud sea válida
func (r *UpdateLocationRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
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
