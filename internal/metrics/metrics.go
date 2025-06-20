package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	PagesFetched = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "crawler_pages_fetched_total",
		Help: "Total number of pages successfully fetched",
	})
	BytesFetched = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "crawler_bytes_fetched_total",
		Help: "Total bytes downloaded",
	})
)

func init() {
	prometheus.MustRegister(PagesFetched, BytesFetched)
}
