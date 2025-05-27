package usecase

import (
	"context"

	"stock/src/location/application/response"
	"stock/src/location/domain/service"
)

// DeactivateLocationUseCase define el caso de uso para desactivar una ubicación
type DeactivateLocationUseCase struct {
	locationService *service.LocationService
}

// NewDeactivateLocationUseCase crea una nueva instancia del caso de uso
func NewDeactivateLocationUseCase(locationService *service.LocationService) *DeactivateLocationUseCase {
	return &DeactivateLocationUseCase{
		locationService: locationService,
	}
}

// Execute ejecuta el caso de uso para desactivar una ubicación
func (uc *DeactivateLocationUseCase) Execute(ctx context.Context, tenantID string, locationID string) (*response.LocationResponse, error) {
	// Obtener la ubicación existente
	location, err := uc.locationService.GetLocationByID(ctx, locationID, tenantID)
	if err != nil {
		return nil, err
	}

	// Desactivar la ubicación
	location.Deactivate()

	// Guardar los cambios
	err = uc.locationService.UpdateLocationEntity(ctx, location)
	if err != nil {
		return nil, err
	}

	// Transformar entidad de dominio en DTO de respuesta
	return response.NewLocationResponse(location), nil
}
