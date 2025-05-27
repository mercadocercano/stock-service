package usecase

import (
	"context"

	"stock/src/location/application/response"
	"stock/src/location/domain/service"
)

// ActivateLocationUseCase define el caso de uso para activar una ubicación
type ActivateLocationUseCase struct {
	locationService *service.LocationService
}

// NewActivateLocationUseCase crea una nueva instancia del caso de uso
func NewActivateLocationUseCase(locationService *service.LocationService) *ActivateLocationUseCase {
	return &ActivateLocationUseCase{
		locationService: locationService,
	}
}

// Execute ejecuta el caso de uso para activar una ubicación
func (uc *ActivateLocationUseCase) Execute(ctx context.Context, tenantID string, locationID string) (*response.LocationResponse, error) {
	// Obtener la ubicación existente
	location, err := uc.locationService.GetLocationByID(ctx, locationID, tenantID)
	if err != nil {
		return nil, err
	}

	// Activar la ubicación
	location.Activate()

	// Guardar los cambios
	err = uc.locationService.UpdateLocationEntity(ctx, location)
	if err != nil {
		return nil, err
	}

	// Transformar entidad de dominio en DTO de respuesta
	return response.NewLocationResponse(location), nil
}
