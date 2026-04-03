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

func TestRevertConsumeUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewRevertConsumeUseCase(mockAvailRepo, mockEntryRepo)

	tenantID := uuid.New()
	availMother := mother.StockAvailabilityMother{}
	avail := availMother.WithQuantities(40, 0)
	avail.TenantID = tenantID

	req := &request.RevertConsumeRequest{
		SKU:       "SKU-001",
		Quantity:  5,
		Reference: "ORDER-CANCEL-001",
	}

	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "SKU-001").Return(avail, nil)
	mockAvailRepo.On("Update", ctx, mock.AnythingOfType("*entity.StockAvailability")).Return(nil)
	mockEntryRepo.On("Save", ctx, mock.AnythingOfType("*entity.StockEntry")).Return(nil)

	// Act
	resp, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "SKU-001", resp.SKU)
	assert.Equal(t, 5, resp.RevertedQty)
	assert.Equal(t, 45, resp.AvailableQty)
	assert.Equal(t, "ORDER-CANCEL-001", resp.Reference)
	mockAvailRepo.AssertExpectations(t)
	mockEntryRepo.AssertExpectations(t)
}

func TestRevertConsumeUseCase_Execute_AvailabilityNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewRevertConsumeUseCase(mockAvailRepo, mockEntryRepo)

	tenantID := uuid.New()
	req := &request.RevertConsumeRequest{
		SKU:       "NONEXISTENT",
		Quantity:  1,
		Reference: "REF-001",
	}

	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "NONEXISTENT").Return(nil, exception.ErrStockAvailabilityNotFound)

	// Act
	_, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stock not found for SKU")
}

func TestRevertConsumeUseCase_Execute_InvalidTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewRevertConsumeUseCase(mockAvailRepo, mockEntryRepo)

	req := &request.RevertConsumeRequest{
		SKU:       "SKU-001",
		Quantity:  5,
		Reference: "REF-001",
	}

	// Act
	_, err := uc.Execute(ctx, "invalid-uuid", req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tenant_id")
}

func TestRevertConsumeUseCase_Execute_UpdateError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewRevertConsumeUseCase(mockAvailRepo, mockEntryRepo)

	tenantID := uuid.New()
	availMother := mother.StockAvailabilityMother{}
	avail := availMother.WithQuantities(40, 0)
	avail.TenantID = tenantID

	req := &request.RevertConsumeRequest{
		SKU:       "SKU-001",
		Quantity:  5,
		Reference: "REF-001",
	}

	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "SKU-001").Return(avail, nil)
	mockAvailRepo.On("Update", ctx, mock.AnythingOfType("*entity.StockAvailability")).Return(errors.New("db error"))

	// Act
	_, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error updating availability")
}

func TestRevertConsumeUseCase_Execute_SaveEntryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewRevertConsumeUseCase(mockAvailRepo, mockEntryRepo)

	tenantID := uuid.New()
	availMother := mother.StockAvailabilityMother{}
	avail := availMother.WithQuantities(40, 0)
	avail.TenantID = tenantID

	req := &request.RevertConsumeRequest{
		SKU:       "SKU-001",
		Quantity:  5,
		Reference: "REF-001",
	}

	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "SKU-001").Return(avail, nil)
	mockAvailRepo.On("Update", ctx, mock.AnythingOfType("*entity.StockAvailability")).Return(nil)
	mockEntryRepo.On("Save", ctx, mock.AnythingOfType("*entity.StockEntry")).Return(errors.New("db error"))

	// Act
	_, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error creating stock entry")
}
