package request

// ProcessSaleRequest representa una petición mínima para procesar una venta
type ProcessSaleRequest struct {
	VariantSKU string  `json:"variant_sku" binding:"required"`
	Quantity   float64 `json:"quantity" binding:"required,gt=0"`
	Reference  string  `json:"reference,omitempty"` // Opcional: referencia externa (POS sale ID, order ID, etc.)
}

// Validate valida la petición
func (r *ProcessSaleRequest) Validate() error {
	if r.VariantSKU == "" {
		return ErrProductSKURequired
	}
	
	if r.Quantity <= 0 {
		return &ValidationError{Message: "quantity must be greater than zero"}
	}
	
	return nil
}
