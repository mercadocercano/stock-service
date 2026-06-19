package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	httpresp "github.com/hornosg/go-shared/infrastructure/response"

	"stock/src/stock_entry/application/request"
	"stock/src/stock_entry/application/usecase"
)

// StockEntryController maneja las peticiones HTTP para entradas de stock
type StockEntryController struct {
	createStockEntryUseCase     *usecase.CreateStockEntryUseCase
	bulkCreateStockEntryUseCase *usecase.BulkCreateStockEntryUseCase
	getAvailabilityUseCase      *usecase.GetAvailabilityUseCase
	listAvailabilityUseCase     *usecase.ListAvailabilityUseCase
	reserveStockUseCase         *usecase.ReserveStockUseCase
	releaseStockUseCase         *usecase.ReleaseStockUseCase
	consumeStockUseCase         *usecase.ConsumeStockUseCase
	revertConsumeUseCase        *usecase.RevertConsumeUseCase
	processSaleUseCase          *usecase.ProcessSaleUseCase
	listSalesUseCase            *usecase.ListSalesUseCase
	compensateSaleUseCase       *usecase.CompensateSaleUseCase
}

// NewStockEntryController crea una nueva instancia del controller
func NewStockEntryController(
	createStockEntryUseCase *usecase.CreateStockEntryUseCase,
	bulkCreateStockEntryUseCase *usecase.BulkCreateStockEntryUseCase,
	getAvailabilityUseCase *usecase.GetAvailabilityUseCase,
	listAvailabilityUseCase *usecase.ListAvailabilityUseCase,
	reserveStockUseCase *usecase.ReserveStockUseCase,
	releaseStockUseCase *usecase.ReleaseStockUseCase,
	consumeStockUseCase *usecase.ConsumeStockUseCase,
	revertConsumeUseCase *usecase.RevertConsumeUseCase,
	processSaleUseCase *usecase.ProcessSaleUseCase,
	listSalesUseCase *usecase.ListSalesUseCase,
	compensateSaleUseCase *usecase.CompensateSaleUseCase,
) *StockEntryController {
	return &StockEntryController{
		createStockEntryUseCase:     createStockEntryUseCase,
		bulkCreateStockEntryUseCase: bulkCreateStockEntryUseCase,
		getAvailabilityUseCase:      getAvailabilityUseCase,
		listAvailabilityUseCase:     listAvailabilityUseCase,
		reserveStockUseCase:         reserveStockUseCase,
		releaseStockUseCase:         releaseStockUseCase,
		consumeStockUseCase:         consumeStockUseCase,
		revertConsumeUseCase:        revertConsumeUseCase,
		processSaleUseCase:          processSaleUseCase,
		listSalesUseCase:            listSalesUseCase,
		compensateSaleUseCase:       compensateSaleUseCase,
	}
}

