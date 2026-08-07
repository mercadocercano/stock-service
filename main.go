package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // registra /debug/pprof/* en http.DefaultServeMux (ver startDebugServer)
	"os"
	"runtime"

	apiConfig "stock/src/api/config"
	locationConfig "stock/src/location/infrastructure/config"
	sharedConfig "stock/src/shared/infrastructure/config"
	stockEntryConfig "stock/src/stock_entry/infrastructure/config"
	_ "stock/src/stock_entry/infrastructure/metrics" // H8: Registra stock_insufficient_total
	stockLocationConfig "stock/src/stock_location/infrastructure/config"
	warehouseConfig "stock/src/warehouse/infrastructure/config"

	"github.com/gin-gonic/gin"
	"github.com/hornosg/go-shared/infrastructure/env"
	tenantmw "github.com/hornosg/go-shared/infrastructure/middleware"
	goshpostgres "github.com/hornosg/go-shared/infrastructure/postgres"
	sharedmigrate "github.com/hornosg/go-shared/migrate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// startDebugServer expone /debug/pprof/* (import _ "net/http/pprof") y /debug/memstats
// en un puerto separado, sin pasar por el router público de Gin ni por Kong (PLAT-E38 T-STK-M1:
// medir HeapAlloc/HeapSys/HeapIdle/NumGC en reposo y bajo carga antes de tocar GOMEMLIMIT).
// Deshabilitado por defecto: solo arranca si DEBUG_PPROF=true.
func startDebugServer() {
	if os.Getenv("DEBUG_PPROF") != "true" {
		return
	}
	http.DefaultServeMux.HandleFunc("/debug/memstats", func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"heap_alloc_bytes":    m.HeapAlloc,
			"heap_sys_bytes":      m.HeapSys,
			"heap_idle_bytes":     m.HeapIdle,
			"heap_released_bytes": m.HeapReleased,
			"heap_inuse_bytes":    m.HeapInuse,
			"sys_bytes":           m.Sys,
			"num_gc":              m.NumGC,
			"gomaxprocs":          runtime.GOMAXPROCS(0),
		})
	})
	go func() {
		log.Println("Debug server (pprof + memstats) escuchando en :6060")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Printf("Debug server detenido: %v", err)
		}
	}()
}

func main() {
	startDebugServer()

	// Configurar el router con Gin
	router := gin.New()

	// Agregar middlewares básicos necesarios
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(tenantmw.TenantValidation(tenantmw.TenantValidationConfig{
		JWTSecret: os.Getenv("JWT_SECRET"),
		ExcludedRoutes: []string{
			"/health",
			"/metrics",
		},
		RejectMissingTenant: true, // cierre de bypass de tenant (rollout verificado 2026-06-19)
	}))

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

	// Obtener configuración de la base de datos de variables de entorno.
	// Sin default inseguro (PLAT-E30 T-STK6, C1 CRÍTICO @dev-security, patrón E27/E28/E29): el
	// viejo default "postgres" es superuser → BYPASSRLS. Con él, el servicio arrancaba SIN error
	// con la RLS de locations/warehouses/stock_locations/stock_entries/stock_availability
	// "activa" pero nunca ejercida (FORCE ROW LEVEL SECURITY no aplica a superusers), sirviendo
	// stock cross-tenant. DB_USER es obligatorio y debe ser un rol NOBYPASSRLS (stock_app). El
	// chequeo de rol vivo se hace en assertNoRLSBypass.
	dbHost := env.Get("DB_HOST", "localhost")
	dbPort := env.Get("DB_PORT", "5432")
	dbUser := env.Get("DB_USER", "")
	if dbUser == "" {
		log.Fatalf("DB_USER is required and must be a NOBYPASSRLS application role " +
			"(e.g. stock_app), never postgres — RULE-09/RULE-10, PLAT-E30 C1")
	}
	dbPassword := env.Get("DB_PASSWORD", "")
	dbName := env.Get("DB_NAME", "stock_db")

	// Conectar a la base de datos usando el helper compartido
	log.Printf("Intentando conectar a postgres://%s:***@%s:%s/%s", dbUser, dbHost, dbPort, dbName)
	db, err := goshpostgres.Connect(goshpostgres.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		SSLMode:  "disable",
	})
	if err != nil {
		log.Fatalf("Error al conectar a la base de datos: %v", err)
	}
	defer db.Close()
	goshpostgres.StartPoolMonitor(context.Background(), db, goshpostgres.MonitorOptions{Service: "stock-service", DBName: dbName})
	log.Println("Conexión a la base de datos establecida con éxito")

	// Fail-fast anti-superuser (PLAT-E30 T-STK6, C1 CRÍTICO @dev-security, patrón E27/E28/E29):
	// el runtime NUNCA debe correr como superuser/BYPASSRLS. FORCE ROW LEVEL SECURITY no aplica
	// a superusers → con un rol privilegiado la RLS de locations/warehouses/stock_locations/
	// stock_entries/stock_availability queda "activa" pero nunca ejercida, sirviendo stock
	// cross-tenant sin error visible. Convierte ese fail-OPEN silencioso en un fail-CLOSED
	// ruidoso. Se corre ANTES de servir tráfico y antes de RunMigrations.
	//
	// Consecuencia operativa del flip (registrada con el sign-off de T-STK3): stock_app no tiene
	// DDL ni escritura sobre schema_migrations, así que una migración PENDIENTE termina en
	// Fatalf abajo y el proceso no arranca. En estado estacionario golang-migrate v4.19.1
	// retorna temprano si schema_migrations existe (hoy v10/dirty=f), así que el reinicio normal
	// no se ve afectado. Toda migración nueva se aplica out-of-band como postgres ANTES de
	// desplegar.
	if err := assertNoRLSBypass(db); err != nil {
		log.Fatalf("%v", err)
	}

	// Migraciones versionadas in-app (ADR-001) — fail-fast antes de servir tráfico.
	if err := sharedmigrate.RunMigrations(db, MigrationsFS, dbName); err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	// API v1 grupo de rutas
	v1 := router.Group("/api/v1")

	// Configurar el módulo API (health check y documentación)
	apiCfg := apiConfig.DefaultAPIConfig()
	apiCfg.DB = db
	apiCfg.Version = "1.0.0"
	apiConfig.SetupAPIModule(router, v1, apiCfg)

	// Configurar módulos
	setupLocationModule(v1, db)
	// setupWarehouseModule(v1, db) // COMENTADO: Conflicto de rutas con location/:id
	setupStockLocationModule(v1, db)
	setupStockEntryModule(v1, db) // HITO 2: Módulo de entradas de stock

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

