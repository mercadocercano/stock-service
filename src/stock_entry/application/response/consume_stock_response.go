package response

// ConsumeStockResponse representa la respuesta de consumo de stock
type ConsumeStockResponse struct {
	SKU         string `json:"sku"`
	ConsumedQty int    `json:"consumed_qty"`
	ReservedQty int    `json:"reserved_qty"`
	Reference   string `json:"reference"`
}