// CreateStockEntry maneja la creación de una entrada de stock
func (ctrl *StockEntryController) CreateStockEntry(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		httpresp.JSON(c, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	var req request.CreateStockEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.JSONWithDetails(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	req.TenantID = tenantID

	response, err := ctrl.createStockEntryUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		httpresp.JSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, response)
}

// BulkCreateStockEntries maneja la creación masiva de entradas
func (ctrl *StockEntryController) BulkCreateStockEntries(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		httpresp.JSON(c, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	var req request.BulkCreateStockEntriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.JSONWithDetails(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	req.TenantID = tenantID

	response, err := ctrl.bulkCreateStockEntryUseCase.Execute(c.Request.Context(), req)
	if err != nil {
		httpresp.JSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	statusCode := http.StatusCreated
	if !response.Success {
		statusCode = http.StatusPartialContent
	}

	c.JSON(statusCode, response)
}

// GetAvailability consulta la disponibilidad de un producto (por SKU) o lista toda la disponibilidad del tenant
func (ctrl *StockEntryController) GetAvailability(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		httpresp.JSON(c, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	productSKU := c.Query("sku")

	if productSKU != "" {
		resp, err := ctrl.getAvailabilityUseCase.Execute(c.Request.Context(), tenantID, productSKU)
		if err != nil {
			httpresp.JSON(c, http.StatusNotFound, err.Error())
			return
		}
		c.JSON(http.StatusOK, resp)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := ctrl.listAvailabilityUseCase.Execute(c.Request.Context(), tenantID, page, pageSize)
	if err != nil {
		httpresp.JSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

// ReserveStock maneja la reserva de stock
func (ctrl *StockEntryController) ReserveStock(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		httpresp.JSON(c, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	var req request.ReserveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.JSONWithDetails(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	response, err := ctrl.reserveStockUseCase.Execute(c.Request.Context(), tenantID, &req)
	if err != nil {
		// Manejar error de stock insuficiente
		if err.Error() == "insufficient stock" {
			httpresp.JSON(c, http.StatusConflict, "Insufficient stock available")
			return
		}
		httpresp.JSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, response)
}

// ReleaseStock maneja la liberación de stock reservado
func (ctrl *StockEntryController) ReleaseStock(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		httpresp.JSON(c, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	var req request.ReleaseStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.JSONWithDetails(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	response, err := ctrl.releaseStockUseCase.Execute(c.Request.Context(), tenantID, &req)
	if err != nil {
		// Manejar error de stock reservado insuficiente
		if contains(err.Error(), "insufficient reserved stock") {
			httpresp.JSON(c, http.StatusConflict, err.Error())
			return
		}
		httpresp.JSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, response)
}

// RegisterRoutes registra las rutas del controller
func (ctrl *StockEntryController) RegisterRoutes(router *gin.RouterGroup) {
	stockEntries := router.Group("")
	{
		stockEntries.POST("/stock-entries", ctrl.CreateStockEntry)
		stockEntries.POST("/stock-entries/bulk", ctrl.BulkCreateStockEntries)
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

	// Endpoint de listado de ventas (reporte POS)
	router.GET("/sales", ctrl.ListSales)

	// Endpoint de compensación (HITO D)
	router.POST("/compensate-sale", ctrl.CompensateSale)
}

// CompensateSale maneja la compensación (reversión) de una venta
// HITO D: Usado para rollback cuando falla persistencia de orden
func (ctrl *StockEntryController) CompensateSale(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		httpresp.JSON(c, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	var req request.CompensateSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.JSONWithDetails(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	response, err := ctrl.compensateSaleUseCase.Execute(c.Request.Context(), tenantID, &req)
	if err != nil {
		httpresp.JSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, response)
}

// RevertConsume maneja la reversión de un consumo de stock (cancelación de orden)
func (ctrl *StockEntryController) RevertConsume(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		httpresp.JSON(c, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	var req request.RevertConsumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.JSONWithDetails(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	response, err := ctrl.revertConsumeUseCase.Execute(c.Request.Context(), tenantID, &req)
	if err != nil {
		httpresp.JSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, response)
}

// ConsumeStock maneja el consumo de stock reservado (confirmación de orden)
func (ctrl *StockEntryController) ConsumeStock(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		httpresp.JSON(c, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	var req request.ConsumeStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.JSONWithDetails(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	response, err := ctrl.consumeStockUseCase.Execute(c.Request.Context(), tenantID, &req)
	if err != nil {
		// Manejar error de stock reservado insuficiente
		if contains(err.Error(), "insufficient reserved stock") {
			httpresp.JSON(c, http.StatusConflict, err.Error())
			return
		}
		httpresp.JSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, response)
}

// ProcessSale maneja el procesamiento de una venta (minimal mock)
func (ctrl *StockEntryController) ProcessSale(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		httpresp.JSON(c, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	var req request.ProcessSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.JSONWithDetails(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	response, err := ctrl.processSaleUseCase.Execute(c.Request.Context(), tenantID, &req)
	if err != nil {
		httpresp.JSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Si la venta no tuvo éxito (stock insuficiente o producto no encontrado), devolver 400
	if !response.Success {
		c.JSON(http.StatusBadRequest, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListSales lista las ventas POS recientes
func (ctrl *StockEntryController) ListSales(c *gin.Context) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		httpresp.JSON(c, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	// Parámetros de paginación (valores por defecto)
	limit := 50
	offset := 0

	sales, err := ctrl.listSalesUseCase.Execute(c.Request.Context(), tenantID, limit, offset)
	if err != nil {
		httpresp.JSON(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       sales,
		"total_count": len(sales),
	})
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
