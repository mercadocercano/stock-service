package response

// ReserveStockResponse representa la respuesta de reserva de stock
type ReserveStockResponse struct {
	SKU          string `json:"sku"`
	ReservedQty  int    `json:"reserved_qty"`
	RemainingQty int    `json:"remaining_qty"`
	Reference    string `json:"reference"`
}
