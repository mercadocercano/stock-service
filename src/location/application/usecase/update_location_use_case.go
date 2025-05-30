package usecase

import (
	"context"

	"stock/src/location/application/request"
	"stock/src/location/application/response"
	"stock/src/location/domain/service"
)

// UpdateLocationUseCase define el caso de uso para actualizar una ubicación
type UpdateLocationUseCase struct {
	locationService *service.LocationService
}

// NewUpdateLocationUseCase crea una nueva instancia del caso de uso
func NewUpdateLocationUseCase(locationService *service.LocationService) *UpdateLocationUseCase {
	return &UpdateLocationUseCase{
		locationService: locationService,
	}
}

// Execute ejecuta el caso de uso para actualizar una ubicación
func (uc *UpdateLocationUseCase) Execute(ctx context.Context, tenantID string, locationID string, req request.UpdateLocationRequest) (*response.LocationResponse, error) {
	// Obtener la ubicación existente
	location, err := uc.locationService.GetLocationByID(ctx, locationID, tenantID)
	if err != nil {
		return nil, err
	}

	// Actualizar los datos
	location.Update(req.Name, req.Address, req.City, req.State, req.Country, req.PostalCode, req.Phone, req.Email)

	// Guardar los cambios
	err = uc.locationService.UpdateLocationEntity(ctx, location)
	if err != nil {
		return nil, err
	}

	// Transformar la entidad de dominio en DTO de respuesta
	return response.NewLocationResponse(location), nil
}
