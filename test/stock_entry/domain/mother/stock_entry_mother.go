package mother

import (
	"time"

	"github.com/google/uuid"

	"stock/src/stock_entry/domain/entity"
)

// StockEntryMother es un factory para crear entidades StockEntry para pruebas
type StockEntryMother struct{}

// Random crea un StockEntry con datos aleatorios de tipo purchase
func (StockEntryMother) Random() *entity.StockEntry {
	now := time.Now()
	return &entity.StockEntry{
		ID:            uuid.New(),
		TenantID:      uuid.New(),
		VariantSKU:    "SKU-" + uuid.New().String()[:8],
		ProductSKU:    "SKU-" + uuid.New().String()[:8],
		EntryType:     entity.EntryTypePurchase,
		Quantity:      10,
		UnitOfMeasure: "unit",
		Status:        entity.EntryStatusConfirmed,
		IsActive:      true,
		Metadata:      make(map[string]interface{}),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// WithTenantID crea un StockEntry con un TenantID especifico
func (m StockEntryMother) WithTenantID(tenantID uuid.UUID) *entity.StockEntry {
	entry := m.Random()
	entry.TenantID = tenantID
	return entry
}

// WithVariantSKU crea un StockEntry con un VariantSKU especifico
func (m StockEntryMother) WithVariantSKU(sku string) *entity.StockEntry {
	entry := m.Random()
	entry.VariantSKU = sku
	entry.ProductSKU = sku
	return entry
}

// WithEntryType crea un StockEntry con un tipo especifico
func (m StockEntryMother) WithEntryType(entryType entity.EntryType) *entity.StockEntry {
	entry := m.Random()
	entry.EntryType = entryType
	return entry
}

// WithQuantity crea un StockEntry con una cantidad especifica
func (m StockEntryMother) WithQuantity(quantity float64) *entity.StockEntry {
	entry := m.Random()
	entry.Quantity = quantity
	return entry
}

// Sale crea un StockEntry de tipo venta
func (m StockEntryMother) Sale() *entity.StockEntry {
	entry := m.Random()
	entry.EntryType = entity.EntryTypeSale
	entry.Quantity = 5
	return entry
}

// Purchase crea un StockEntry de tipo compra
func (m StockEntryMother) Purchase() *entity.StockEntry {
	entry := m.Random()
	entry.EntryType = entity.EntryTypePurchase
	entry.Quantity = 20
	return entry
}

// InitialStock crea un StockEntry de tipo stock inicial
func (m StockEntryMother) InitialStock() *entity.StockEntry {
	entry := m.Random()
	entry.EntryType = entity.EntryTypeInitialStock
	entry.Quantity = 100
	return entry
}

// Pending crea un StockEntry en estado pendiente
func (m StockEntryMother) Pending() *entity.StockEntry {
	entry := m.Random()
	entry.Status = entity.EntryStatusPending
	return entry
}

// Cancelled crea un StockEntry en estado cancelado
func (m StockEntryMother) Cancelled() *entity.StockEntry {
	entry := m.Random()
	entry.Status = entity.EntryStatusCancelled
	entry.IsActive = false
	return entry
}

// Return crea un StockEntry de tipo devolucion
func (m StockEntryMother) Return() *entity.StockEntry {
	entry := m.Random()
	entry.EntryType = entity.EntryTypeReturn
	entry.Quantity = 3
	return entry
}

// TransferIn crea un StockEntry de tipo transferencia de entrada
func (m StockEntryMother) TransferIn() *entity.StockEntry {
	entry := m.Random()
	entry.EntryType = entity.EntryTypeTransferIn
	entry.Quantity = 15
	return entry
}

// TransferOut crea un StockEntry de tipo transferencia de salida
func (m StockEntryMother) TransferOut() *entity.StockEntry {
	entry := m.Random()
	entry.EntryType = entity.EntryTypeTransferOut
	entry.Quantity = 8
	return entry
}

// Adjustment crea un StockEntry de tipo ajuste con cantidad positiva
func (m StockEntryMother) Adjustment(quantity float64) *entity.StockEntry {
	entry := m.Random()
	entry.EntryType = entity.EntryTypeAdjustment
	entry.Quantity = quantity
	return entry
}
