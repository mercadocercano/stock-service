package controller

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	
	"stock/src/stock_entry/application/request"
	"stock/src/stock_entry/application/usecase"
)

// StockEntryController maneja las peticiones HTTP para entradas de stock
type StockEntryController struct {
	createStockEntryUseCase     *usecase.CreateStockEntryUseCase
	bulkCreateStockEntryUseCase *usecase.BulkCreateStockEntryUseCase
	getAvailabilityUseCase      *usecase.GetAvailabilityUseCase
	reserveStockUseCase         *usecase.ReserveStockUseCase
	releaseStockUseCase         *usecase.ReleaseStockUseCase
	consumeStockUseCase         *usecase.ConsumeStockUseCase
	revertConsumeUseCase        *usecase.RevertConsumeUseCase
	processSaleUseCase          *usecase.ProcessSaleUseCase
}

// NewStockEntryController crea una nueva instancia del controller
func NewStockEntryController(
	createStockEntryUseCase *usecase.CreateStockEntryUseCase,
	bulkCreateStockEntryUseCase *usecase.BulkCreateStockEntryUseCase,
	getAvailabilityUseCase *usecase.GetAvailabilityUseCase,
	reserveStockUseCase *usecase.ReserveStockUseCase,
	releaseStockUseCase *usecase.ReleaseStockUseCase,
	consumeStockUseCase *usecase.ConsumeStockUseCase,
	revertConsumeUseCase *usecase.RevertConsumeUseCase,
	processSaleUseCase *usecase.ProcessSaleUseCase,
) *StockEntryController {
	return &StockEntryController{
		createStockEntryUseCase:     createStockEntryUseCase,
		bulkCreateStockEntryUseCase: bulkCreateStockEntryUseCase,
		getAvailabilityUseCase:      getAvailabilityUseCase,
		reserveStockUseCase:         reserveStockUseCase,
		releaseStockUseCase:         releaseStockUseCase,
		consumeStockUseCase:         consumeStockUseCase,
		revertConsumeUseCase:        revertConsumeUseCase,
		processSaleUseCase:          processSaleUseCase,
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

// ReserveStock maneja la reserva de stock
func (ctrl *StockEntryController) ReserveStock(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	var req request.ReserveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	response, err := ctrl.reserveStockUseCase.Execute(c.Request.Context(), tenantID, &req)
	if err != nil {
		// Manejar error de stock insuficiente
		if err.Error() == "insufficient stock" {
			c.JSON(http.StatusConflict, gin.H{"error": "Insufficient stock available"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ReleaseStock maneja la liberación de stock reservado
func (ctrl *StockEntryController) ReleaseStock(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	var req request.ReleaseStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	response, err := ctrl.releaseStockUseCase.Execute(c.Request.Context(), tenantID, &req)
	if err != nil {
		// Manejar error de stock reservado insuficiente
		if contains(err.Error(), "insufficient reserved stock") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	
	// Endpoint de reserva
	router.POST("/reserve", ctrl.ReserveStock)
	
	// Endpoint de liberación
	router.POST("/release", ctrl.ReleaseStock)
	
	// Endpoint de consumo
	router.POST("/consume", ctrl.ConsumeStock)
	
	// Endpoint de reversión de consumo
	router.POST("/revert-consume", ctrl.RevertConsume)
	
	// Endpoint de venta (minimal mock)
	router.POST("/sale", ctrl.ProcessSale)
}

// RevertConsume maneja la reversión de un consumo de stock (cancelación de orden)
func (ctrl *StockEntryController) RevertConsume(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	var req request.RevertConsumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	response, err := ctrl.revertConsumeUseCase.Execute(c.Request.Context(), tenantID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ConsumeStock maneja el consumo de stock reservado (confirmación de orden)
func (ctrl *StockEntryController) ConsumeStock(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	var req request.ConsumeStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	response, err := ctrl.consumeStockUseCase.Execute(c.Request.Context(), tenantID, &req)
	if err != nil {
		// Manejar error de stock reservado insuficiente
		if contains(err.Error(), "insufficient reserved stock") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ProcessSale maneja el procesamiento de una venta (minimal mock)
func (ctrl *StockEntryController) ProcessSale(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Tenant-ID header is required"})
		return
	}

	var req request.ProcessSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	response, err := ctrl.processSaleUseCase.Execute(c.Request.Context(), tenantID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Si la venta no tuvo éxito (stock insuficiente o producto no encontrado), devolver 400
	if !response.Success {
		c.JSON(http.StatusBadRequest, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// contains helper para verificar substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if len(s[i:]) >= len(substr) && s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

