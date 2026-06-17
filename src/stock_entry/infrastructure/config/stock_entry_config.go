package config

import (
	"database/sql"

	"stock/src/stock_entry/application/usecase"
	"stock/src/stock_entry/infrastructure/controller"
	"stock/src/stock_entry/infrastructure/logging"
	"stock/src/stock_entry/infrastructure/persistence"
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

	// Logger canónico (ADR-001)
	stockLogger := logging.NewStockLogger("stock")

	// Use cases
	createStockEntryUseCase := usecase.NewCreateStockEntryUseCase(stockEntryRepo, stockLogger)
	bulkCreateStockEntryUseCase := usecase.NewBulkCreateStockEntryUseCase(stockEntryRepo)
	getAvailabilityUseCase := usecase.NewGetAvailabilityUseCase(stockAvailabilityRepo)
	listAvailabilityUseCase := usecase.NewListAvailabilityUseCase(stockAvailabilityRepo)
	reserveStockUseCase := usecase.NewReserveStockUseCase(stockAvailabilityRepo, stockEntryRepo)
	releaseStockUseCase := usecase.NewReleaseStockUseCase(stockAvailabilityRepo, stockEntryRepo)
	consumeStockUseCase := usecase.NewConsumeStockUseCase(stockAvailabilityRepo, stockEntryRepo)
	revertConsumeUseCase := usecase.NewRevertConsumeUseCase(stockAvailabilityRepo, stockEntryRepo)
	processSaleUseCase := usecase.NewProcessSaleUseCase(stockEntryRepo, stockAvailabilityRepo, stockLogger)
	listSalesUseCase := usecase.NewListSalesUseCase(stockEntryRepo)
	compensateSaleUseCase := usecase.NewCompensateSaleUseCase(stockEntryRepo, stockLogger)

	// Controller
	stockEntryController := controller.NewStockEntryController(
		createStockEntryUseCase,
		bulkCreateStockEntryUseCase,
		getAvailabilityUseCase,
		listAvailabilityUseCase,
		reserveStockUseCase,
		releaseStockUseCase,
		consumeStockUseCase,
		revertConsumeUseCase,
		processSaleUseCase,
		listSalesUseCase,
		compensateSaleUseCase,
	)
	
	return &StockEntryConfig{
		Controller: stockEntryController,
	}
}

