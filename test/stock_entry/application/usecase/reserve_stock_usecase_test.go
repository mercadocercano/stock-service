package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"stock/src/stock_entry/application/request"
	"stock/src/stock_entry/application/usecase"
	"stock/src/stock_entry/domain/exception"
	mockRepo "stock/test/stock_entry/infrastructure/persistence/repository"
	"stock/test/stock_entry/domain/mother"
)

func TestReserveStockUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewReserveStockUseCase(mockAvailRepo, mockEntryRepo)

	tenantID := uuid.New()
	availMother := mother.StockAvailabilityMother{}
	avail := availMother.WithQuantities(50, 0)
	avail.TenantID = tenantID

	req := &request.ReserveStockRequest{
		SKU:       "SKU-001",
		Quantity:  10,
		Reference: "ORDER-001",
	}

	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "SKU-001").Return(avail, nil)
	mockAvailRepo.On("Update", ctx, mock.AnythingOfType("*entity.StockAvailability")).Return(nil)

	// Act
	resp, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "SKU-001", resp.SKU)
	assert.Equal(t, 10, resp.ReservedQty)
	assert.Equal(t, 40, resp.RemainingQty)
	assert.Equal(t, "ORDER-001", resp.Reference)
	mockAvailRepo.AssertExpectations(t)
}

func TestReserveStockUseCase_Execute_InsufficientStock(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewReserveStockUseCase(mockAvailRepo, mockEntryRepo)

	tenantID := uuid.New()
	availMother := mother.StockAvailabilityMother{}
	avail := availMother.WithQuantities(5, 0)
	avail.TenantID = tenantID

	req := &request.ReserveStockRequest{
		SKU:       "SKU-001",
		Quantity:  20,
		Reference: "ORDER-001",
	}

	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "SKU-001").Return(avail, nil)

	// Act
	_, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrInsufficientStock)
	mockAvailRepo.AssertNotCalled(t, "Update")
}

func TestReserveStockUseCase_Execute_AvailabilityNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewReserveStockUseCase(mockAvailRepo, mockEntryRepo)

	tenantID := uuid.New()
	req := &request.ReserveStockRequest{
		SKU:       "NONEXISTENT",
		Quantity:  1,
		Reference: "ORDER-001",
	}

	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "NONEXISTENT").Return(nil, exception.ErrStockAvailabilityNotFound)

	// Act
	_, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stock not found for SKU")
}

func TestReserveStockUseCase_Execute_InvalidTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewReserveStockUseCase(mockAvailRepo, mockEntryRepo)

	req := &request.ReserveStockRequest{
		SKU:       "SKU-001",
		Quantity:  10,
		Reference: "ORDER-001",
	}

	// Act
	_, err := uc.Execute(ctx, "invalid-uuid", req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tenant_id")
}

func TestReserveStockUseCase_Execute_UpdateError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewReserveStockUseCase(mockAvailRepo, mockEntryRepo)

	tenantID := uuid.New()
	availMother := mother.StockAvailabilityMother{}
	avail := availMother.WithQuantities(50, 0)
	avail.TenantID = tenantID

	req := &request.ReserveStockRequest{
		SKU:       "SKU-001",
		Quantity:  10,
		Reference: "ORDER-001",
	}

	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "SKU-001").Return(avail, nil)
	mockAvailRepo.On("Update", ctx, mock.AnythingOfType("*entity.StockAvailability")).Return(errors.New("db error"))

	// Act
	_, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error updating availability")
}
