//go:build integration
// +build integration

package persistence_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/src/stock_entry/domain/entity"
	"stock/src/stock_entry/domain/exception"
	"stock/src/stock_entry/infrastructure/persistence"
)

// TestProcessSaleAtomic_Success verifica que la venta atómica funciona correctamente
func TestProcessSaleAtomic_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tenantID := uuid.New()
	variantSKU := "TEST-SKU-001"

	// Setup: crear stock inicial
	setupInitialStock(t, db, tenantID, variantSKU, 10.0)

	// Crear repositorio
	repo := persistence.NewPostgresStockEntryRepository(db)

	// Ejecutar venta atómica
	stockEntry, err := repo.ProcessSaleAtomic(
		context.Background(),
		tenantID,
		variantSKU,
		3.0,
		"TEST-SALE-001",
	)

	// Verificaciones
	require.NoError(t, err)
	require.NotNil(t, stockEntry)
	assert.Equal(t, variantSKU, stockEntry.VariantSKU)
	assert.Equal(t, 3.0, stockEntry.Quantity)
	assert.Equal(t, entity.EntryTypeSale, stockEntry.EntryType)

	// Verificar stock actualizado
	availability := getAvailability(t, db, tenantID, variantSKU)
	assert.Equal(t, 7.0, availability.AvailableQuantity, "Stock should be 10 - 3 = 7")
	assert.Equal(t, 7.0, availability.TotalQuantity)
}

// TestProcessSaleAtomic_StockNotInitialized verifica error cuando no hay stock inicial
func TestProcessSaleAtomic_StockNotInitialized(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tenantID := uuid.New()
	variantSKU := "NEVER-EXISTED-SKU"

	repo := persistence.NewPostgresStockEntryRepository(db)

	// Intentar vender producto sin stock inicial
	_, err := repo.ProcessSaleAtomic(
		context.Background(),
		tenantID,
		variantSKU,
		1.0,
		"TEST-SALE",
	)

	// Debe fallar con error específico
	require.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrStockNotInitialized)
}

// TestProcessSaleAtomic_InsufficientStock verifica error cuando no hay stock suficiente
func TestProcessSaleAtomic_InsufficientStock(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tenantID := uuid.New()
	variantSKU := "TEST-SKU-002"

	// Setup: stock inicial de 5 unidades
	setupInitialStock(t, db, tenantID, variantSKU, 5.0)

	repo := persistence.NewPostgresStockEntryRepository(db)

	// Intentar vender 10 unidades (más de lo disponible)
	_, err := repo.ProcessSaleAtomic(
		context.Background(),
		tenantID,
		variantSKU,
		10.0,
		"TEST-SALE",
	)

	// Debe fallar con error de stock insuficiente
	require.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrInsufficientStock)

	// Verificar que NO se descontó nada (rollback correcto)
	availability := getAvailability(t, db, tenantID, variantSKU)
	assert.Equal(t, 5.0, availability.AvailableQuantity, "Stock should remain unchanged")
}

// TestProcessSaleAtomic_ConcurrentSales_NoRaceCondition 
// Prueba crítica: 3 goroutines intentan vender 3 unidades de un stock de 5
// Solo una venta debe tener éxito, las otras 2 deben fallar con stock insuficiente
func TestProcessSaleAtomic_ConcurrentSales_NoRaceCondition(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tenantID := uuid.New()
	variantSKU := "CONCURRENT-TEST-SKU"

	// Setup: stock inicial de 5 unidades
	setupInitialStock(t, db, tenantID, variantSKU, 5.0)

	repo := persistence.NewPostgresStockEntryRepository(db)

	var wg sync.WaitGroup
	results := make(chan error, 3)

	// Lanzar 3 goroutines que intentan vender 3 unidades cada una
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()

			ctx := context.Background()
			_, err := repo.ProcessSaleAtomic(
				ctx,
				tenantID,
				variantSKU,
				3.0,
				fmt.Sprintf("CONCURRENT-SALE-%d", threadID),
			)
			results <- err
		}(i)
	}

	wg.Wait()
	close(results)

	// Analizar resultados
	successCount := 0
	insufficientCount := 0
	var otherErrors []error

	for err := range results {
		if err == nil {
			successCount++
		} else if errors.Is(err, exception.ErrInsufficientStock) {
			insufficientCount++
		} else {
			otherErrors = append(otherErrors, err)
		}
	}

	// Verificaciones críticas
	assert.Empty(t, otherErrors, "No should have unexpected errors")
	assert.Equal(t, 1, successCount, "Exactly ONE sale should succeed (5 / 3 = 1)")
	assert.Equal(t, 2, insufficientCount, "Exactly TWO sales should fail with insufficient stock")

	// Verificar stock final
	availability := getAvailability(t, db, tenantID, variantSKU)
	assert.Equal(t, 2.0, availability.AvailableQuantity, "Final stock should be 5 - 3 = 2")
	assert.Equal(t, 2.0, availability.TotalQuantity)
}

