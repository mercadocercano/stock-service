package service

import (
	"context"

	"github.com/hornosg/go-shared/criteria"
	"stock/src/stock_location/domain/entity"
	"stock/src/stock_location/domain/port"
)

// StockLocationService representa el servicio de dominio para las ubicaciones de stock
type StockLocationService struct {
	repository port.StockLocationRepository
}

// NewStockLocationService crea una nueva instancia del servicio de ubicaciones de stock
func NewStockLocationService(repository port.StockLocationRepository) *StockLocationService {
	return &StockLocationService{
		repository: repository,
	}
}

// CreateStockLocation crea una nueva ubicación de stock
func (s *StockLocationService) CreateStockLocation(
	ctx context.Context,
	tenantID string,
	warehouseID string,
	parentID *string,
	name string,
	code string,
	description string,
) (*entity.StockLocation, error) {
	// Crear nueva entidad de ubicación de stock
	stockLocation := entity.NewStockLocation(
		tenantID,
		warehouseID,
		parentID,
		name,
		code,
		description,
	)

	// Guardar en el repositorio
	err := s.repository.Save(ctx, stockLocation)
	if err != nil {
		return nil, err
	}

	return stockLocation, nil
}

// GetStockLocationByID obtiene una ubicación de stock por su ID
func (s *StockLocationService) GetStockLocationByID(ctx context.Context, id string, tenantID string) (*entity.StockLocation, error) {
	return s.repository.FindByID(ctx, id, tenantID)
}

// UpdateStockLocationEntity actualiza la entidad de ubicación de stock en el repositorio
func (s *StockLocationService) UpdateStockLocationEntity(ctx context.Context, stockLocation *entity.StockLocation) error {
	return s.repository.Update(ctx, stockLocation)
}

// DeleteStockLocation elimina una ubicación de stock por su ID
func (s *StockLocationService) DeleteStockLocation(ctx context.Context, id string, tenantID string) error {
	return s.repository.Delete(ctx, id, tenantID)
}

// FindStockLocationsByCriteria busca ubicaciones de stock según criterios
func (s *StockLocationService) FindStockLocationsByCriteria(ctx context.Context, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error) {
	return s.repository.FindByCriteria(ctx, tenantID, crit)
}

// FindStockLocationsByWarehouseID busca ubicaciones de stock por el ID del almacén
func (s *StockLocationService) FindStockLocationsByWarehouseID(ctx context.Context, warehouseID string, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error) {
	return s.repository.FindByWarehouseID(ctx, warehouseID, tenantID, crit)
}

// FindChildrenStockLocations busca ubicaciones de stock hijas de una ubicación padre
func (s *StockLocationService) FindChildrenStockLocations(ctx context.Context, parentID string, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error) {
	return s.repository.FindChildren(ctx, parentID, tenantID, crit)
}

// FindRootStockLocations busca ubicaciones de stock de nivel raíz en un almacén
func (s *StockLocationService) FindRootStockLocations(ctx context.Context, warehouseID string, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error) {
	return s.repository.FindRoots(ctx, warehouseID, tenantID, crit)
}
