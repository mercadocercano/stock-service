package mother

import (
	"stock/src/warehouse/domain/entity"
	"time"

	"github.com/google/uuid"
)

// WarehouseMother es un factory para crear entidades Warehouse para pruebas
type WarehouseMother struct{}

// Random crea un Warehouse con datos aleatorios
func (WarehouseMother) Random() *entity.Warehouse {
	return &entity.Warehouse{
		ID:          uuid.New().String(),
		TenantID:    "tenant-" + uuid.New().String(),
		LocationID:  "location-" + uuid.New().String(),
		Name:        "Test Warehouse",
		Code:        "WH-TEST",
		Type:        entity.RegularWarehouseType,
		Description: "Test warehouse description",
		Priority:    1,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// WithID crea un Warehouse con un ID específico
func (m WarehouseMother) WithID(id string) *entity.Warehouse {
	warehouse := m.Random()
	warehouse.ID = id
	return warehouse
}

// WithTenantID crea un Warehouse con un TenantID específico
func (m WarehouseMother) WithTenantID(tenantID string) *entity.Warehouse {
	warehouse := m.Random()
	warehouse.TenantID = tenantID
	return warehouse
}

// WithLocationID crea un Warehouse con un LocationID específico
func (m WarehouseMother) WithLocationID(locationID string) *entity.Warehouse {
	warehouse := m.Random()
	warehouse.LocationID = locationID
	return warehouse
}

// WithName crea un Warehouse con un nombre específico
func (m WarehouseMother) WithName(name string) *entity.Warehouse {
	warehouse := m.Random()
	warehouse.Name = name
	return warehouse
}

// WithType crea un Warehouse con un tipo específico
func (m WarehouseMother) WithType(warehouseType entity.WarehouseType) *entity.Warehouse {
	warehouse := m.Random()
	warehouse.Type = warehouseType
	return warehouse
}

// RegularType crea un Warehouse de tipo regular
func (m WarehouseMother) RegularType() *entity.Warehouse {
	warehouse := m.Random()
	warehouse.Type = entity.RegularWarehouseType
	return warehouse
}

// SpecialType crea un Warehouse de tipo especial
func (m WarehouseMother) SpecialType() *entity.Warehouse {
	warehouse := m.Random()
	warehouse.Type = entity.SpecialWarehouseType
	return warehouse
}

// VirtualType crea un Warehouse de tipo virtual
func (m WarehouseMother) VirtualType() *entity.Warehouse {
	warehouse := m.Random()
	warehouse.Type = entity.VirtualWarehouseType
	return warehouse
}

// Inactive crea un Warehouse inactivo
func (m WarehouseMother) Inactive() *entity.Warehouse {
	warehouse := m.Random()
	warehouse.Active = false
	return warehouse
}

// WithPriority crea un Warehouse con una prioridad específica
func (m WarehouseMother) WithPriority(priority int) *entity.Warehouse {
	warehouse := m.Random()
	warehouse.Priority = priority
	return warehouse
}

// Complete crea un Warehouse con todos los datos personalizados
func (m WarehouseMother) Complete(
	id string,
	tenantID string,
	locationID string,
	name string,
	code string,
	warehouseType entity.WarehouseType,
	description string,
	priority int,
	active bool,
	createdAt time.Time,
	updatedAt time.Time,
) *entity.Warehouse {
	return &entity.Warehouse{
		ID:          id,
		TenantID:    tenantID,
		LocationID:  locationID,
		Name:        name,
		Code:        code,
		Type:        warehouseType,
		Description: description,
		Priority:    priority,
		Active:      active,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
