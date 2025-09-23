package builder

import (
	"github.com/prometheus/client_golang/prometheus"
)

const metricsNamespace = "k6build"

type metrics struct {
	requestCounter       prometheus.Counter
	requestTimeHistogram prometheus.Histogram
	buildCounter         prometheus.Counter
	storeHitsCounter     prometheus.Counter
	buildsFailedCounter  prometheus.Counter
	buildsInvalidCounter prometheus.Counter
	buildTimeHistogram   prometheus.Histogram
}

func newMetrics() *metrics {
	requestCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Name:      "requests_total",
		Help:      "The total number of builds requests",
	})

	requestDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: metricsNamespace,
		Name:      "request_duration_seconds",
		Help:      "The duration of the build request in seconds",
		Buckets:   []float64{0.1, 0.5, 1, 2.5, 5, 10, 20, 30, 60, 120, 300},
	})

	buildCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Name:      "builds_total",
		Help:      "The total number of builds",
	})

	buildsFailedCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Name:      "builds_failed_total",
		Help:      "The total number of failed builds",
	})

	buildsInvalidCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Name:      "builds_invalid_total",
		Help:      "The total number of builds with invalid parameters",
	})

	storeHitsCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Name:      "object_store_hits_total",
		Help:      "The total number of object store hits",
	})

	buildTimeHistogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: metricsNamespace,
		Name:      "build_duration_seconds",
		Help:      "The duration of the build in seconds",
		Buckets:   []float64{1, 2.5, 5, 10, 20, 30, 60, 120, 300},
	})

	return &metrics{
		requestCounter:       requestCounter,
		requestTimeHistogram: requestDuration,
		buildCounter:         buildCounter,
		buildsFailedCounter:  buildsFailedCounter,
		buildsInvalidCounter: buildsInvalidCounter,
		storeHitsCounter:     storeHitsCounter,
		buildTimeHistogram:   buildTimeHistogram,
	}
}

func (m *metrics) register(registerer prometheus.Registerer) error {
	if err := registerer.Register(m.requestCounter); err != nil {
		return err
	}

	if err := registerer.Register(m.requestTimeHistogram); err != nil {
		return err
	}

	if err := registerer.Register(m.buildCounter); err != nil {
		return err
	}

	if err := registerer.Register(m.buildsFailedCounter); err != nil {
		return err
	}

	if err := registerer.Register(m.buildsInvalidCounter); err != nil {
		return err
	}

	if err := registerer.Register(m.storeHitsCounter); err != nil {
		return err
	}

	if err := registerer.Register(m.buildTimeHistogram); err != nil {
		return err
	}

	return nil
}
