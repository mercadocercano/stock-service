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
	"stock/src/stock_entry/domain/exception"
	mockRepo "stock/test/stock_entry/infrastructure/persistence/repository"
	"stock/test/stock_entry/domain/mother"
)

func TestProcessSaleUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewProcessSaleUseCase(mockEntryRepo, mockAvailRepo, nil)

	tenantID := uuid.New()
	entryMother := mother.StockEntryMother{}
	saleEntry := entryMother.Sale()
	saleEntry.TenantID = tenantID

	availMother := mother.StockAvailabilityMother{}
	avail := availMother.WithQuantities(47, 0)
	avail.TenantID = tenantID

	req := &request.ProcessSaleRequest{
		VariantSKU: "SKU-001",
		Quantity:   3,
		Reference:  "SALE-001",
	}

	mockEntryRepo.On("ProcessSaleAtomic", ctx, tenantID, "SKU-001", 3.0, "SALE-001").Return(saleEntry, nil)
	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "SKU-001").Return(avail, nil)

	// Act
	resp, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "SKU-001", resp.VariantSKU)
	assert.Equal(t, 3.0, resp.QuantitySold)
	assert.Contains(t, resp.Message, "successfully")
	mockEntryRepo.AssertExpectations(t)
	mockAvailRepo.AssertExpectations(t)
}

func TestProcessSaleUseCase_Execute_StockNotInitialized(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewProcessSaleUseCase(mockEntryRepo, mockAvailRepo, nil)

	tenantID := uuid.New()
	req := &request.ProcessSaleRequest{
		VariantSKU: "NEW-SKU",
		Quantity:   1,
	}

	mockEntryRepo.On("ProcessSaleAtomic", ctx, tenantID, "NEW-SKU", 1.0, mock.AnythingOfType("string")).Return(nil, exception.ErrStockNotInitialized)

	// Act
	resp, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.NoError(t, err) // No retorna error, retorna response con success=false
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Message, "Stock not initialized")
	assert.Equal(t, "NEW-SKU", resp.VariantSKU)
}

func TestProcessSaleUseCase_Execute_InsufficientStock(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewProcessSaleUseCase(mockEntryRepo, mockAvailRepo, nil)

	tenantID := uuid.New()
	req := &request.ProcessSaleRequest{
		VariantSKU: "SKU-001",
		Quantity:   100,
		Reference:  "SALE-BIG",
	}

	mockEntryRepo.On("ProcessSaleAtomic", ctx, tenantID, "SKU-001", 100.0, "SALE-BIG").Return(nil, exception.ErrInsufficientStock)

	// Act
	resp, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.NoError(t, err) // No retorna error, retorna response con success=false
	assert.False(t, resp.Success)
	assert.Equal(t, "SKU-001", resp.VariantSKU)
}

func TestProcessSaleUseCase_Execute_TechnicalError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewProcessSaleUseCase(mockEntryRepo, mockAvailRepo, nil)

	tenantID := uuid.New()
	req := &request.ProcessSaleRequest{
		VariantSKU: "SKU-001",
		Quantity:   5,
		Reference:  "SALE-001",
	}

	mockEntryRepo.On("ProcessSaleAtomic", ctx, tenantID, "SKU-001", 5.0, "SALE-001").Return(nil, errors.New("database connection lost"))

	// Act
	_, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to process sale atomically")
}

func TestProcessSaleUseCase_Execute_ValidationError_MissingSKU(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewProcessSaleUseCase(mockEntryRepo, mockAvailRepo, nil)

	req := &request.ProcessSaleRequest{
		VariantSKU: "",
		Quantity:   5,
	}

	// Act
	_, err := uc.Execute(ctx, uuid.New().String(), req)

	// Assert
	require.Error(t, err)
	mockEntryRepo.AssertNotCalled(t, "ProcessSaleAtomic")
}

func TestProcessSaleUseCase_Execute_ValidationError_ZeroQuantity(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewProcessSaleUseCase(mockEntryRepo, mockAvailRepo, nil)

	req := &request.ProcessSaleRequest{
		VariantSKU: "SKU-001",
		Quantity:   0,
	}

	// Act
	_, err := uc.Execute(ctx, uuid.New().String(), req)

	// Assert
	require.Error(t, err)
	mockEntryRepo.AssertNotCalled(t, "ProcessSaleAtomic")
}

func TestProcessSaleUseCase_Execute_InvalidTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewProcessSaleUseCase(mockEntryRepo, mockAvailRepo, nil)

	req := &request.ProcessSaleRequest{
		VariantSKU: "SKU-001",
		Quantity:   5,
	}

	// Act
	_, err := uc.Execute(ctx, "bad-uuid", req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tenant_id")
}

func TestProcessSaleUseCase_Execute_GeneratesReferenceWhenEmpty(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewProcessSaleUseCase(mockEntryRepo, mockAvailRepo, nil)

	tenantID := uuid.New()
	entryMother := mother.StockEntryMother{}
	saleEntry := entryMother.Sale()
	saleEntry.TenantID = tenantID
	saleEntry.EntryType = entity.EntryTypeSale

	availMother := mother.StockAvailabilityMother{}
	avail := availMother.WithQuantities(50, 0)
	avail.TenantID = tenantID

	req := &request.ProcessSaleRequest{
		VariantSKU: "SKU-001",
		Quantity:   2,
		Reference:  "", // Sin referencia
	}

	// El use case generara un reference automatico tipo "SALE-xxxx"
	mockEntryRepo.On("ProcessSaleAtomic", ctx, tenantID, "SKU-001", 2.0, mock.MatchedBy(func(ref string) bool {
		return len(ref) > 0 // Solo verificar que se genero una referencia
	})).Return(saleEntry, nil)
	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "SKU-001").Return(avail, nil)

	// Act
	resp, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.NoError(t, err)
	assert.True(t, resp.Success)
	mockEntryRepo.AssertExpectations(t)
}

func TestProcessSaleUseCase_Execute_AvailabilityReadFail_StillSuccess(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockEntryRepo := new(mockRepo.MockStockEntryRepository)
	mockAvailRepo := new(mockRepo.MockStockAvailabilityRepository)

	uc := usecase.NewProcessSaleUseCase(mockEntryRepo, mockAvailRepo, nil)

	tenantID := uuid.New()
	entryMother := mother.StockEntryMother{}
	saleEntry := entryMother.Sale()
	saleEntry.TenantID = tenantID

	req := &request.ProcessSaleRequest{
		VariantSKU: "SKU-001",
		Quantity:   5,
		Reference:  "SALE-001",
	}

	mockEntryRepo.On("ProcessSaleAtomic", ctx, tenantID, "SKU-001", 5.0, "SALE-001").Return(saleEntry, nil)
	mockAvailRepo.On("FindByTenantAndSKU", ctx, tenantID, "SKU-001").Return(nil, errors.New("availability read error"))

	// Act
	resp, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Contains(t, resp.Message, "availability read failed")
}
