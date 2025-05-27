package config

import (
	"database/sql"

	"stock/src/location/application/usecase"
	"stock/src/location/domain/service"
	"stock/src/location/infrastructure/controller"
	"stock/src/location/infrastructure/persistence/repository"
)

// LocationConfig contiene la configuración para el módulo Location
type LocationConfig struct {
	DB                 *sql.DB
	LocationController *controller.LocationController
	LocationService    *service.LocationService
}

// LocationController es una interfaz temporal para el controlador de Location
type LocationController struct {
	db *sql.DB
}

// RegisterRoutes registra las rutas del controlador
func (lc *LocationController) RegisterRoutes(router interface{}) {
	// Implementación temporal
}

// NewLocationConfig crea una nueva configuración para el módulo Location
func NewLocationConfig(db *sql.DB) *LocationConfig {
	// Crear repositorio
	locationRepository := repository.NewPostgresLocationRepository(db)

	// Crear servicio de dominio
	locationService := service.NewLocationService(locationRepository)

	// Crear casos de uso
	createLocationUseCase := usecase.NewCreateLocationUseCase(locationService)
	listLocationsUseCase := usecase.NewListLocationsUseCase(locationService)
	getLocationUseCase := usecase.NewGetLocationUseCase(locationService)
	updateLocationUseCase := usecase.NewUpdateLocationUseCase(locationService)
	activateLocationUseCase := usecase.NewActivateLocationUseCase(locationService)
	deactivateLocationUseCase := usecase.NewDeactivateLocationUseCase(locationService)
	deleteLocationUseCase := usecase.NewDeleteLocationUseCase(locationService)

	// Crear controlador
	locationController := controller.NewLocationController(
		createLocationUseCase,
		listLocationsUseCase,
		getLocationUseCase,
		updateLocationUseCase,
		activateLocationUseCase,
		deactivateLocationUseCase,
		deleteLocationUseCase,
	)

	return &LocationConfig{
		DB:                 db,
		LocationController: locationController,
		LocationService:    locationService,
	}
}
