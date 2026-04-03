package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/src/stock_entry/application/usecase"
	"stock/src/stock_entry/domain/entity"
	mockRepo "stock/test/stock_entry/infrastructure/persistence/repository"
)

func TestListSalesUseCase_Execute_Success_FiltersSalesOnly(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewListSalesUseCase(mockEntryRepo)

	tenantID := uuid.New()
	now := time.Now()

	entries := []*entity.StockEntry{
		{
			ID:         uuid.New(),
			TenantID:   tenantID,
			VariantSKU: "SKU-001",
			ProductSKU: "SKU-001",
			EntryType:  entity.EntryTypeSale,
			Quantity:   5,
			Status:     entity.EntryStatusConfirmed,
			IsActive:   true,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         uuid.New(),
			TenantID:   tenantID,
			VariantSKU: "SKU-002",
			ProductSKU: "SKU-002",
			EntryType:  entity.EntryTypePurchase, // No es venta
			Quantity:   20,
			Status:     entity.EntryStatusConfirmed,
			IsActive:   true,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         uuid.New(),
			TenantID:   tenantID,
			VariantSKU: "SKU-003",
			ProductSKU: "SKU-003",
			EntryType:  entity.EntryTypeSale,
			Quantity:   3,
			Status:     entity.EntryStatusConfirmed,
			IsActive:   true,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	mockEntryRepo.On("FindByTenant", ctx, tenantID, 50, 0).Return(entries, nil)

	// Act
	sales, err := uc.Execute(ctx, tenantID.String(), 50, 0)

	// Assert
	require.NoError(t, err)
	assert.Len(t, sales, 2) // Solo las 2 ventas, no la compra
	assert.Equal(t, string(entity.EntryTypeSale), sales[0].EntryType)
	assert.Equal(t, string(entity.EntryTypeSale), sales[1].EntryType)
	mockEntryRepo.AssertExpectations(t)
}

func TestListSalesUseCase_Execute_NoSales(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewListSalesUseCase(mockEntryRepo)

	tenantID := uuid.New()
	now := time.Now()

	entries := []*entity.StockEntry{
		{
			ID:         uuid.New(),
			TenantID:   tenantID,
			VariantSKU: "SKU-001",
			ProductSKU: "SKU-001",
			EntryType:  entity.EntryTypePurchase,
			Quantity:   20,
			Status:     entity.EntryStatusConfirmed,
			IsActive:   true,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	mockEntryRepo.On("FindByTenant", ctx, tenantID, 50, 0).Return(entries, nil)

	// Act
	sales, err := uc.Execute(ctx, tenantID.String(), 50, 0)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, sales)
	mockEntryRepo.AssertExpectations(t)
}

func TestListSalesUseCase_Execute_EmptyEntries(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewListSalesUseCase(mockEntryRepo)

	tenantID := uuid.New()

	mockEntryRepo.On("FindByTenant", ctx, tenantID, 50, 0).Return([]*entity.StockEntry{}, nil)

	// Act
	sales, err := uc.Execute(ctx, tenantID.String(), 50, 0)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, sales)
}

func TestListSalesUseCase_Execute_InvalidTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewListSalesUseCase(mockEntryRepo)

	// Act
	_, err := uc.Execute(ctx, "bad-uuid", 50, 0)

	// Assert
	require.Error(t, err)
}

func TestListSalesUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewListSalesUseCase(mockEntryRepo)

	tenantID := uuid.New()
	mockEntryRepo.On("FindByTenant", ctx, tenantID, 50, 0).Return(nil, errors.New("db error"))

	// Act
	_, err := uc.Execute(ctx, tenantID.String(), 50, 0)

	// Assert
	require.Error(t, err)
	mockEntryRepo.AssertExpectations(t)
}
