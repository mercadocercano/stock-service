package logging

import (
	"io"

	"stock/src/stock_entry/domain/port"

	sharedlog "github.com/hornosg/go-shared/infrastructure/logging"
)

// StockLogger implementa port.StockEventLogger emitiendo una línea JSON canónica
// (ADR-001) por evento, delegando el envelope (ts/level/service/event + campos flat
// omitempty) en go-shared CanonicalLogger (>= v0.8.0). El mapeo struct→fields y las
// reglas de nivel por evento viven acá; el formato canónico es compartido por la flota.
type StockLogger struct {
	canonical *sharedlog.CanonicalLogger
}

// NewStockLogger crea el adapter escribiendo a stdout. El service se fija acá, nunca por-call.
func NewStockLogger(service string) *StockLogger {
	return &StockLogger{canonical: sharedlog.NewCanonicalLogger(service)}
}

// NewStockLoggerWithWriter permite inyectar un io.Writer (tests).
func NewStockLoggerWithWriter(service string, w io.Writer) *StockLogger {
	return &StockLogger{canonical: sharedlog.NewCanonicalLoggerWithWriter(service, w)}
}

// levelFor aplica las reglas de nivel del ADR-001 por tipo de evento.
func levelFor(event string) string {
	switch event {
	case "stock.entry_created":
		return "info"
	case "stock.bulk_entry_created":
		return "info"
	case "stock.sale_processed":
		return "info"
	case "stock.sale_rejected":
		return "warn"
	case "stock.compensated":
		return "info"
	case "stock.compensate_failed":
		return "error"
	case "stock.adjusted":
		return "info"
	default:
		return "info"
	}
}

func (l *StockLogger) Log(e port.StockEvent) {
	fields := map[string]any{
		"tenant_id":      e.TenantID,
		"user_id":        e.UserID,
		"sku":            e.SKU,
		"stock_entry_id": e.StockEntryID,
		"entry_type":     e.EntryType,
		"reference":      e.Reference,
		"reason":         e.Reason,
	}
	if e.Quantity != 0 {
		fields["quantity"] = e.Quantity
	}
	l.canonical.Emit(levelFor(e.Event), e.Event, fields)
}