// TestProcessSaleAtomic_MultipleSequentialSales verifica ventas secuenciales
func TestProcessSaleAtomic_MultipleSequentialSales(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tenantID := uuid.New()
	variantSKU := "SEQUENTIAL-TEST-SKU"

	// Setup: stock inicial de 100 unidades
	setupInitialStock(t, db, tenantID, variantSKU, 100.0)

	repo := persistence.NewPostgresStockEntryRepository(db)
	ctx := context.Background()

	// Vender 10 veces, 8 unidades cada vez
	for i := 0; i < 10; i++ {
		_, err := repo.ProcessSaleAtomic(
			ctx,
			tenantID,
			variantSKU,
			8.0,
			fmt.Sprintf("SALE-%d", i),
		)
		require.NoError(t, err, "Sale %d should succeed", i)
	}

	// Verificar stock final
	availability := getAvailability(t, db, tenantID, variantSKU)
	assert.Equal(t, 20.0, availability.AvailableQuantity, "Stock should be 100 - (10 * 8) = 20")

	// La venta 11 de 25 unidades debe fallar
	_, err := repo.ProcessSaleAtomic(ctx, tenantID, variantSKU, 25.0, "SALE-11")
	assert.ErrorIs(t, err, exception.ErrInsufficientStock)
}

// TestCompensateSale_Success verifica que la compensación funciona correctamente
func TestCompensateSale_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tenantID := uuid.New()
	variantSKU := "COMPENSATE-TEST-SKU"

	// Setup: stock inicial de 10
	setupInitialStock(t, db, tenantID, variantSKU, 10.0)

	repo := persistence.NewPostgresStockEntryRepository(db)
	ctx := context.Background()

	// 1. Vender 3 unidades
	stockEntry, err := repo.ProcessSaleAtomic(ctx, tenantID, variantSKU, 3.0, "SALE-TO-COMPENSATE")
	require.NoError(t, err)

	// Verificar stock = 7
	avail := getAvailability(t, db, tenantID, variantSKU)
	assert.Equal(t, 7.0, avail.AvailableQuantity)

	// 2. Compensar la venta (revertir)
	err = repo.CompensateSale(ctx, tenantID, stockEntry.ID, "order_creation_failed")
	require.NoError(t, err)

	// 3. Verificar que el stock volvió a 10
	avail = getAvailability(t, db, tenantID, variantSKU)
	assert.Equal(t, 10.0, avail.AvailableQuantity, "Stock should be restored after compensation")
}

// TestCompensateSale_OnlyForSaleEntries verifica que solo se pueden compensar ventas
func TestCompensateSale_OnlyForSaleEntries(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tenantID := uuid.New()
	variantSKU := "COMPENSATE-FAIL-TEST"

	// Setup: crear un movimiento de compra (no venta)
	setupInitialStock(t, db, tenantID, variantSKU, 10.0)

	// Obtener el ID del stock_entry inicial (tipo 'initial_stock')
	var entryID uuid.UUID
	err := db.QueryRow(`
		SELECT id FROM stock_entries 
		WHERE tenant_id = $1 AND variant_sku = $2 
		ORDER BY created_at ASC LIMIT 1
	`, tenantID, variantSKU).Scan(&entryID)
	require.NoError(t, err)

	repo := persistence.NewPostgresStockEntryRepository(db)

	// Intentar compensar un entry que NO es tipo 'sale'
	err = repo.CompensateSale(context.Background(), tenantID, entryID, "test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "can only compensate sale entries")
}

// ========== HELPERS ==========

func setupTestDB(t *testing.T) *sql.DB {
	// Conectar a base de datos de test
	// Ajustar según tu configuración de testing
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=stock_test sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	// Limpiar tablas antes de cada test
	_, err = db.Exec("TRUNCATE TABLE stock_entries, stock_availability CASCADE")
	require.NoError(t, err)

	return db
}

func setupInitialStock(t *testing.T, db *sql.DB, tenantID uuid.UUID, variantSKU string, quantity float64) {
	entryID := uuid.New()

	// Insertar stock_entry inicial
	_, err := db.Exec(`
		INSERT INTO stock_entries (
			id, tenant_id, variant_sku, product_sku, 
			entry_type, quantity, status, is_active, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, 'confirmed', true, NOW(), NOW())
	`, entryID, tenantID, variantSKU, variantSKU, "initial_stock", quantity)
	require.NoError(t, err)

	// Esperar a que el trigger actualice stock_availability
	time.Sleep(50 * time.Millisecond)
}

func getAvailability(t *testing.T, db *sql.DB, tenantID uuid.UUID, variantSKU string) struct {
	AvailableQuantity float64
	TotalQuantity     float64
} {
	var result struct {
		AvailableQuantity float64
		TotalQuantity     float64
	}

	err := db.QueryRow(`
		SELECT available_quantity, total_quantity 
		FROM stock_availability 
		WHERE tenant_id = $1 AND variant_sku = $2
	`, tenantID, variantSKU).Scan(&result.AvailableQuantity, &result.TotalQuantity)

	require.NoError(t, err)
	return result
}
