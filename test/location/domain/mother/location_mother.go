package mother

import (
	"stock/src/location/domain/entity"
	"time"

	"github.com/google/uuid"
)

// LocationMother es un factory para crear entidades Location para pruebas
type LocationMother struct{}

// Random crea una Location con datos aleatorios
func (LocationMother) Random() *entity.Location {
	return &entity.Location{
		ID:         uuid.New().String(),
		TenantID:   "tenant-" + uuid.New().String(),
		Name:       "Test Location",
		Type:       entity.StoreType,
		Address:    "123 Test St",
		City:       "Test City",
		State:      "Test State",
		Country:    "Test Country",
		PostalCode: "12345",
		Phone:      "+1234567890",
		Email:      "test@example.com",
		Active:     true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// WithID crea una Location con un ID específico
func (m LocationMother) WithID(id string) *entity.Location {
	location := m.Random()
	location.ID = id
	return location
}

// WithTenantID crea una Location con un TenantID específico
func (m LocationMother) WithTenantID(tenantID string) *entity.Location {
	location := m.Random()
	location.TenantID = tenantID
	return location
}

// WithName crea una Location con un nombre específico
func (m LocationMother) WithName(name string) *entity.Location {
	location := m.Random()
	location.Name = name
	return location
}

// WithType crea una Location con un tipo específico
func (m LocationMother) WithType(locationType entity.LocationType) *entity.Location {
	location := m.Random()
	location.Type = locationType
	return location
}

// Store crea una Location de tipo tienda
func (m LocationMother) Store() *entity.Location {
	location := m.Random()
	location.Type = entity.StoreType
	return location
}

// DistributionCenter crea una Location de tipo centro de distribución
func (m LocationMother) DistributionCenter() *entity.Location {
	location := m.Random()
	location.Type = entity.DistributionCenterType
	return location
}

// Inactive crea una Location inactiva
func (m LocationMother) Inactive() *entity.Location {
	location := m.Random()
	location.Active = false
	return location
}

// Complete crea una Location con todos los datos personalizados
func (m LocationMother) Complete(
	id string,
	tenantID string,
	name string,
	locationType entity.LocationType,
	address string,
	city string,
	state string,
	country string,
	postalCode string,
	phone string,
	email string,
	active bool,
	createdAt time.Time,
	updatedAt time.Time,
) *entity.Location {
	return &entity.Location{
		ID:         id,
		TenantID:   tenantID,
		Name:       name,
		Type:       locationType,
		Address:    address,
		City:       city,
		State:      state,
		Country:    country,
		PostalCode: postalCode,
		Phone:      phone,
		Email:      email,
		Active:     active,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}
}
