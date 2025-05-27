package exception

import "fmt"

// LocationNotFound representa el error cuando una ubicación no se encuentra
type LocationNotFound struct {
	ID       string
	TenantID string
}

// Error implementa la interfaz error
func (e *LocationNotFound) Error() string {
	return fmt.Sprintf("Location with ID '%s' not found for tenant '%s'", e.ID, e.TenantID)
}

// NewLocationNotFound crea una nueva instancia de LocationNotFound
func NewLocationNotFound(id string, tenantID string) *LocationNotFound {
	return &LocationNotFound{
		ID:       id,
		TenantID: tenantID,
	}
}
