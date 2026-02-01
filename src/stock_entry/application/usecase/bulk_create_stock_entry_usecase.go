package usecase

import (
	"context"
	"fmt"
	
	"github.com/google/uuid"
	
	"stock-service/src/stock_entry/application/request"
	"stock-service/src/stock_entry/application/response"
	"stock-service/src/stock_entry/domain/entity"
	"stock-service/src/stock_entry/domain/port"
)

// BulkCreateStockEntryUseCase caso de uso para crear múltiples entradas de stock
type BulkCreateStockEntryUseCase struct {
	stockEntryRepo port.StockEntryRepository
}

// NewBulkCreateStockEntryUseCase crea una nueva instancia
func NewBulkCreateStockEntryUseCase(stockEntryRepo port.StockEntryRepository) *BulkCreateStockEntryUseCase {
	return &BulkCreateStockEntryUseCase{
		stockEntryRepo: stockEntryRepo,
	}
}

// Execute ejecuta la creación masiva
func (uc *BulkCreateStockEntryUseCase) Execute(ctx context.Context, req request.BulkCreateStockEntriesRequest) (*response.BulkCreateResponse, error) {
	// Parsear tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}
	
	resp := &response.BulkCreateResponse{
		Success:        true,
		TotalEntries:   len(req.Entries),
		EntriesCreated: 0,
		EntriesFailed:  0,
		Errors:         make([]string, 0),
		CreatedEntries: make([]response.StockEntryResponse, 0),
	}
	
	entries := make([]*entity.StockEntry, 0, len(req.Entries))
	
	// Procesar cada entrada
	for idx, entryReq := range req.Entries {
		// Asegurar que tenga el tenant ID
		entryReq.TenantID = req.TenantID
		
		// Crear entidad
		stockEntry, err := uc.createEntryFromRequest(tenantID, entryReq)
		if err != nil {
			resp.EntriesFailed++
			resp.Errors = append(resp.Errors, fmt.Sprintf(
				"Entry %d (SKU: %s): %v",
				idx+1,
				entryReq.ProductSKU,
				err,
			))
			continue
		}
		
		entries = append(entries, stockEntry)
	}
	
	// Guardar todas las entradas exitosas
	if len(entries) > 0 {
		if err := uc.stockEntryRepo.SaveBulk(ctx, entries); err != nil {
			return nil, fmt.Errorf("error saving stock entries: %w", err)
		}
		
		resp.EntriesCreated = len(entries)
		
		// Convertir a responses
		for _, entry := range entries {
			resp.CreatedEntries = append(resp.CreatedEntries, response.FromStockEntry(entry))
		}
	}
	
	// Si todos fallaron, marcar como no exitoso
	if resp.EntriesCreated == 0 {
		resp.Success = false
	}
	
	return resp, nil
}

// createEntryFromRequest crea una entrada desde el request
func (uc *BulkCreateStockEntryUseCase) createEntryFromRequest(tenantID uuid.UUID, req request.CreateStockEntryRequest) (*entity.StockEntry, error) {
	// Validar
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	// Crear entidad
	stockEntry, err := entity.NewStockEntry(
		tenantID,
		req.ProductSKU,
		entity.EntryType(req.EntryType),
		req.Quantity,
	)
	if err != nil {
		return nil, err
	}
	
	// Configurar campos adicionales
	if req.ProductName != "" {
		productID, _ := req.ParseProductID()
		stockEntry.SetProductInfo(productID, req.ProductName)
	}
	
	if req.LocationID != "" {
		locationID, _ := req.ParseLocationID()
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
	
	return stockEntry, nil
}

