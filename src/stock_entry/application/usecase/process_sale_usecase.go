package usecase

import (
	"context"
	"fmt"
	
	"github.com/google/uuid"
	
	"stock/src/stock_entry/application/request"
	"stock/src/stock_entry/application/response"
	"stock/src/stock_entry/domain/entity"
	"stock/src/stock_entry/domain/port"
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

// Execute ejecuta el caso de uso de venta
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
	
	// 1. Verificar disponibilidad
	availability, err := uc.availabilityRepo.FindByTenantAndSKU(ctx, tenantUUID, req.VariantSKU)
	if err != nil {
		return &response.ProcessSaleResponse{
			Success:    false,
			Message:    fmt.Sprintf("Product not found: %s", req.VariantSKU),
			VariantSKU: req.VariantSKU,
		}, nil
	}
	
	// 2. Verificar que hay stock suficiente
	if availability.AvailableQuantity < req.Quantity {
		return &response.ProcessSaleResponse{
			Success:        false,
			Message:        fmt.Sprintf("Insufficient stock. Available: %.2f, Requested: %.2f", availability.AvailableQuantity, req.Quantity),
			VariantSKU:     req.VariantSKU,
			RemainingStock: availability.AvailableQuantity,
		}, nil
	}
	
	// 3. Crear entrada de stock tipo "sale" con cantidad negativa
	stockEntry, err := entity.NewStockEntry(
		tenantUUID,
		req.VariantSKU,
		entity.EntryTypeSale,
		req.Quantity, // La cantidad se registra positiva, el método CalculatedQuantity() la convertirá a negativa
	)
	if err != nil {
		return nil, fmt.Errorf("error creating stock entry: %w", err)
	}
	
	// Agregar referencia de la venta
	refNumber := fmt.Sprintf("SALE-%s", uuid.New().String()[:8])
	stockEntry.SetReference(refNumber)
	stockEntry.SetNotes("Sale processed via minimal mock endpoint")
	
	// 4. Guardar en repositorio
	if err := uc.stockEntryRepo.Save(ctx, stockEntry); err != nil {
		return nil, fmt.Errorf("error saving stock entry: %w", err)
	}
	
	// 5. Calcular stock restante
	remainingStock := availability.AvailableQuantity - req.Quantity
	
	return &response.ProcessSaleResponse{
		Success:        true,
		Message:        "Sale processed successfully",
		VariantSKU:     req.VariantSKU,
		QuantitySold:   req.Quantity,
		RemainingStock: remainingStock,
		StockEntryID:   stockEntry.ID.String(),
	}, nil
}
