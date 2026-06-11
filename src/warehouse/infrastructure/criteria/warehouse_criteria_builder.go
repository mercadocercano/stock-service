package criteria

import (
	"net/url"

	"github.com/gin-gonic/gin"
	crit "github.com/hornosg/go-shared/criteria"
)

// WarehouseCriteriaBuilder construye criterios específicos para almacenes
type WarehouseCriteriaBuilder struct {
	*crit.CriteriaBuilder
	helper *crit.EntityCriteriaHelper
}

// NewWarehouseCriteriaBuilder crea un nuevo builder para criterios de almacenes
func NewWarehouseCriteriaBuilder() *WarehouseCriteriaBuilder {
	return &WarehouseCriteriaBuilder{
		CriteriaBuilder: crit.NewCriteriaBuilder(),
		helper:          crit.NewEntityCriteriaHelper(),
	}
}

// BuildFromContext construye criterios desde el contexto de Gin con filtros específicos de almacenes
func (b *WarehouseCriteriaBuilder) BuildFromContext(c *gin.Context) *WarehouseCriteriaBuilder {
	// Construir criterios base desde query parameters
	b.CriteriaBuilder = b.helper.BuildBaseFromContext(c)

	// Agregar filtros específicos de almacenes
	b.addWarehouseFilters(c.Request.URL.Query())

	return b
}

// BuildValidated construye y valida criterios desde el contexto
func (b *WarehouseCriteriaBuilder) BuildValidated(c *gin.Context) crit.Criteria {
	criteria := b.BuildFromContext(c).Build()
	return b.helper.ValidateAndSanitizeCriteria(criteria, b.GetAllowedFields())
}

// addWarehouseFilters agrega filtros específicos de almacenes
func (b *WarehouseCriteriaBuilder) addWarehouseFilters(values url.Values) {
	// Filtro por tenant_id (obligatorio)
	if tenantID := values.Get("tenant_id"); tenantID != "" {
		b.AddUUIDFilter("tenant_id", tenantID)
	}

	// Filtros de búsqueda por texto
	if name := values.Get("name"); name != "" {
		b.AddLikeFilter("name", name)
	}

	if code := values.Get("code"); code != "" {
		b.AddLikeFilter("code", code)
	}

	if description := values.Get("description"); description != "" {
		b.AddLikeFilter("description", description)
	}

	// Filtros exactos
	if warehouseType := values.Get("type"); warehouseType != "" {
		b.AddEqualFilter("type", warehouseType)
	}

	if locationID := values.Get("location_id"); locationID != "" {
		b.AddUUIDFilter("location_id", locationID)
	}

	// Filtros booleanos
	if active := values.Get("active"); active != "" {
		b.AddBoolFilter("active", active)
	}

	// Filtros especiales
	if activeOnly := values.Get("active_only"); activeOnly == "true" {
		b.AddEqualFilter("active", true)
	}

	// Filtros de capacidad (si el almacén tiene campos de capacidad)
	if minCapacity := values.Get("min_capacity"); minCapacity != "" {
		b.AddFilter("capacity", crit.OpGreaterThanOrEqual, minCapacity)
	}

	if maxCapacity := values.Get("max_capacity"); maxCapacity != "" {
		b.AddFilter("capacity", crit.OpLessThanOrEqual, maxCapacity)
	}

	// Filtro de prioridad
	if priority := values.Get("priority"); priority != "" {
		b.AddEqualFilter("priority", priority)
	}
}

// GetAllowedFields retorna los campos permitidos para filtrado y ordenamiento
func (b *WarehouseCriteriaBuilder) GetAllowedFields() []string {
	return []string{
		"id",
		"tenant_id",
		"location_id",
		"name",
		"code",
		"type",
		"description",
		"priority",
		"capacity",
		"active",
		"created_at",
		"updated_at",
	}
}

// GetDefaultSortField retorna el campo de ordenamiento por defecto
func (b *WarehouseCriteriaBuilder) GetDefaultSortField() string {
	return "created_at"
}

// GetDefaultSortDirection retorna la dirección de ordenamiento por defecto
func (b *WarehouseCriteriaBuilder) GetDefaultSortDirection() crit.OrderDirection {
	return crit.OrderDesc
}

// Métodos de filtrado específicos

func (b *WarehouseCriteriaBuilder) AddNameFilter(name string) *WarehouseCriteriaBuilder {
	if name != "" {
		b.AddLikeFilter("name", name)
	}
	return b
}

func (b *WarehouseCriteriaBuilder) AddCodeFilter(code string) *WarehouseCriteriaBuilder {
	if code != "" {
		b.AddLikeFilter("code", code)
	}
	return b
}

func (b *WarehouseCriteriaBuilder) AddTypeFilter(warehouseType string) *WarehouseCriteriaBuilder {
	if warehouseType != "" {
		b.AddEqualFilter("type", warehouseType)
	}
	return b
}

func (b *WarehouseCriteriaBuilder) AddLocationIDFilter(locationID string) *WarehouseCriteriaBuilder {
	if locationID != "" {
		b.AddUUIDFilter("location_id", locationID)
	}
	return b
}

func (b *WarehouseCriteriaBuilder) AddActiveFilter(active string) *WarehouseCriteriaBuilder {
	if active != "" {
		b.AddBoolFilter("active", active)
	}
	return b
}

// FromURLValues inicializa el builder desde url.Values
func (b *WarehouseCriteriaBuilder) FromURLValues(values url.Values) *WarehouseCriteriaBuilder {
	// Construir criterios base
	b.CriteriaBuilder = b.CriteriaBuilder.FromURLValues(values)

	// Filtros
	b.AddNameFilter(values.Get("name"))
	b.AddCodeFilter(values.Get("code"))
	b.AddTypeFilter(values.Get("type"))
	b.AddLocationIDFilter(values.Get("location_id"))
	b.AddActiveFilter(values.Get("active"))

	return b
}
