package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	
	"stock/src/stock_entry/application/request"
	"stock/src/stock_entry/application/response"
	"stock/src/stock_entry/domain/exception"
	"stock/src/stock_entry/domain/port"
	"stock/src/stock_entry/infrastructure/metrics"
)

// ProcessSaleUseCase caso de uso mínimo para procesar una venta
type ProcessSaleUseCase struct {
	stockEntryRepo   port.StockEntryRepository
	availabilityRepo port.StockAvailabilityRepository
}

// NewProcessSaleUseCase crea una nueva instancia del caso de uso
func NewProcessSaleUseCase(
	stockEntryRepo port.StockEntryRepository,
	availabilityRepo port.StockAvailabilityRepository,
) *ProcessSaleUseCase {
	return &ProcessSaleUseCase{
		stockEntryRepo:   stockEntryRepo,
		availabilityRepo: availabilityRepo,
	}
}

// Execute ejecuta el caso de uso de venta con operación atómica
// HITO D: Refactorizado para eliminar race conditions
func (uc *ProcessSaleUseCase) Execute(ctx context.Context, tenantID string, req *request.ProcessSaleRequest) (*response.ProcessSaleResponse, error) {
	// Validar request
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// Parsear tenant ID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}
	
	// Generar referencia única (o usar la provista)
	reference := req.Reference
	if reference == "" {
		reference = fmt.Sprintf("SALE-%s", uuid.New().String()[:8])
	}

	// OPERACIÓN ATÓMICA: lock + validar + descontar en una sola transacción
	// Elimina race condition entre CheckAvailability y ProcessSale
	stockEntry, err := uc.stockEntryRepo.ProcessSaleAtomic(
		ctx,
		tenantUUID,
		req.VariantSKU,
		req.Quantity,
		reference,
	)
	
	if err != nil {
		// Distinguir errores de negocio (400/409) vs errores técnicos (500)
		
		// Stock no inicializado → producto sin movimientos previos
		if errors.Is(err, exception.ErrStockNotInitialized) {
			return &response.ProcessSaleResponse{
				Success:    false,
				Message:    fmt.Sprintf("Stock not initialized for SKU: %s. Product needs initial stock entry.", req.VariantSKU),
				VariantSKU: req.VariantSKU,
			}, nil
		}
		
		// Stock insuficiente → validación de negocio
		if errors.Is(err, exception.ErrInsufficientStock) {
			metrics.StockInsufficient.Inc()
			return &response.ProcessSaleResponse{
				Success:    false,
				Message:    err.Error(),
				VariantSKU: req.VariantSKU,
			}, nil
		}
		
		// Error técnico (DB connection, transacción fallida, etc.)
		return nil, fmt.Errorf("failed to process sale atomically: %w", err)
	}

	// Leer stock actualizado (post-commit del trigger)
	availability, err := uc.availabilityRepo.FindByTenantAndSKU(ctx, tenantUUID, req.VariantSKU)
	if err != nil {
		// S001: emit movement even if availability read fails
		metrics.MCStockMovementsTotal.WithLabelValues(tenantID, "sale").Inc()
		// Si la venta se procesó correctamente pero no podemos leer availability,
		// aún retornamos success pero sin remaining stock
		return &response.ProcessSaleResponse{
			Success:      true,
			Message:      "Sale processed successfully (availability read failed)",
			VariantSKU:   req.VariantSKU,
			QuantitySold: req.Quantity,
			StockEntryID: stockEntry.ID.String(),
			Timestamp:    time.Now().UTC(),
		}, nil
	}

	// S001: actualizar nivel de stock y contar movimiento
	metrics.MCStockLevel.WithLabelValues(tenantID, req.VariantSKU).Set(availability.AvailableQuantity)
	metrics.MCStockMovementsTotal.WithLabelValues(tenantID, "sale").Inc()

	return &response.ProcessSaleResponse{
		Success:        true,
		Message:        "Sale processed successfully",
		VariantSKU:     req.VariantSKU,
		QuantitySold:   req.Quantity,
		RemainingStock: availability.AvailableQuantity,
		TotalQuantity:  availability.TotalQuantity,
		StockEntryID:   stockEntry.ID.String(),
		Timestamp:      time.Now().UTC(),
	}, nil
}
