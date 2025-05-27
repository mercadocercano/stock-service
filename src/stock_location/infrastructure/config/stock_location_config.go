package config

import (
	"database/sql"

	"stock/src/stock_location/application/usecase"
	"stock/src/stock_location/domain/service"
	"stock/src/stock_location/infrastructure/controller"
	"stock/src/stock_location/infrastructure/persistence/repository"
)

// StockLocationConfig contiene la configuración para el módulo StockLocation
type StockLocationConfig struct {
	DB                      *sql.DB
	StockLocationController *controller.StockLocationController
	StockLocationService    *service.StockLocationService
}

// NewStockLocationConfig crea una nueva configuración para el módulo StockLocation
func NewStockLocationConfig(db *sql.DB) *StockLocationConfig {
	// Crear repositorio
	stockLocationRepository := repository.NewPostgresStockLocationRepository(db)

	// Crear servicio de dominio
	stockLocationService := service.NewStockLocationService(stockLocationRepository)

	// Crear casos de uso
	createStockLocationUseCase := usecase.NewCreateStockLocationUseCase(stockLocationService)
	listStockLocationsUseCase := usecase.NewListStockLocationsUseCase(stockLocationService)
	getStockLocationUseCase := usecase.NewGetStockLocationUseCase(stockLocationService)
	updateStockLocationUseCase := usecase.NewUpdateStockLocationUseCase(stockLocationService)
	activateStockLocationUseCase := usecase.NewActivateStockLocationUseCase(stockLocationService)
	deactivateStockLocationUseCase := usecase.NewDeactivateStockLocationUseCase(stockLocationService)
	deleteStockLocationUseCase := usecase.NewDeleteStockLocationUseCase(stockLocationService)

	// Crear controlador
	stockLocationController := controller.NewStockLocationController(
		createStockLocationUseCase,
		listStockLocationsUseCase,
		getStockLocationUseCase,
		updateStockLocationUseCase,
		activateStockLocationUseCase,
		deactivateStockLocationUseCase,
		deleteStockLocationUseCase,
	)

	return &StockLocationConfig{
		DB:                      db,
		StockLocationController: stockLocationController,
		StockLocationService:    stockLocationService,
	}
}
