package main

import (
	"database/sql"
	"log"
	"os"

	apiConfig "stock/src/api/config"
	locationConfig "stock/src/location/infrastructure/config"
	sharedConfig "stock/src/shared/infrastructure/config"
	stockLocationConfig "stock/src/stock_location/infrastructure/config"
	warehouseConfig "stock/src/warehouse/infrastructure/config"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // Driver de PostgreSQL
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// getEnv obtiene una variable de entorno o devuelve un valor por defecto
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	// Configurar el router con Gin
	router := gin.New()

	// Agregar middlewares básicos necesarios
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Configurar Prometheus metrics si está habilitado
	prometheusEnabled := os.Getenv("PROMETHEUS_ENABLED")
	log.Printf("PROMETHEUS_ENABLED value: '%s'", prometheusEnabled)

	if prometheusEnabled == "true" {
		log.Println("Registering /metrics endpoint for Stock service")
		// Endpoint de métricas usando la librería oficial de Prometheus
		router.GET("/metrics", gin.WrapH(promhttp.Handler()))
		log.Println("/metrics endpoint registered successfully for Stock service")
	} else {
		log.Println("Prometheus metrics disabled for Stock service")
	}

	// Configurar GZIP y otros middlewares compartidos
	gzipSharedCfg := sharedConfig.DefaultSharedConfig()
	sharedConfig.SetupSharedMiddleware(router, gzipSharedCfg)

	// Obtener configuración de la base de datos de variables de entorno
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "stock_db")

	// Crear string de conexión
	connStr := "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=disable"
	log.Printf("Intentando conectar a %s", connStr)

	// Conectar a la base de datos
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error al conectar a la base de datos: %v", err)
	}
	defer db.Close()

	// Comprobar la conexión
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error al verificar la conexión a la base de datos: %v", err)
	}
	log.Println("Conexión a la base de datos establecida con éxito")

	// API v1 grupo de rutas
	v1 := router.Group("/api/v1")

	// Configurar el módulo API (health check y documentación)
	apiCfg := apiConfig.DefaultAPIConfig()
	apiCfg.DB = db
	apiCfg.Version = "1.0.0"
	apiConfig.SetupAPIModule(router, v1, apiCfg)

	// Configurar módulos
	setupLocationModule(v1, db)
	setupWarehouseModule(v1, db)
	setupStockLocationModule(v1, db)

	// Iniciar el servidor
	log.Println("Servidor iniciando en http://localhost:8080")
	router.Run(":8080")
}

// setupLocationModule configura el módulo Location
func setupLocationModule(router *gin.RouterGroup, db *sql.DB) {
	log.Println("Configurando módulo Location...")

	// Crear configuración del módulo Location
	locationCfg := locationConfig.NewLocationConfig(db)

	// Registrar rutas
	locationCfg.LocationController.RegisterRoutes(router)

	log.Println("Módulo Location configurado exitosamente")
	log.Println("Rutas Location disponibles:")
	log.Println("  POST   /api/v1/locations")
	log.Println("  GET    /api/v1/locations")
	log.Println("  GET    /api/v1/locations/:id")
	log.Println("  PUT    /api/v1/locations/:id")
	log.Println("  DELETE /api/v1/locations/:id")
	log.Println("  PATCH  /api/v1/locations/:id/activate")
	log.Println("  PATCH  /api/v1/locations/:id/deactivate")
	log.Println("  GET    /api/v1/locations/stores")
	log.Println("  GET    /api/v1/locations/distribution-centers")
}

// setupWarehouseModule configura el módulo Warehouse
func setupWarehouseModule(router *gin.RouterGroup, db *sql.DB) {
	log.Println("Configurando módulo Warehouse...")

	// Crear configuración del módulo Warehouse
	warehouseCfg := warehouseConfig.NewWarehouseConfig(db)

	// Registrar rutas
	warehouseCfg.WarehouseController.RegisterRoutes(router)

	log.Println("Módulo Warehouse configurado exitosamente")
	log.Println("Rutas Warehouse disponibles:")
	log.Println("  POST   /api/v1/warehouses")
	log.Println("  GET    /api/v1/warehouses")
	log.Println("  GET    /api/v1/warehouses/:id")
	log.Println("  PUT    /api/v1/warehouses/:id")
	log.Println("  DELETE /api/v1/warehouses/:id")
	log.Println("  PATCH  /api/v1/warehouses/:id/activate")
	log.Println("  PATCH  /api/v1/warehouses/:id/deactivate")
	log.Println("  GET    /api/v1/locations/:location_id/warehouses")
}

// setupStockLocationModule configura el módulo StockLocation
func setupStockLocationModule(router *gin.RouterGroup, db *sql.DB) {
	log.Println("Configurando módulo StockLocation...")

	// Crear configuración del módulo StockLocation
	stockLocationCfg := stockLocationConfig.NewStockLocationConfig(db)

	// Registrar rutas
	stockLocationCfg.StockLocationController.RegisterRoutes(router)

	log.Println("Módulo StockLocation configurado exitosamente")
	log.Println("Rutas StockLocation disponibles:")
	log.Println("  POST   /api/v1/stock-locations")
	log.Println("  GET    /api/v1/stock-locations")
	log.Println("  GET    /api/v1/stock-locations/:id")
	log.Println("  PUT    /api/v1/stock-locations/:id")
	log.Println("  DELETE /api/v1/stock-locations/:id")
	log.Println("  GET    /api/v1/warehouses/:warehouse_id/stock-locations")
}
// Test change for git hook
