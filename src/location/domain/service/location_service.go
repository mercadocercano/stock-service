package service

import (
	"context"

	"stock/src/location/domain/entity"
	"stock/src/location/domain/port"
	"stock/src/shared/domain/criteria"
)

// LocationService representa el servicio de dominio para las ubicaciones
type LocationService struct {
	repository port.LocationRepository
}

// NewLocationService crea una nueva instancia del servicio de ubicaciones
func NewLocationService(repository port.LocationRepository) *LocationService {
	return &LocationService{
		repository: repository,
	}
}

// CreateLocation crea una nueva ubicación
func (s *LocationService) CreateLocation(ctx context.Context, tenantID, name string, locationType entity.LocationType,
	address, city, state, country, postalCode, phone, email string) (*entity.Location, error) {

	// Crear nueva entidad de ubicación
	location := entity.NewLocation(
		tenantID,
		name,
		locationType,
		address,
		city,
		state,
		country,
		postalCode,
		phone,
		email,
	)

	// Guardar en el repositorio
	err := s.repository.Save(ctx, location)
	if err != nil {
		return nil, err
	}

	return location, nil
}

// GetLocationByID obtiene una ubicación por su ID
func (s *LocationService) GetLocationByID(ctx context.Context, id string, tenantID string) (*entity.Location, error) {
	return s.repository.FindByID(ctx, id, tenantID)
}

// UpdateLocationEntity actualiza la entidad de ubicación en el repositorio
func (s *LocationService) UpdateLocationEntity(ctx context.Context, location *entity.Location) error {
	return s.repository.Update(ctx, location)
}

// DeleteLocation elimina una ubicación
func (s *LocationService) DeleteLocation(ctx context.Context, id string, tenantID string) error {
	return s.repository.Delete(ctx, id, tenantID)
}

// FindLocationsByCriteria busca ubicaciones según criterios
func (s *LocationService) FindLocationsByCriteria(ctx context.Context, tenantID string, criteria criteria.Criteria) ([]*entity.Location, int, error) {
	return s.repository.FindByCriteria(ctx, tenantID, criteria)
}

// FindStores busca solo ubicaciones de tipo tienda
func (s *LocationService) FindStores(ctx context.Context, tenantID string, criteria criteria.Criteria) ([]*entity.Location, int, error) {
	return s.repository.FindStores(ctx, tenantID, criteria)
}

// FindDistributionCenters busca solo ubicaciones de tipo centro de distribución
func (s *LocationService) FindDistributionCenters(ctx context.Context, tenantID string, criteria criteria.Criteria) ([]*entity.Location, int, error) {
	return s.repository.FindDistributionCenters(ctx, tenantID, criteria)
}
