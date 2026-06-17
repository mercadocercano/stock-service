package logging_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"stock/src/stock_entry/domain/port"
	stocklog "stock/src/stock_entry/infrastructure/logging"

	"github.com/stretchr/testify/assert"
)

// ADR-001: cada evento produce UNA línea JSON canónica con envelope ts/level/service/event.
func parseLine(t *testing.T, b []byte) map[string]any {
	t.Helper()
	lines := bytes.Split(bytes.TrimSpace(b), []byte("\n"))
	assert.Len(t, lines, 1, "debe ser exactamente una línea por evento")
	var m map[string]any
	assert.NoError(t, json.Unmarshal(lines[0], &m))
	return m
}

func TestStockLogger_SaleProcessed_EnvelopeAndInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := stocklog.NewStockLoggerWithWriter("stock-test", &buf)

	logger.Log(port.StockEvent{
		Event:        "stock.sale_processed",
		TenantID:     "t-123",
		SKU:          "SKU-001",
		StockEntryID: "e-456",
		Quantity:     3,
		Reference:    "POS-abc-123",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "stock.sale_processed", line["event"])
	assert.Equal(t, "info", line["level"])
	assert.Equal(t, "stock-test", line["service"])
	assert.NotEmpty(t, line["ts"], "ts (RFC3339 UTC) siempre presente")
	assert.Equal(t, "t-123", line["tenant_id"])
	assert.Equal(t, "SKU-001", line["sku"])
	assert.Equal(t, "e-456", line["stock_entry_id"])
	assert.Equal(t, float64(3), line["quantity"])
	assert.Equal(t, "POS-abc-123", line["reference"])
}

func TestStockLogger_SaleRejected_WarnLevel_OmitsEmptyFields(t *testing.T) {
	var buf bytes.Buffer
	logger := stocklog.NewStockLoggerWithWriter("stock-test", &buf)

	logger.Log(port.StockEvent{
		Event:    "stock.sale_rejected",
		TenantID: "t-123",
		SKU:      "SKU-001",
		Reason:   "insufficient_stock",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "warn", line["level"])
	assert.Equal(t, "insufficient_stock", line["reason"])
	// omitempty: campos vacíos no aparecen
	_, hasEntryID := line["stock_entry_id"]
	assert.False(t, hasEntryID, "stock_entry_id vacío debe omitirse")
	_, hasQty := line["quantity"]
	assert.False(t, hasQty, "quantity=0 debe omitirse")
}

func TestStockLogger_Compensated_InfoLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := stocklog.NewStockLoggerWithWriter("stock-test", &buf)

	logger.Log(port.StockEvent{
		Event:        "stock.compensated",
		TenantID:     "t-1",
		StockEntryID: "e-789",
		Reason:       "pos_sale_persistence_failed",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "info", line["level"])
	assert.Equal(t, "stock.compensated", line["event"])
	assert.Equal(t, "e-789", line["stock_entry_id"])
}

func TestStockLogger_CompensateFailed_ErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := stocklog.NewStockLoggerWithWriter("stock-test", &buf)

	logger.Log(port.StockEvent{Event: "stock.compensate_failed", TenantID: "t-1", StockEntryID: "e-999", Reason: "db timeout"})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "error", line["level"])
	assert.Equal(t, "stock.compensate_failed", line["event"])
}

func TestStockLogger_EntryCreated_InfoLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := stocklog.NewStockLoggerWithWriter("stock-test", &buf)

	logger.Log(port.StockEvent{
		Event:        "stock.entry_created",
		TenantID:     "t-123",
		SKU:          "SKU-002",
		StockEntryID: "e-111",
		Quantity:     10,
		EntryType:    "purchase",
		Reference:    "PO-001",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "info", line["level"])
	assert.Equal(t, "stock.entry_created", line["event"])
	assert.Equal(t, "SKU-002", line["sku"])
	assert.Equal(t, float64(10), line["quantity"])
	assert.Equal(t, "purchase", line["entry_type"])
	assert.Equal(t, "PO-001", line["reference"])
}
