package exception

import "fmt"

// StockLocationNotFound representa el error cuando no se encuentra una ubicación de stock
type StockLocationNotFound struct {
	ID       string
	TenantID string
}

// Error devuelve el mensaje de error
func (e *StockLocationNotFound) Error() string {
	return fmt.Sprintf("stock location with ID '%s' not found for tenant '%s'", e.ID, e.TenantID)
} 