package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// EntryType representa el tipo de movimiento de stock
type EntryType string

const (
	EntryTypeInitialStock EntryType = "initial_stock"
	EntryTypePurchase     EntryType = "purchase"
	EntryTypeAdjustment   EntryType = "adjustment"
	EntryTypeTransferIn   EntryType = "transfer_in"
	EntryTypeTransferOut  EntryType = "transfer_out"
	EntryTypeSale         EntryType = "sale"
	EntryTypeReturn       EntryType = "return"
)

// EntryStatus representa el estado de una entrada de stock
type EntryStatus string

const (
	EntryStatusPending   EntryStatus = "pending"
	EntryStatusConfirmed EntryStatus = "confirmed"
	EntryStatusCancelled EntryStatus = "cancelled"
)

// StockEntry representa un movimiento de entrada/salida de stock
type StockEntry struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	
	// Variante del producto (HITO 2.1)
	VariantSKU  string      // Campo principal - SKU de la variante
	ProductSKU  string      // Alias para compatibilidad (mismo valor que VariantSKU)
	ProductID   *uuid.UUID  // Opcional - ID del producto padre
	ProductName string
	
	// Location
	LocationID *uuid.UUID  // Opcional
	
	// Tipo y cantidades
	EntryType       EntryType
	Quantity        float64
	UnitOfMeasure   string
	
	// Costos
	UnitCost  *float64
	TotalCost *float64
	
	// Metadata
	ReferenceNumber *string
	Notes           *string
	Metadata        map[string]interface{}
	
	// Estado
	Status   EntryStatus
	IsActive bool
	
	// Auditoría
	CreatedBy *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewStockEntry crea una nueva entrada de stock (HITO 2.1 - variant_sku)
func NewStockEntry(
	tenantID uuid.UUID,
	variantSKU string,
	entryType EntryType,
	quantity float64,
) (*StockEntry, error) {
	// Validaciones
	if variantSKU == "" {
		return nil, errors.New("variant SKU is required")
	}
	
	if quantity == 0 {
		return nil, errors.New("quantity cannot be zero")
	}
	
	if !isValidEntryType(entryType) {
		return nil, errors.New("invalid entry type")
	}
	
	now := time.Now()
	
	return &StockEntry{
		ID:            uuid.New(),
		TenantID:      tenantID,
		VariantSKU:    variantSKU,
		ProductSKU:    variantSKU,  // Mantener sincronizado para compatibilidad
		EntryType:     entryType,
		Quantity:      quantity,
		UnitOfMeasure: "unit",  // Default
		Status:        EntryStatusConfirmed,
		IsActive:      true,
		Metadata:      make(map[string]interface{}),
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// SetProductInfo establece información adicional del producto
func (se *StockEntry) SetProductInfo(productID *uuid.UUID, productName string) {
	se.ProductID = productID
	se.ProductName = productName
	se.UpdatedAt = time.Now()
}

// SetLocation establece la ubicación del stock
func (se *StockEntry) SetLocation(locationID uuid.UUID) {
	se.LocationID = &locationID
	se.UpdatedAt = time.Now()
}

// SetCosts establece los costos unitario y total
func (se *StockEntry) SetCosts(unitCost, totalCost float64) {
	se.UnitCost = &unitCost
	se.TotalCost = &totalCost
	se.UpdatedAt = time.Now()
}

// SetReference establece el número de referencia
func (se *StockEntry) SetReference(refNumber string) {
	se.ReferenceNumber = &refNumber
	se.UpdatedAt = time.Now()
}

// SetNotes establece notas adicionales
func (se *StockEntry) SetNotes(notes string) {
	se.Notes = &notes
	se.UpdatedAt = time.Now()
}

// Confirm confirma la entrada de stock
func (se *StockEntry) Confirm() error {
	if se.Status == EntryStatusCancelled {
		return errors.New("cannot confirm a cancelled entry")
	}
	
	se.Status = EntryStatusConfirmed
	se.UpdatedAt = time.Now()
	return nil
}

// Cancel cancela la entrada de stock
func (se *StockEntry) Cancel() error {
	if se.Status == EntryStatusConfirmed {
		return errors.New("cannot cancel a confirmed entry")
	}
	
	se.Status = EntryStatusCancelled
	se.IsActive = false
	se.UpdatedAt = time.Now()
	return nil
}

// IsPositiveMovement indica si el movimiento suma al stock
func (se *StockEntry) IsPositiveMovement() bool {
	return se.EntryType == EntryTypeInitialStock ||
		se.EntryType == EntryTypePurchase ||
		se.EntryType == EntryTypeAdjustment && se.Quantity > 0 ||
		se.EntryType == EntryTypeTransferIn ||
		se.EntryType == EntryTypeReturn
}

// IsNegativeMovement indica si el movimiento resta del stock
func (se *StockEntry) IsNegativeMovement() bool {
	return se.EntryType == EntryTypeTransferOut ||
		se.EntryType == EntryTypeSale ||
		se.EntryType == EntryTypeAdjustment && se.Quantity < 0
}

// CalculatedQuantity retorna la cantidad con su signo según el tipo de movimiento
func (se *StockEntry) CalculatedQuantity() float64 {
	if se.IsNegativeMovement() {
		return -se.Quantity
	}
	return se.Quantity
}

// Validate valida la entrada de stock
func (se *StockEntry) Validate() error {
	if se.TenantID == uuid.Nil {
		return errors.New("tenant ID is required")
	}
	
	if se.ProductSKU == "" {
		return errors.New("product SKU is required")
	}
	
	if se.Quantity == 0 {
		return errors.New("quantity cannot be zero")
	}
	
	if !isValidEntryType(se.EntryType) {
		return errors.New("invalid entry type")
	}
	
	if !isValidStatus(se.Status) {
		return errors.New("invalid status")
	}
	
	return nil
}

// Helper functions

func isValidEntryType(et EntryType) bool {
	switch et {
	case EntryTypeInitialStock, EntryTypePurchase, EntryTypeAdjustment,
		EntryTypeTransferIn, EntryTypeTransferOut, EntryTypeSale, EntryTypeReturn:
		return true
	}
	return false
}

func isValidStatus(status EntryStatus) bool {
	switch status {
	case EntryStatusPending, EntryStatusConfirmed, EntryStatusCancelled:
		return true
	}
	return false
}

