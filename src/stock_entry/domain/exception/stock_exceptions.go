package exception

import "errors"

var (
	// ErrStockEntryNotFound cuando no se encuentra una entrada de stock
	ErrStockEntryNotFound = errors.New("stock entry not found")
	
	// ErrStockAvailabilityNotFound cuando no se encuentra disponibilidad para un producto
	ErrStockAvailabilityNotFound = errors.New("stock availability not found")
	
	// ErrStockNotInitialized cuando el producto nunca tuvo movimientos de stock
	ErrStockNotInitialized = errors.New("stock not initialized for this product")
	
	// ErrInsufficientStock cuando no hay suficiente stock disponible
	ErrInsufficientStock = errors.New("insufficient stock available")
	
	// ErrInvalidQuantity cuando la cantidad es inválida
	ErrInvalidQuantity = errors.New("invalid quantity")
	
	// ErrInvalidEntryType cuando el tipo de entrada es inválido
	ErrInvalidEntryType = errors.New("invalid entry type")
	
	// ErrInvalidStatus cuando el status es inválido
	ErrInvalidStatus = errors.New("invalid status")
	
	// ErrProductSKURequired cuando el SKU del producto es requerido
	ErrProductSKURequired = errors.New("product SKU is required")
	
	// ErrTenantIDRequired cuando el tenant ID es requerido
	ErrTenantIDRequired = errors.New("tenant ID is required")
	
	// ErrCannotCancelConfirmed cuando se intenta cancelar una entrada confirmada
	ErrCannotCancelConfirmed = errors.New("cannot cancel a confirmed entry")
	
	// ErrCannotConfirmCancelled cuando se intenta confirmar una entrada cancelada
	ErrCannotConfirmCancelled = errors.New("cannot confirm a cancelled entry")
)

