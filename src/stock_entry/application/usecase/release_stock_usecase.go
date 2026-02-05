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

// ReleaseStockUseCase caso de uso para liberar stock reservado
type ReleaseStockUseCase struct {
	availabilityRepo port.StockAvailabilityRepository
	stockEntryRepo   port.StockEntryRepository
}

// NewReleaseStockUseCase crea una nueva instancia
func NewReleaseStockUseCase(
	availabilityRepo port.StockAvailabilityRepository,
	stockEntryRepo port.StockEntryRepository,
) *ReleaseStockUseCase {
	return &ReleaseStockUseCase{
		availabilityRepo: availabilityRepo,
		stockEntryRepo:   stockEntryRepo,
	}
}

// Execute ejecuta la liberación de stock
func (uc *ReleaseStockUseCase) Execute(ctx context.Context, tenantID string, req *request.ReleaseStockRequest) (*response.ReleaseStockResponse, error) {
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

	// Liberar cantidad
	availability.Release(float64(req.Quantity))

	// Actualizar disponibilidad (solo mueve de reserved a available, sin afectar total)
	if err := uc.availabilityRepo.Update(ctx, availability); err != nil {
		return nil, fmt.Errorf("error updating availability: %w", err)
	}

	// NO crear StockEntry para liberaciones - solo es un movimiento lógico
	// El total_quantity NO cambia, solo se mueve de reserved a available
	// StockEntry solo se crea en movimientos físicos de inventario

	// Construir respuesta
	return &response.ReleaseStockResponse{
		SKU:          req.SKU,
		ReleasedQty:  req.Quantity,
		AvailableQty: int(availability.AvailableQuantity),
		ReservedQty:  int(availability.ReservedQuantity),
		Reference:    req.Reference,
	}, nil
}
