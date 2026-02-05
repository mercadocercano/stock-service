package request

// ReserveStockRequest representa la petición para reservar stock
type ReserveStockRequest struct {
	SKU       string `json:"sku" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	Reference string `json:"reference" binding:"required"`
}
