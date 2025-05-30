package controller

import (
	"net/http"

	"stock/src/location/application/request"
	"stock/src/location/application/usecase"
	"stock/src/location/domain/exception"
	"stock/src/location/infrastructure/criteria"

	"github.com/gin-gonic/gin"
)

// LocationController maneja las peticiones HTTP relacionadas con ubicaciones
type LocationController struct {
	createLocationUseCase     *usecase.CreateLocationUseCase
	listLocationsUseCase      *usecase.ListLocationsUseCase
	getLocationUseCase        *usecase.GetLocationUseCase
	updateLocationUseCase     *usecase.UpdateLocationUseCase
	activateLocationUseCase   *usecase.ActivateLocationUseCase
	deactivateLocationUseCase *usecase.DeactivateLocationUseCase
	deleteLocationUseCase     *usecase.DeleteLocationUseCase
}

// NewLocationController crea una nueva instancia del controlador
func NewLocationController(
	createLocationUseCase *usecase.CreateLocationUseCase,
	listLocationsUseCase *usecase.ListLocationsUseCase,
	getLocationUseCase *usecase.GetLocationUseCase,
	updateLocationUseCase *usecase.UpdateLocationUseCase,
	activateLocationUseCase *usecase.ActivateLocationUseCase,
	deactivateLocationUseCase *usecase.DeactivateLocationUseCase,
	deleteLocationUseCase *usecase.DeleteLocationUseCase,
) *LocationController {
	return &LocationController{
		createLocationUseCase:     createLocationUseCase,
		listLocationsUseCase:      listLocationsUseCase,
		getLocationUseCase:        getLocationUseCase,
		updateLocationUseCase:     updateLocationUseCase,
		activateLocationUseCase:   activateLocationUseCase,
		deactivateLocationUseCase: deactivateLocationUseCase,
		deleteLocationUseCase:     deleteLocationUseCase,
	}
}

// RegisterRoutes registra las rutas del controlador en el router
func (c *LocationController) RegisterRoutes(router *gin.RouterGroup) {
	locations := router.Group("/locations")
	{
		locations.POST("", c.CreateLocation)
		locations.GET("", c.ListLocations)
		locations.GET("/:id", c.GetLocation)
		locations.PUT("/:id", c.UpdateLocation)
		locations.DELETE("/:id", c.DeleteLocation)
		locations.PATCH("/:id/activate", c.ActivateLocation)
		locations.PATCH("/:id/deactivate", c.DeactivateLocation)
		locations.GET("/stores", c.ListStores)
		locations.GET("/distribution-centers", c.ListDistributionCenters)
	}
}

// CreateLocation maneja la petición de creación de una ubicación
func (c *LocationController) CreateLocation(ctx *gin.Context) {
	var req request.CreateLocationRequest

	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}
	req.TenantID = tenantID.(string)

	// Parsear el cuerpo de la petición
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ejecutar el caso de uso
	response, err := c.createLocationUseCase.Execute(ctx, req)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.LocationNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusCreated, response)
}

