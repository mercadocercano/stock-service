package criteria

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"stock/src/shared/domain/criteria"
)

// LocationCriteriaBuilder construye criterios específicos para ubicaciones
type LocationCriteriaBuilder struct {
	filters    []criteria.Filter
	orderField string
	orderDir   criteria.OrderType
	limit      *int
	offset     *int
}

// NewLocationCriteriaBuilder crea una nueva instancia del builder
func NewLocationCriteriaBuilder() *LocationCriteriaBuilder {
	return &LocationCriteriaBuilder{
		filters:    make([]criteria.Filter, 0),
		orderField: "created_at",
		orderDir:   criteria.DESC,
	}
}

// FromContext construye criterios desde el contexto de Gin
func (b *LocationCriteriaBuilder) FromContext(c *gin.Context) *LocationCriteriaBuilder {
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

	// Filtros específicos de ubicaciones
	b.AddNameFilter(c.Query("name"))
	b.AddTypeFilter(c.Query("type"))
	b.AddCityFilter(c.Query("city"))
	b.AddCountryFilter(c.Query("country"))
	b.AddActiveFilter(c.Query("active"))

	return b
}

// BuildValidated construye y valida los criterios
func (b *LocationCriteriaBuilder) BuildValidated(c *gin.Context) criteria.Criteria {
	return b.FromContext(c).Build()
}

// Build construye los criterios sin validación adicional
func (b *LocationCriteriaBuilder) Build() criteria.Criteria {
	filters := criteria.NewFilters(b.filters...)
	order := criteria.NewOrder(b.orderField, b.orderDir)

	return criteria.NewCriteria(filters, order, b.limit, b.offset)
}

// GetAllowedFields retorna los campos permitidos para filtrado y ordenamiento
func (b *LocationCriteriaBuilder) GetAllowedFields() []string {
	return []string{
		"id",
		"tenant_id",
		"name",
		"type",
		"address",
		"city",
		"state",
		"country",
		"postal_code",
		"active",
		"created_at",
		"updated_at",
	}
}

// GetDefaultSortField retorna el campo de ordenamiento por defecto
func (b *LocationCriteriaBuilder) GetDefaultSortField() string {
	return "created_at"
}

// GetDefaultSortDirection retorna la dirección de ordenamiento por defecto
func (b *LocationCriteriaBuilder) GetDefaultSortDirection() criteria.OrderType {
	return criteria.DESC
}

// Métodos de filtrado específicos

func (b *LocationCriteriaBuilder) AddNameFilter(name string) *LocationCriteriaBuilder {
	if name != "" {
		b.filters = append(b.filters, criteria.NewFilter("name", "LIKE", "%"+name+"%"))
	}
	return b
}

func (b *LocationCriteriaBuilder) AddTypeFilter(locationType string) *LocationCriteriaBuilder {
	if locationType != "" {
		b.filters = append(b.filters, criteria.NewFilter("type", "=", locationType))
	}
	return b
}

func (b *LocationCriteriaBuilder) AddCityFilter(city string) *LocationCriteriaBuilder {
	if city != "" {
		b.filters = append(b.filters, criteria.NewFilter("city", "LIKE", "%"+city+"%"))
	}
	return b
}

func (b *LocationCriteriaBuilder) AddCountryFilter(country string) *LocationCriteriaBuilder {
	if country != "" {
		b.filters = append(b.filters, criteria.NewFilter("country", "=", country))
	}
	return b
}

func (b *LocationCriteriaBuilder) AddActiveFilter(active string) *LocationCriteriaBuilder {
	if active != "" {
		isActive := active == "true"
		b.filters = append(b.filters, criteria.NewFilter("active", "=", isActive))
	}
	return b
}

// FromURLValues inicializa el builder desde url.Values
func (b *LocationCriteriaBuilder) FromURLValues(values url.Values) *LocationCriteriaBuilder {
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
	b.AddTypeFilter(values.Get("type"))
	b.AddCityFilter(values.Get("city"))
	b.AddCountryFilter(values.Get("country"))
	b.AddActiveFilter(values.Get("active"))

	return b
}
