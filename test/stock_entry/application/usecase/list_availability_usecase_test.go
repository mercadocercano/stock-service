package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/src/stock_entry/application/usecase"
	"stock/src/stock_entry/domain/entity"
	mockRepo "stock/test/stock_entry/infrastructure/persistence/repository"
	"stock/test/stock_entry/domain/mother"
)

func TestListAvailabilityUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewListAvailabilityUseCase(mockAvailRepo)

	tenantID := uuid.New()
	availMother := mother.StockAvailabilityMother{}

	avail1 := availMother.WithQuantities(50, 0)
	avail1.TenantID = tenantID
	avail2 := availMother.WithQuantities(30, 5)
	avail2.TenantID = tenantID

	mockAvailRepo.On("CountByTenant", ctx, tenantID).Return(2, nil)
	mockAvailRepo.On("FindByTenant", ctx, tenantID, 20, 0).Return([]*entity.StockAvailability{avail1, avail2}, nil)

	// Act
	result, err := uc.Execute(ctx, tenantID.String(), 1, 20)

	// Assert
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 2, result.TotalCount)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PageSize)
	assert.Equal(t, 1, result.TotalPages)
	mockAvailRepo.AssertExpectations(t)
}

func TestListAvailabilityUseCase_Execute_DefaultPagination(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewListAvailabilityUseCase(mockAvailRepo)

	tenantID := uuid.New()

	mockAvailRepo.On("CountByTenant", ctx, tenantID).Return(0, nil)
	mockAvailRepo.On("FindByTenant", ctx, tenantID, 20, 0).Return([]*entity.StockAvailability{}, nil)

	// Act: pagina 0, pageSize 0 => se corrigen a 1 y 20
	result, err := uc.Execute(ctx, tenantID.String(), 0, 0)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PageSize)
	mockAvailRepo.AssertExpectations(t)
}

func TestListAvailabilityUseCase_Execute_PageSizeLimit(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewListAvailabilityUseCase(mockAvailRepo)

	tenantID := uuid.New()

	mockAvailRepo.On("CountByTenant", ctx, tenantID).Return(0, nil)
	mockAvailRepo.On("FindByTenant", ctx, tenantID, 20, 0).Return([]*entity.StockAvailability{}, nil)

	// Act: pageSize > 500 => se corrige a 20
	result, err := uc.Execute(ctx, tenantID.String(), 1, 1000)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 20, result.PageSize)
	mockAvailRepo.AssertExpectations(t)
}

func TestListAvailabilityUseCase_Execute_InvalidTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewListAvailabilityUseCase(mockAvailRepo)

	// Act
	_, err := uc.Execute(ctx, "bad-uuid", 1, 20)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tenant_id")
}

func TestListAvailabilityUseCase_Execute_CountError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewListAvailabilityUseCase(mockAvailRepo)

	tenantID := uuid.New()
	mockAvailRepo.On("CountByTenant", ctx, tenantID).Return(0, errors.New("db error"))

	// Act
	_, err := uc.Execute(ctx, tenantID.String(), 1, 20)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error counting availability")
}

func TestListAvailabilityUseCase_Execute_FindError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewListAvailabilityUseCase(mockAvailRepo)

	tenantID := uuid.New()
	mockAvailRepo.On("CountByTenant", ctx, tenantID).Return(5, nil)
	mockAvailRepo.On("FindByTenant", ctx, tenantID, 20, 0).Return(nil, errors.New("db error"))

	// Act
	_, err := uc.Execute(ctx, tenantID.String(), 1, 20)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error listing availability")
}

func TestListAvailabilityUseCase_Execute_TotalPagesCalculation(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewListAvailabilityUseCase(mockAvailRepo)

	tenantID := uuid.New()

	// 25 registros con pageSize 10 => 3 paginas
	mockAvailRepo.On("CountByTenant", ctx, tenantID).Return(25, nil)
	mockAvailRepo.On("FindByTenant", ctx, tenantID, 10, 0).Return([]*entity.StockAvailability{}, nil)

	// Act
	result, err := uc.Execute(ctx, tenantID.String(), 1, 10)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 3, result.TotalPages)
	assert.Equal(t, 25, result.TotalCount)
}
