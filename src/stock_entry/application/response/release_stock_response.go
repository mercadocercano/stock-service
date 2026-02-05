package response

// ReleaseStockResponse representa la respuesta de liberación de stock
type ReleaseStockResponse struct {
	SKU          string `json:"sku"`
	ReleasedQty  int    `json:"released_qty"`
	AvailableQty int    `json:"available_qty"`
	ReservedQty  int    `json:"reserved_qty"`
	Reference    string `json:"reference"`
}
