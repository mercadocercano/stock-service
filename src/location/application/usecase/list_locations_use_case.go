package usecase

import (
	"context"

	"stock/src/location/application/response"
	"stock/src/location/domain/service"
	"github.com/mercadocercano/criteria"
)

// ListLocationsUseCase define el caso de uso para listar ubicaciones
type ListLocationsUseCase struct {
	locationService *service.LocationService
}

// NewListLocationsUseCase crea una nueva instancia del caso de uso
func NewListLocationsUseCase(locationService *service.LocationService) *ListLocationsUseCase {
	return &ListLocationsUseCase{
		locationService: locationService,
	}
}

// Execute ejecuta el caso de uso para listar ubicaciones
func (uc *ListLocationsUseCase) Execute(ctx context.Context, tenantID string, crit criteria.Criteria) (*response.LocationListResponse, error) {
	// Obtener ubicaciones del servicio de dominio
	locations, total, err := uc.locationService.FindLocationsByCriteria(ctx, tenantID, crit)
	if err != nil {
		return nil, err
	}

	// Transformar entidades de dominio en DTO de respuesta
	locationDTOs := make([]response.LocationDTO, 0, len(locations))
	for _, location := range locations {
		locationDTOs = append(locationDTOs, response.LocationDTO{
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
		})
	}

	// Construir respuesta
	response := &response.LocationListResponse{
		Total:     total,
		Count:     len(locationDTOs),
		Locations: locationDTOs,
	}

	return response, nil
}
