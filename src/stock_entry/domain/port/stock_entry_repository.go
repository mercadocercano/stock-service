package port

import (
	"context"
	
	"github.com/google/uuid"
	
	"stock-service/src/stock_entry/domain/entity"
)

// StockEntryRepository define las operaciones del repositorio de entradas de stock
type StockEntryRepository interface {
	// Save guarda una entrada de stock
	Save(ctx context.Context, entry *entity.StockEntry) error
	
	// SaveBulk guarda múltiples entradas de stock
	SaveBulk(ctx context.Context, entries []*entity.StockEntry) error
	
	// FindByID busca una entrada por ID
	FindByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*entity.StockEntry, error)
	
	// FindByTenantAndSKU busca entradas por tenant y SKU
	FindByTenantAndSKU(ctx context.Context, tenantID uuid.UUID, productSKU string) ([]*entity.StockEntry, error)
	
	// FindByTenant busca entradas por tenant con paginación
	FindByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.StockEntry, error)
	
	// Delete elimina una entrada de stock (soft delete)
	Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
}

// StockAvailabilityRepository define las operaciones del repositorio de disponibilidad
type StockAvailabilityRepository interface {
	// FindByTenantAndSKU busca disponibilidad por tenant y SKU
	FindByTenantAndSKU(ctx context.Context, tenantID uuid.UUID, productSKU string) (*entity.StockAvailability, error)
	
	// FindByTenant busca disponibilidad por tenant
	FindByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.StockAvailability, error)
	
	// FindLowStock busca productos con bajo stock
	FindLowStock(ctx context.Context, tenantID uuid.UUID) ([]*entity.StockAvailability, error)
	
	// FindOutOfStock busca productos sin stock
	FindOutOfStock(ctx context.Context, tenantID uuid.UUID) ([]*entity.StockAvailability, error)
	
	// Save guarda o actualiza disponibilidad
	Save(ctx context.Context, availability *entity.StockAvailability) error
	
	// Update actualiza disponibilidad existente
	Update(ctx context.Context, availability *entity.StockAvailability) error
}

