package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"stock/src/stock_entry/application/request"
	"stock/src/stock_entry/application/response"
	"stock/src/stock_entry/domain/exception"
	"stock/src/stock_entry/domain/port"
)

// ReserveStockUseCase caso de uso para reservar stock
type ReserveStockUseCase struct {
	availabilityRepo port.StockAvailabilityRepository
	stockEntryRepo   port.StockEntryRepository
}

// NewReserveStockUseCase crea una nueva instancia
func NewReserveStockUseCase(
	availabilityRepo port.StockAvailabilityRepository,
	stockEntryRepo port.StockEntryRepository,
) *ReserveStockUseCase {
	return &ReserveStockUseCase{
		availabilityRepo: availabilityRepo,
		stockEntryRepo:   stockEntryRepo,
	}
}

// Execute ejecuta la reserva de stock
func (uc *ReserveStockUseCase) Execute(ctx context.Context, tenantID string, req *request.ReserveStockRequest) (*response.ReserveStockResponse, error) {
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

	// Verificar si hay stock suficiente
	if availability.AvailableQuantity < float64(req.Quantity) {
		return nil, exception.ErrInsufficientStock
	}

	// Reservar cantidad
	if err := availability.Reserve(float64(req.Quantity)); err != nil {
		return nil, err
	}

	// Actualizar disponibilidad (solo mueve de available a reserved, sin afectar total)
	if err := uc.availabilityRepo.Update(ctx, availability); err != nil {
		return nil, fmt.Errorf("error updating availability: %w", err)
	}

	// NO crear StockEntry para reservas - solo es un movimiento lógico
	// El total_quantity NO cambia, solo se mueve de available a reserved
	// StockEntry solo se crea en movimientos físicos (purchase, sale, adjustment de inventario)

	// Construir respuesta
	return &response.ReserveStockResponse{
		SKU:          req.SKU,
		ReservedQty:  req.Quantity,
		RemainingQty: int(availability.AvailableQuantity),
		Reference:    req.Reference,
	}, nil
}
