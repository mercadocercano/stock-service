package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"stock/src/stock_entry/application/request"
	"stock/src/stock_entry/application/response"
	"stock/src/stock_entry/domain/entity"
	"stock/src/stock_entry/domain/exception"
	"stock/src/stock_entry/domain/port"
)

// RevertConsumeUseCase caso de uso para revertir un consumo de stock (cancelación de orden)
type RevertConsumeUseCase struct {
	availabilityRepo port.StockAvailabilityRepository
	stockEntryRepo   port.StockEntryRepository
}

// NewRevertConsumeUseCase crea una nueva instancia
func NewRevertConsumeUseCase(
	availabilityRepo port.StockAvailabilityRepository,
	stockEntryRepo port.StockEntryRepository,
) *RevertConsumeUseCase {
	return &RevertConsumeUseCase{
		availabilityRepo: availabilityRepo,
		stockEntryRepo:   stockEntryRepo,
	}
}

// Execute ejecuta la reversión del consumo de stock
func (uc *RevertConsumeUseCase) Execute(ctx context.Context, tenantID string, req *request.RevertConsumeRequest) (*response.RevertConsumeResponse, error) {
	// Parsear tenant ID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	// Buscar disponibilidad actual
	availability, err := uc.availabilityRepo.FindByTenantAndSKU(ctx, tenantUUID, req.SKU)
	if err != nil {
		if err == exception.ErrStockAvailabilityNotFound {
			return nil, fmt.Errorf("stock not found for SKU %s", req.SKU)
		}
		return nil, fmt.Errorf("error finding availability: %w", err)
	}

	// Revertir consumo: aumentar available_quantity
	// NO modificar reserved_quantity
	availability.AvailableQuantity += float64(req.Quantity)
	availability.UpdatedAt = time.Now()

	// Actualizar disponibilidad
	if err := uc.availabilityRepo.Update(ctx, availability); err != nil {
		return nil, fmt.Errorf("error updating availability: %w", err)
	}

	// Registrar movimiento de reversión (devolución)
	reference := req.Reference
	notes := fmt.Sprintf("Stock reverted (order canceled): %s", req.Reference)
	stockEntry := &entity.StockEntry{
		ID:              uuid.New(),
		TenantID:        tenantUUID,
		VariantSKU:      req.SKU,
		ProductSKU:      req.SKU,
		ProductID:       availability.ProductID,
		ProductName:     availability.ProductName,
		LocationID:      availability.LocationID,
		EntryType:       entity.EntryTypeReturn, // Tipo "return" para reversión
		Quantity:        float64(req.Quantity),  // Positivo (return)
		UnitOfMeasure:   availability.UnitOfMeasure,
		ReferenceNumber: &reference,
		Notes:           &notes,
		Status:          entity.EntryStatusConfirmed,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := uc.stockEntryRepo.Save(ctx, stockEntry); err != nil {
		return nil, fmt.Errorf("error creating stock entry: %w", err)
	}

	// Construir respuesta
	return &response.RevertConsumeResponse{
		SKU:          req.SKU,
		RevertedQty:  req.Quantity,
		AvailableQty: int(availability.AvailableQuantity),
		Reference:    req.Reference,
	}, nil
}
