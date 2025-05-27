package usecase

import (
	"context"

	"stock/src/location/application/request"
	"stock/src/location/application/response"
	"stock/src/location/domain/entity"
	"stock/src/location/domain/service"
)

// CreateLocationUseCase representa el caso de uso para crear una ubicación
type CreateLocationUseCase struct {
	locationService *service.LocationService
}

// NewCreateLocationUseCase crea una nueva instancia del caso de uso
func NewCreateLocationUseCase(locationService *service.LocationService) *CreateLocationUseCase {
	return &CreateLocationUseCase{
		locationService: locationService,
	}
}

// Execute ejecuta el caso de uso
func (uc *CreateLocationUseCase) Execute(ctx context.Context, req request.CreateLocationRequest) (*response.LocationResponse, error) {
	// Validar la solicitud
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Crear la ubicación en el dominio
	location, err := uc.locationService.CreateLocation(
		ctx,
		req.TenantID,
		req.Name,
		entity.LocationType(req.Type),
		req.Address,
		req.City,
		req.State,
		req.Country,
		req.PostalCode,
		req.Phone,
		req.Email,
	)

	if err != nil {
		return nil, err
	}

	// Crear la respuesta
	return response.NewLocationResponse(location), nil
}
