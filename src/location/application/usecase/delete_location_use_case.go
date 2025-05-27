package usecase

import (
	"context"

	"stock/src/location/domain/service"
)

// DeleteLocationUseCase define el caso de uso para eliminar una ubicación
type DeleteLocationUseCase struct {
	locationService *service.LocationService
}

// NewDeleteLocationUseCase crea una nueva instancia del caso de uso
func NewDeleteLocationUseCase(locationService *service.LocationService) *DeleteLocationUseCase {
	return &DeleteLocationUseCase{
		locationService: locationService,
	}
}

// Execute ejecuta el caso de uso para eliminar una ubicación
func (uc *DeleteLocationUseCase) Execute(ctx context.Context, tenantID string, locationID string) error {
	// Delegar la eliminación al servicio de dominio
	return uc.locationService.DeleteLocation(ctx, locationID, tenantID)
}
