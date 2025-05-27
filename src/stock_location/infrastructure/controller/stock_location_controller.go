package controller

import (
	"net/http"

	"stock/src/stock_location/application/request"
	"stock/src/stock_location/application/usecase"
	"stock/src/stock_location/domain/exception"
	"stock/src/stock_location/infrastructure/criteria"

	"github.com/gin-gonic/gin"
)

// StockLocationController maneja las peticiones HTTP relacionadas con ubicaciones de stock
type StockLocationController struct {
	createStockLocationUseCase     *usecase.CreateStockLocationUseCase
	listStockLocationsUseCase      *usecase.ListStockLocationsUseCase
	getStockLocationUseCase        *usecase.GetStockLocationUseCase
	updateStockLocationUseCase     *usecase.UpdateStockLocationUseCase
	activateStockLocationUseCase   *usecase.ActivateStockLocationUseCase
	deactivateStockLocationUseCase *usecase.DeactivateStockLocationUseCase
	deleteStockLocationUseCase     *usecase.DeleteStockLocationUseCase
}

// NewStockLocationController crea una nueva instancia del controlador
func NewStockLocationController(
	createStockLocationUseCase *usecase.CreateStockLocationUseCase,
	listStockLocationsUseCase *usecase.ListStockLocationsUseCase,
	getStockLocationUseCase *usecase.GetStockLocationUseCase,
	updateStockLocationUseCase *usecase.UpdateStockLocationUseCase,
	activateStockLocationUseCase *usecase.ActivateStockLocationUseCase,
	deactivateStockLocationUseCase *usecase.DeactivateStockLocationUseCase,
	deleteStockLocationUseCase *usecase.DeleteStockLocationUseCase,
) *StockLocationController {
	return &StockLocationController{
		createStockLocationUseCase:     createStockLocationUseCase,
		listStockLocationsUseCase:      listStockLocationsUseCase,
		getStockLocationUseCase:        getStockLocationUseCase,
		updateStockLocationUseCase:     updateStockLocationUseCase,
		activateStockLocationUseCase:   activateStockLocationUseCase,
		deactivateStockLocationUseCase: deactivateStockLocationUseCase,
		deleteStockLocationUseCase:     deleteStockLocationUseCase,
	}
}

// RegisterRoutes registra las rutas del controlador en el router
func (c *StockLocationController) RegisterRoutes(router *gin.RouterGroup) {
	stockLocations := router.Group("/stock-locations")
	{
		stockLocations.POST("", c.CreateStockLocation)
		stockLocations.GET("", c.ListStockLocations)
		stockLocations.GET("/:id", c.GetStockLocation)
		stockLocations.PUT("/:id", c.UpdateStockLocation)
		stockLocations.DELETE("/:id", c.DeleteStockLocation)
		stockLocations.PATCH("/:id/activate", c.ActivateStockLocation)
		stockLocations.PATCH("/:id/deactivate", c.DeactivateStockLocation)
	}

	// Rutas adicionales para listar ubicaciones por almacén
	warehouses := router.Group("/warehouses")
	{
		warehouses.GET("/:warehouse_id/stock-locations", c.ListStockLocationsByWarehouse)
		warehouses.GET("/:warehouse_id/stock-locations/roots", c.ListRootStockLocations)
	}

	// Rutas para listar ubicaciones hijas
	stockLocations.GET("/:id/children", c.ListChildrenStockLocations)
}

// CreateStockLocation maneja la petición de creación de una ubicación de stock
func (c *StockLocationController) CreateStockLocation(ctx *gin.Context) {
	var req request.CreateStockLocationRequest

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
	response, err := c.createStockLocationUseCase.Execute(ctx, req)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.StockLocationNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusCreated, response)
}

