package criteria

import (
	"net/url"

	"github.com/gin-gonic/gin"
	crit "github.com/mercadocercano/criteria"
)

// StockLocationCriteriaBuilder construye criterios específicos para ubicaciones de stock
type StockLocationCriteriaBuilder struct {
	*crit.CriteriaBuilder
	helper *crit.EntityCriteriaHelper
}

// NewStockLocationCriteriaBuilder crea un nuevo builder para criterios de ubicaciones de stock
func NewStockLocationCriteriaBuilder() *StockLocationCriteriaBuilder {
	return &StockLocationCriteriaBuilder{
		CriteriaBuilder: crit.NewCriteriaBuilder(),
		helper:          crit.NewEntityCriteriaHelper(),
	}
}

// BuildFromContext construye criterios desde el contexto de Gin con filtros específicos de ubicaciones de stock
func (b *StockLocationCriteriaBuilder) BuildFromContext(c *gin.Context) *StockLocationCriteriaBuilder {
	// Construir criterios base desde query parameters
	b.CriteriaBuilder = b.helper.BuildBaseFromContext(c)

	// Agregar filtros específicos de ubicaciones de stock
	b.addStockLocationFilters(c.Request.URL.Query())

	return b
}

// BuildValidated construye y valida criterios desde el contexto
func (b *StockLocationCriteriaBuilder) BuildValidated(c *gin.Context) crit.Criteria {
	criteria := b.BuildFromContext(c).Build()
	return b.helper.ValidateAndSanitizeCriteria(criteria, b.GetAllowedFields())
}

// addStockLocationFilters agrega filtros específicos de ubicaciones de stock
func (b *StockLocationCriteriaBuilder) addStockLocationFilters(values url.Values) {
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

	// Filtros por relaciones
	if warehouseID := values.Get("warehouse_id"); warehouseID != "" {
		b.AddUUIDFilter("warehouse_id", warehouseID)
	}

	if parentID := values.Get("parent_id"); parentID != "" {
		b.AddUUIDFilter("parent_id", parentID)
	}

	// Filtros exactos
	if locationType := values.Get("type"); locationType != "" {
		b.AddEqualFilter("type", locationType)
	}

	// Filtros booleanos
	if active := values.Get("active"); active != "" {
		b.AddBoolFilter("active", active)
	}

	// Filtros especiales
	if activeOnly := values.Get("active_only"); activeOnly == "true" {
		b.AddEqualFilter("active", true)
	}

	// Filtro por capacidad
	if minCapacity := values.Get("min_capacity"); minCapacity != "" {
		b.AddFilter("capacity", crit.OpGreaterThanOrEqual, minCapacity)
	}

	if maxCapacity := values.Get("max_capacity"); maxCapacity != "" {
		b.AddFilter("capacity", crit.OpLessThanOrEqual, maxCapacity)
	}
}

// GetAllowedFields retorna los campos permitidos para filtrado y ordenamiento
func (b *StockLocationCriteriaBuilder) GetAllowedFields() []string {
	return []string{
		"id",
		"tenant_id",
		"warehouse_id",
		"parent_id",
		"name",
		"code",
		"type",
		"description",
		"capacity",
		"active",
		"created_at",
		"updated_at",
	}
}

// GetDefaultSortField retorna el campo de ordenamiento por defecto
func (b *StockLocationCriteriaBuilder) GetDefaultSortField() string {
	return "name"
}

// GetDefaultSortDirection retorna la dirección de ordenamiento por defecto
func (b *StockLocationCriteriaBuilder) GetDefaultSortDirection() crit.OrderDirection {
	return crit.OrderAsc
}

// Métodos de filtrado específicos

func (b *StockLocationCriteriaBuilder) AddNameFilter(name string) *StockLocationCriteriaBuilder {
	if name != "" {
		b.AddLikeFilter("name", name)
	}
	return b
}

func (b *StockLocationCriteriaBuilder) AddCodeFilter(code string) *StockLocationCriteriaBuilder {
	if code != "" {
		b.AddLikeFilter("code", code)
	}
	return b
}

func (b *StockLocationCriteriaBuilder) AddActiveFilter(active string) *StockLocationCriteriaBuilder {
	if active != "" {
		b.AddBoolFilter("active", active)
	}
	return b
}

func (b *StockLocationCriteriaBuilder) AddWarehouseFilter(warehouseID string) *StockLocationCriteriaBuilder {
	if warehouseID != "" {
		b.AddUUIDFilter("warehouse_id", warehouseID)
	}
	return b
}

func (b *StockLocationCriteriaBuilder) AddParentFilter(parentID string) *StockLocationCriteriaBuilder {
	if parentID != "" {
		b.AddUUIDFilter("parent_id", parentID)
	}
	return b
}

// FromURLValues inicializa el builder desde url.Values
func (b *StockLocationCriteriaBuilder) FromURLValues(values url.Values) *StockLocationCriteriaBuilder {
	// Construir criterios base
	b.CriteriaBuilder = b.CriteriaBuilder.FromURLValues(values)

	// Filtros específicos
	b.AddNameFilter(values.Get("name"))
	b.AddCodeFilter(values.Get("code"))
	b.AddActiveFilter(values.Get("active"))
	b.AddWarehouseFilter(values.Get("warehouse_id"))
	b.AddParentFilter(values.Get("parent_id"))

	return b
}
