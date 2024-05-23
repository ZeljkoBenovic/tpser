package prom

import (
	"fmt"
	"github.com/ZeljkoBenovic/tpser/pkg/conf"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type Prom struct {
	conf conf.Conf
	log  logger.Logger

	metrics metrics
}

type metrics struct {
	transactionHttpRequestDurationHistogram prometheus.Histogram
	transactionNumberPerInterval            prometheus.Gauge
	transactionSendInterval                 prometheus.Gauge
	transactionErrorCount                   prometheus.Counter
}

func NewPrometheus(conf conf.Conf, log logger.Logger) *Prom {
	return &Prom{
		conf: conf,
		log:  log,
		metrics: metrics{
			transactionHttpRequestDurationHistogram: promauto.NewHistogram(prometheus.HistogramOpts{
				Namespace: "tpser",
				Name:      "tx_request_duration_milliseconds",
				Help:      "transaction HTTP request duration in milliseconds",
			}),
			transactionNumberPerInterval: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace: "tpser",
				Name:      "tx_number_per_interval",
				Help:      "total number of transactions sent per defined interval",
			}),
			transactionSendInterval: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace: "tpser",
				Name:      "tx_send_interval_seconds",
				Help:      "transaction send interval in seconds",
			}),
			transactionErrorCount: promauto.NewCounter(prometheus.CounterOpts{
				Namespace: "tpser",
				Name:      "tx_send_error_count",
				Help:      "the number of transaction send errors",
			}),
		},
	}
}

func (p *Prom) ServeHTTP() error {
	p.log.Info("Starting metrics server", "port", p.conf.MetricsPort)

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", p.conf.MetricsPort), nil); err != nil {
		return err
	}

	return nil
}

func (p *Prom) ObserveTxRequestDuration(observable float64) {
	p.metrics.transactionHttpRequestDurationHistogram.Observe(observable)
}

func (p *Prom) SetTxNumberPerInterval(txNumber float64) {
	p.metrics.transactionNumberPerInterval.Set(txNumber)
}

func (p *Prom) SetTxSendInterval(txSendInterval float64) {
	p.metrics.transactionSendInterval.Set(txSendInterval)
}

func (p *Prom) IncreaseTxErrorCount() {
	p.metrics.transactionErrorCount.Inc()
}
