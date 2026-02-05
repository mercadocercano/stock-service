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

// CreateStockEntryUseCase caso de uso para crear una entrada de stock
type CreateStockEntryUseCase struct {
	stockEntryRepo port.StockEntryRepository
}

// NewCreateStockEntryUseCase crea una nueva instancia del caso de uso
func NewCreateStockEntryUseCase(stockEntryRepo port.StockEntryRepository) *CreateStockEntryUseCase {
	return &CreateStockEntryUseCase{
		stockEntryRepo: stockEntryRepo,
	}
}

// Execute ejecuta el caso de uso
func (uc *CreateStockEntryUseCase) Execute(ctx context.Context, req request.CreateStockEntryRequest) (*response.StockEntryResponse, error) {
	// Validar request
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// Parsear tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}
	
	// Crear entidad
	stockEntry, err := entity.NewStockEntry(
		tenantID,
		req.ProductSKU,
		entity.EntryType(req.EntryType),
		req.Quantity,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating stock entry: %w", err)
	}
	
	// Establecer información adicional
	if req.ProductName != "" {
		productID, _ := req.ParseProductID()
		stockEntry.SetProductInfo(productID, req.ProductName)
	}
	
	if req.LocationID != "" {
		locationID, err := req.ParseLocationID()
		if err != nil {
			return nil, fmt.Errorf("invalid location_id: %w", err)
		}
		if locationID != nil {
			stockEntry.SetLocation(*locationID)
		}
	}
	
	if req.UnitCost > 0 {
		totalCost := req.UnitCost * req.Quantity
		stockEntry.SetCosts(req.UnitCost, totalCost)
	}
	
	if req.UnitOfMeasure != "" {
		stockEntry.UnitOfMeasure = req.UnitOfMeasure
	}
	
	if req.ReferenceNumber != "" {
		stockEntry.SetReference(req.ReferenceNumber)
	}
	
	if req.Notes != "" {
		stockEntry.SetNotes(req.Notes)
	}
	
	// Guardar en repositorio
	if err := uc.stockEntryRepo.Save(ctx, stockEntry); err != nil {
		return nil, fmt.Errorf("error saving stock entry: %w", err)
	}
	
	// Convertir a response
	resp := response.FromStockEntry(stockEntry)
	return &resp, nil
}

