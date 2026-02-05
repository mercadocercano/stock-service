package request

// ConsumeStockRequest representa la petición para consumir stock reservado
type ConsumeStockRequest struct {
	SKU       string `json:"sku" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	Reference string `json:"reference" binding:"required"`
}
