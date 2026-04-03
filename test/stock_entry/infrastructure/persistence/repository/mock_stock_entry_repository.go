package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"stock/src/stock_entry/domain/entity"
)

// MockStockEntryRepository es un mock del repositorio de entradas de stock
type MockStockEntryRepository struct {
	mock.Mock
}

func (m *MockStockEntryRepository) Save(ctx context.Context, entry *entity.StockEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockStockEntryRepository) SaveBulk(ctx context.Context, entries []*entity.StockEntry) error {
	args := m.Called(ctx, entries)
	return args.Error(0)
}

func (m *MockStockEntryRepository) FindByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*entity.StockEntry, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.StockEntry), args.Error(1)
}

func (m *MockStockEntryRepository) FindByTenantAndSKU(ctx context.Context, tenantID uuid.UUID, productSKU string) ([]*entity.StockEntry, error) {
	args := m.Called(ctx, tenantID, productSKU)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.StockEntry), args.Error(1)
}

func (m *MockStockEntryRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.StockEntry, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.StockEntry), args.Error(1)
}

func (m *MockStockEntryRepository) Delete(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	args := m.Called(ctx, id, tenantID)
	return args.Error(0)
}

func (m *MockStockEntryRepository) ProcessSaleAtomic(ctx context.Context, tenantID uuid.UUID, variantSKU string, quantity float64, reference string) (*entity.StockEntry, error) {
	args := m.Called(ctx, tenantID, variantSKU, quantity, reference)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.StockEntry), args.Error(1)
}

func (m *MockStockEntryRepository) CompensateSale(ctx context.Context, tenantID uuid.UUID, stockEntryID uuid.UUID, reason string) error {
	args := m.Called(ctx, tenantID, stockEntryID, reason)
	return args.Error(0)
}

// MockStockAvailabilityRepository es un mock del repositorio de disponibilidad
type MockStockAvailabilityRepository struct {
	mock.Mock
}

func (m *MockStockAvailabilityRepository) FindByTenantAndSKU(ctx context.Context, tenantID uuid.UUID, productSKU string) (*entity.StockAvailability, error) {
	args := m.Called(ctx, tenantID, productSKU)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.StockAvailability), args.Error(1)
}

func (m *MockStockAvailabilityRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.StockAvailability, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.StockAvailability), args.Error(1)
}

func (m *MockStockAvailabilityRepository) FindLowStock(ctx context.Context, tenantID uuid.UUID) ([]*entity.StockAvailability, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.StockAvailability), args.Error(1)
}

func (m *MockStockAvailabilityRepository) FindOutOfStock(ctx context.Context, tenantID uuid.UUID) ([]*entity.StockAvailability, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.StockAvailability), args.Error(1)
}

func (m *MockStockAvailabilityRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	args := m.Called(ctx, tenantID)
	return args.Int(0), args.Error(1)
}

func (m *MockStockAvailabilityRepository) Save(ctx context.Context, availability *entity.StockAvailability) error {
	args := m.Called(ctx, availability)
	return args.Error(0)
}

func (m *MockStockAvailabilityRepository) Update(ctx context.Context, availability *entity.StockAvailability) error {
	args := m.Called(ctx, availability)
	return args.Error(0)
}
