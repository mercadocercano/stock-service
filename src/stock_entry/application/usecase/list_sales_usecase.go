package usecase

import (
	"context"
	
	"github.com/google/uuid"
	
	"stock/src/stock_entry/application/response"
	"stock/src/stock_entry/domain/entity"
	"stock/src/stock_entry/domain/port"
)

// ListSalesUseCase caso de uso para listar ventas POS
type ListSalesUseCase struct {
	stockEntryRepository port.StockEntryRepository
}

// NewListSalesUseCase crea una nueva instancia
func NewListSalesUseCase(repo port.StockEntryRepository) *ListSalesUseCase {
	return &ListSalesUseCase{
		stockEntryRepository: repo,
	}
}

// Execute ejecuta el caso de uso
func (uc *ListSalesUseCase) Execute(ctx context.Context, tenantID string, limit, offset int) ([]response.StockEntryResponse, error) {
	// Parsear tenant ID
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}
	
	// Obtener todas las entradas del tenant
	entries, err := uc.stockEntryRepository.FindByTenant(ctx, tid, limit, offset)
	if err != nil {
		return nil, err
	}
	
	// Filtrar solo las ventas (type = sale)
	sales := make([]response.StockEntryResponse, 0)
	for _, entry := range entries {
		if entry.EntryType == entity.EntryTypeSale {
			sales = append(sales, response.StockEntryResponse{
				ID:              entry.ID.String(),
				TenantID:        entry.TenantID.String(),
				VariantSKU:      entry.VariantSKU,
				ProductSKU:      entry.ProductSKU,
				ProductName:     entry.ProductName,
				EntryType:       string(entry.EntryType),
				Quantity:        entry.Quantity,
				UnitOfMeasure:   entry.UnitOfMeasure,
				UnitCost:        entry.UnitCost,
				TotalCost:       entry.TotalCost,
				ReferenceNumber: entry.ReferenceNumber,
				Notes:           entry.Notes,
				Status:          string(entry.Status),
				IsActive:        entry.IsActive,
				CreatedAt:       entry.CreatedAt,
				UpdatedAt:       entry.UpdatedAt,
			})
		}
	}
	
	return sales, nil
}
