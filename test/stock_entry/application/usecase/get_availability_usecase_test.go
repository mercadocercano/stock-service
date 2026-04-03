package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/src/stock_entry/application/usecase"
	mockRepo "stock/test/stock_entry/infrastructure/persistence/repository"
	"stock/test/stock_entry/domain/mother"
)

func TestGetAvailabilityUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewGetAvailabilityUseCase(mockAvailRepo)

	tenantID := uuid.New()
	availMother := mother.StockAvailabilityMother{}
	avail := availMother.WithQuantities(100, 10)
	avail.TenantID = tenantID
	avail.ProductSKU = "SKU-001"

	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "SKU-001").Return(avail, nil)

	// Act
	resp, err := uc.Execute(ctx, tenantID.String(), "SKU-001")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "SKU-001", resp.ProductSKU)
	assert.Equal(t, 90.0, resp.AvailableQuantity)
	assert.Equal(t, 10.0, resp.ReservedQuantity)
	assert.Equal(t, 100.0, resp.TotalQuantity)
	mockAvailRepo.AssertExpectations(t)
}

func TestGetAvailabilityUseCase_Execute_InvalidTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewGetAvailabilityUseCase(mockAvailRepo)

	// Act
	_, err := uc.Execute(ctx, "bad-uuid", "SKU-001")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tenant_id")
}

func TestGetAvailabilityUseCase_Execute_EmptySKU(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewGetAvailabilityUseCase(mockAvailRepo)

	// Act
	_, err := uc.Execute(ctx, uuid.New().String(), "")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "product_sku is required")
}

func TestGetAvailabilityUseCase_Execute_NotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewGetAvailabilityUseCase(mockAvailRepo)

	tenantID := uuid.New()
	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "NONEXISTENT").Return(nil, errors.New("not found"))

	// Act
	_, err := uc.Execute(ctx, tenantID.String(), "NONEXISTENT")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error finding availability")
}

func TestGetAvailabilityUseCase_ExecuteMultiple_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewGetAvailabilityUseCase(mockAvailRepo)

	tenantID := uuid.New()
	availMother := mother.StockAvailabilityMother{}

	avail1 := availMother.WithQuantities(50, 0)
	avail1.TenantID = tenantID
	avail1.ProductSKU = "SKU-001"

	avail2 := availMother.WithQuantities(30, 5)
	avail2.TenantID = tenantID
	avail2.ProductSKU = "SKU-002"

	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "SKU-001").Return(avail1, nil)
	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "SKU-002").Return(avail2, nil)
	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "SKU-MISSING").Return(nil, errors.New("not found"))

	// Act
	results, err := uc.ExecuteMultiple(ctx, tenantID.String(), []string{"SKU-001", "SKU-002", "SKU-MISSING"})

	// Assert
	require.NoError(t, err)
	assert.Len(t, results, 2) // SKU-MISSING se salta silenciosamente
	mockAvailRepo.AssertExpectations(t)
}

func TestGetAvailabilityUseCase_ExecuteMultiple_InvalidTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewGetAvailabilityUseCase(mockAvailRepo)

	// Act
	_, err := uc.ExecuteMultiple(ctx, "bad-uuid", []string{"SKU-001"})

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tenant_id")
}
