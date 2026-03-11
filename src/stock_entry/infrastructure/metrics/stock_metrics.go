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
)

func init() {
	prometheus.MustRegister(StockInsufficient)
}
