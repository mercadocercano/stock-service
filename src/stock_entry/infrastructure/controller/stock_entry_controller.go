package controller

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	
	"stock-service/src/stock_entry/application/request"
	"stock-service/src/stock_entry/application/usecase"
)

// StockEntryController maneja las peticiones HTTP para entradas de stock
type StockEntryController struct {
	createStockEntryUseCase     *usecase.CreateStockEntryUseCase
	bulkCreateStockEntryUseCase *usecase.BulkCreateStockEntryUseCase
	getAvailabilityUseCase      *usecase.GetAvailabilityUseCase
}

// NewStockEntryController crea una nueva instancia del controller
func NewStockEntryController(
	createStockEntryUseCase *usecase.CreateStockEntryUseCase,
	bulkCreateStockEntryUseCase *usecase.BulkCreateStockEntryUseCase,
	getAvailabilityUseCase *usecase.GetAvailabilityUseCase,
) *StockEntryController {
	return &StockEntryController{
		createStockEntryUseCase:     createStockEntryUseCase,
		bulkCreateStockEntryUseCase: bulkCreateStockEntryUseCase,
		getAvailabilityUseCase:      getAvailabilityUseCase,
	}
}

// CreateStockEntry maneja la creación de una entrada de stock
func (ctrl *StockEntryController) CreateStockEntry(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}
	
	var req request.CreateStockEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}
	
	req.TenantID = tenantID
	
	response, err := ctrl.createStockEntryUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, response)
}

// BulkCreateStockEntries maneja la creación masiva de entradas
func (ctrl *StockEntryController) BulkCreateStockEntries(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}
	
	var req request.BulkCreateStockEntriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}
	
	req.TenantID = tenantID
	
	response, err := ctrl.bulkCreateStockEntryUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	statusCode := http.StatusCreated
	if !response.Success {
		statusCode = http.StatusPartialContent
	}
	
	c.JSON(statusCode, response)
}

// GetAvailability consulta la disponibilidad de un producto
func (ctrl *StockEntryController) GetAvailability(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}
	
	productSKU := c.Query("sku")
	if productSKU == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku query parameter is required"})
		return
	}
	
	response, err := ctrl.getAvailabilityUseCase.Execute(c.Request.Context(), tenantID, productSKU)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// RegisterRoutes registra las rutas del controller
func (ctrl *StockEntryController) RegisterRoutes(router *gin.RouterGroup) {
	stockEntries := router.Group("/stock-entries")
	{
		stockEntries.POST("", ctrl.CreateStockEntry)
		stockEntries.POST("/bulk", ctrl.BulkCreateStockEntries)
	}
	
	// Endpoint de disponibilidad
	router.GET("/availability", ctrl.GetAvailability)
}

