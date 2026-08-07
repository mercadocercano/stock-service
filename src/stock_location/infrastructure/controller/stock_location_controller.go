package controller

import (
	"net/http"

	"stock/src/stock_location/application/request"
	"stock/src/stock_location/application/usecase"
	"stock/src/stock_location/domain/exception"
	"stock/src/stock_location/infrastructure/criteria"

	"github.com/gin-gonic/gin"
	httpresp "github.com/hornosg/go-shared/infrastructure/response"
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

	// tenant_id SIEMPRE del claim JWT verificado por tenantmw.TenantValidation (PLAT-E30 D1,
	// patrón E29 tenant-service) — nunca del header X-Tenant-ID crudo. Fail-closed 401: si el
	// claim no está en el contexto, RejectMissingTenant (main.go:75) ya abortó antes.
	tenantID := ctx.GetString("tenant_id")
	if tenantID == "" {
		httpresp.JSON(ctx, http.StatusUnauthorized, "tenant_id missing from request context")
		return
	}
	req.TenantID = tenantID

	// Parsear el cuerpo de la petición
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httpresp.JSON(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// Ejecutar el caso de uso
	response, err := c.createStockLocationUseCase.Execute(ctx, req)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.StockLocationNotFound:
			httpresp.JSON(ctx, http.StatusNotFound, err.Error())
		default:
			httpresp.JSON(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusCreated, response)
}

// ListStockLocations maneja la petición para listar ubicaciones de stock con filtros y paginación
func (c *StockLocationController) ListStockLocations(ctx *gin.Context) {
	// tenant_id SIEMPRE del claim JWT verificado por tenantmw.TenantValidation (PLAT-E30 D1,
	// patrón E29 tenant-service) — nunca del header X-Tenant-ID crudo. Fail-closed 401: si el
	// claim no está en el contexto, RejectMissingTenant (main.go:75) ya abortó antes.
	tenantID := ctx.GetString("tenant_id")
	if tenantID == "" {
		httpresp.JSON(ctx, http.StatusUnauthorized, "tenant_id missing from request context")
		return
	}

	// Agregar tenant_id a los query parameters para el filtrado
	query := ctx.Request.URL.Query()
	query.Set("tenant_id", tenantID)
	ctx.Request.URL.RawQuery = query.Encode()

	// Utilizar el criteria builder para construir los criterios desde la petición
	criteriaBuilder := criteria.NewStockLocationCriteriaBuilder()
	crit := criteriaBuilder.BuildValidated(ctx)

	// Ejecutar el caso de uso para listar ubicaciones de stock
	response, err := c.listStockLocationsUseCase.Execute(ctx, tenantID, crit)

	// Manejar errores
	if err != nil {
		httpresp.JSON(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// GetStockLocation maneja la petición para obtener una ubicación de stock por su ID
func (c *StockLocationController) GetStockLocation(ctx *gin.Context) {
	// tenant_id SIEMPRE del claim JWT verificado por tenantmw.TenantValidation (PLAT-E30 D1,
	// patrón E29 tenant-service) — nunca del header X-Tenant-ID crudo. Fail-closed 401: si el
	// claim no está en el contexto, RejectMissingTenant (main.go:75) ya abortó antes.
	tenantID := ctx.GetString("tenant_id")
	if tenantID == "" {
		httpresp.JSON(ctx, http.StatusUnauthorized, "tenant_id missing from request context")
		return
	}

	// Obtener el ID de la ubicación de stock de los parámetros de la URL
	stockLocationID := ctx.Param("id")
	if stockLocationID == "" {
		httpresp.JSON(ctx, http.StatusBadRequest, "Stock Location ID is required")
		return
	}

	// Ejecutar el caso de uso para obtener una ubicación de stock
	response, err := c.getStockLocationUseCase.Execute(ctx, tenantID, stockLocationID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.StockLocationNotFound:
			httpresp.JSON(ctx, http.StatusNotFound, err.Error())
		default:
			httpresp.JSON(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// UpdateStockLocation maneja la petición para actualizar una ubicación de stock
func (c *StockLocationController) UpdateStockLocation(ctx *gin.Context) {
	// tenant_id SIEMPRE del claim JWT verificado por tenantmw.TenantValidation (PLAT-E30 D1,
	// patrón E29 tenant-service) — nunca del header X-Tenant-ID crudo. Fail-closed 401: si el
	// claim no está en el contexto, RejectMissingTenant (main.go:75) ya abortó antes.
	tenantID := ctx.GetString("tenant_id")
	if tenantID == "" {
		httpresp.JSON(ctx, http.StatusUnauthorized, "tenant_id missing from request context")
		return
	}

	// Obtener el ID de la ubicación de stock de los parámetros de la URL
	stockLocationID := ctx.Param("id")
	if stockLocationID == "" {
		httpresp.JSON(ctx, http.StatusBadRequest, "Stock Location ID is required")
		return
	}

	// Parsear el cuerpo de la petición
	var req request.UpdateStockLocationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httpresp.JSON(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// Ejecutar el caso de uso para actualizar una ubicación de stock
	response, err := c.updateStockLocationUseCase.Execute(ctx, tenantID, stockLocationID, req)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.StockLocationNotFound:
			httpresp.JSON(ctx, http.StatusNotFound, err.Error())
		default:
			httpresp.JSON(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// ActivateStockLocation maneja la petición para activar una ubicación de stock
func (c *StockLocationController) ActivateStockLocation(ctx *gin.Context) {
	// tenant_id SIEMPRE del claim JWT verificado por tenantmw.TenantValidation (PLAT-E30 D1,
	// patrón E29 tenant-service) — nunca del header X-Tenant-ID crudo. Fail-closed 401: si el
	// claim no está en el contexto, RejectMissingTenant (main.go:75) ya abortó antes.
	tenantID := ctx.GetString("tenant_id")
	if tenantID == "" {
		httpresp.JSON(ctx, http.StatusUnauthorized, "tenant_id missing from request context")
		return
	}

	// Obtener el ID de la ubicación de stock de los parámetros de la URL
	stockLocationID := ctx.Param("id")
	if stockLocationID == "" {
		httpresp.JSON(ctx, http.StatusBadRequest, "Stock Location ID is required")
		return
	}

	// Ejecutar el caso de uso para activar una ubicación de stock
	response, err := c.activateStockLocationUseCase.Execute(ctx, tenantID, stockLocationID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.StockLocationNotFound:
			httpresp.JSON(ctx, http.StatusNotFound, err.Error())
		default:
			httpresp.JSON(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// DeactivateStockLocation maneja la petición para desactivar una ubicación de stock
func (c *StockLocationController) DeactivateStockLocation(ctx *gin.Context) {
	// tenant_id SIEMPRE del claim JWT verificado por tenantmw.TenantValidation (PLAT-E30 D1,
	// patrón E29 tenant-service) — nunca del header X-Tenant-ID crudo. Fail-closed 401: si el
	// claim no está en el contexto, RejectMissingTenant (main.go:75) ya abortó antes.
	tenantID := ctx.GetString("tenant_id")
	if tenantID == "" {
		httpresp.JSON(ctx, http.StatusUnauthorized, "tenant_id missing from request context")
		return
	}

	// Obtener el ID de la ubicación de stock de los parámetros de la URL
	stockLocationID := ctx.Param("id")
	if stockLocationID == "" {
		httpresp.JSON(ctx, http.StatusBadRequest, "Stock Location ID is required")
		return
	}

	// Ejecutar el caso de uso para desactivar una ubicación de stock
	response, err := c.deactivateStockLocationUseCase.Execute(ctx, tenantID, stockLocationID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.StockLocationNotFound:
			httpresp.JSON(ctx, http.StatusNotFound, err.Error())
		default:
			httpresp.JSON(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// DeleteStockLocation maneja la petición para eliminar una ubicación de stock
func (c *StockLocationController) DeleteStockLocation(ctx *gin.Context) {
	// tenant_id SIEMPRE del claim JWT verificado por tenantmw.TenantValidation (PLAT-E30 D1,
	// patrón E29 tenant-service) — nunca del header X-Tenant-ID crudo. Fail-closed 401: si el
	// claim no está en el contexto, RejectMissingTenant (main.go:75) ya abortó antes.
	tenantID := ctx.GetString("tenant_id")
	if tenantID == "" {
		httpresp.JSON(ctx, http.StatusUnauthorized, "tenant_id missing from request context")
		return
	}

	// Obtener el ID de la ubicación de stock de los parámetros de la URL
	stockLocationID := ctx.Param("id")
	if stockLocationID == "" {
		httpresp.JSON(ctx, http.StatusBadRequest, "Stock Location ID is required")
		return
	}

	// Ejecutar el caso de uso para eliminar una ubicación de stock
	err := c.deleteStockLocationUseCase.Execute(ctx, tenantID, stockLocationID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.StockLocationNotFound:
			httpresp.JSON(ctx, http.StatusNotFound, err.Error())
		default:
			httpresp.JSON(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Devolver respuesta exitosa (sin contenido)
	ctx.Status(http.StatusNoContent)
}

// ListStockLocationsByWarehouse maneja la petición para listar ubicaciones de stock por almacén
func (c *StockLocationController) ListStockLocationsByWarehouse(ctx *gin.Context) {
	// tenant_id SIEMPRE del claim JWT verificado por tenantmw.TenantValidation (PLAT-E30 D1,
	// patrón E29 tenant-service) — nunca del header X-Tenant-ID crudo. Fail-closed 401: si el
	// claim no está en el contexto, RejectMissingTenant (main.go:75) ya abortó antes.
	tenantID := ctx.GetString("tenant_id")
	if tenantID == "" {
		httpresp.JSON(ctx, http.StatusUnauthorized, "tenant_id missing from request context")
		return
	}

	// Obtener el ID del almacén de los parámetros de la URL
	warehouseID := ctx.Param("warehouse_id")
	if warehouseID == "" {
		httpresp.JSON(ctx, http.StatusBadRequest, "Warehouse ID is required")
		return
	}

	// Agregar tenant_id a los query parameters para el filtrado
	query := ctx.Request.URL.Query()
	query.Set("tenant_id", tenantID)
	ctx.Request.URL.RawQuery = query.Encode()

	// Utilizar el criteria builder para construir los criterios desde la petición
	criteriaBuilder := criteria.NewStockLocationCriteriaBuilder()
	crit := criteriaBuilder.BuildValidated(ctx)

	// Ejecutar el caso de uso para listar ubicaciones de stock por almacén
	response, err := c.listStockLocationsUseCase.ExecuteByWarehouseID(ctx, warehouseID, tenantID, crit)

	// Manejar errores
	if err != nil {
		httpresp.JSON(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// ListRootStockLocations maneja la petición para listar ubicaciones de stock raíz por almacén
func (c *StockLocationController) ListRootStockLocations(ctx *gin.Context) {
	// tenant_id SIEMPRE del claim JWT verificado por tenantmw.TenantValidation (PLAT-E30 D1,
	// patrón E29 tenant-service) — nunca del header X-Tenant-ID crudo. Fail-closed 401: si el
	// claim no está en el contexto, RejectMissingTenant (main.go:75) ya abortó antes.
	tenantID := ctx.GetString("tenant_id")
	if tenantID == "" {
		httpresp.JSON(ctx, http.StatusUnauthorized, "tenant_id missing from request context")
		return
	}

	// Obtener el ID del almacén de los parámetros de la URL
	warehouseID := ctx.Param("warehouse_id")
	if warehouseID == "" {
		httpresp.JSON(ctx, http.StatusBadRequest, "Warehouse ID is required")
		return
	}

	// Agregar tenant_id a los query parameters para el filtrado
	query := ctx.Request.URL.Query()
	query.Set("tenant_id", tenantID)
	ctx.Request.URL.RawQuery = query.Encode()

	// Utilizar el criteria builder para construir los criterios desde la petición
	criteriaBuilder := criteria.NewStockLocationCriteriaBuilder()
	crit := criteriaBuilder.BuildValidated(ctx)

	// Ejecutar el caso de uso para listar ubicaciones de stock raíz por almacén
	response, err := c.listStockLocationsUseCase.ExecuteRoots(ctx, warehouseID, tenantID, crit)

	// Manejar errores
	if err != nil {
		httpresp.JSON(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// ListChildrenStockLocations maneja la petición para listar ubicaciones de stock hijas
func (c *StockLocationController) ListChildrenStockLocations(ctx *gin.Context) {
	// tenant_id SIEMPRE del claim JWT verificado por tenantmw.TenantValidation (PLAT-E30 D1,
	// patrón E29 tenant-service) — nunca del header X-Tenant-ID crudo. Fail-closed 401: si el
	// claim no está en el contexto, RejectMissingTenant (main.go:75) ya abortó antes.
	tenantID := ctx.GetString("tenant_id")
	if tenantID == "" {
		httpresp.JSON(ctx, http.StatusUnauthorized, "tenant_id missing from request context")
		return
	}

	// Obtener el ID de la ubicación padre de los parámetros de la URL
	parentID := ctx.Param("id")
	if parentID == "" {
		httpresp.JSON(ctx, http.StatusBadRequest, "Parent Stock Location ID is required")
		return
	}

	// Agregar tenant_id a los query parameters para el filtrado
	query := ctx.Request.URL.Query()
	query.Set("tenant_id", tenantID)
	ctx.Request.URL.RawQuery = query.Encode()

	// Utilizar el criteria builder para construir los criterios desde la petición
	criteriaBuilder := criteria.NewStockLocationCriteriaBuilder()
	crit := criteriaBuilder.BuildValidated(ctx)

	// Ejecutar el caso de uso para listar ubicaciones de stock hijas
	response, err := c.listStockLocationsUseCase.ExecuteChildren(ctx, parentID, tenantID, crit)

	// Manejar errores
	if err != nil {
		httpresp.JSON(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}
