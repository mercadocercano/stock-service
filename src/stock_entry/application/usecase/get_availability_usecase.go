package usecase

import (
	"context"
	"fmt"
	
	"github.com/google/uuid"
	
	"stock/src/stock_entry/application/response"
	"stock/src/stock_entry/domain/port"
)

// GetAvailabilityUseCase caso de uso para consultar disponibilidad de stock
type GetAvailabilityUseCase struct {
	availabilityRepo port.StockAvailabilityRepository
}

// NewGetAvailabilityUseCase crea una nueva instancia
func NewGetAvailabilityUseCase(availabilityRepo port.StockAvailabilityRepository) *GetAvailabilityUseCase {
	return &GetAvailabilityUseCase{
		availabilityRepo: availabilityRepo,
	}
}

// Execute consulta la disponibilidad de un producto por SKU
func (uc *GetAvailabilityUseCase) Execute(ctx context.Context, tenantID, productSKU string) (*response.StockAvailabilityResponse, error) {
	// Parsear tenant ID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}
	
	// Validar SKU
	if productSKU == "" {
		return nil, fmt.Errorf("product_sku is required")
	}
	
	// Buscar disponibilidad
	availability, err := uc.availabilityRepo.FindByTenantAndSKU(ctx, tenantUUID, productSKU)
	if err != nil {
		return nil, fmt.Errorf("error finding availability: %w", err)
	}
	
	// Convertir a response
	resp := response.FromStockAvailability(availability)
	return &resp, nil
}

// ExecuteMultiple consulta disponibilidad de múltiples productos
func (uc *GetAvailabilityUseCase) ExecuteMultiple(ctx context.Context, tenantID string, productSKUs []string) ([]response.StockAvailabilityResponse, error) {
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}
	
	results := make([]response.StockAvailabilityResponse, 0, len(productSKUs))
	
	for _, sku := range productSKUs {
		availability, err := uc.availabilityRepo.FindByTenantAndSKU(ctx, tenantUUID, sku)
		if err != nil {
			// Si no encuentra, continuar con el siguiente
			continue
		}
		
		results = append(results, response.FromStockAvailability(availability))
	}
	
	return results, nil
}

