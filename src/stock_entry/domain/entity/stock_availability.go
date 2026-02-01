package entity

import (
	"time"

	"github.com/google/uuid"
	
	"stock-service/src/stock_entry/domain/exception"
)

// StockAvailability representa la disponibilidad consolidada de un producto
type StockAvailability struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	
	// Producto
	ProductSKU  string
	ProductID   *uuid.UUID
	ProductName string
	
	// Location (opcional)
	LocationID *uuid.UUID
	
	// Cantidades
	AvailableQuantity float64
	ReservedQuantity  float64
	TotalQuantity     float64
	UnitOfMeasure     string
	
	// Valores
	AvgUnitCost *float64
	TotalValue  *float64
	
	// Niveles de alerta
	MinStockLevel  float64
	MaxStockLevel  *float64
	IsLowStock     bool
	IsOutOfStock   bool
	
	// Metadata
	LastEntryAt      *time.Time
	LastMovementType *string
	
	// Auditoría
	UpdatedAt time.Time
}

// NewStockAvailability crea una nueva instancia de disponibilidad
func NewStockAvailability(
	tenantID uuid.UUID,
	productSKU string,
	totalQuantity float64,
) *StockAvailability {
	now := time.Now()
	
	return &StockAvailability{
		ID:                uuid.New(),
		TenantID:          tenantID,
		ProductSKU:        productSKU,
		AvailableQuantity: totalQuantity,
		ReservedQuantity:  0,
		TotalQuantity:     totalQuantity,
		UnitOfMeasure:     "unit",
		MinStockLevel:     0,
		IsLowStock:        totalQuantity < 10,  // Simplificado
		IsOutOfStock:      totalQuantity <= 0,
		UpdatedAt:         now,
	}
}

// UpdateQuantity actualiza las cantidades
func (sa *StockAvailability) UpdateQuantity(total, reserved float64) {
	sa.TotalQuantity = total
	sa.ReservedQuantity = reserved
	sa.AvailableQuantity = total - reserved
	sa.IsOutOfStock = total <= 0
	sa.IsLowStock = total > 0 && total < sa.MinStockLevel
	sa.UpdatedAt = time.Now()
}

// Reserve reserva una cantidad
func (sa *StockAvailability) Reserve(quantity float64) error {
	if quantity > sa.AvailableQuantity {
		return exception.ErrInsufficientStock
	}
	
	sa.ReservedQuantity += quantity
	sa.AvailableQuantity = sa.TotalQuantity - sa.ReservedQuantity
	sa.UpdatedAt = time.Now()
	return nil
}

// Release libera una cantidad reservada
func (sa *StockAvailability) Release(quantity float64) {
	sa.ReservedQuantity -= quantity
	if sa.ReservedQuantity < 0 {
		sa.ReservedQuantity = 0
	}
	sa.AvailableQuantity = sa.TotalQuantity - sa.ReservedQuantity
	sa.UpdatedAt = time.Now()
}

// SetStockLevels configura los niveles mínimo y máximo
func (sa *StockAvailability) SetStockLevels(min, max float64) {
	sa.MinStockLevel = min
	maxVal := max
	sa.MaxStockLevel = &maxVal
	sa.IsLowStock = sa.TotalQuantity > 0 && sa.TotalQuantity < min
	sa.UpdatedAt = time.Now()
}

// UpdateValue actualiza el valor total basado en costo promedio
func (sa *StockAvailability) UpdateValue(avgCost float64) {
	sa.AvgUnitCost = &avgCost
	totalVal := sa.TotalQuantity * avgCost
	sa.TotalValue = &totalVal
	sa.UpdatedAt = time.Now()
}

