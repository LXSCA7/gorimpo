package telemetry

import (
	"github.com/LXSCA7/gorimpo/internal/core/ports"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// compile-time interface check
var _ ports.Metrics = (*PrometheusMetrics)(nil)

type PrometheusMetrics struct {
	discarded    *prometheus.CounterVec
	valid        *prometheus.CounterVec
	scrapedTotal *prometheus.CounterVec
	sentTotal    *prometheus.CounterVec
}

func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		discarded: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "gorimpo_discarded_total",
			Help: "Total discarded offers by term and reason",
		}, []string{"term", "reason"}),

		valid: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "gorimpo_valid_total",
			Help: "Total valid offers found by term",
		}, []string{"term"}),

		scrapedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gorimpo_scraped_total",
				Help: "Total raw items scraped from the platform",
			},
			[]string{"term"}),

		sentTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gorimpo_sent_total",
				Help: "Total offers successfully sent to Telegram",
			},
			[]string{"term"}),
	}
}

func (p *PrometheusMetrics) RecordDiscarded(term, reason string, count int) {
	p.discarded.WithLabelValues(term, reason).Add(float64(count))
}

func (p *PrometheusMetrics) RecordValid(term string, count int) {
	p.valid.WithLabelValues(term).Add(float64(count))
}

func (p *PrometheusMetrics) RecordScraped(term string, count int) {
	p.scrapedTotal.WithLabelValues(term).Add(float64(count))
}

func (p *PrometheusMetrics) RecordSent(term string, count int) {
	p.sentTotal.WithLabelValues(term).Add(float64(count))
}
