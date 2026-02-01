package config

import (
	"database/sql"
	
	"stock-service/src/stock_entry/application/usecase"
	"stock-service/src/stock_entry/infrastructure/controller"
	"stock-service/src/stock_entry/infrastructure/persistence"
)

// StockEntryConfig configuración del módulo stock_entry
type StockEntryConfig struct {
	Controller *controller.StockEntryController
}

// NewStockEntryConfig crea una nueva configuración del módulo
func NewStockEntryConfig(db *sql.DB) *StockEntryConfig {
	// Repositorios
	stockEntryRepo := persistence.NewPostgresStockEntryRepository(db)
	stockAvailabilityRepo := persistence.NewPostgresStockAvailabilityRepository(db)
	
	// Use cases
	createStockEntryUseCase := usecase.NewCreateStockEntryUseCase(stockEntryRepo)
	bulkCreateStockEntryUseCase := usecase.NewBulkCreateStockEntryUseCase(stockEntryRepo)
	getAvailabilityUseCase := usecase.NewGetAvailabilityUseCase(stockAvailabilityRepo)
	
	// Controller
	stockEntryController := controller.NewStockEntryController(
		createStockEntryUseCase,
		bulkCreateStockEntryUseCase,
		getAvailabilityUseCase,
	)
	
	return &StockEntryConfig{
		Controller: stockEntryController,
	}
}

