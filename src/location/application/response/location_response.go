package response

import (
	"time"

	"stock/src/location/domain/entity"
)

// LocationDTO representa una ubicación en el formato de respuesta
type LocationDTO struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Address    string    `json:"address"`
	City       string    `json:"city"`
	State      string    `json:"state"`
	Country    string    `json:"country"`
	PostalCode string    `json:"postal_code"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// LocationListResponse representa la respuesta para listar ubicaciones
type LocationListResponse struct {
	Total     int           `json:"total"`
	Count     int           `json:"count"`
	Locations []LocationDTO `json:"locations"`
}

// LocationResponse representa la respuesta para una única ubicación
type LocationResponse struct {
	Location LocationDTO `json:"location"`
}

// NewLocationResponse crea una nueva respuesta de ubicación a partir de una entidad
func NewLocationResponse(location *entity.Location) *LocationResponse {
	return &LocationResponse{
		Location: LocationDTO{
			ID:         location.ID,
			TenantID:   location.TenantID,
			Name:       location.Name,
			Type:       string(location.Type),
			Address:    location.Address,
			City:       location.City,
			State:      location.State,
			Country:    location.Country,
			PostalCode: location.PostalCode,
			Phone:      location.Phone,
			Email:      location.Email,
			Active:     location.Active,
			CreatedAt:  location.CreatedAt,
			UpdatedAt:  location.UpdatedAt,
		},
	}
}

// NewLocationResponses convierte una lista de entidades en una lista de respuestas
func NewLocationResponses(locations []*entity.Location) []*LocationResponse {
	responses := make([]*LocationResponse, len(locations))
	for i, location := range locations {
		responses[i] = NewLocationResponse(location)
	}
	return responses
}