// ListStockLocations maneja la petición para listar ubicaciones de stock con filtros y paginación
func (c *StockLocationController) ListStockLocations(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Utilizar el criteria builder para construir los criterios desde la petición
	criteriaBuilder := criteria.NewStockLocationCriteriaBuilder()
	crit := criteriaBuilder.BuildValidated(ctx)

	// Ejecutar el caso de uso para listar ubicaciones de stock
	response, err := c.listStockLocationsUseCase.Execute(ctx, tenantID.(string), crit)

	// Manejar errores
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// GetStockLocation maneja la petición para obtener una ubicación de stock por su ID
func (c *StockLocationController) GetStockLocation(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID de la ubicación de stock de los parámetros de la URL
	stockLocationID := ctx.Param("id")
	if stockLocationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Stock Location ID is required"})
		return
	}

	// Ejecutar el caso de uso para obtener una ubicación de stock
	response, err := c.getStockLocationUseCase.Execute(ctx, tenantID.(string), stockLocationID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.StockLocationNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// UpdateStockLocation maneja la petición para actualizar una ubicación de stock
func (c *StockLocationController) UpdateStockLocation(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID de la ubicación de stock de los parámetros de la URL
	stockLocationID := ctx.Param("id")
	if stockLocationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Stock Location ID is required"})
		return
	}

	// Parsear el cuerpo de la petición
	var req request.UpdateStockLocationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ejecutar el caso de uso para actualizar una ubicación de stock
	response, err := c.updateStockLocationUseCase.Execute(ctx, tenantID.(string), stockLocationID, req)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.StockLocationNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// ActivateStockLocation maneja la petición para activar una ubicación de stock
func (c *StockLocationController) ActivateStockLocation(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID de la ubicación de stock de los parámetros de la URL
	stockLocationID := ctx.Param("id")
	if stockLocationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Stock Location ID is required"})
		return
	}

	// Ejecutar el caso de uso para activar una ubicación de stock
	response, err := c.activateStockLocationUseCase.Execute(ctx, tenantID.(string), stockLocationID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.StockLocationNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// DeactivateStockLocation maneja la petición para desactivar una ubicación de stock
func (c *StockLocationController) DeactivateStockLocation(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID de la ubicación de stock de los parámetros de la URL
	stockLocationID := ctx.Param("id")
	if stockLocationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Stock Location ID is required"})
		return
	}

	// Ejecutar el caso de uso para desactivar una ubicación de stock
	response, err := c.deactivateStockLocationUseCase.Execute(ctx, tenantID.(string), stockLocationID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.StockLocationNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// DeleteStockLocation maneja la petición para eliminar una ubicación de stock
func (c *StockLocationController) DeleteStockLocation(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID de la ubicación de stock de los parámetros de la URL
	stockLocationID := ctx.Param("id")
	if stockLocationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Stock Location ID is required"})
		return
	}

	// Ejecutar el caso de uso para eliminar una ubicación de stock
	err := c.deleteStockLocationUseCase.Execute(ctx, tenantID.(string), stockLocationID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.StockLocationNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa (sin contenido)
	ctx.Status(http.StatusNoContent)
}

// ListStockLocationsByWarehouse maneja la petición para listar ubicaciones de stock por almacén
func (c *StockLocationController) ListStockLocationsByWarehouse(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID del almacén de los parámetros de la URL
	warehouseID := ctx.Param("warehouse_id")
	if warehouseID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Warehouse ID is required"})
		return
	}

	// Utilizar el criteria builder para construir los criterios desde la petición
	criteriaBuilder := criteria.NewStockLocationCriteriaBuilder()
	crit := criteriaBuilder.BuildValidated(ctx)

	// Ejecutar el caso de uso para listar ubicaciones de stock por almacén
	response, err := c.listStockLocationsUseCase.ExecuteByWarehouseID(ctx, warehouseID, tenantID.(string), crit)

	// Manejar errores
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// ListRootStockLocations maneja la petición para listar ubicaciones de stock raíz por almacén
func (c *StockLocationController) ListRootStockLocations(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID del almacén de los parámetros de la URL
	warehouseID := ctx.Param("warehouse_id")
	if warehouseID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Warehouse ID is required"})
		return
	}

	// Utilizar el criteria builder para construir los criterios desde la petición
	criteriaBuilder := criteria.NewStockLocationCriteriaBuilder()
	crit := criteriaBuilder.BuildValidated(ctx)

	// Ejecutar el caso de uso para listar ubicaciones de stock raíz por almacén
	response, err := c.listStockLocationsUseCase.ExecuteRoots(ctx, warehouseID, tenantID.(string), crit)

	// Manejar errores
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// ListChildrenStockLocations maneja la petición para listar ubicaciones de stock hijas
func (c *StockLocationController) ListChildrenStockLocations(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID de la ubicación padre de los parámetros de la URL
	parentID := ctx.Param("id")
	if parentID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Parent Stock Location ID is required"})
		return
	}

	// Utilizar el criteria builder para construir los criterios desde la petición
	criteriaBuilder := criteria.NewStockLocationCriteriaBuilder()
	crit := criteriaBuilder.BuildValidated(ctx)

	// Ejecutar el caso de uso para listar ubicaciones de stock hijas
	response, err := c.listStockLocationsUseCase.ExecuteChildren(ctx, parentID, tenantID.(string), crit)

	// Manejar errores
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}
