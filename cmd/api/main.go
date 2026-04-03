package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Métricas Prometheus
var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "stock_http_requests_total",
			Help: "Total number of HTTP requests by method and path",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "stock_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "stock_active_connections",
			Help: "Current number of active connections",
		},
	)
)

// Middleware para métricas
func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		activeConnections.Inc()
		defer activeConnections.Dec()

		c.Next()

		status := fmt.Sprintf("%d", c.Writer.Status())
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)

		// Registro detallado para logs
		log.Printf("[%s] %s %s %d %s", method, path, c.ClientIP(), c.Writer.Status(), time.Since(start))
	}
}

func main() {
	// Configuración
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Configuración de Prometheus
	prometheusEnabled := os.Getenv("PROMETHEUS_ENABLED") == "true"
	prometheusPort := os.Getenv("PROMETHEUS_PORT")
	if prometheusPort == "" {
		prometheusPort = "2114"
	}

	// Configurar el modo Gin basado en la variable de entorno
	ginMode := os.Getenv("GIN_MODE")
	if ginMode != "" {
		gin.SetMode(ginMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Inicializar Gin
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(prometheusMiddleware()) // Agregar middleware para métricas

	// Rutas principales
	healthHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	}
	r.GET("/health", healthHandler)
	r.GET("/api/v1/health", healthHandler)

	// Endpoint para métricas directamente en el servicio
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Crear grupo de API
	api := r.Group("/api/v1")
	{
		// Rutas para ubicaciones
		locations := api.Group("/locations")
		{
			locations.GET("", func(c *gin.Context) {
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Listing locations for tenant: %s", tenantID)

				c.JSON(http.StatusOK, gin.H{
					"locations": []gin.H{},
				})
			})
			locations.POST("", func(c *gin.Context) {
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Creating location for tenant: %s", tenantID)

				c.JSON(http.StatusCreated, gin.H{
					"id":        "sample-id",
					"name":      "Sample Location",
					"code":      "SL-001",
					"tenant_id": tenantID,
					"active":    true,
				})
			})
			locations.GET("/:id", func(c *gin.Context) {
				id := c.Param("id")
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Getting location %s for tenant: %s", id, tenantID)

				c.JSON(http.StatusOK, gin.H{
					"id":        id,
					"name":      "Sample Location",
					"code":      "SL-001",
					"tenant_id": tenantID,
					"active":    true,
				})
			})
			locations.PUT("/:id", func(c *gin.Context) {
				id := c.Param("id")
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Updating location %s for tenant: %s", id, tenantID)

				c.JSON(http.StatusOK, gin.H{
					"id":        id,
					"name":      "Updated Location",
					"code":      "UL-001",
					"tenant_id": tenantID,
					"active":    true,
				})
			})
			locations.DELETE("/:id", func(c *gin.Context) {
				id := c.Param("id")
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Deleting location %s for tenant: %s", id, tenantID)

				c.Status(http.StatusNoContent)
			})
			locations.PUT("/:id/activate", func(c *gin.Context) {
				id := c.Param("id")
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Activating location %s for tenant: %s", id, tenantID)

				c.JSON(http.StatusOK, gin.H{
					"id":        id,
					"name":      "Sample Location",
					"code":      "SL-001",
					"tenant_id": tenantID,
					"active":    true,
				})
			})
			locations.PUT("/:id/deactivate", func(c *gin.Context) {
				id := c.Param("id")
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Deactivating location %s for tenant: %s", id, tenantID)

				c.JSON(http.StatusOK, gin.H{
					"id":        id,
					"name":      "Sample Location",
					"code":      "SL-001",
					"tenant_id": tenantID,
					"active":    false,
				})
			})
		}

		// Rutas para almacenes
		warehouses := api.Group("/warehouses")
		{
			warehouses.GET("", func(c *gin.Context) {
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Listing warehouses for tenant: %s", tenantID)

				c.JSON(http.StatusOK, gin.H{
					"warehouses": []gin.H{},
				})
			})
			warehouses.POST("", func(c *gin.Context) {
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Creating warehouse for tenant: %s", tenantID)

				c.JSON(http.StatusCreated, gin.H{
					"id":          "sample-id",
					"name":        "Sample Warehouse",
					"code":        "SW-001",
					"tenant_id":   tenantID,
					"location_id": "location-id",
					"active":      true,
				})
			})
			warehouses.GET("/:id", func(c *gin.Context) {
				id := c.Param("id")
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Getting warehouse %s for tenant: %s", id, tenantID)

				c.JSON(http.StatusOK, gin.H{
					"id":          id,
					"name":        "Sample Warehouse",
					"code":        "SW-001",
					"tenant_id":   tenantID,
					"location_id": "location-id",
					"active":      true,
				})
			})
			warehouses.PUT("/:id", func(c *gin.Context) {
				id := c.Param("id")
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Updating warehouse %s for tenant: %s", id, tenantID)

				c.JSON(http.StatusOK, gin.H{
					"id":          id,
					"name":        "Updated Warehouse",
					"code":        "UW-001",
					"tenant_id":   tenantID,
					"location_id": "location-id",
					"active":      true,
				})
			})
			warehouses.DELETE("/:id", func(c *gin.Context) {
				id := c.Param("id")
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Deleting warehouse %s for tenant: %s", id, tenantID)

				c.Status(http.StatusNoContent)
			})
			warehouses.PUT("/:id/activate", func(c *gin.Context) {
				id := c.Param("id")
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Activating warehouse %s for tenant: %s", id, tenantID)

				c.JSON(http.StatusOK, gin.H{
					"id":          id,
					"name":        "Sample Warehouse",
					"code":        "SW-001",
					"tenant_id":   tenantID,
					"location_id": "location-id",
					"active":      true,
				})
			})
			warehouses.PUT("/:id/deactivate", func(c *gin.Context) {
				id := c.Param("id")
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Deactivating warehouse %s for tenant: %s", id, tenantID)

				c.JSON(http.StatusOK, gin.H{
					"id":          id,
					"name":        "Sample Warehouse",
					"code":        "SW-001",
					"tenant_id":   tenantID,
					"location_id": "location-id",
					"active":      false,
				})
			})
		}

		// Rutas para ubicaciones de stock
		stockLocations := api.Group("/stock-locations")
		{
			stockLocations.GET("", func(c *gin.Context) {
				tenantID := c.GetHeader("X-Tenant-ID")
				warehouseID := c.Query("warehouse_id")
				log.Printf("Listing stock locations for tenant: %s, warehouse: %s", tenantID, warehouseID)

				c.JSON(http.StatusOK, gin.H{
					"stock_locations": []gin.H{},
				})
			})
			stockLocations.POST("", func(c *gin.Context) {
				tenantID := c.GetHeader("X-Tenant-ID")
				var body map[string]interface{}
				if err := c.ShouldBindJSON(&body); err == nil {
					warehouseID, _ := body["warehouse_id"].(string)
					log.Printf("Creating stock location for tenant: %s, warehouse: %s", tenantID, warehouseID)
				}

				c.JSON(http.StatusCreated, gin.H{
					"id":           "sample-id",
					"name":         "Sample Stock Location",
					"code":         "SSL-001",
					"tenant_id":    tenantID,
					"warehouse_id": "warehouse-id",
					"parent_id":    "",
					"path":         "sample-id",
					"active":       true,
				})
			})
			stockLocations.GET("/:id", func(c *gin.Context) {
				id := c.Param("id")
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Getting stock location %s for tenant: %s", id, tenantID)

				c.JSON(http.StatusOK, gin.H{
					"id":           id,
					"name":         "Sample Stock Location",
					"code":         "SSL-001",
					"tenant_id":    tenantID,
					"warehouse_id": "warehouse-id",
					"parent_id":    "",
					"path":         id,
					"active":       true,
				})
			})
			stockLocations.PUT("/:id", func(c *gin.Context) {
				id := c.Param("id")
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Updating stock location %s for tenant: %s", id, tenantID)

				c.JSON(http.StatusOK, gin.H{
					"id":           id,
					"name":         "Updated Stock Location",
					"code":         "USL-001",
					"tenant_id":    tenantID,
					"warehouse_id": "warehouse-id",
					"parent_id":    "",
					"path":         id,
					"active":       true,
				})
			})
			stockLocations.DELETE("/:id", func(c *gin.Context) {
				id := c.Param("id")
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Deleting stock location %s for tenant: %s", id, tenantID)

				c.Status(http.StatusNoContent)
			})
			stockLocations.PUT("/:id/activate", func(c *gin.Context) {
				id := c.Param("id")
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Activating stock location %s for tenant: %s", id, tenantID)

				c.JSON(http.StatusOK, gin.H{
					"id":           id,
					"name":         "Sample Stock Location",
					"code":         "SSL-001",
					"tenant_id":    tenantID,
					"warehouse_id": "warehouse-id",
					"parent_id":    "",
					"path":         id,
					"active":       true,
				})
			})
			stockLocations.PUT("/:id/deactivate", func(c *gin.Context) {
				id := c.Param("id")
				tenantID := c.GetHeader("X-Tenant-ID")
				log.Printf("Deactivating stock location %s for tenant: %s", id, tenantID)

				c.JSON(http.StatusOK, gin.H{
					"id":           id,
					"name":         "Sample Stock Location",
					"code":         "SSL-001",
					"tenant_id":    tenantID,
					"warehouse_id": "warehouse-id",
					"parent_id":    "",
					"path":         id,
					"active":       false,
				})
			})
		}
	}

	// Iniciar servidor HTTP
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: r,
	}

	// Iniciar servidor Prometheus si está habilitado
	var promSrv *http.Server
	if prometheusEnabled {
		promHandler := http.NewServeMux()
		promHandler.Handle("/metrics", promhttp.Handler())

		promSrv = &http.Server{
			Addr:    fmt.Sprintf(":%s", prometheusPort),
			Handler: promHandler,
		}

		go func() {
			log.Printf("Starting Prometheus metrics server on :%s", prometheusPort)
			if err := promSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start Prometheus metrics server: %v", err)
			}
		}()
	}

	// Iniciar servidor en goroutine separada
	go func() {
		log.Printf("Starting API server on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Configurar logging mejorado
	log.Printf("Stock service started. Environment: %s", os.Getenv("GIN_MODE"))
	log.Printf("Connected to database: %s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	// Esperar señal de finalización
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Crear contexto con timeout para apagado ordenado
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Apagar servidor API
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Apagar servidor Prometheus si está habilitado
	if prometheusEnabled && promSrv != nil {
		if err := promSrv.Shutdown(ctx); err != nil {
			log.Fatalf("Prometheus server forced to shutdown: %v", err)
		}
	}

	log.Println("Server exiting")
}
