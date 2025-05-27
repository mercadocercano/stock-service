package criteria

import (
	"strconv"
	"strings"

	"stock/src/shared/domain/criteria"

	"github.com/gin-gonic/gin"
)

// StockLocationCriteriaBuilder construye criterios para buscar ubicaciones de stock
type StockLocationCriteriaBuilder struct {
	filters    []criteria.Filter
	orderField string
	orderDir   criteria.OrderType
	limit      *int
	offset     *int
	warehouse  string
	parent     string
}

// NewStockLocationCriteriaBuilder crea un nuevo builder de criterios
func NewStockLocationCriteriaBuilder() *StockLocationCriteriaBuilder {
	return &StockLocationCriteriaBuilder{
		filters:    make([]criteria.Filter, 0),
		orderField: "name",
		orderDir:   criteria.ASC,
	}
}

// FromContext obtiene los parámetros de la petición HTTP para construir el criterio
func (b *StockLocationCriteriaBuilder) FromContext(c *gin.Context) *StockLocationCriteriaBuilder {
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
	orderField := c.DefaultQuery("order_by", "name")
	orderType := c.DefaultQuery("order_type", "asc")

	if orderField != "" {
		b.orderField = orderField
	}

	if strings.ToLower(orderType) == "desc" {
		b.orderDir = criteria.DESC
	} else {
		b.orderDir = criteria.ASC
	}

	// Filtros específicos
	b.AddNameFilter(c.Query("name"))
	b.AddCodeFilter(c.Query("code"))
	b.AddActiveFilter(c.Query("active"))

	// Filtros por almacén y padre
	warehouse := c.Query("warehouse_id")
	if warehouse != "" {
		b.AddWarehouseFilter(warehouse)
		b.warehouse = warehouse
	}

	parent := c.Query("parent_id")
	if parent != "" {
		b.AddParentFilter(parent)
		b.parent = parent
	}

	return b
}

// Build construye el criterio final
func (b *StockLocationCriteriaBuilder) Build() criteria.Criteria {
	filters := criteria.NewFilters(b.filters...)
	order := criteria.NewOrder(b.orderField, b.orderDir)

	return criteria.NewCriteria(filters, order, b.limit, b.offset)
}

// BuildValidated construye el criterio con validación
func (b *StockLocationCriteriaBuilder) BuildValidated(c *gin.Context) criteria.Criteria {
	return b.FromContext(c).Build()
}

// GetWarehouseID devuelve el ID del almacén si se ha especificado
func (b *StockLocationCriteriaBuilder) GetWarehouseID() string {
	return b.warehouse
}

// GetParentID devuelve el ID del padre si se ha especificado
func (b *StockLocationCriteriaBuilder) GetParentID() string {
	return b.parent
}

// Métodos de filtrado específicos

// AddNameFilter añade un filtro por nombre
func (b *StockLocationCriteriaBuilder) AddNameFilter(name string) *StockLocationCriteriaBuilder {
	if name != "" {
		b.filters = append(b.filters, criteria.NewFilter("name", "LIKE", "%"+name+"%"))
	}
	return b
}

// AddCodeFilter añade un filtro por código
func (b *StockLocationCriteriaBuilder) AddCodeFilter(code string) *StockLocationCriteriaBuilder {
	if code != "" {
		b.filters = append(b.filters, criteria.NewFilter("code", "LIKE", "%"+code+"%"))
	}
	return b
}

// AddActiveFilter añade un filtro por estado
func (b *StockLocationCriteriaBuilder) AddActiveFilter(active string) *StockLocationCriteriaBuilder {
	if active != "" {
		isActive := active == "true"
		b.filters = append(b.filters, criteria.NewFilter("active", "=", isActive))
	}
	return b
}

// AddWarehouseFilter añade un filtro por almacén
func (b *StockLocationCriteriaBuilder) AddWarehouseFilter(warehouseID string) *StockLocationCriteriaBuilder {
	if warehouseID != "" {
		b.filters = append(b.filters, criteria.NewFilter("warehouse_id", "=", warehouseID))
		b.warehouse = warehouseID
	}
	return b
}

// AddParentFilter añade un filtro por padre
func (b *StockLocationCriteriaBuilder) AddParentFilter(parentID string) *StockLocationCriteriaBuilder {
	if parentID != "" {
		b.filters = append(b.filters, criteria.NewFilter("parent_id", "=", parentID))
		b.parent = parentID
	}
	return b
}
