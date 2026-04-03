package usecase_test

import (
	"context"
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

func TestConsumeStockUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewConsumeStockUseCase(mockAvailRepo, mockEntryRepo)

	tenantID := uuid.New()
	availMother := mother.StockAvailabilityMother{}
	avail := availMother.WithReservation(50, 20)
	avail.TenantID = tenantID

	req := &request.ConsumeStockRequest{
		SKU:       "SKU-001",
		Quantity:  10,
		Reference: "ORDER-CONFIRMED-001",
	}

	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "SKU-001").Return(avail, nil)
	mockAvailRepo.On("Update", ctx, mock.AnythingOfType("*entity.StockAvailability")).Return(nil)
	mockEntryRepo.On("Save", ctx, mock.AnythingOfType("*entity.StockEntry")).Return(nil)

	// Act
	resp, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "SKU-001", resp.SKU)
	assert.Equal(t, 10, resp.ConsumedQty)
	assert.Equal(t, "ORDER-CONFIRMED-001", resp.Reference)
	mockAvailRepo.AssertExpectations(t)
	mockEntryRepo.AssertExpectations(t)
}

func TestConsumeStockUseCase_Execute_InsufficientReservedStock(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewConsumeStockUseCase(mockAvailRepo, mockEntryRepo)

	tenantID := uuid.New()
	availMother := mother.StockAvailabilityMother{}
	avail := availMother.WithReservation(50, 5)
	avail.TenantID = tenantID

	req := &request.ConsumeStockRequest{
		SKU:       "SKU-001",
		Quantity:  20,
		Reference: "ORDER-001",
	}

	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "SKU-001").Return(avail, nil)

	// Act
	_, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient reserved stock")
	mockAvailRepo.AssertNotCalled(t, "Update")
	mockEntryRepo.AssertNotCalled(t, "Save")
}

func TestConsumeStockUseCase_Execute_AvailabilityNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewConsumeStockUseCase(mockAvailRepo, mockEntryRepo)

	tenantID := uuid.New()
	req := &request.ConsumeStockRequest{
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

func TestConsumeStockUseCase_Execute_InvalidTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewConsumeStockUseCase(mockAvailRepo, mockEntryRepo)

	req := &request.ConsumeStockRequest{
		SKU:       "SKU-001",
		Quantity:  10,
		Reference: "ORDER-001",
	}

	// Act
	_, err := uc.Execute(ctx, "bad-uuid", req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tenant_id")
}
