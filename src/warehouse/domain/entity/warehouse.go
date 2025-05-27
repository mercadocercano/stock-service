package entity

import (
	"time"

	"github.com/google/uuid"
)

// WarehouseType representa el tipo de almacén
type WarehouseType string

const (
	// RegularWarehouseType representa un almacén regular
	RegularWarehouseType WarehouseType = "regular"
	// SpecialWarehouseType representa un almacén especial (refrigerado, seguridad, etc.)
	SpecialWarehouseType WarehouseType = "special"
	// VirtualWarehouseType representa un almacén virtual (sin ubicación física real)
	VirtualWarehouseType WarehouseType = "virtual"
)

// Warehouse representa un almacén dentro de una ubicación física
type Warehouse struct {
	ID          string        `json:"id"`
	TenantID    string        `json:"tenant_id"`
	LocationID  string        `json:"location_id"`
	Name        string        `json:"name"`
	Code        string        `json:"code"`
	Type        WarehouseType `json:"type"`
	Description string        `json:"description"`
	Priority    int           `json:"priority"`
	Active      bool          `json:"active"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// NewWarehouse crea un nuevo almacén
func NewWarehouse(
	tenantID string,
	locationID string,
	name string,
	code string,
	warehouseType WarehouseType,
	description string,
	priority int,
) *Warehouse {
	now := time.Now()
	return &Warehouse{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		LocationID:  locationID,
		Name:        name,
		Code:        code,
		Type:        warehouseType,
		Description: description,
		Priority:    priority,
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Update actualiza los datos del almacén
func (w *Warehouse) Update(name, code string, warehouseType WarehouseType, description string, priority int) {
	w.Name = name
	w.Code = code
	w.Type = warehouseType
	w.Description = description
	w.Priority = priority
	w.UpdatedAt = time.Now()
}

// Activate activa el almacén
func (w *Warehouse) Activate() {
	w.Active = true
	w.UpdatedAt = time.Now()
}

// Deactivate desactiva el almacén
func (w *Warehouse) Deactivate() {
	w.Active = false
	w.UpdatedAt = time.Now()
}

// IsRegular verifica si el almacén es de tipo regular
func (w *Warehouse) IsRegular() bool {
	return w.Type == RegularWarehouseType
}

// IsSpecial verifica si el almacén es de tipo especial
func (w *Warehouse) IsSpecial() bool {
	return w.Type == SpecialWarehouseType
}

// IsVirtual verifica si el almacén es de tipo virtual
func (w *Warehouse) IsVirtual() bool {
	return w.Type == VirtualWarehouseType
}
