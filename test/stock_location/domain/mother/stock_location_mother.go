package mother

import (
	"stock/src/stock_location/domain/entity"
	"time"

	"github.com/google/uuid"
)

// StockLocationMother es un factory para crear entidades StockLocation para pruebas
type StockLocationMother struct{}

// Random crea un StockLocation con datos aleatorios
func (StockLocationMother) Random() *entity.StockLocation {
	return &entity.StockLocation{
		ID:          uuid.New().String(),
		TenantID:    "tenant-" + uuid.New().String(),
		WarehouseID: "warehouse-" + uuid.New().String(),
		ParentID:    nil,
		Name:        "Test Stock Location",
		Code:        "SL-TEST",
		Path:        uuid.New().String(),
		Level:       1,
		Description: "Test stock location description",
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// WithID crea un StockLocation con un ID específico
func (m StockLocationMother) WithID(id string) *entity.StockLocation {
	stockLocation := m.Random()
	stockLocation.ID = id
	return stockLocation
}

// WithTenantID crea un StockLocation con un TenantID específico
func (m StockLocationMother) WithTenantID(tenantID string) *entity.StockLocation {
	stockLocation := m.Random()
	stockLocation.TenantID = tenantID
	return stockLocation
}

// WithWarehouseID crea un StockLocation con un WarehouseID específico
func (m StockLocationMother) WithWarehouseID(warehouseID string) *entity.StockLocation {
	stockLocation := m.Random()
	stockLocation.WarehouseID = warehouseID
	return stockLocation
}

// WithParentID crea un StockLocation con un ParentID específico
func (m StockLocationMother) WithParentID(parentID string) *entity.StockLocation {
	stockLocation := m.Random()
	stockLocation.ParentID = &parentID
	stockLocation.Path = parentID + "/" + stockLocation.ID
	stockLocation.Level = 2
	return stockLocation
}

// WithName crea un StockLocation con un nombre específico
func (m StockLocationMother) WithName(name string) *entity.StockLocation {
	stockLocation := m.Random()
	stockLocation.Name = name
	return stockLocation
}

// WithCode crea un StockLocation con un código específico
func (m StockLocationMother) WithCode(code string) *entity.StockLocation {
	stockLocation := m.Random()
	stockLocation.Code = code
	return stockLocation
}

// WithLevel crea un StockLocation con un nivel específico
func (m StockLocationMother) WithLevel(level int) *entity.StockLocation {
	stockLocation := m.Random()
	stockLocation.Level = level
	return stockLocation
}

// Root crea un StockLocation de nivel raíz
func (m StockLocationMother) Root() *entity.StockLocation {
	stockLocation := m.Random()
	stockLocation.ParentID = nil
	stockLocation.Path = stockLocation.ID
	stockLocation.Level = 1
	return stockLocation
}

// Child crea un StockLocation hijo de otro
func (m StockLocationMother) Child(parent *entity.StockLocation) *entity.StockLocation {
	stockLocation := m.Random()
	stockLocation.ParentID = &parent.ID
	stockLocation.WarehouseID = parent.WarehouseID
	stockLocation.Path = parent.Path + "/" + stockLocation.ID
	stockLocation.Level = parent.Level + 1
	return stockLocation
}

// Inactive crea un StockLocation inactivo
func (m StockLocationMother) Inactive() *entity.StockLocation {
	stockLocation := m.Random()
	stockLocation.Active = false
	return stockLocation
}

// Complete crea un StockLocation con todos los datos personalizados
func (m StockLocationMother) Complete(
	id string,
	tenantID string,
	warehouseID string,
	parentID *string,
	name string,
	code string,
	path string,
	level int,
	description string,
	active bool,
	createdAt time.Time,
	updatedAt time.Time,
) *entity.StockLocation {
	return &entity.StockLocation{
		ID:          id,
		TenantID:    tenantID,
		WarehouseID: warehouseID,
		ParentID:    parentID,
		Name:        name,
		Code:        code,
		Path:        path,
		Level:       level,
		Description: description,
		Active:      active,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
