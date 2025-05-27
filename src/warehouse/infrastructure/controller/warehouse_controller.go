package controller

import (
	"net/http"

	"stock/src/warehouse/application/request"
	"stock/src/warehouse/application/usecase"
	"stock/src/warehouse/domain/exception"
	"stock/src/warehouse/infrastructure/criteria"

	"github.com/gin-gonic/gin"
)

// WarehouseController maneja las peticiones HTTP relacionadas con almacenes
type WarehouseController struct {
	createWarehouseUseCase     *usecase.CreateWarehouseUseCase
	listWarehousesUseCase      *usecase.ListWarehousesUseCase
	getWarehouseUseCase        *usecase.GetWarehouseUseCase
	updateWarehouseUseCase     *usecase.UpdateWarehouseUseCase
	activateWarehouseUseCase   *usecase.ActivateWarehouseUseCase
	deactivateWarehouseUseCase *usecase.DeactivateWarehouseUseCase
	deleteWarehouseUseCase     *usecase.DeleteWarehouseUseCase
}

// NewWarehouseController crea una nueva instancia del controlador
func NewWarehouseController(
	createWarehouseUseCase *usecase.CreateWarehouseUseCase,
	listWarehousesUseCase *usecase.ListWarehousesUseCase,
	getWarehouseUseCase *usecase.GetWarehouseUseCase,
	updateWarehouseUseCase *usecase.UpdateWarehouseUseCase,
	activateWarehouseUseCase *usecase.ActivateWarehouseUseCase,
	deactivateWarehouseUseCase *usecase.DeactivateWarehouseUseCase,
	deleteWarehouseUseCase *usecase.DeleteWarehouseUseCase,
) *WarehouseController {
	return &WarehouseController{
		createWarehouseUseCase:     createWarehouseUseCase,
		listWarehousesUseCase:      listWarehousesUseCase,
		getWarehouseUseCase:        getWarehouseUseCase,
		updateWarehouseUseCase:     updateWarehouseUseCase,
		activateWarehouseUseCase:   activateWarehouseUseCase,
		deactivateWarehouseUseCase: deactivateWarehouseUseCase,
		deleteWarehouseUseCase:     deleteWarehouseUseCase,
	}
}

// RegisterRoutes registra las rutas del controlador en el router
func (c *WarehouseController) RegisterRoutes(router *gin.RouterGroup) {
	warehouses := router.Group("/warehouses")
	{
		warehouses.POST("", c.CreateWarehouse)
		warehouses.GET("", c.ListWarehouses)
		warehouses.GET("/:id", c.GetWarehouse)
		warehouses.PUT("/:id", c.UpdateWarehouse)
		warehouses.DELETE("/:id", c.DeleteWarehouse)
		warehouses.PATCH("/:id/activate", c.ActivateWarehouse)
		warehouses.PATCH("/:id/deactivate", c.DeactivateWarehouse)
	}

	// Rutas adicionales para listar almacenes por ubicación
	locations := router.Group("/locations")
	{
		locations.GET("/:location_id/warehouses", c.ListWarehousesByLocation)
	}
}

// CreateWarehouse maneja la petición de creación de un almacén
func (c *WarehouseController) CreateWarehouse(ctx *gin.Context) {
	var req request.CreateWarehouseRequest

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
	response, err := c.createWarehouseUseCase.Execute(ctx, req)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.WarehouseNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusCreated, response)
}

// ListWarehouses maneja la petición para listar almacenes con filtros y paginación
func (c *WarehouseController) ListWarehouses(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Utilizar el criteria builder para construir los criterios desde la petición
	criteriaBuilder := criteria.NewWarehouseCriteriaBuilder()
	crit := criteriaBuilder.BuildValidated(ctx)

	// Ejecutar el caso de uso para listar almacenes
	response, err := c.listWarehousesUseCase.Execute(ctx, tenantID.(string), crit)

	// Manejar errores
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// GetWarehouse maneja la petición para obtener un almacén por su ID
func (c *WarehouseController) GetWarehouse(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID del almacén de los parámetros de la URL
	warehouseID := ctx.Param("id")
	if warehouseID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Warehouse ID is required"})
		return
	}

	// Ejecutar el caso de uso para obtener un almacén
	response, err := c.getWarehouseUseCase.Execute(ctx, tenantID.(string), warehouseID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.WarehouseNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// UpdateWarehouse maneja la petición para actualizar un almacén
func (c *WarehouseController) UpdateWarehouse(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID del almacén de los parámetros de la URL
	warehouseID := ctx.Param("id")
	if warehouseID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Warehouse ID is required"})
		return
	}

	// Parsear el cuerpo de la petición
	var req request.UpdateWarehouseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ejecutar el caso de uso para actualizar un almacén
	response, err := c.updateWarehouseUseCase.Execute(ctx, tenantID.(string), warehouseID, req)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.WarehouseNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// ListWarehousesByLocation maneja la petición para listar almacenes por ubicación
func (c *WarehouseController) ListWarehousesByLocation(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID de la ubicación de los parámetros de la URL
	locationID := ctx.Param("location_id")
	if locationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Location ID is required"})
		return
	}

	// Utilizar el criteria builder para construir los criterios desde la petición
	criteriaBuilder := criteria.NewWarehouseCriteriaBuilder()
	crit := criteriaBuilder.BuildValidated(ctx)

	// Ejecutar el caso de uso para listar almacenes por ubicación
	response, err := c.listWarehousesUseCase.ExecuteByLocationID(ctx, locationID, tenantID.(string), crit)

	// Manejar errores
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// ActivateWarehouse maneja la petición para activar un almacén
func (c *WarehouseController) ActivateWarehouse(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID del almacén de los parámetros de la URL
	warehouseID := ctx.Param("id")
	if warehouseID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Warehouse ID is required"})
		return
	}

	// Ejecutar el caso de uso para activar un almacén
	response, err := c.activateWarehouseUseCase.Execute(ctx, tenantID.(string), warehouseID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.WarehouseNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// DeactivateWarehouse maneja la petición para desactivar un almacén
func (c *WarehouseController) DeactivateWarehouse(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID del almacén de los parámetros de la URL
	warehouseID := ctx.Param("id")
	if warehouseID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Warehouse ID is required"})
		return
	}

	// Ejecutar el caso de uso para desactivar un almacén
	response, err := c.deactivateWarehouseUseCase.Execute(ctx, tenantID.(string), warehouseID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.WarehouseNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa
	ctx.JSON(http.StatusOK, response)
}

// DeleteWarehouse maneja la petición para eliminar un almacén
func (c *WarehouseController) DeleteWarehouse(ctx *gin.Context) {
	// Obtener el tenant ID del contexto
	tenantID, exists := ctx.Get("tenantID")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Obtener el ID del almacén de los parámetros de la URL
	warehouseID := ctx.Param("id")
	if warehouseID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Warehouse ID is required"})
		return
	}

	// Ejecutar el caso de uso para eliminar un almacén
	err := c.deleteWarehouseUseCase.Execute(ctx, tenantID.(string), warehouseID)

	// Manejar errores
	if err != nil {
		switch err.(type) {
		case *exception.WarehouseNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Devolver respuesta exitosa (sin contenido)
	ctx.Status(http.StatusNoContent)
}
