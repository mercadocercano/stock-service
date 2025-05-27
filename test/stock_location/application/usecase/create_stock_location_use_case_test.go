package usecase_test

import (
	"context"
	"testing"
	"time"

	"stock/src/shared/domain/bus/event"
	"stock/src/stock_location/application/usecase"
	"stock/src/stock_location/domain/entity"
	"stock/src/stock_location/domain/exception"
	"stock/src/stock_location/domain/repository"
	"stock/test/stock_location/domain/mother"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStockLocationRepository es un mock del repositorio de ubicaciones de stock
type MockStockLocationRepository struct {
	mock.Mock
}

func (m *MockStockLocationRepository) Create(ctx context.Context, stockLocation *entity.StockLocation) error {
	args := m.Called(ctx, stockLocation)
	return args.Error(0)
}

func (m *MockStockLocationRepository) Update(ctx context.Context, stockLocation *entity.StockLocation) error {
	args := m.Called(ctx, stockLocation)
	return args.Error(0)
}

func (m *MockStockLocationRepository) Delete(ctx context.Context, id string, tenantID string) error {
	args := m.Called(ctx, id, tenantID)
	return args.Error(0)
}

func (m *MockStockLocationRepository) GetByID(ctx context.Context, id string, tenantID string) (*entity.StockLocation, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.StockLocation), args.Error(1)
}

func (m *MockStockLocationRepository) FindByCriteria(ctx context.Context, criteria repository.StockLocationCriteria) ([]*entity.StockLocation, error) {
	args := m.Called(ctx, criteria)
	return args.Get(0).([]*entity.StockLocation), args.Error(1)
}

// MockEventBus es un mock del bus de eventos
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, events []event.Event) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func TestCreateStockLocationUseCase_Execute(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockStockLocationRepository)
	mockEventBus := new(MockEventBus)

	createUseCase := usecase.NewCreateStockLocationUseCase(mockRepo, mockEventBus)

	tenantID := "tenant-" + uuid.New().String()
	warehouseID := "warehouse-" + uuid.New().String()
	name := "Test Stock Location"
	code := "SL-TEST"
	description := "Test stock location description"

	request := usecase.CreateStockLocationRequest{
		TenantID:    tenantID,
		WarehouseID: warehouseID,
		ParentID:    nil,
		Name:        name,
		Code:        code,
		Description: description,
	}

	// El repositorio debe llamar a Create una vez y retornar nil (sin error)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.StockLocation")).Return(nil)

	// El bus de eventos debe llamar a Publish una vez y retornar nil (sin error)
	mockEventBus.On("Publish", ctx, mock.AnythingOfType("[]event.Event")).Return(nil)

	// Act
	response, err := createUseCase.Execute(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.ID)

	// Verificar que se llamaron los métodos esperados en los mocks
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestCreateStockLocationUseCase_Execute_WithParent(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockStockLocationRepository)
	mockEventBus := new(MockEventBus)

	createUseCase := usecase.NewCreateStockLocationUseCase(mockRepo, mockEventBus)

	tenantID := "tenant-" + uuid.New().String()
	warehouseID := "warehouse-" + uuid.New().String()
	parentID := "parent-" + uuid.New().String()
	name := "Child Stock Location"
	code := "SL-CHILD"
	description := "Child stock location description"

	// Crear un parent para simular su existencia
	parentStockLocation := mother.StockLocationMother{}.Complete(
		parentID,
		tenantID,
		warehouseID,
		nil,
		"Parent Stock Location",
		"SL-PARENT",
		parentID,
		1,
		"Parent stock location description",
		true,
		time.Now(),
		time.Now(),
	)

	request := usecase.CreateStockLocationRequest{
		TenantID:    tenantID,
		WarehouseID: warehouseID,
		ParentID:    &parentID,
		Name:        name,
		Code:        code,
		Description: description,
	}

	// El repositorio debe buscar el parent y luego crear el nuevo stockLocation
	mockRepo.On("GetByID", ctx, parentID, tenantID).Return(parentStockLocation, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.StockLocation")).Return(nil)

	// El bus de eventos debe llamar a Publish una vez y retornar nil (sin error)
	mockEventBus.On("Publish", ctx, mock.AnythingOfType("[]event.Event")).Return(nil)

	// Act
	response, err := createUseCase.Execute(ctx, request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.ID)

	// Verificar que se llamaron los métodos esperados en los mocks
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestCreateStockLocationUseCase_Execute_ParentNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockStockLocationRepository)
	mockEventBus := new(MockEventBus)

	createUseCase := usecase.NewCreateStockLocationUseCase(mockRepo, mockEventBus)

	tenantID := "tenant-" + uuid.New().String()
	warehouseID := "warehouse-" + uuid.New().String()
	parentID := "nonexistent-parent"
	name := "Test Stock Location"
	code := "SL-TEST"
	description := "Test stock location description"

	request := usecase.CreateStockLocationRequest{
		TenantID:    tenantID,
		WarehouseID: warehouseID,
		ParentID:    &parentID,
		Name:        name,
		Code:        code,
		Description: description,
	}

	// El repositorio debe retornar un error al buscar el parent
	expectedError := exception.NewStockLocationNotFoundError(parentID, tenantID)
	mockRepo.On("GetByID", ctx, parentID, tenantID).Return(nil, expectedError)

	// Act
	response, err := createUseCase.Execute(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &exception.StockLocationNotFoundError{}, err)
	assert.Nil(t, response)

	// Verificar que se llamó GetByID pero no Create ni Publish
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Create")
	mockEventBus.AssertNotCalled(t, "Publish")
}

func TestCreateStockLocationUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockStockLocationRepository)
	mockEventBus := new(MockEventBus)

	createUseCase := usecase.NewCreateStockLocationUseCase(mockRepo, mockEventBus)

	tenantID := "tenant-" + uuid.New().String()
	warehouseID := "warehouse-" + uuid.New().String()
	name := "Test Stock Location"
	code := "SL-TEST"
	description := "Test stock location description"

	request := usecase.CreateStockLocationRequest{
		TenantID:    tenantID,
		WarehouseID: warehouseID,
		ParentID:    nil,
		Name:        name,
		Code:        code,
		Description: description,
	}

	// El repositorio debe retornar un error
	expectedError := exception.NewStockLocationCreationError("test error")
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.StockLocation")).Return(expectedError)

	// Act
	response, err := createUseCase.Execute(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, response)

	// Verificar que se llamó Create pero no Publish
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertNotCalled(t, "Publish")
}
