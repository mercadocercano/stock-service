package port

// StockEvent es el payload canónico para eventos de dominio de stock (ADR-001).
// Campos flat, named. Los nombres comunes (tenant_id, user_id, sku) son idénticos al resto
// de la flota para que el LogQL cross-service funcione. Todos opcionales salvo Event.
type StockEvent struct {
	Event        string // <domain>.<action>_<result>, p.ej. "stock.entry_created"
	TenantID     string
	UserID       string
	SKU          string
	StockEntryID string
	Quantity     float64
	EntryType    string
	Reference    string
	Reason       string
}

// StockEventLogger es el puerto para emitir eventos canónicos de stock.
// El código de aplicación depende de esta interfaz; el adapter (JSON a stdout,
// Loki push, etc.) la implementa. Nunca al revés.
type StockEventLogger interface {
	Log(e StockEvent)
}
