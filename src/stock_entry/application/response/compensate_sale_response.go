package response

// CompensateSaleResponse representa la respuesta de compensación
type CompensateSaleResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	StockEntryID string `json:"stock_entry_id"`
	Reason       string `json:"reason"`
}
