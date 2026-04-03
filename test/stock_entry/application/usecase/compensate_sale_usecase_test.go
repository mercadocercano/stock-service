package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/src/stock_entry/application/request"
	"stock/src/stock_entry/application/usecase"
	mockRepo "stock/test/stock_entry/infrastructure/persistence/repository"
)

func TestCompensateSaleUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewCompensateSaleUseCase(mockStockEntryRepo)

	tenantID := uuid.New()
	stockEntryID := uuid.New()

	req := &request.CompensateSaleRequest{
		StockEntryID: stockEntryID.String(),
		Reason:       "order_creation_failed",
	}

	mockStockEntryRepo.On("CompensateSale", ctx, tenantID, stockEntryID, "order_creation_failed").Return(nil)

	// Act
	resp, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, stockEntryID.String(), resp.StockEntryID)
	assert.Equal(t, "order_creation_failed", resp.Reason)
	assert.Contains(t, resp.Message, "compensated successfully")
	mockStockEntryRepo.AssertExpectations(t)
}

func TestCompensateSaleUseCase_Execute_InvalidRequest_MissingEntryID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewCompensateSaleUseCase(mockStockEntryRepo)

	req := &request.CompensateSaleRequest{
		StockEntryID: "",
		Reason:       "reason",
	}

	// Act
	_, err := uc.Execute(ctx, uuid.New().String(), req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stock_entry_id is required")
	mockStockEntryRepo.AssertNotCalled(t, "CompensateSale")
}

func TestCompensateSaleUseCase_Execute_InvalidRequest_MissingReason(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewCompensateSaleUseCase(mockStockEntryRepo)

	req := &request.CompensateSaleRequest{
		StockEntryID: uuid.New().String(),
		Reason:       "",
	}

	// Act
	_, err := uc.Execute(ctx, uuid.New().String(), req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reason is required")
	mockStockEntryRepo.AssertNotCalled(t, "CompensateSale")
}

func TestCompensateSaleUseCase_Execute_InvalidTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewCompensateSaleUseCase(mockStockEntryRepo)

	req := &request.CompensateSaleRequest{
		StockEntryID: uuid.New().String(),
		Reason:       "test",
	}

	// Act
	_, err := uc.Execute(ctx, "invalid-uuid", req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tenant_id")
}

func TestCompensateSaleUseCase_Execute_InvalidStockEntryID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewCompensateSaleUseCase(mockStockEntryRepo)

	req := &request.CompensateSaleRequest{
		StockEntryID: "bad-uuid",
		Reason:       "test",
	}

	// Act
	_, err := uc.Execute(ctx, uuid.New().String(), req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid stock_entry_id")
}

func TestCompensateSaleUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockStockEntryRepo := new(mockRepo.MockStockEntryRepository)

	uc := usecase.NewCompensateSaleUseCase(mockStockEntryRepo)

	tenantID := uuid.New()
	stockEntryID := uuid.New()

	req := &request.CompensateSaleRequest{
		StockEntryID: stockEntryID.String(),
		Reason:       "test_reason",
	}

	mockStockEntryRepo.On("CompensateSale", ctx, tenantID, stockEntryID, "test_reason").Return(errors.New("can only compensate sale entries"))

	// Act
	_, err := uc.Execute(ctx, tenantID.String(), req)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to compensate sale")
	mockStockEntryRepo.AssertExpectations(t)
}
