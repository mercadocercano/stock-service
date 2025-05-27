package criteria

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"stock/src/shared/domain/criteria"
)

// WarehouseCriteriaBuilder construye criterios específicos para almacenes
type WarehouseCriteriaBuilder struct {
	filters    []criteria.Filter
	orderField string
	orderDir   criteria.OrderType
	limit      *int
	offset     *int
}

// NewWarehouseCriteriaBuilder crea una nueva instancia del builder
func NewWarehouseCriteriaBuilder() *WarehouseCriteriaBuilder {
	return &WarehouseCriteriaBuilder{
		filters:    make([]criteria.Filter, 0),
		orderField: "created_at",
		orderDir:   criteria.DESC,
	}
}

// FromContext construye criterios desde el contexto de Gin
func (b *WarehouseCriteriaBuilder) FromContext(c *gin.Context) *WarehouseCriteriaBuilder {
	// Paginación
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	limit := pageSize

	b.limit = &limit
	b.offset = &offset

	// Ordenamiento
	orderField := c.DefaultQuery("order_by", "created_at")
	orderType := c.DefaultQuery("order_type", "desc")

	if orderField != "" {
		b.orderField = orderField
	}

	if strings.ToLower(orderType) == "asc" {
		b.orderDir = criteria.ASC
	} else {
		b.orderDir = criteria.DESC
	}

	// Filtros específicos de almacenes
	b.AddNameFilter(c.Query("name"))
	b.AddCodeFilter(c.Query("code"))
	b.AddTypeFilter(c.Query("type"))
	b.AddLocationIDFilter(c.Query("location_id"))
	b.AddActiveFilter(c.Query("active"))

	return b
}

// BuildValidated construye y valida los criterios
func (b *WarehouseCriteriaBuilder) BuildValidated(c *gin.Context) criteria.Criteria {
	return b.FromContext(c).Build()
}

// Build construye los criterios sin validación adicional
func (b *WarehouseCriteriaBuilder) Build() criteria.Criteria {
	filters := criteria.NewFilters(b.filters...)
	order := criteria.NewOrder(b.orderField, b.orderDir)

	return criteria.NewCriteria(filters, order, b.limit, b.offset)
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
func (b *WarehouseCriteriaBuilder) GetDefaultSortDirection() criteria.OrderType {
	return criteria.DESC
}

// Métodos de filtrado específicos

func (b *WarehouseCriteriaBuilder) AddNameFilter(name string) *WarehouseCriteriaBuilder {
	if name != "" {
		b.filters = append(b.filters, criteria.NewFilter("name", "LIKE", "%"+name+"%"))
	}
	return b
}

func (b *WarehouseCriteriaBuilder) AddCodeFilter(code string) *WarehouseCriteriaBuilder {
	if code != "" {
		b.filters = append(b.filters, criteria.NewFilter("code", "LIKE", "%"+code+"%"))
	}
	return b
}

func (b *WarehouseCriteriaBuilder) AddTypeFilter(warehouseType string) *WarehouseCriteriaBuilder {
	if warehouseType != "" {
		b.filters = append(b.filters, criteria.NewFilter("type", "=", warehouseType))
	}
	return b
}

func (b *WarehouseCriteriaBuilder) AddLocationIDFilter(locationID string) *WarehouseCriteriaBuilder {
	if locationID != "" {
		b.filters = append(b.filters, criteria.NewFilter("location_id", "=", locationID))
	}
	return b
}

func (b *WarehouseCriteriaBuilder) AddActiveFilter(active string) *WarehouseCriteriaBuilder {
	if active != "" {
		isActive := active == "true"
		b.filters = append(b.filters, criteria.NewFilter("active", "=", isActive))
	}
	return b
}

// FromURLValues inicializa el builder desde url.Values
func (b *WarehouseCriteriaBuilder) FromURLValues(values url.Values) *WarehouseCriteriaBuilder {
	// Paginación
	if page := values.Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			pageSize := 10
			if pageSizeStr := values.Get("page_size"); pageSizeStr != "" {
				if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
					pageSize = ps
				}
			}

			offset := (p - 1) * pageSize
			limit := pageSize

			b.limit = &limit
			b.offset = &offset
		}
	}

	// Ordenamiento
	if sortBy := values.Get("order_by"); sortBy != "" {
		b.orderField = sortBy
	}

	if sortDir := values.Get("order_type"); sortDir != "" {
		if strings.ToLower(sortDir) == "asc" {
			b.orderDir = criteria.ASC
		} else {
			b.orderDir = criteria.DESC
		}
	}

	// Filtros
	b.AddNameFilter(values.Get("name"))
	b.AddCodeFilter(values.Get("code"))
	b.AddTypeFilter(values.Get("type"))
	b.AddLocationIDFilter(values.Get("location_id"))
	b.AddActiveFilter(values.Get("active"))

	return b
} 