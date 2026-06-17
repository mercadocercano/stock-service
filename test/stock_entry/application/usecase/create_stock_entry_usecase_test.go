package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"

	"stock/src/stock_entry/application/request"
	"stock/src/stock_entry/application/usecase"
	"stock/src/stock_entry/domain/entity"
	mockRepo "stock/test/stock_entry/infrastructure/persistence/repository"
)

func TestCreateStockEntryUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewCreateStockEntryUseCase(mockStockEntryRepo, nil)

	tenantID := uuid.New()
	req := request.CreateStockEntryRequest{
		TenantID:   tenantID.String(),
		ProductSKU: "SKU-001",
		EntryType:  string(entity.EntryTypePurchase),
		Quantity:   10,
	}

	mockStockEntryRepo.On("Save", ctx, mock.AnythingOfType("*entity.StockEntry")).Return(nil)

	// Act
	resp, err := uc.Execute(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "SKU-001", resp.ProductSKU)
	assert.Equal(t, 10.0, resp.Quantity)
	assert.Equal(t, string(entity.EntryTypePurchase), resp.EntryType)
	mockStockEntryRepo.AssertExpectations(t)
}

func TestCreateStockEntryUseCase_Execute_ValidationError_MissingSKU(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewCreateStockEntryUseCase(mockStockEntryRepo, nil)

	req := request.CreateStockEntryRequest{
		TenantID:  uuid.New().String(),
		EntryType: string(entity.EntryTypePurchase),
		Quantity:  10,
	}

	// Act
	resp, err := uc.Execute(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, resp)
	mockStockEntryRepo.AssertNotCalled(t, "Save")
}

func TestCreateStockEntryUseCase_Execute_ValidationError_ZeroQuantity(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewCreateStockEntryUseCase(mockStockEntryRepo, nil)

	req := request.CreateStockEntryRequest{
		TenantID:   uuid.New().String(),
		ProductSKU: "SKU-001",
		EntryType:  string(entity.EntryTypePurchase),
		Quantity:   0,
	}

	// Act
	resp, err := uc.Execute(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, resp)
	mockStockEntryRepo.AssertNotCalled(t, "Save")
}

func TestCreateStockEntryUseCase_Execute_InvalidTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewCreateStockEntryUseCase(mockStockEntryRepo, nil)

	req := request.CreateStockEntryRequest{
		TenantID:   "invalid-uuid",
		ProductSKU: "SKU-001",
		EntryType:  string(entity.EntryTypePurchase),
		Quantity:   10,
	}

	// Act
	resp, err := uc.Execute(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid tenant_id")
	mockStockEntryRepo.AssertNotCalled(t, "Save")
}

func TestCreateStockEntryUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewCreateStockEntryUseCase(mockStockEntryRepo, nil)

	req := request.CreateStockEntryRequest{
		TenantID:   uuid.New().String(),
		ProductSKU: "SKU-001",
		EntryType:  string(entity.EntryTypePurchase),
		Quantity:   10,
	}

	mockStockEntryRepo.On("Save", ctx, mock.AnythingOfType("*entity.StockEntry")).Return(errors.New("db connection error"))

	// Act
	resp, err := uc.Execute(ctx, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "error saving stock entry")
	mockStockEntryRepo.AssertExpectations(t)
}

func TestCreateStockEntryUseCase_Execute_WithOptionalFields(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewCreateStockEntryUseCase(mockStockEntryRepo, nil)

	tenantID := uuid.New()
	locationID := uuid.New()
	req := request.CreateStockEntryRequest{
		TenantID:        tenantID.String(),
		ProductSKU:      "SKU-001",
		ProductName:     "Producto Test",
		LocationID:      locationID.String(),
		EntryType:       string(entity.EntryTypePurchase),
		Quantity:        10,
		UnitOfMeasure:   "kg",
		UnitCost:        25.50,
		ReferenceNumber: "PO-001",
		Notes:           "Compra de prueba",
	}

	mockStockEntryRepo.On("Save", ctx, mock.AnythingOfType("*entity.StockEntry")).Return(nil)

	// Act
	resp, err := uc.Execute(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "SKU-001", resp.ProductSKU)
	mockStockEntryRepo.AssertExpectations(t)
}
