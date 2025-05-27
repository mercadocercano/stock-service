package usecase_test

import (
	"context"
	"testing"

	"stock/src/shared/domain/bus/event"
	"stock/src/warehouse/application/usecase"
	"stock/src/warehouse/domain/entity"
	"stock/src/warehouse/domain/exception"
	"stock/src/warehouse/domain/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWarehouseRepository es un mock del repositorio de almacenes
type MockWarehouseRepository struct {
	mock.Mock
}

func (m *MockWarehouseRepository) Create(ctx context.Context, warehouse *entity.Warehouse) error {
	args := m.Called(ctx, warehouse)
	return args.Error(0)
}

func (m *MockWarehouseRepository) Update(ctx context.Context, warehouse *entity.Warehouse) error {
	args := m.Called(ctx, warehouse)
	return args.Error(0)
}

func (m *MockWarehouseRepository) Delete(ctx context.Context, id string, tenantID string) error {
	args := m.Called(ctx, id, tenantID)
	return args.Error(0)
}

func (m *MockWarehouseRepository) GetByID(ctx context.Context, id string, tenantID string) (*entity.Warehouse, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Warehouse), args.Error(1)
}

func (m *MockWarehouseRepository) FindByCriteria(ctx context.Context, criteria repository.WarehouseCriteria) ([]*entity.Warehouse, error) {
	args := m.Called(ctx, criteria)
	return args.Get(0).([]*entity.Warehouse), args.Error(1)
}

// MockEventBus es un mock del bus de eventos
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, events []event.Event) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func TestCreateWarehouseUseCase_Execute(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockWarehouseRepository)
	mockEventBus := new(MockEventBus)

	createUseCase := usecase.NewCreateWarehouseUseCase(mockRepo, mockEventBus)

	tenantID := "tenant-" + uuid.New().String()
	locationID := "location-" + uuid.New().String()
	name := "Test Warehouse"
	code := "WH-TEST"
	warehouseType := entity.RegularWarehouseType
	description := "Test warehouse description"
	priority := 1

	request := usecase.CreateWarehouseRequest{
		TenantID:    tenantID,
		LocationID:  locationID,
		Name:        name,
		Code:        code,
		Type:        string(warehouseType),
		Description: description,
		Priority:    priority,
	}

	// El repositorio debe llamar a Create una vez y retornar nil (sin error)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.Warehouse")).Return(nil)

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

func TestCreateWarehouseUseCase_Execute_InvalidType(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockWarehouseRepository)
	mockEventBus := new(MockEventBus)

	createUseCase := usecase.NewCreateWarehouseUseCase(mockRepo, mockEventBus)

	request := usecase.CreateWarehouseRequest{
		TenantID:    "tenant-" + uuid.New().String(),
		LocationID:  "location-" + uuid.New().String(),
		Name:        "Test Warehouse",
		Code:        "WH-TEST",
		Type:        "invalid_type", // Tipo inválido
		Description: "Test warehouse description",
		Priority:    1,
	}

	// Act
	response, err := createUseCase.Execute(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &exception.InvalidWarehouseType{}, err)
	assert.Nil(t, response)

	// Verificar que no se llamaron los métodos en los mocks
	mockRepo.AssertNotCalled(t, "Create")
	mockEventBus.AssertNotCalled(t, "Publish")
}

func TestCreateWarehouseUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockWarehouseRepository)
	mockEventBus := new(MockEventBus)

	createUseCase := usecase.NewCreateWarehouseUseCase(mockRepo, mockEventBus)

	tenantID := "tenant-" + uuid.New().String()
	locationID := "location-" + uuid.New().String()
	request := usecase.CreateWarehouseRequest{
		TenantID:    tenantID,
		LocationID:  locationID,
		Name:        "Test Warehouse",
		Code:        "WH-TEST",
		Type:        string(entity.RegularWarehouseType),
		Description: "Test warehouse description",
		Priority:    1,
	}

	// El repositorio debe retornar un error
	expectedError := exception.NewWarehouseCreationError("test error")
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.Warehouse")).Return(expectedError)

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
