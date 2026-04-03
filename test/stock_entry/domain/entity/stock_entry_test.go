package entity_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/src/stock_entry/domain/entity"
	"stock/test/stock_entry/domain/mother"
)

func TestNewStockEntry_Success(t *testing.T) {
	// Arrange
	tenantID := uuid.New()
	variantSKU := "SKU-001"
	entryType := entity.EntryTypePurchase
	quantity := 10.0

	// Act
	entry, err := entity.NewStockEntry(tenantID, variantSKU, entryType, quantity)

	// Assert
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, entry.ID)
	assert.Equal(t, tenantID, entry.TenantID)
	assert.Equal(t, variantSKU, entry.VariantSKU)
	assert.Equal(t, variantSKU, entry.ProductSKU)
	assert.Equal(t, entryType, entry.EntryType)
	assert.Equal(t, quantity, entry.Quantity)
	assert.Equal(t, "unit", entry.UnitOfMeasure)
	assert.Equal(t, entity.EntryStatusConfirmed, entry.Status)
	assert.True(t, entry.IsActive)
	assert.NotZero(t, entry.CreatedAt)
	assert.NotZero(t, entry.UpdatedAt)
}

func TestNewStockEntry_EmptySKU_ReturnsError(t *testing.T) {
	// Act
	entry, err := entity.NewStockEntry(uuid.New(), "", entity.EntryTypePurchase, 10)

	// Assert
	require.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "variant SKU is required")
}

func TestNewStockEntry_ZeroQuantity_ReturnsError(t *testing.T) {
	// Act
	entry, err := entity.NewStockEntry(uuid.New(), "SKU-001", entity.EntryTypePurchase, 0)

	// Assert
	require.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "quantity cannot be zero")
}

func TestNewStockEntry_InvalidEntryType_ReturnsError(t *testing.T) {
	// Act
	entry, err := entity.NewStockEntry(uuid.New(), "SKU-001", entity.EntryType("invalid"), 10)

	// Assert
	require.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "invalid entry type")
}

func TestStockEntry_SetProductInfo(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Random()
	productID := uuid.New()

	// Act
	entry.SetProductInfo(&productID, "Test Product")

	// Assert
	assert.Equal(t, &productID, entry.ProductID)
	assert.Equal(t, "Test Product", entry.ProductName)
}

func TestStockEntry_SetLocation(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Random()
	locationID := uuid.New()

	// Act
	entry.SetLocation(locationID)

	// Assert
	require.NotNil(t, entry.LocationID)
	assert.Equal(t, locationID, *entry.LocationID)
}

func TestStockEntry_SetCosts(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Random()

	// Act
	entry.SetCosts(15.50, 155.00)

	// Assert
	require.NotNil(t, entry.UnitCost)
	require.NotNil(t, entry.TotalCost)
	assert.Equal(t, 15.50, *entry.UnitCost)
	assert.Equal(t, 155.00, *entry.TotalCost)
}

func TestStockEntry_SetReference(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Random()

	// Act
	entry.SetReference("REF-001")

	// Assert
	require.NotNil(t, entry.ReferenceNumber)
	assert.Equal(t, "REF-001", *entry.ReferenceNumber)
}

func TestStockEntry_SetNotes(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Random()

	// Act
	entry.SetNotes("Nota de prueba")

	// Assert
	require.NotNil(t, entry.Notes)
	assert.Equal(t, "Nota de prueba", *entry.Notes)
}

func TestStockEntry_Confirm_FromPending_Success(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Pending()

	// Act
	err := entry.Confirm()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, entity.EntryStatusConfirmed, entry.Status)
}

func TestStockEntry_Confirm_FromCancelled_ReturnsError(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Cancelled()

	// Act
	err := entry.Confirm()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot confirm a cancelled entry")
}

func TestStockEntry_Cancel_FromPending_Success(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Pending()

	// Act
	err := entry.Cancel()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, entity.EntryStatusCancelled, entry.Status)
	assert.False(t, entry.IsActive)
}

func TestStockEntry_Cancel_FromConfirmed_ReturnsError(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Random() // Random es confirmed

	// Act
	err := entry.Cancel()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel a confirmed entry")
}

func TestStockEntry_IsPositiveMovement(t *testing.T) {
	m := mother.StockEntryMother{}

	tests := []struct {
		name     string
		entry    *entity.StockEntry
		expected bool
	}{
		{"initial_stock es positivo", m.InitialStock(), true},
		{"purchase es positivo", m.Purchase(), true},
		{"return es positivo", m.Return(), true},
		{"transfer_in es positivo", m.TransferIn(), true},
		{"sale no es positivo", m.Sale(), false},
		{"transfer_out no es positivo", m.TransferOut(), false},
		{"adjustment positivo es positivo", m.Adjustment(5), true},
		{"adjustment negativo no es positivo", m.Adjustment(-5), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.entry.IsPositiveMovement())
		})
	}
}

func TestStockEntry_IsNegativeMovement(t *testing.T) {
	m := mother.StockEntryMother{}

	tests := []struct {
		name     string
		entry    *entity.StockEntry
		expected bool
	}{
		{"sale es negativo", m.Sale(), true},
		{"transfer_out es negativo", m.TransferOut(), true},
		{"adjustment negativo es negativo", m.Adjustment(-5), true},
		{"purchase no es negativo", m.Purchase(), false},
		{"initial_stock no es negativo", m.InitialStock(), false},
		{"return no es negativo", m.Return(), false},
		{"transfer_in no es negativo", m.TransferIn(), false},
		{"adjustment positivo no es negativo", m.Adjustment(5), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.entry.IsNegativeMovement())
		})
	}
}

func TestStockEntry_CalculatedQuantity_Positive(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Purchase()

	// Act & Assert
	assert.Equal(t, entry.Quantity, entry.CalculatedQuantity())
}

func TestStockEntry_CalculatedQuantity_Negative(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Sale()

	// Act & Assert
	assert.Equal(t, -entry.Quantity, entry.CalculatedQuantity())
}

func TestStockEntry_Validate_Success(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Random()

	// Act
	err := entry.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestStockEntry_Validate_MissingTenantID(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Random()
	entry.TenantID = uuid.Nil

	// Act
	err := entry.Validate()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant ID is required")
}

func TestStockEntry_Validate_MissingProductSKU(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Random()
	entry.ProductSKU = ""

	// Act
	err := entry.Validate()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "product SKU is required")
}

func TestStockEntry_Validate_ZeroQuantity(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Random()
	entry.Quantity = 0

	// Act
	err := entry.Validate()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quantity cannot be zero")
}

func TestStockEntry_Validate_InvalidEntryType(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Random()
	entry.EntryType = entity.EntryType("bad_type")

	// Act
	err := entry.Validate()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid entry type")
}

func TestStockEntry_Validate_InvalidStatus(t *testing.T) {
	// Arrange
	m := mother.StockEntryMother{}
	entry := m.Random()
	entry.Status = entity.EntryStatus("bad_status")

	// Act
	err := entry.Validate()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}
