package usecase

import (
	"context"

	"stock/src/location/application/response"
	"stock/src/location/domain/service"
)

// GetLocationUseCase define el caso de uso para obtener una ubicación por su ID
type GetLocationUseCase struct {
	locationService *service.LocationService
}

// NewGetLocationUseCase crea una nueva instancia del caso de uso
func NewGetLocationUseCase(locationService *service.LocationService) *GetLocationUseCase {
	return &GetLocationUseCase{
		locationService: locationService,
	}
}

// Execute ejecuta el caso de uso para obtener una ubicación por su ID
func (uc *GetLocationUseCase) Execute(ctx context.Context, tenantID string, locationID string) (*response.LocationResponse, error) {
	// Obtener ubicación del servicio de dominio
	location, err := uc.locationService.GetLocationByID(ctx, locationID, tenantID)
	if err != nil {
		return nil, err
	}

	// Transformar entidad de dominio en DTO de respuesta
	return response.NewLocationResponse(location), nil
}
