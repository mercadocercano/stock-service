package usecase_test

import (
	"context"
	"testing"

	"stock/src/location/application/usecase"
	"stock/src/location/domain/entity"
	"stock/src/location/domain/exception"
	"stock/src/location/domain/repository"
	"stock/src/shared/domain/bus/event"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLocationRepository es un mock del repositorio de ubicaciones
type MockLocationRepository struct {
	mock.Mock
}

func (m *MockLocationRepository) Create(ctx context.Context, location *entity.Location) error {
	args := m.Called(ctx, location)
	return args.Error(0)
}

func (m *MockLocationRepository) Update(ctx context.Context, location *entity.Location) error {
	args := m.Called(ctx, location)
	return args.Error(0)
}

func (m *MockLocationRepository) Delete(ctx context.Context, id string, tenantID string) error {
	args := m.Called(ctx, id, tenantID)
	return args.Error(0)
}

func (m *MockLocationRepository) GetByID(ctx context.Context, id string, tenantID string) (*entity.Location, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Location), args.Error(1)
}

func (m *MockLocationRepository) FindByCriteria(ctx context.Context, criteria repository.LocationCriteria) ([]*entity.Location, error) {
	args := m.Called(ctx, criteria)
	return args.Get(0).([]*entity.Location), args.Error(1)
}

// MockEventBus es un mock del bus de eventos
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, events []event.Event) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func TestCreateLocationUseCase_Execute(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockLocationRepository)
	mockEventBus := new(MockEventBus)

	createUseCase := usecase.NewCreateLocationUseCase(mockRepo, mockEventBus)

	tenantID := "tenant-" + uuid.New().String()
	name := "Test Location"
	locationType := entity.StoreType
	address := "123 Test St"
	city := "Test City"
	state := "Test State"
	country := "Test Country"
	postalCode := "12345"
	phone := "+1234567890"
	email := "test@example.com"

	request := usecase.CreateLocationRequest{
		TenantID:   tenantID,
		Name:       name,
		Type:       string(locationType),
		Address:    address,
		City:       city,
		State:      state,
		Country:    country,
		PostalCode: postalCode,
		Phone:      phone,
		Email:      email,
	}

	// El repositorio debe llamar a Create una vez y retornar nil (sin error)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.Location")).Return(nil)

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

func TestCreateLocationUseCase_Execute_InvalidType(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockLocationRepository)
	mockEventBus := new(MockEventBus)

	createUseCase := usecase.NewCreateLocationUseCase(mockRepo, mockEventBus)

	request := usecase.CreateLocationRequest{
		TenantID:   "tenant-" + uuid.New().String(),
		Name:       "Test Location",
		Type:       "invalid_type", // Tipo inválido
		Address:    "123 Test St",
		City:       "Test City",
		State:      "Test State",
		Country:    "Test Country",
		PostalCode: "12345",
		Phone:      "+1234567890",
		Email:      "test@example.com",
	}

	// Act
	response, err := createUseCase.Execute(ctx, request)

	// Assert
	assert.Error(t, err)
	assert.IsType(t, &exception.InvalidLocationType{}, err)
	assert.Nil(t, response)

	// Verificar que no se llamaron los métodos en los mocks
	mockRepo.AssertNotCalled(t, "Create")
	mockEventBus.AssertNotCalled(t, "Publish")
}

func TestCreateLocationUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockLocationRepository)
	mockEventBus := new(MockEventBus)

	createUseCase := usecase.NewCreateLocationUseCase(mockRepo, mockEventBus)

	tenantID := "tenant-" + uuid.New().String()
	request := usecase.CreateLocationRequest{
		TenantID:   tenantID,
		Name:       "Test Location",
		Type:       string(entity.StoreType),
		Address:    "123 Test St",
		City:       "Test City",
		State:      "Test State",
		Country:    "Test Country",
		PostalCode: "12345",
		Phone:      "+1234567890",
		Email:      "test@example.com",
	}

	// El repositorio debe retornar un error
	expectedError := exception.NewLocationCreationError("test error")
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.Location")).Return(expectedError)

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
