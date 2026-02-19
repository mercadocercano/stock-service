package request

import "errors"

// CompensateSaleRequest representa el request para compensar una venta
type CompensateSaleRequest struct {
	StockEntryID string `json:"stock_entry_id" binding:"required"`
	Reason       string `json:"reason" binding:"required"`
}

// Validate valida la petición
func (r *CompensateSaleRequest) Validate() error {
	if r.StockEntryID == "" {
		return errors.New("stock_entry_id is required")
	}

	if r.Reason == "" {
		return errors.New("reason is required")
	}

	return nil
}
