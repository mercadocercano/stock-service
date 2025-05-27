package entity

import (
	"time"

	"github.com/google/uuid"
)

// LocationType representa el tipo de ubicación
type LocationType string

const (
	// StoreType representa un punto de venta
	StoreType LocationType = "store"
	// DistributionCenterType representa un centro de distribución
	DistributionCenterType LocationType = "distribution_center"
)

// Location representa una ubicación física (tienda o centro de distribución)
type Location struct {
	ID         string       `json:"id"`
	TenantID   string       `json:"tenant_id"`
	Name       string       `json:"name"`
	Type       LocationType `json:"type"`
	Address    string       `json:"address"`
	City       string       `json:"city"`
	State      string       `json:"state"`
	Country    string       `json:"country"`
	PostalCode string       `json:"postal_code"`
	Phone      string       `json:"phone"`
	Email      string       `json:"email"`
	Active     bool         `json:"active"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

// NewLocation crea una nueva ubicación
func NewLocation(tenantID, name string, locationType LocationType, address, city, state, country, postalCode, phone, email string) *Location {
	now := time.Now()
	return &Location{
		ID:         uuid.New().String(),
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
		Active:     true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Update actualiza los datos de la ubicación
func (l *Location) Update(name, address, city, state, country, postalCode, phone, email string) {
	l.Name = name
	l.Address = address
	l.City = city
	l.State = state
	l.Country = country
	l.PostalCode = postalCode
	l.Phone = phone
	l.Email = email
	l.UpdatedAt = time.Now()
}

// Activate activa la ubicación
func (l *Location) Activate() {
	l.Active = true
	l.UpdatedAt = time.Now()
}

// Deactivate desactiva la ubicación
func (l *Location) Deactivate() {
	l.Active = false
	l.UpdatedAt = time.Now()
}

// IsStore verifica si la ubicación es una tienda
func (l *Location) IsStore() bool {
	return l.Type == StoreType
}

// IsDistributionCenter verifica si la ubicación es un centro de distribución
func (l *Location) IsDistributionCenter() bool {
	return l.Type == DistributionCenterType
}