// setupStockEntryModule configura el módulo StockEntry (HITO 2)
func setupStockEntryModule(router *gin.RouterGroup, db *sql.DB) {
	log.Println("Configurando módulo StockEntry...")

	// Crear configuración del módulo StockEntry
	stockEntryCfg := stockEntryConfig.NewStockEntryConfig(db)

	// Registrar rutas
	stockEntryCfg.Controller.RegisterRoutes(router)

	log.Println("Módulo StockEntry configurado exitosamente")
	log.Println("Rutas StockEntry disponibles:")
	log.Println("  POST   /api/v1/stock-entries")
	log.Println("  POST   /api/v1/stock-entries/bulk")
	log.Println("  GET    /api/v1/availability")
	log.Println("  POST   /api/v1/sale              (minimal mock endpoint)")
}

// Test change for git hook

// assertNoRLSBypass aborta el arranque si el rol de base de datos con el que conectamos es
// superuser o tiene el atributo BYPASSRLS (PLAT-E30 T-STK6, C1 CRÍTICO @dev-security). Con un rol
// así, FORCE ROW LEVEL SECURITY no se aplica y la RLS de locations, warehouses, stock_locations,
// stock_entries y stock_availability (inventario y disponibilidad por tenant) queda inerte: el
// servicio serviría stock cross-tenant sin ningún error visible. Convierte ese fail-OPEN
// silencioso en un fail-CLOSED ruidoso. Stock tiene una sola conexión (stock_db) — a diferencia de
// tenant-service (E29) no publica eventos, sin segunda conexión que proteger. El criterio "Hecho
// cuando" exige verificar `SELECT current_user` desde el proceso vivo (C1): este guard lo emite
// por log, pero la conformidad real se confirma consultando current_user desde el binario
// arrancado, no leyendo el log. ALLOW_SUPERUSER_DB=true es un escape hatch explícito para tareas
// admin locales — jamás debe usarse en producción.
func assertNoRLSBypass(db *sql.DB) error {
	if env.Get("ALLOW_SUPERUSER_DB", "false") == "true" {
		log.Println("⚠️  ALLOW_SUPERUSER_DB=true — se omite el chequeo NOBYPASSRLS (solo admin/local, NUNCA prod)")
		return nil
	}

	var privileged bool
	if err := db.QueryRow(
		`SELECT rolsuper OR rolbypassrls FROM pg_roles WHERE rolname = current_user`,
	).Scan(&privileged); err != nil {
		return fmt.Errorf("no se pudo verificar los privilegios del rol de DB (current_user): %w", err)
	}
	if privileged {
		return fmt.Errorf("negativa a arrancar: el rol de DB actual es SUPERUSER o BYPASSRLS y " +
			"eludiría la row-level security de locations/warehouses/stock_locations/" +
			"stock_entries/stock_availability (stock por tenant — RULE-09/RULE-10, PLAT-E30 C1). " +
			"Usá un rol NOBYPASSRLS como stock_app, o exportá ALLOW_SUPERUSER_DB=true solo para " +
			"tareas admin locales")
	}

	log.Println("RLS guard OK: el rol de DB es NOBYPASSRLS (locations, warehouses, stock_locations, stock_entries y stock_availability protegidas por FORCE ROW LEVEL SECURITY)")
	return nil
}
