package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// StockInsufficient ventas rechazadas por stock insuficiente
	StockInsufficient = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "stock_insufficient_total",
			Help: "Number of sales rejected due to insufficient stock",
		},
	)

	// MCStockLevel nivel de stock actual por tenant y SKU.
	// Gauge — se actualiza tras cada movimiento de stock.
	// Usado por metrics-gateway KPI: stock-alerts (mc_stock_level < 5).
	MCStockLevel = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mc_stock_level",
			Help: "Current stock level per tenant and SKU",
		},
		[]string{"tenant_id", "sku"},
	)

	// MCStockMovementsTotal cuenta movimientos de stock (ventas, entradas, ajustes).
	// Usado por metrics-gateway para tendencias de rotación.
	MCStockMovementsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mc_stock_movements_total",
			Help: "Total stock movements by tenant and type",
		},
		[]string{"tenant_id", "movement_type"},
	)
)

func init() {
	prometheus.MustRegister(
		StockInsufficient,
		MCStockLevel,
		MCStockMovementsTotal,
	)
}
