package response

// RevertConsumeResponse representa la respuesta de reversión de consumo
type RevertConsumeResponse struct {
	SKU          string `json:"sku"`
	RevertedQty  int    `json:"reverted_qty"`
	AvailableQty int    `json:"available_qty"`
	Reference    string `json:"reference"`
}
