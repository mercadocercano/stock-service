package entity_test

import (
	"testing"
	"time"

	"stock/src/stock_location/domain/entity"
	"stock/test/stock_location/domain/mother"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewStockLocation(t *testing.T) {
	// Arrange
	tenantID := "tenant-" + uuid.New().String()
	warehouseID := "warehouse-" + uuid.New().String()
	var parentID *string
	name := "Test Stock Location"
	code := "SL-TEST"
	description := "Test stock location description"

	// Act
	stockLocation := entity.NewStockLocation(
		tenantID,
		warehouseID,
		parentID,
		name,
		code,
		description,
	)

	// Assert
	assert.NotEmpty(t, stockLocation.ID)
	assert.Equal(t, tenantID, stockLocation.TenantID)
	assert.Equal(t, warehouseID, stockLocation.WarehouseID)
	assert.Nil(t, stockLocation.ParentID)
	assert.Equal(t, name, stockLocation.Name)
	assert.Equal(t, code, stockLocation.Code)
	assert.Equal(t, stockLocation.ID, stockLocation.Path)
	assert.Equal(t, 1, stockLocation.Level)
	assert.Equal(t, description, stockLocation.Description)
	assert.True(t, stockLocation.Active)
	assert.NotZero(t, stockLocation.CreatedAt)
	assert.NotZero(t, stockLocation.UpdatedAt)
}

func TestNewStockLocation_WithParent(t *testing.T) {
	// Arrange
	tenantID := "tenant-" + uuid.New().String()
	warehouseID := "warehouse-" + uuid.New().String()
	parentID := uuid.New().String()
	parentIDPtr := &parentID
	name := "Child Stock Location"
	code := "SL-CHILD"
	description := "Child stock location description"

	// Act
	stockLocation := entity.NewStockLocation(
		tenantID,
		warehouseID,
		parentIDPtr,
		name,
		code,
		description,
	)

	// Assert
	assert.NotEmpty(t, stockLocation.ID)
	assert.Equal(t, tenantID, stockLocation.TenantID)
	assert.Equal(t, warehouseID, stockLocation.WarehouseID)
	assert.Equal(t, parentIDPtr, stockLocation.ParentID)
	assert.Equal(t, name, stockLocation.Name)
	assert.Equal(t, code, stockLocation.Code)
	assert.Contains(t, stockLocation.Path, parentID+"/")
	assert.Contains(t, stockLocation.Path, stockLocation.ID)
	assert.Equal(t, 2, stockLocation.Level)
	assert.Equal(t, description, stockLocation.Description)
	assert.True(t, stockLocation.Active)
	assert.NotZero(t, stockLocation.CreatedAt)
	assert.NotZero(t, stockLocation.UpdatedAt)
}

func TestStockLocation_Update(t *testing.T) {
	// Arrange
	stockLocationMother := mother.StockLocationMother{}
	stockLocation := stockLocationMother.Random()
	originalUpdatedAt := stockLocation.UpdatedAt

	// Esperar un momento para que el tiempo de actualización sea diferente
	time.Sleep(time.Millisecond * 10)

	newName := "Updated Stock Location"
	newCode := "SL-UPD"
	newDescription := "Updated stock location description"

	// Act
	stockLocation.Update(
		newName,
		newCode,
		newDescription,
	)

	// Assert
	assert.Equal(t, newName, stockLocation.Name)
	assert.Equal(t, newCode, stockLocation.Code)
	assert.Equal(t, newDescription, stockLocation.Description)
	assert.True(t, stockLocation.UpdatedAt.After(originalUpdatedAt))
}

func TestStockLocation_Activate(t *testing.T) {
	// Arrange
	stockLocationMother := mother.StockLocationMother{}
	stockLocation := stockLocationMother.Inactive()
	originalUpdatedAt := stockLocation.UpdatedAt

	// Esperar un momento para que el tiempo de actualización sea diferente
	time.Sleep(time.Millisecond * 10)

	// Act
	stockLocation.Activate()

	// Assert
	assert.True(t, stockLocation.Active)
	assert.True(t, stockLocation.UpdatedAt.After(originalUpdatedAt))
}

func TestStockLocation_Deactivate(t *testing.T) {
	// Arrange
	stockLocationMother := mother.StockLocationMother{}
	stockLocation := stockLocationMother.Random()
	originalUpdatedAt := stockLocation.UpdatedAt

	// Esperar un momento para que el tiempo de actualización sea diferente
	time.Sleep(time.Millisecond * 10)

	// Act
	stockLocation.Deactivate()

	// Assert
	assert.False(t, stockLocation.Active)
	assert.True(t, stockLocation.UpdatedAt.After(originalUpdatedAt))
}

func TestStockLocation_IsRoot(t *testing.T) {
	// Arrange
	stockLocationMother := mother.StockLocationMother{}

	// Act & Assert
	assert.True(t, stockLocationMother.Root().IsRoot())

	parentID := uuid.New().String()
	stockLocation := stockLocationMother.WithParentID(parentID)
	assert.False(t, stockLocation.IsRoot())
}
