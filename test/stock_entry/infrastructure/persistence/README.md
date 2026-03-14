# Tests de Transacciones Atómicas - HITO D

**Nota:** Estos tests tienen build tag `integration` y se omiten por defecto (`go test ./...`). El flujo ProcessSaleAtomic está cubierto por E2E (`test_order_confirm_stock_e2e.py`, `test-pos-sale-complete-dto.sh`). Para ejecutarlos: `go test -tags=integration ./test/stock_entry/infrastructure/persistence/...` (requiere `stock_test` creada).

## 🎯 Objetivo

Validar que la operación `ProcessSaleAtomic` elimina completamente las race conditions en ventas concurrentes usando `SELECT FOR UPDATE` y transacciones.

## 📋 Tests Implementados

### 1. `TestProcessSaleAtomic_Success`
Valida el caso feliz de una venta atómica exitosa.

**Escenario:**
- Stock inicial: 10 unidades
- Venta: 3 unidades
- Stock final esperado: 7 unidades

**Verifica:**
- ✅ Stock entry creado correctamente
- ✅ Stock actualizado por trigger
- ✅ Sin errores

---

### 2. `TestProcessSaleAtomic_StockNotInitialized`
Valida que NO se puede vender un producto sin stock inicializado.

**Escenario:**
- Producto nunca tuvo movimientos de stock
- Intento de venta: 1 unidad

**Verifica:**
- ✅ Error: `ErrStockNotInitialized`
- ✅ No se crea movimiento de venta
- ✅ No se materializa fila fantasma en stock_availability

**Principio arquitectónico:** Stock Service solo vende productos con existencia física confirmada.

---

### 3. `TestProcessSaleAtomic_InsufficientStock`
Valida que NO se puede vender más stock del disponible.

**Escenario:**
- Stock inicial: 5 unidades
- Intento de venta: 10 unidades

**Verifica:**
- ✅ Error: `ErrInsufficientStock`
- ✅ Stock NO se descuenta (rollback correcto)
- ✅ Stock permanece en 5 unidades

---

### 4. `TestProcessSaleAtomic_ConcurrentSales_NoRaceCondition` ⚡ **CRÍTICO**

**ESTE ES EL TEST CLAVE QUE VALIDA QUE NO HAY RACE CONDITION.**

**Escenario:**
- Stock inicial: 5 unidades
- 3 goroutines intentan vender 3 unidades cada una simultáneamente
- Matemáticamente: 5 / 3 = 1 venta posible (sobran 2 unidades)

**Verifica:**
- ✅ **Exactamente 1 venta exitosa**
- ✅ **Exactamente 2 ventas fallan con `ErrInsufficientStock`**
- ✅ Stock final = 2 unidades (5 - 3)
- ✅ No hay sobreventa (stock negativo)
- ✅ No hay errores de deadlock o timeout

**Comportamiento esperado:**

```
Thread A: BEGIN → SELECT FOR UPDATE → LOCK ROW (available=5)
Thread B: BEGIN → SELECT FOR UPDATE → WAIT (bloqueado por A)
Thread C: BEGIN → SELECT FOR UPDATE → WAIT (bloqueado por A)

Thread A: Valida OK (5 >= 3) → INSERT sale → COMMIT
          Trigger actualiza stock a 2

Thread B: DESBLOQUEA → SELECT devuelve available=2
          Valida FAIL (2 < 3) → ROLLBACK → Error

Thread C: DESBLOQUEA → SELECT devuelve available=2
          Valida FAIL (2 < 3) → ROLLBACK → Error
```

**Sin `SELECT FOR UPDATE` (código viejo):**
```
Thread A: SELECT available → 5 (sin lock)
Thread B: SELECT available → 5 (sin lock)
Thread C: SELECT available → 5 (sin lock)

Thread A: Valida OK → INSERT -3
Thread B: Valida OK → INSERT -3
Thread C: Valida OK → INSERT -3

Resultado: Stock = -4 ❌ SOBREVENTA
```

---

### 5. `TestProcessSaleAtomic_MultipleSequentialSales`
Valida ventas secuenciales sin concurrencia.

**Escenario:**
- Stock inicial: 100 unidades
- 10 ventas de 8 unidades cada una
- Intento de venta 11: 25 unidades

**Verifica:**
- ✅ 10 ventas exitosas
- ✅ Stock final: 20 unidades (100 - 80)
- ✅ Venta 11 falla con `ErrInsufficientStock`

