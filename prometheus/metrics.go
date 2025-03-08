package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ProductInventoryGauge tracks the inventory count for each product
	ProductInventoryGauge *prometheus.GaugeVec

	// ScrapeDurationSeconds tracks the time spent scraping product pages
	ScrapeDurationSeconds *prometheus.HistogramVec

	// ScrapeSuccessCounter tracks successful scrapes
	ScrapeSuccessCounter *prometheus.CounterVec

	// ScrapeFailureCounter tracks failed scrapes
	ScrapeFailureCounter *prometheus.CounterVec
)

// InitMetrics initializes all Prometheus metrics
func InitMetrics() {
	ProductInventoryGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "microcenter_product_inventory_count",
			Help: "The current inventory count for a product at a specific store",
		},
		[]string{"store_id", "product_name", "product_url"},
	)

	ScrapeDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "microcenter_scrape_duration_seconds",
			Help:    "Duration of inventory scrape in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"product_url"},
	)

	ScrapeSuccessCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "microcenter_scrape_success_total",
			Help: "Total number of successful scrapes",
		},
		[]string{"product_url"},
	)

	ScrapeFailureCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "microcenter_scrape_failure_total",
			Help: "Total number of failed scrapes",
		},
		[]string{"product_url", "error"},
	)
}
