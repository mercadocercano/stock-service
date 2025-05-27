package exception

import "fmt"

// WarehouseNotFound representa el error cuando no se encuentra un almacén
type WarehouseNotFound struct {
	ID       string
	TenantID string
}

// Error implementa la interfaz error
func (e *WarehouseNotFound) Error() string {
	return fmt.Sprintf("warehouse with id '%s' not found for tenant '%s'", e.ID, e.TenantID)
}

// NewWarehouseNotFound crea una nueva instancia del error
func NewWarehouseNotFound(id, tenantID string) *WarehouseNotFound {
	return &WarehouseNotFound{
		ID:       id,
		TenantID: tenantID,
	}
}
