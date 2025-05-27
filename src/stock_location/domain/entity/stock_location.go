package entity

import (
	"time"

	"github.com/google/uuid"
)

// StockLocation representa una ubicación dentro de un almacén
type StockLocation struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	WarehouseID string    `json:"warehouse_id"`
	ParentID    *string   `json:"parent_id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Path        string    `json:"path"`
	Level       int       `json:"level"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewStockLocation crea una nueva ubicación de stock
func NewStockLocation(tenantID, warehouseID string, parentID *string, name, code, description string) *StockLocation {
	now := time.Now()
	id := uuid.New().String()

	// Calcular path y level
	path := id
	level := 1

	if parentID != nil && *parentID != "" {
		// Para una implementación real, necesitaríamos recuperar el parent y su path
		// Aquí simplemente concatenamos el ID del parent con el nuevo ID
		path = *parentID + "/" + id
		level = 2 // Simplificado, en realidad debería ser parent.Level + 1
	}

	return &StockLocation{
		ID:          id,
		TenantID:    tenantID,
		WarehouseID: warehouseID,
		ParentID:    parentID,
		Name:        name,
		Code:        code,
		Path:        path,
		Level:       level,
		Description: description,
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Update actualiza los datos de la ubicación de stock
func (sl *StockLocation) Update(name, code, description string) {
	sl.Name = name
	sl.Code = code
	sl.Description = description
	sl.UpdatedAt = time.Now()
}

// Activate activa la ubicación de stock
func (sl *StockLocation) Activate() {
	sl.Active = true
	sl.UpdatedAt = time.Now()
}

// Deactivate desactiva la ubicación de stock
func (sl *StockLocation) Deactivate() {
	sl.Active = false
	sl.UpdatedAt = time.Now()
}

// IsRoot verifica si esta ubicación es raíz (no tiene padre)
func (sl *StockLocation) IsRoot() bool {
	return sl.ParentID == nil || *sl.ParentID == ""
}
