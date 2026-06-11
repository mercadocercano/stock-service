package usecase_test

import (
	"context"
	"testing"

	"github.com/hornosg/go-shared/criteria"
	"stock/src/warehouse/application/request"
	"stock/src/warehouse/application/usecase"
	"stock/src/warehouse/domain/entity"
	"stock/test/warehouse/domain/mother"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWarehouseService es un mock del servicio de almacenes
type MockWarehouseService struct {
	mock.Mock
}

func (m *MockWarehouseService) CreateWarehouse(ctx context.Context, tenantID, locationID, name, code string, warehouseType entity.WarehouseType, description string, priority int) (*entity.Warehouse, error) {
	args := m.Called(ctx, tenantID, locationID, name, code, warehouseType, description, priority)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Warehouse), args.Error(1)
}

func (m *MockWarehouseService) GetWarehouseByID(ctx context.Context, id, tenantID string) (*entity.Warehouse, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Warehouse), args.Error(1)
}

func (m *MockWarehouseService) UpdateWarehouseEntity(ctx context.Context, warehouse *entity.Warehouse) error {
	args := m.Called(ctx, warehouse)
	return args.Error(0)
}

func (m *MockWarehouseService) DeleteWarehouse(ctx context.Context, id, tenantID string) error {
	args := m.Called(ctx, id, tenantID)
	return args.Error(0)
}

func (m *MockWarehouseService) ActivateWarehouse(ctx context.Context, id, tenantID string) (*entity.Warehouse, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Warehouse), args.Error(1)
}

func (m *MockWarehouseService) DeactivateWarehouse(ctx context.Context, id, tenantID string) (*entity.Warehouse, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Warehouse), args.Error(1)
}

func (m *MockWarehouseService) FindWarehousesByCriteria(ctx context.Context, tenantID string, crit criteria.Criteria) ([]*entity.Warehouse, int, error) {
	args := m.Called(ctx, tenantID, crit)
	return args.Get(0).([]*entity.Warehouse), args.Int(1), args.Error(2)
}

func (m *MockWarehouseService) FindWarehousesByLocationID(ctx context.Context, locationID, tenantID string, crit criteria.Criteria) ([]*entity.Warehouse, int, error) {
	args := m.Called(ctx, locationID, tenantID, crit)
	return args.Get(0).([]*entity.Warehouse), args.Int(1), args.Error(2)
}

func TestCreateWarehouseUseCase_Execute(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockService := new(MockWarehouseService)

	createUseCase := usecase.NewCreateWarehouseUseCase(mockService)

	tenantID := "tenant-" + uuid.New().String()
	locationID := "location-" + uuid.New().String()
	name := "Test Warehouse"
	code := "WH-TEST"
	warehouseType := entity.RegularWarehouseType
	description := "Test warehouse description"
	priority := 1

	req := request.CreateWarehouseRequest{
		TenantID:    tenantID,
		LocationID:  locationID,
		Name:        name,
		Code:        code,
		Type:        string(warehouseType),
		Description: description,
		Priority:    priority,
	}

	warehouseMother := mother.WarehouseMother{}
	expectedWarehouse := warehouseMother.Random()

	// El servicio debe crear el warehouse y retornarlo
	mockService.On("CreateWarehouse", ctx, tenantID, locationID, name, code, warehouseType, description, priority).Return(expectedWarehouse, nil)

	// Act
	response, err := createUseCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, expectedWarehouse.ID, response.Warehouse.ID)

	// Verificar que se llamó CreateWarehouse
	mockService.AssertExpectations(t)
}

func TestCreateWarehouseUseCase_Execute_ServiceError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockService := new(MockWarehouseService)

	createUseCase := usecase.NewCreateWarehouseUseCase(mockService)

	tenantID := "tenant-" + uuid.New().String()
	locationID := "location-" + uuid.New().String()
	name := "Test Warehouse"
	code := "WH-TEST"
	warehouseType := entity.RegularWarehouseType
	description := "Test warehouse description"
	priority := 1

	req := request.CreateWarehouseRequest{
		TenantID:    tenantID,
		LocationID:  locationID,
		Name:        name,
		Code:        code,
		Type:        string(warehouseType),
		Description: description,
		Priority:    priority,
	}

	// El servicio debe retornar error
	expectedError := assert.AnError
	mockService.On("CreateWarehouse", ctx, tenantID, locationID, name, code, warehouseType, description, priority).Return(nil, expectedError)

	// Act
	response, err := createUseCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, response)

	// Verificar que se llamó CreateWarehouse
	mockService.AssertExpectations(t)
}
