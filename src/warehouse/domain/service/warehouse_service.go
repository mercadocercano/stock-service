package service

import (
	"context"

	"stock/src/shared/domain/criteria"
	"stock/src/warehouse/domain/entity"
	"stock/src/warehouse/domain/port"
)

// WarehouseService representa el servicio de dominio para los almacenes
type WarehouseService struct {
	repository port.WarehouseRepository
}

// NewWarehouseService crea una nueva instancia del servicio de almacenes
func NewWarehouseService(repository port.WarehouseRepository) *WarehouseService {
	return &WarehouseService{
		repository: repository,
	}
}

// CreateWarehouse crea un nuevo almacén
func (s *WarehouseService) CreateWarehouse(
	ctx context.Context,
	tenantID string,
	locationID string,
	name string,
	code string,
	warehouseType entity.WarehouseType,
	description string,
	priority int,
) (*entity.Warehouse, error) {
	// Crear nueva entidad de almacén
	warehouse := entity.NewWarehouse(
		tenantID,
		locationID,
		name,
		code,
		warehouseType,
		description,
		priority,
	)

	// Guardar en el repositorio
	err := s.repository.Save(ctx, warehouse)
	if err != nil {
		return nil, err
	}

	return warehouse, nil
}

// GetWarehouseByID obtiene un almacén por su ID
func (s *WarehouseService) GetWarehouseByID(ctx context.Context, id string, tenantID string) (*entity.Warehouse, error) {
	return s.repository.FindByID(ctx, id, tenantID)
}

// UpdateWarehouseEntity actualiza la entidad del almacén en el repositorio
func (s *WarehouseService) UpdateWarehouseEntity(ctx context.Context, warehouse *entity.Warehouse) error {
	return s.repository.Update(ctx, warehouse)
}

// DeleteWarehouse elimina un almacén por su ID
func (s *WarehouseService) DeleteWarehouse(ctx context.Context, id string, tenantID string) error {
	return s.repository.Delete(ctx, id, tenantID)
}

// ActivateWarehouse activa un almacén
func (s *WarehouseService) ActivateWarehouse(ctx context.Context, id string, tenantID string) (*entity.Warehouse, error) {
	// Obtener el almacén existente
	warehouse, err := s.repository.FindByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	// Activar el almacén
	warehouse.Activate()

	// Guardar los cambios
	err = s.repository.Update(ctx, warehouse)
	if err != nil {
		return nil, err
	}

	return warehouse, nil
}

// DeactivateWarehouse desactiva un almacén
func (s *WarehouseService) DeactivateWarehouse(ctx context.Context, id string, tenantID string) (*entity.Warehouse, error) {
	// Obtener el almacén existente
	warehouse, err := s.repository.FindByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	// Desactivar el almacén
	warehouse.Deactivate()

	// Guardar los cambios
	err = s.repository.Update(ctx, warehouse)
	if err != nil {
		return nil, err
	}

	return warehouse, nil
}

// FindWarehousesByCriteria busca almacenes según criterios
func (s *WarehouseService) FindWarehousesByCriteria(ctx context.Context, tenantID string, crit criteria.Criteria) ([]*entity.Warehouse, int, error) {
	return s.repository.FindByCriteria(ctx, tenantID, crit)
}

// FindWarehousesByLocationID busca almacenes por el ID de su ubicación
func (s *WarehouseService) FindWarehousesByLocationID(ctx context.Context, locationID string, tenantID string, crit criteria.Criteria) ([]*entity.Warehouse, int, error) {
	return s.repository.FindByLocationID(ctx, locationID, tenantID, crit)
}