// ListLocations maneja la petición para listar ubicaciones con filtros y paginación
func (c *LocationController) ListLocations(ctx *gin.Context) {
	// Obtener el tenantID del header y agregarlo a los query parameters
	tenantID := ctx.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		// Fallback: intentar obtener del contexto (middleware)
		if tenant, exists := ctx.Get("tenantID"); exists {
			tenantID = tenant.(string)
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header es requerido"})
			return
		}
	}

	// Agregar tenant_id a los query parameters para el filtrado
	query := ctx.Request.URL.Query()
	query.Set("tenant_id", tenantID)
	ctx.Request.URL.RawQuery = query.Encode()

	// Utilizar el criteria builder para construir los criterios desde la petición
	criteriaBuilder := criteria.NewLocationCriteriaBuilder()
	crit := criteriaBuilder.BuildValidated(ctx)

	// Ejecutar el caso de uso para listar ubicaciones
	response, err := c.listLocationsUseCase.Execute(ctx, tenantID, crit)

	// Manejar errores
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// GetLocation maneja la petición para obtener una ubicación por su ID
func (c *LocationController) GetLocation(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID de la ubicación de los parámetros de la URL
	locationID := ctx.Param("id")
	if locationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Location ID is required"})
		return
	}

	// Ejecutar el caso de uso para obtener una ubicación
	response, err := c.getLocationUseCase.Execute(ctx, tenantID.(string), locationID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.LocationNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// UpdateLocation maneja la petición para actualizar una ubicación
func (c *LocationController) UpdateLocation(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID de la ubicación de los parámetros de la URL
	locationID := ctx.Param("id")
	if locationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Location ID is required"})
		return
	}

	// Parsear el cuerpo de la petición
	var req request.UpdateLocationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ejecutar el caso de uso para actualizar una ubicación
	response, err := c.updateLocationUseCase.Execute(ctx, tenantID.(string), locationID, req)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.LocationNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// ActivateLocation maneja la petición para activar una ubicación
func (c *LocationController) ActivateLocation(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID de la ubicación de los parámetros de la URL
	locationID := ctx.Param("id")
	if locationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Location ID is required"})
		return
	}

	// Ejecutar el caso de uso para activar una ubicación
	response, err := c.activateLocationUseCase.Execute(ctx, tenantID.(string), locationID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.LocationNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// DeactivateLocation maneja la petición para desactivar una ubicación
func (c *LocationController) DeactivateLocation(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID de la ubicación de los parámetros de la URL
	locationID := ctx.Param("id")
	if locationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Location ID is required"})
		return
	}

	// Ejecutar el caso de uso para desactivar una ubicación
	response, err := c.deactivateLocationUseCase.Execute(ctx, tenantID.(string), locationID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.LocationNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// DeleteLocation maneja la petición para eliminar una ubicación
func (c *LocationController) DeleteLocation(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID de la ubicación de los parámetros de la URL
	locationID := ctx.Param("id")
	if locationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Location ID is required"})
		return
	}

	// Ejecutar el caso de uso para eliminar una ubicación
	err := c.deleteLocationUseCase.Execute(ctx, tenantID.(string), locationID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.LocationNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa (sin contenido)
	ctx.Status(http.StatusNoContent)
}

// ListStores maneja la petición para listar solo ubicaciones de tipo tienda
func (c *LocationController) ListStores(ctx *gin.Context) {
	// Obtener el tenantID del header y agregarlo a los query parameters
	tenantID := ctx.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		// Fallback: intentar obtener del contexto (middleware)
		if tenant, exists := ctx.Get("tenantID"); exists {
			tenantID = tenant.(string)
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header es requerido"})
			return
		}
	}

	// Agregar tenant_id a los query parameters para el filtrado
	query := ctx.Request.URL.Query()
	query.Set("tenant_id", tenantID)
	ctx.Request.URL.RawQuery = query.Encode()

	// Construir criterios con filtro de tipo = 'store'
	criteriaBuilder := criteria.NewLocationCriteriaBuilder()
	criteriaBuilder.BuildFromContext(ctx)

	// Añadir filtro de tipo 'store' (sobrescribe cualquier otro filtro de tipo)
	crit := criteriaBuilder.AddTypeFilter("store").Build()

	// Ejecutar el caso de uso
	response, err := c.listLocationsUseCase.Execute(ctx, tenantID, crit)

	// Manejar errores
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// ListDistributionCenters maneja la petición para listar solo centros de distribución
func (c *LocationController) ListDistributionCenters(ctx *gin.Context) {
	// Obtener el tenantID del header y agregarlo a los query parameters
	tenantID := ctx.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		// Fallback: intentar obtener del contexto (middleware)
		if tenant, exists := ctx.Get("tenantID"); exists {
			tenantID = tenant.(string)
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header es requerido"})
			return
		}
	}

	// Agregar tenant_id a los query parameters para el filtrado
	query := ctx.Request.URL.Query()
	query.Set("tenant_id", tenantID)
	ctx.Request.URL.RawQuery = query.Encode()

	// Construir criterios con filtro de tipo = 'distribution_center'
	criteriaBuilder := criteria.NewLocationCriteriaBuilder()
	criteriaBuilder.BuildFromContext(ctx)

	// Añadir filtro de tipo 'distribution_center' (sobrescribe cualquier otro filtro de tipo)
	crit := criteriaBuilder.AddTypeFilter("distribution_center").Build()

	// Ejecutar el caso de uso
	response, err := c.listLocationsUseCase.Execute(ctx, tenantID, crit)

	// Manejar errores
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}
