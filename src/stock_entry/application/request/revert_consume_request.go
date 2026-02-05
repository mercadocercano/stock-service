package request

// RevertConsumeRequest representa la petición para revertir un consumo de stock
type RevertConsumeRequest struct {
	SKU       string `json:"sku" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	Reference string `json:"reference" binding:"required"`
}
