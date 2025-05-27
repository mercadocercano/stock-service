package entity_test

import (
	"testing"
	"time"

	"stock/src/warehouse/domain/entity"
	"stock/test/warehouse/domain/mother"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewWarehouse(t *testing.T) {
	// Arrange
	tenantID := "tenant-" + uuid.New().String()
	locationID := "location-" + uuid.New().String()
	name := "Test Warehouse"
	code := "WH-TEST"
	warehouseType := entity.RegularWarehouseType
	description := "Test warehouse description"
	priority := 1

	// Act
	warehouse := entity.NewWarehouse(
		tenantID,
		locationID,
		name,
		code,
		warehouseType,
		description,
		priority,
	)

	// Assert
	assert.NotEmpty(t, warehouse.ID)
	assert.Equal(t, tenantID, warehouse.TenantID)
	assert.Equal(t, locationID, warehouse.LocationID)
	assert.Equal(t, name, warehouse.Name)
	assert.Equal(t, code, warehouse.Code)
	assert.Equal(t, warehouseType, warehouse.Type)
	assert.Equal(t, description, warehouse.Description)
	assert.Equal(t, priority, warehouse.Priority)
	assert.True(t, warehouse.Active)
	assert.NotZero(t, warehouse.CreatedAt)
	assert.NotZero(t, warehouse.UpdatedAt)
}

func TestWarehouse_Update(t *testing.T) {
	// Arrange
	warehouseMother := mother.WarehouseMother{}
	warehouse := warehouseMother.Random()
	originalUpdatedAt := warehouse.UpdatedAt

	// Esperar un momento para que el tiempo de actualización sea diferente
	time.Sleep(time.Millisecond * 10)

	newName := "Updated Warehouse"
	newCode := "WH-UPD"
	newType := entity.SpecialWarehouseType
	newDescription := "Updated warehouse description"
	newPriority := 2

	// Act
	warehouse.Update(
		newName,
		newCode,
		newType,
		newDescription,
		newPriority,
	)

	// Assert
	assert.Equal(t, newName, warehouse.Name)
	assert.Equal(t, newCode, warehouse.Code)
	assert.Equal(t, newType, warehouse.Type)
	assert.Equal(t, newDescription, warehouse.Description)
	assert.Equal(t, newPriority, warehouse.Priority)
	assert.True(t, warehouse.UpdatedAt.After(originalUpdatedAt))
}

func TestWarehouse_Activate(t *testing.T) {
	// Arrange
	warehouseMother := mother.WarehouseMother{}
	warehouse := warehouseMother.Inactive()
	originalUpdatedAt := warehouse.UpdatedAt

	// Esperar un momento para que el tiempo de actualización sea diferente
	time.Sleep(time.Millisecond * 10)

	// Act
	warehouse.Activate()

	// Assert
	assert.True(t, warehouse.Active)
	assert.True(t, warehouse.UpdatedAt.After(originalUpdatedAt))
}

func TestWarehouse_Deactivate(t *testing.T) {
	// Arrange
	warehouseMother := mother.WarehouseMother{}
	warehouse := warehouseMother.Random()
	originalUpdatedAt := warehouse.UpdatedAt

	// Esperar un momento para que el tiempo de actualización sea diferente
	time.Sleep(time.Millisecond * 10)

	// Act
	warehouse.Deactivate()

	// Assert
	assert.False(t, warehouse.Active)
	assert.True(t, warehouse.UpdatedAt.After(originalUpdatedAt))
}

func TestWarehouse_IsRegular(t *testing.T) {
	// Arrange
	warehouseMother := mother.WarehouseMother{}

	// Act & Assert
	assert.True(t, warehouseMother.RegularType().IsRegular())
	assert.False(t, warehouseMother.SpecialType().IsRegular())
	assert.False(t, warehouseMother.VirtualType().IsRegular())
}

func TestWarehouse_IsSpecial(t *testing.T) {
	// Arrange
	warehouseMother := mother.WarehouseMother{}

	// Act & Assert
	assert.True(t, warehouseMother.SpecialType().IsSpecial())
	assert.False(t, warehouseMother.RegularType().IsSpecial())
	assert.False(t, warehouseMother.VirtualType().IsSpecial())
}

func TestWarehouse_IsVirtual(t *testing.T) {
	// Arrange
	warehouseMother := mother.WarehouseMother{}

	// Act & Assert
	assert.True(t, warehouseMother.VirtualType().IsVirtual())
	assert.False(t, warehouseMother.RegularType().IsVirtual())
	assert.False(t, warehouseMother.SpecialType().IsVirtual())
}
