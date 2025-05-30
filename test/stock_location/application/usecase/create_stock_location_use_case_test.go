package usecase_test

import (
	"context"
	"testing"

	"stock/src/shared/domain/criteria"
	"stock/src/stock_location/application/request"
	"stock/src/stock_location/application/usecase"
	"stock/src/stock_location/domain/entity"
	"stock/src/stock_location/domain/exception"
	"stock/test/stock_location/domain/mother"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStockLocationService es un mock del servicio de ubicaciones de stock
type MockStockLocationService struct {
	mock.Mock
}

func (m *MockStockLocationService) CreateStockLocation(ctx context.Context, tenantID, warehouseID string, parentID *string, name, code, description string) (*entity.StockLocation, error) {
	args := m.Called(ctx, tenantID, warehouseID, parentID, name, code, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.StockLocation), args.Error(1)
}

func (m *MockStockLocationService) GetStockLocationByID(ctx context.Context, id, tenantID string) (*entity.StockLocation, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.StockLocation), args.Error(1)
}

func (m *MockStockLocationService) UpdateStockLocationEntity(ctx context.Context, stockLocation *entity.StockLocation) error {
	args := m.Called(ctx, stockLocation)
	return args.Error(0)
}

func (m *MockStockLocationService) DeleteStockLocation(ctx context.Context, id, tenantID string) error {
	args := m.Called(ctx, id, tenantID)
	return args.Error(0)
}

func (m *MockStockLocationService) FindStockLocationsByCriteria(ctx context.Context, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error) {
	args := m.Called(ctx, tenantID, crit)
	return args.Get(0).([]*entity.StockLocation), args.Int(1), args.Error(2)
}

func (m *MockStockLocationService) FindStockLocationsByWarehouseID(ctx context.Context, warehouseID, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error) {
	args := m.Called(ctx, warehouseID, tenantID, crit)
	return args.Get(0).([]*entity.StockLocation), args.Int(1), args.Error(2)
}

func (m *MockStockLocationService) FindChildrenStockLocations(ctx context.Context, parentID, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error) {
	args := m.Called(ctx, parentID, tenantID, crit)
	return args.Get(0).([]*entity.StockLocation), args.Int(1), args.Error(2)
}

func (m *MockStockLocationService) FindRootStockLocations(ctx context.Context, warehouseID, tenantID string, crit criteria.Criteria) ([]*entity.StockLocation, int, error) {
	args := m.Called(ctx, warehouseID, tenantID, crit)
	return args.Get(0).([]*entity.StockLocation), args.Int(1), args.Error(2)
}

func TestCreateStockLocationUseCase_Execute(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockService := new(MockStockLocationService)

	createUseCase := usecase.NewCreateStockLocationUseCase(mockService)

	tenantID := "tenant-" + uuid.New().String()
	warehouseID := "warehouse-" + uuid.New().String()
	name := "Test Stock Location"
	code := "SL-TEST"
	description := "Test stock location description"

	req := request.CreateStockLocationRequest{
		TenantID:    tenantID,
		WarehouseID: warehouseID,
		ParentID:    nil,
		Name:        name,
		Code:        code,
		Description: description,
	}

	stockLocationMother := mother.StockLocationMother{}
	expectedStockLocation := stockLocationMother.Random()

	// El servicio debe crear la stock location y retornarla
	mockService.On("CreateStockLocation", ctx, tenantID, warehouseID, (*string)(nil), name, code, description).Return(expectedStockLocation, nil)

	// Act
	response, err := createUseCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, expectedStockLocation.ID, response.ID)

	// Verificar que se llamó CreateStockLocation
	mockService.AssertExpectations(t)
}

func TestCreateStockLocationUseCase_Execute_WithParent(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockService := new(MockStockLocationService)

	createUseCase := usecase.NewCreateStockLocationUseCase(mockService)

	tenantID := "tenant-" + uuid.New().String()
	warehouseID := "warehouse-" + uuid.New().String()
	parentID := "parent-" + uuid.New().String()
	name := "Child Stock Location"
	code := "SL-CHILD"
	description := "Child stock location description"

	req := request.CreateStockLocationRequest{
		TenantID:    tenantID,
		WarehouseID: warehouseID,
		ParentID:    &parentID,
		Name:        name,
		Code:        code,
		Description: description,
	}

	stockLocationMother := mother.StockLocationMother{}
	expectedStockLocation := stockLocationMother.Random()

	// El servicio debe crear la stock location con parent
	mockService.On("CreateStockLocation", ctx, tenantID, warehouseID, &parentID, name, code, description).Return(expectedStockLocation, nil)

	// Act
	response, err := createUseCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, expectedStockLocation.ID, response.ID)

	// Verificar que se llamó CreateStockLocation
	mockService.AssertExpectations(t)
}

func TestCreateStockLocationUseCase_Execute_ServiceError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockService := new(MockStockLocationService)

	createUseCase := usecase.NewCreateStockLocationUseCase(mockService)

	tenantID := "tenant-" + uuid.New().String()
	warehouseID := "warehouse-" + uuid.New().String()
	name := "Test Stock Location"
	code := "SL-TEST"
	description := "Test stock location description"

	req := request.CreateStockLocationRequest{
		TenantID:    tenantID,
		WarehouseID: warehouseID,
		ParentID:    nil,
		Name:        name,
		Code:        code,
		Description: description,
	}

	// El servicio debe retornar error
	expectedError := exception.NewStockLocationNotFoundError("parent-id", tenantID)
	mockService.On("CreateStockLocation", ctx, tenantID, warehouseID, (*string)(nil), name, code, description).Return(nil, expectedError)

	// Act
	response, err := createUseCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, response)

	// Verificar que se llamó CreateStockLocation
	mockService.AssertExpectations(t)
}
