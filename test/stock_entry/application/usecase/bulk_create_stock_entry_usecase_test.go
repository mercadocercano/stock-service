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
	"stock/src/stock_entry/domain/entity"
	mockRepo "stock/test/stock_entry/infrastructure/persistence/repository"
)

func TestBulkCreateStockEntryUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewBulkCreateStockEntryUseCase(mockStockEntryRepo)

	tenantID := uuid.New()
	req := request.BulkCreateStockEntriesRequest{
		TenantID: tenantID.String(),
		Entries: []request.CreateStockEntryRequest{
			{ProductSKU: "SKU-001", EntryType: string(entity.EntryTypePurchase), Quantity: 10},
			{ProductSKU: "SKU-002", EntryType: string(entity.EntryTypePurchase), Quantity: 20},
		},
	}

	mockStockEntryRepo.On("SaveBulk", ctx, mock.AnythingOfType("[]*entity.StockEntry")).Return(nil)

	// Act
	resp, err := uc.Execute(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, 2, resp.TotalEntries)
	assert.Equal(t, 2, resp.EntriesCreated)
	assert.Equal(t, 0, resp.EntriesFailed)
	assert.Empty(t, resp.Errors)
	assert.Len(t, resp.CreatedEntries, 2)
	mockStockEntryRepo.AssertExpectations(t)
}

func TestBulkCreateStockEntryUseCase_Execute_PartialFailure(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewBulkCreateStockEntryUseCase(mockStockEntryRepo)

	tenantID := uuid.New()
	req := request.BulkCreateStockEntriesRequest{
		TenantID: tenantID.String(),
		Entries: []request.CreateStockEntryRequest{
			{ProductSKU: "SKU-001", EntryType: string(entity.EntryTypePurchase), Quantity: 10},
			{ProductSKU: "", EntryType: string(entity.EntryTypePurchase), Quantity: 5}, // Sin SKU
			{ProductSKU: "SKU-003", EntryType: string(entity.EntryTypePurchase), Quantity: 15},
		},
	}

	mockStockEntryRepo.On("SaveBulk", ctx, mock.AnythingOfType("[]*entity.StockEntry")).Return(nil)

	// Act
	resp, err := uc.Execute(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, 3, resp.TotalEntries)
	assert.Equal(t, 2, resp.EntriesCreated)
	assert.Equal(t, 1, resp.EntriesFailed)
	assert.Len(t, resp.Errors, 1)
	mockStockEntryRepo.AssertExpectations(t)
}

func TestBulkCreateStockEntryUseCase_Execute_AllFail(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewBulkCreateStockEntryUseCase(mockStockEntryRepo)

	tenantID := uuid.New()
	req := request.BulkCreateStockEntriesRequest{
		TenantID: tenantID.String(),
		Entries: []request.CreateStockEntryRequest{
			{ProductSKU: "", EntryType: string(entity.EntryTypePurchase), Quantity: 10},
			{ProductSKU: "SKU-001", EntryType: "", Quantity: 5},
		},
	}

	// Act
	resp, err := uc.Execute(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, 0, resp.EntriesCreated)
	assert.Equal(t, 2, resp.EntriesFailed)
	mockStockEntryRepo.AssertNotCalled(t, "SaveBulk")
}

func TestBulkCreateStockEntryUseCase_Execute_InvalidTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewBulkCreateStockEntryUseCase(mockStockEntryRepo)

	req := request.BulkCreateStockEntriesRequest{
		TenantID: "invalid-uuid",
		Entries: []request.CreateStockEntryRequest{
			{ProductSKU: "SKU-001", EntryType: string(entity.EntryTypePurchase), Quantity: 10},
		},
	}

	// Act
	resp, err := uc.Execute(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid tenant_id")
}

func TestBulkCreateStockEntryUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewBulkCreateStockEntryUseCase(mockStockEntryRepo)

	tenantID := uuid.New()
	req := request.BulkCreateStockEntriesRequest{
		TenantID: tenantID.String(),
		Entries: []request.CreateStockEntryRequest{
			{ProductSKU: "SKU-001", EntryType: string(entity.EntryTypePurchase), Quantity: 10},
		},
	}

	mockStockEntryRepo.On("SaveBulk", ctx, mock.AnythingOfType("[]*entity.StockEntry")).Return(errors.New("db error"))

	// Act
	resp, err := uc.Execute(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "error saving stock entries")
	mockStockEntryRepo.AssertExpectations(t)
}
