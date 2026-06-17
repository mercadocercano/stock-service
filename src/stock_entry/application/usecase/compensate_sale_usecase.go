package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"stock/src/stock_entry/application/request"
	"stock/src/stock_entry/application/response"
	"stock/src/stock_entry/domain/port"
)

// CompensateSaleUseCase caso de uso para compensar (revertir) una venta
// HITO D: Usado para rollback cuando falla persistencia de orden/sale
type CompensateSaleUseCase struct {
	stockEntryRepo port.StockEntryRepository
	logger         port.StockEventLogger
}

// NewCompensateSaleUseCase crea una nueva instancia del caso de uso
func NewCompensateSaleUseCase(stockEntryRepo port.StockEntryRepository, logger port.StockEventLogger) *CompensateSaleUseCase {
	return &CompensateSaleUseCase{
		stockEntryRepo: stockEntryRepo,
		logger:         logger,
	}
}

// logEvent emite un evento canónico si hay logger inyectado (nil-safe).
func (uc *CompensateSaleUseCase) logEvent(e port.StockEvent) {
	if uc.logger != nil {
		uc.logger.Log(e)
	}
}

// Execute ejecuta la compensación de una venta
func (uc *CompensateSaleUseCase) Execute(
	ctx context.Context,
	tenantID string,
	req *request.CompensateSaleRequest,
) (*response.CompensateSaleResponse, error) {
	// Validar request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Parsear tenant ID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	// Parsear stock entry ID
	stockEntryID, err := uuid.Parse(req.StockEntryID)
	if err != nil {
		return nil, fmt.Errorf("invalid stock_entry_id: %w", err)
	}

	// Ejecutar compensación (método del repositorio implementado en HITO D)
	if err := uc.stockEntryRepo.CompensateSale(ctx, tenantUUID, stockEntryID, req.Reason); err != nil {
		uc.logEvent(port.StockEvent{
			Event:        "stock.compensate_failed",
			TenantID:     tenantID,
			StockEntryID: req.StockEntryID,
			Reason:       err.Error(),
		})
		return nil, fmt.Errorf("failed to compensate sale: %w", err)
	}

	uc.logEvent(port.StockEvent{
		Event:        "stock.compensated",
		TenantID:     tenantID,
		StockEntryID: req.StockEntryID,
		Reason:       req.Reason,
	})

	return &response.CompensateSaleResponse{
		Success:      true,
		Message:      "Sale compensated successfully",
		StockEntryID: req.StockEntryID,
		Reason:       req.Reason,
	}, nil
}
