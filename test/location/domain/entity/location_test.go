package entity_test

import (
	"testing"
	"time"

	"stock/src/location/domain/entity"
	"stock/test/location/domain/mother"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewLocation(t *testing.T) {
	// Arrange
	tenantID := "tenant-" + uuid.New().String()
	name := "Test Location"
	locationType := entity.StoreType
	address := "123 Test St"
	city := "Test City"
	state := "Test State"
	country := "Test Country"
	postalCode := "12345"
	phone := "+1234567890"
	email := "test@example.com"

	// Act
	location := entity.NewLocation(
		tenantID,
		name,
		locationType,
		address,
		city,
		state,
		country,
		postalCode,
		phone,
		email,
	)

	// Assert
	assert.NotEmpty(t, location.ID)
	assert.Equal(t, tenantID, location.TenantID)
	assert.Equal(t, name, location.Name)
	assert.Equal(t, locationType, location.Type)
	assert.Equal(t, address, location.Address)
	assert.Equal(t, city, location.City)
	assert.Equal(t, state, location.State)
	assert.Equal(t, country, location.Country)
	assert.Equal(t, postalCode, location.PostalCode)
	assert.Equal(t, phone, location.Phone)
	assert.Equal(t, email, location.Email)
	assert.True(t, location.Active)
	assert.NotZero(t, location.CreatedAt)
	assert.NotZero(t, location.UpdatedAt)
}

func TestLocation_Update(t *testing.T) {
	// Arrange
	locationMother := mother.LocationMother{}
	location := locationMother.Random()
	originalUpdatedAt := location.UpdatedAt

	// Esperar un momento para que el tiempo de actualización sea diferente
	time.Sleep(time.Millisecond * 10)

	newName := "Updated Location"
	newAddress := "456 Updated St"
	newCity := "Updated City"
	newState := "Updated State"
	newCountry := "Updated Country"
	newPostalCode := "54321"
	newPhone := "+0987654321"
	newEmail := "updated@example.com"

	// Act
	location.Update(
		newName,
		newAddress,
		newCity,
		newState,
		newCountry,
		newPostalCode,
		newPhone,
		newEmail,
	)

	// Assert
	assert.Equal(t, newName, location.Name)
	assert.Equal(t, newAddress, location.Address)
	assert.Equal(t, newCity, location.City)
	assert.Equal(t, newState, location.State)
	assert.Equal(t, newCountry, location.Country)
	assert.Equal(t, newPostalCode, location.PostalCode)
	assert.Equal(t, newPhone, location.Phone)
	assert.Equal(t, newEmail, location.Email)
	assert.True(t, location.UpdatedAt.After(originalUpdatedAt))
}

func TestLocation_Activate(t *testing.T) {
	// Arrange
	locationMother := mother.LocationMother{}
	location := locationMother.Inactive()
	originalUpdatedAt := location.UpdatedAt

	// Esperar un momento para que el tiempo de actualización sea diferente
	time.Sleep(time.Millisecond * 10)

	// Act
	location.Activate()

	// Assert
	assert.True(t, location.Active)
	assert.True(t, location.UpdatedAt.After(originalUpdatedAt))
}

func TestLocation_Deactivate(t *testing.T) {
	// Arrange
	locationMother := mother.LocationMother{}
	location := locationMother.Random()
	originalUpdatedAt := location.UpdatedAt

	// Esperar un momento para que el tiempo de actualización sea diferente
	time.Sleep(time.Millisecond * 10)

	// Act
	location.Deactivate()

	// Assert
	assert.False(t, location.Active)
	assert.True(t, location.UpdatedAt.After(originalUpdatedAt))
}

func TestLocation_IsStore(t *testing.T) {
	// Arrange
	locationMother := mother.LocationMother{}

	// Act & Assert
	assert.True(t, locationMother.Store().IsStore())
	assert.False(t, locationMother.DistributionCenter().IsStore())
}

func TestLocation_IsDistributionCenter(t *testing.T) {
	// Arrange
	locationMother := mother.LocationMother{}

	// Act & Assert
	assert.True(t, locationMother.DistributionCenter().IsDistributionCenter())
	assert.False(t, locationMother.Store().IsDistributionCenter())
}
