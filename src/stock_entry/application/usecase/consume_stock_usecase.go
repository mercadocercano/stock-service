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

// ConsumeStockUseCase caso de uso para consumir stock reservado (confirmación de orden)
type ConsumeStockUseCase struct {
	availabilityRepo port.StockAvailabilityRepository
	stockEntryRepo   port.StockEntryRepository
}

// NewConsumeStockUseCase crea una nueva instancia
func NewConsumeStockUseCase(
	availabilityRepo port.StockAvailabilityRepository,
	stockEntryRepo port.StockEntryRepository,
) *ConsumeStockUseCase {
	return &ConsumeStockUseCase{
		availabilityRepo: availabilityRepo,
		stockEntryRepo:   stockEntryRepo,
	}
}

// Execute ejecuta el consumo de stock reservado
func (uc *ConsumeStockUseCase) Execute(ctx context.Context, tenantID string, req *request.ConsumeStockRequest) (*response.ConsumeStockResponse, error) {
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

	// Verificar si hay suficiente stock reservado
	if availability.ReservedQuantity < float64(req.Quantity) {
		return nil, fmt.Errorf("insufficient reserved stock: have %.0f, requested %d", availability.ReservedQuantity, req.Quantity)
	}

	// Consumir reserva (solo disminuir reserved, NO aumentar available)
	availability.ReservedQuantity -= float64(req.Quantity)
	availability.UpdatedAt = time.Now()

	// Actualizar disponibilidad
	if err := uc.availabilityRepo.Update(ctx, availability); err != nil {
		return nil, fmt.Errorf("error updating availability: %w", err)
	}

	// Registrar movimiento de consumo (venta)
	reference := req.Reference
	notes := fmt.Sprintf("Stock consumed (order confirmed): %s", req.Reference)
	stockEntry := &entity.StockEntry{
		ID:              uuid.New(),
		TenantID:        tenantUUID,
		VariantSKU:      req.SKU,
		ProductSKU:      req.SKU,
		ProductID:       availability.ProductID,
		ProductName:     availability.ProductName,
		LocationID:      availability.LocationID,
		EntryType:       entity.EntryTypeSale, // Tipo "sale" para consumo definitivo
		Quantity:        -float64(req.Quantity), // Negativo (sale)
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
	return &response.ConsumeStockResponse{
		SKU:         req.SKU,
		ConsumedQty: req.Quantity,
		ReservedQty: int(availability.ReservedQuantity),
		Reference:   req.Reference,
	}, nil
}
