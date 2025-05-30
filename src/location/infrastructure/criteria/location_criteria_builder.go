package criteria

import (
	"net/url"

	domainCriteria "stock/src/shared/domain/criteria"
	sharedCriteria "stock/src/shared/infrastructure/criteria"

	"github.com/gin-gonic/gin"
)

// LocationCriteriaBuilder construye criterios específicos para ubicaciones
type LocationCriteriaBuilder struct {
	*domainCriteria.CriteriaBuilder
	helper *sharedCriteria.EntityCriteriaHelper
}

// NewLocationCriteriaBuilder crea un nuevo builder para criterios de ubicaciones
func NewLocationCriteriaBuilder() *LocationCriteriaBuilder {
	return &LocationCriteriaBuilder{
		CriteriaBuilder: domainCriteria.NewCriteriaBuilder(),
		helper:          sharedCriteria.NewEntityCriteriaHelper(),
	}
}

// BuildFromContext construye criterios desde el contexto de Gin con filtros específicos de ubicaciones
func (b *LocationCriteriaBuilder) BuildFromContext(c *gin.Context) *LocationCriteriaBuilder {
	// Construir criterios base desde query parameters
	b.CriteriaBuilder = b.helper.BuildBaseFromContext(c)

	// Agregar filtros específicos de ubicaciones
	b.addLocationFilters(c.Request.URL.Query())

	return b
}

// BuildValidated construye y valida criterios desde el contexto
func (b *LocationCriteriaBuilder) BuildValidated(c *gin.Context) domainCriteria.Criteria {
	criteria := b.BuildFromContext(c).Build()
	return b.helper.ValidateAndSanitizeCriteria(criteria, b.GetAllowedFields())
}

// addLocationFilters agrega filtros específicos de ubicaciones
func (b *LocationCriteriaBuilder) addLocationFilters(values url.Values) {
	// Filtro por tenant_id (obligatorio)
	if tenantID := values.Get("tenant_id"); tenantID != "" {
		b.AddUUIDFilter("tenant_id", tenantID)
	}

	// Filtros de búsqueda por texto
	if name := values.Get("name"); name != "" {
		b.AddLikeFilter("name", name)
	}

	if address := values.Get("address"); address != "" {
		b.AddLikeFilter("address", address)
	}

	if city := values.Get("city"); city != "" {
		b.AddLikeFilter("city", city)
	}

	if state := values.Get("state"); state != "" {
		b.AddLikeFilter("state", state)
	}

	// Filtros exactos
	if locationType := values.Get("type"); locationType != "" {
		b.AddEqualFilter("type", locationType)
	}

	if country := values.Get("country"); country != "" {
		b.AddEqualFilter("country", country)
	}

	if postalCode := values.Get("postal_code"); postalCode != "" {
		b.AddEqualFilter("postal_code", postalCode)
	}

	// Filtros booleanos
	if active := values.Get("active"); active != "" {
		b.AddBoolFilter("active", active)
	}

	// Filtros especiales
	if activeOnly := values.Get("active_only"); activeOnly == "true" {
		b.AddEqualFilter("active", true)
	}
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
func (b *LocationCriteriaBuilder) GetDefaultSortDirection() domainCriteria.OrderType {
	return domainCriteria.DESC
}

// Métodos de filtrado específicos

func (b *LocationCriteriaBuilder) AddNameFilter(name string) *LocationCriteriaBuilder {
	if name != "" {
		b.AddLikeFilter("name", name)
	}
	return b
}

func (b *LocationCriteriaBuilder) AddTypeFilter(locationType string) *LocationCriteriaBuilder {
	if locationType != "" {
		b.AddEqualFilter("type", locationType)
	}
	return b
}

func (b *LocationCriteriaBuilder) AddCityFilter(city string) *LocationCriteriaBuilder {
	if city != "" {
		b.AddLikeFilter("city", city)
	}
	return b
}

func (b *LocationCriteriaBuilder) AddCountryFilter(country string) *LocationCriteriaBuilder {
	if country != "" {
		b.AddEqualFilter("country", country)
	}
	return b
}

func (b *LocationCriteriaBuilder) AddActiveFilter(active string) *LocationCriteriaBuilder {
	if active != "" {
		b.AddBoolFilter("active", active)
	}
	return b
}

// FromURLValues inicializa el builder desde url.Values
func (b *LocationCriteriaBuilder) FromURLValues(values url.Values) *LocationCriteriaBuilder {
	// Construir criterios base
	b.CriteriaBuilder = b.CriteriaBuilder.FromURLValues(values)

	// Filtros específicos
	b.AddNameFilter(values.Get("name"))
	b.AddTypeFilter(values.Get("type"))
	b.AddCityFilter(values.Get("city"))
	b.AddCountryFilter(values.Get("country"))
	b.AddActiveFilter(values.Get("active"))

	return b
}
