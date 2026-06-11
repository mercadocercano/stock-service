package usecase_test

import (
	"context"
	"testing"

	"stock/src/location/application/request"
	"stock/src/location/application/usecase"
	"stock/src/location/domain/entity"
	"github.com/hornosg/go-shared/criteria"
	"stock/test/location/domain/mother"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLocationService es un mock del servicio de ubicaciones
type MockLocationService struct {
	mock.Mock
}

func (m *MockLocationService) CreateLocation(ctx context.Context, tenantID, name string, locationType entity.LocationType,
	address, city, state, country, postalCode, phone, email string) (*entity.Location, error) {
	args := m.Called(ctx, tenantID, name, locationType, address, city, state, country, postalCode, phone, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Location), args.Error(1)
}

func (m *MockLocationService) GetLocationByID(ctx context.Context, id, tenantID string) (*entity.Location, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Location), args.Error(1)
}

func (m *MockLocationService) UpdateLocationEntity(ctx context.Context, location *entity.Location) error {
	args := m.Called(ctx, location)
	return args.Error(0)
}

func (m *MockLocationService) DeleteLocation(ctx context.Context, id, tenantID string) error {
	args := m.Called(ctx, id, tenantID)
	return args.Error(0)
}

func (m *MockLocationService) FindLocationsByCriteria(ctx context.Context, tenantID string, criteria criteria.Criteria) ([]*entity.Location, int, error) {
	args := m.Called(ctx, tenantID, criteria)
	return args.Get(0).([]*entity.Location), args.Int(1), args.Error(2)
}

func (m *MockLocationService) FindStores(ctx context.Context, tenantID string, criteria criteria.Criteria) ([]*entity.Location, int, error) {
	args := m.Called(ctx, tenantID, criteria)
	return args.Get(0).([]*entity.Location), args.Int(1), args.Error(2)
}

func (m *MockLocationService) FindDistributionCenters(ctx context.Context, tenantID string, criteria criteria.Criteria) ([]*entity.Location, int, error) {
	args := m.Called(ctx, tenantID, criteria)
	return args.Get(0).([]*entity.Location), args.Int(1), args.Error(2)
}

func TestCreateLocationUseCase_Execute(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockService := new(MockLocationService)

	createUseCase := usecase.NewCreateLocationUseCase(mockService)

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

	req := request.CreateLocationRequest{
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

	locationMother := mother.LocationMother{}
	expectedLocation := locationMother.Random()

	// El servicio debe crear la location y retornarla
	mockService.On("CreateLocation", ctx, tenantID, name, locationType, address, city, state, country, postalCode, phone, email).Return(expectedLocation, nil)

	// Act
	response, err := createUseCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, expectedLocation.ID, response.Location.ID)

	// Verificar que se llamó CreateLocation
	mockService.AssertExpectations(t)
}

func TestCreateLocationUseCase_Execute_ServiceError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockService := new(MockLocationService)

	createUseCase := usecase.NewCreateLocationUseCase(mockService)

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

	req := request.CreateLocationRequest{
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

	// El servicio debe retornar error
	expectedError := assert.AnError
	mockService.On("CreateLocation", ctx, tenantID, name, locationType, address, city, state, country, postalCode, phone, email).Return(nil, expectedError)

	// Act
	response, err := createUseCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, response)

	// Verificar que se llamó CreateLocation
	mockService.AssertExpectations(t)
}
