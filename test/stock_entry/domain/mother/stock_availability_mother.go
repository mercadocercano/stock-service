package mother

import (
	"time"

	"github.com/google/uuid"

	"stock/src/stock_entry/domain/entity"
)

// StockAvailabilityMother es un factory para crear entidades StockAvailability para pruebas
type StockAvailabilityMother struct{}

// Random crea un StockAvailability con datos aleatorios
func (StockAvailabilityMother) Random() *entity.StockAvailability {
	return &entity.StockAvailability{
		ID:                uuid.New(),
		TenantID:          uuid.New(),
		VariantSKU:        "SKU-" + uuid.New().String()[:8],
		ProductSKU:        "SKU-" + uuid.New().String()[:8],
		AvailableQuantity: 50,
		ReservedQuantity:  0,
		TotalQuantity:     50,
		UnitOfMeasure:     "unit",
		MinStockLevel:     5,
		IsLowStock:        false,
		IsOutOfStock:      false,
		UpdatedAt:         time.Now(),
	}
}

// WithTenantID crea un StockAvailability con un TenantID especifico
func (m StockAvailabilityMother) WithTenantID(tenantID uuid.UUID) *entity.StockAvailability {
	avail := m.Random()
	avail.TenantID = tenantID
	return avail
}

// WithSKU crea un StockAvailability con un SKU especifico
func (m StockAvailabilityMother) WithSKU(sku string) *entity.StockAvailability {
	avail := m.Random()
	avail.VariantSKU = sku
	avail.ProductSKU = sku
	return avail
}

// WithQuantities crea un StockAvailability con cantidades especificas
func (m StockAvailabilityMother) WithQuantities(total, reserved float64) *entity.StockAvailability {
	avail := m.Random()
	avail.TotalQuantity = total
	avail.ReservedQuantity = reserved
	avail.AvailableQuantity = total - reserved
	avail.IsOutOfStock = total <= 0
	avail.IsLowStock = total > 0 && total < avail.MinStockLevel
	return avail
}

// OutOfStock crea un StockAvailability sin stock
func (m StockAvailabilityMother) OutOfStock() *entity.StockAvailability {
	avail := m.Random()
	avail.TotalQuantity = 0
	avail.AvailableQuantity = 0
	avail.ReservedQuantity = 0
	avail.IsOutOfStock = true
	avail.IsLowStock = false
	return avail
}

// LowStock crea un StockAvailability con bajo stock
func (m StockAvailabilityMother) LowStock() *entity.StockAvailability {
	avail := m.Random()
	avail.TotalQuantity = 3
	avail.AvailableQuantity = 3
	avail.ReservedQuantity = 0
	avail.MinStockLevel = 5
	avail.IsLowStock = true
	avail.IsOutOfStock = false
	return avail
}

// WithReservation crea un StockAvailability con reserva
func (m StockAvailabilityMother) WithReservation(total, reserved float64) *entity.StockAvailability {
	avail := m.Random()
	avail.TotalQuantity = total
	avail.ReservedQuantity = reserved
	avail.AvailableQuantity = total - reserved
	return avail
}