---

### 6. `TestCompensateSale_Success`
Valida que la compensación revierte correctamente una venta.

**Escenario:**
- Stock inicial: 10 unidades
- Venta: 3 unidades → Stock = 7
- Compensación (rollback): +3 unidades

**Verifica:**
- ✅ Stock vuelve a 10 unidades
- ✅ Se crea movimiento tipo `return`
- ✅ Referencia correcta: `COMPENSATION-{sale_id}`

**Caso de uso:** Si falla la persistencia de la orden/sale en PIM, compensamos el descuento de stock.

---

### 7. `TestCompensateSale_OnlyForSaleEntries`
Valida que solo se pueden compensar movimientos de tipo `sale`.

**Escenario:**
- Intento de compensar un movimiento tipo `initial_stock`

**Verifica:**
- ✅ Error: "can only compensate sale entries"
- ✅ No se crea movimiento de compensación

---

## 🚀 Cómo Ejecutar

### Configuración de Base de Datos de Test

```bash
# Crear base de datos de test
createdb stock_test

# Ejecutar migraciones
psql -d stock_test -f migrations/001_initial_schema.sql
psql -d stock_test -f migrations/004_create_stock_entries.sql
psql -d stock_test -f migrations/005_rename_product_sku_to_variant_sku.sql
psql -d stock_test -f migrations/006_fix_reserved_quantity_trigger.sql
psql -d stock_test -f migrations/007_fix_sale_calculation_trigger.sql
```

### Ejecutar Tests

```bash
# Requiere build tag integration (tests excluidos por defecto)
cd services/stock-service

# Todos los tests de integración (requiere stock_test creada)
go test -tags=integration ./test/stock_entry/infrastructure/persistence/... -v

# Solo el test crítico de concurrencia
go test -tags=integration ./test/stock_entry/infrastructure/persistence/... -v -run TestProcessSaleAtomic_ConcurrentSales

# Con race detector (recomendado)
go test -tags=integration ./test/stock_entry/infrastructure/persistence/... -v -race

# Múltiples ejecuciones para detectar race conditions intermitentes
go test -tags=integration ./test/stock_entry/infrastructure/persistence/... -v -count=10 -run TestProcessSaleAtomic_ConcurrentSales
```

### Variables de Entorno (opcional)

```bash
export STOCK_TEST_DB="host=localhost port=5432 user=postgres password=postgres dbname=stock_test sslmode=disable"
```

---

## 📊 Resultados Esperados

### Test Exitoso
```
=== RUN   TestProcessSaleAtomic_ConcurrentSales_NoRaceCondition
--- PASS: TestProcessSaleAtomic_ConcurrentSales_NoRaceCondition (0.15s)
    process_sale_atomic_test.go:142: Exactly ONE sale should succeed (5 / 3 = 1)
    process_sale_atomic_test.go:143: Exactly TWO sales should fail with insufficient stock
    process_sale_atomic_test.go:147: Final stock should be 5 - 3 = 2
PASS
```

### Test Fallido (Race Condition Detectada)
```
=== RUN   TestProcessSaleAtomic_ConcurrentSales_NoRaceCondition
--- FAIL: TestProcessSaleAtomic_ConcurrentSales_NoRaceCondition (0.12s)
    process_sale_atomic_test.go:142: Expected 1 successful sale, got 2
    process_sale_atomic_test.go:147: Final stock should be 2, got -1
FAIL
```

---

## 🔍 Debugging

### Ver logs de transacciones en PostgreSQL

```sql
-- Habilitar logging de transacciones
ALTER SYSTEM SET log_statement = 'all';
SELECT pg_reload_conf();

-- Ver locks activos durante test
SELECT 
    pid,
    locktype,
    mode,
    granted,
    relation::regclass
FROM pg_locks
WHERE NOT granted;
```

### Forzar deadlock (test de robustez)

Modificar timeout de transacción:

```sql
SET lock_timeout = '1s';
```

---

## 🧪 Coverage Esperado

```bash
go test ./test/stock_entry/infrastructure/persistence/... -cover

PASS
coverage: 87.5% of statements
```

---

## 📚 Referencias

- PostgreSQL SELECT FOR UPDATE: https://www.postgresql.org/docs/current/sql-select.html#SQL-FOR-UPDATE-SHARE
- Go race detector: https://go.dev/blog/race-detector
- ACID Transactions: https://en.wikipedia.org/wiki/ACID
