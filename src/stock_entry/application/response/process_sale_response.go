package response

// ProcessSaleResponse representa la respuesta de una venta procesada
type ProcessSaleResponse struct {
	Success           bool    `json:"success"`
	Message           string  `json:"message"`
	VariantSKU        string  `json:"variant_sku"`
	QuantitySold      float64 `json:"quantity_sold"`
	RemainingStock    float64 `json:"remaining_stock"`
	StockEntryID      string  `json:"stock_entry_id"`
}
