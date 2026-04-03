package entity_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/src/stock_entry/domain/entity"
	"stock/src/stock_entry/domain/exception"
	"stock/test/stock_entry/domain/mother"
)

func TestNewStockAvailability(t *testing.T) {
	// Arrange
	tenantID := uuid.New()
	variantSKU := "SKU-001"
	totalQuantity := 50.0

	// Act
	avail := entity.NewStockAvailability(tenantID, variantSKU, totalQuantity)

	// Assert
	assert.NotEqual(t, uuid.Nil, avail.ID)
	assert.Equal(t, tenantID, avail.TenantID)
	assert.Equal(t, variantSKU, avail.VariantSKU)
	assert.Equal(t, variantSKU, avail.ProductSKU)
	assert.Equal(t, totalQuantity, avail.AvailableQuantity)
	assert.Equal(t, 0.0, avail.ReservedQuantity)
	assert.Equal(t, totalQuantity, avail.TotalQuantity)
	assert.Equal(t, "unit", avail.UnitOfMeasure)
	assert.False(t, avail.IsOutOfStock)
	assert.NotZero(t, avail.UpdatedAt)
}

func TestNewStockAvailability_ZeroQuantity_IsOutOfStock(t *testing.T) {
	// Act
	avail := entity.NewStockAvailability(uuid.New(), "SKU-001", 0)

	// Assert
	assert.True(t, avail.IsOutOfStock)
	assert.Equal(t, 0.0, avail.AvailableQuantity)
}

func TestNewStockAvailability_LowQuantity_IsLowStock(t *testing.T) {
	// Act
	avail := entity.NewStockAvailability(uuid.New(), "SKU-001", 5)

	// Assert
	// Default threshold is 10 in NewStockAvailability
	assert.True(t, avail.IsLowStock)
	assert.False(t, avail.IsOutOfStock)
}

func TestStockAvailability_UpdateQuantity(t *testing.T) {
	// Arrange
	m := mother.StockAvailabilityMother{}
	avail := m.Random()

	// Act
	avail.UpdateQuantity(100, 20)

	// Assert
	assert.Equal(t, 100.0, avail.TotalQuantity)
	assert.Equal(t, 20.0, avail.ReservedQuantity)
	assert.Equal(t, 80.0, avail.AvailableQuantity)
	assert.False(t, avail.IsOutOfStock)
}

func TestStockAvailability_UpdateQuantity_OutOfStock(t *testing.T) {
	// Arrange
	m := mother.StockAvailabilityMother{}
	avail := m.Random()

	// Act
	avail.UpdateQuantity(0, 0)

	// Assert
	assert.True(t, avail.IsOutOfStock)
	assert.Equal(t, 0.0, avail.AvailableQuantity)
}

func TestStockAvailability_Reserve_Success(t *testing.T) {
	// Arrange
	m := mother.StockAvailabilityMother{}
	avail := m.WithQuantities(50, 0)

	// Act
	err := avail.Reserve(10)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 10.0, avail.ReservedQuantity)
	assert.Equal(t, 40.0, avail.AvailableQuantity)
	assert.Equal(t, 50.0, avail.TotalQuantity)
}

func TestStockAvailability_Reserve_InsufficientStock(t *testing.T) {
	// Arrange
	m := mother.StockAvailabilityMother{}
	avail := m.WithQuantities(10, 0)

	// Act
	err := avail.Reserve(20)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrInsufficientStock)
	// Cantidades no deben cambiar
	assert.Equal(t, 0.0, avail.ReservedQuantity)
	assert.Equal(t, 10.0, avail.AvailableQuantity)
}

func TestStockAvailability_Reserve_MultipleReservations(t *testing.T) {
	// Arrange
	m := mother.StockAvailabilityMother{}
	avail := m.WithQuantities(100, 0)

	// Act
	err1 := avail.Reserve(30)
	err2 := avail.Reserve(40)

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Equal(t, 70.0, avail.ReservedQuantity)
	assert.Equal(t, 30.0, avail.AvailableQuantity)
}

func TestStockAvailability_Release(t *testing.T) {
	// Arrange
	m := mother.StockAvailabilityMother{}
	avail := m.WithReservation(50, 20)

	// Act
	avail.Release(10)

	// Assert
	assert.Equal(t, 10.0, avail.ReservedQuantity)
	assert.Equal(t, 40.0, avail.AvailableQuantity)
	assert.Equal(t, 50.0, avail.TotalQuantity)
}

func TestStockAvailability_Release_CannotGoBelowZero(t *testing.T) {
	// Arrange
	m := mother.StockAvailabilityMother{}
	avail := m.WithReservation(50, 5)

	// Act
	avail.Release(20) // Liberar mas de lo reservado

	// Assert
	assert.Equal(t, 0.0, avail.ReservedQuantity)
	assert.Equal(t, 50.0, avail.AvailableQuantity)
}

func TestStockAvailability_SetStockLevels(t *testing.T) {
	// Arrange
	m := mother.StockAvailabilityMother{}
	avail := m.WithQuantities(50, 0)

	// Act
	avail.SetStockLevels(10, 200)

	// Assert
	assert.Equal(t, 10.0, avail.MinStockLevel)
	require.NotNil(t, avail.MaxStockLevel)
	assert.Equal(t, 200.0, *avail.MaxStockLevel)
	assert.False(t, avail.IsLowStock)
}

func TestStockAvailability_SetStockLevels_LowStock(t *testing.T) {
	// Arrange
	m := mother.StockAvailabilityMother{}
	avail := m.WithQuantities(3, 0)

	// Act
	avail.SetStockLevels(10, 200)

	// Assert
	assert.True(t, avail.IsLowStock)
}

func TestStockAvailability_UpdateValue(t *testing.T) {
	// Arrange
	m := mother.StockAvailabilityMother{}
	avail := m.WithQuantities(100, 0)

	// Act
	avail.UpdateValue(25.50)

	// Assert
	require.NotNil(t, avail.AvgUnitCost)
	require.NotNil(t, avail.TotalValue)
	assert.Equal(t, 25.50, *avail.AvgUnitCost)
	assert.Equal(t, 2550.0, *avail.TotalValue)
}
