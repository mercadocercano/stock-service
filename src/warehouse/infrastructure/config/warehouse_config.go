package config

import (
	"database/sql"

	"stock/src/warehouse/application/usecase"
	"stock/src/warehouse/domain/service"
	"stock/src/warehouse/infrastructure/controller"
	"stock/src/warehouse/infrastructure/persistence/repository"
)

// WarehouseConfig contiene la configuración para el módulo Warehouse
type WarehouseConfig struct {
	DB                  *sql.DB
	WarehouseController *controller.WarehouseController
	WarehouseService    *service.WarehouseService
}

// NewWarehouseConfig crea una nueva configuración para el módulo Warehouse
func NewWarehouseConfig(db *sql.DB) *WarehouseConfig {
	// Crear repositorio
	warehouseRepository := repository.NewPostgresWarehouseRepository(db)

	// Crear servicio de dominio
	warehouseService := service.NewWarehouseService(warehouseRepository)

	// Crear casos de uso
	createWarehouseUseCase := usecase.NewCreateWarehouseUseCase(warehouseService)
	listWarehousesUseCase := usecase.NewListWarehousesUseCase(warehouseService)
	getWarehouseUseCase := usecase.NewGetWarehouseUseCase(warehouseService)
	updateWarehouseUseCase := usecase.NewUpdateWarehouseUseCase(warehouseService)
	activateWarehouseUseCase := usecase.NewActivateWarehouseUseCase(warehouseService)
	deactivateWarehouseUseCase := usecase.NewDeactivateWarehouseUseCase(warehouseService)
	deleteWarehouseUseCase := usecase.NewDeleteWarehouseUseCase(warehouseService)

	// Crear controlador
	warehouseController := controller.NewWarehouseController(
		createWarehouseUseCase,
		listWarehousesUseCase,
		getWarehouseUseCase,
		updateWarehouseUseCase,
		activateWarehouseUseCase,
		deactivateWarehouseUseCase,
		deleteWarehouseUseCase,
	)

	return &WarehouseConfig{
		DB:                  db,
		WarehouseController: warehouseController,
		WarehouseService:    warehouseService,
	}
}
