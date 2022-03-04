package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	up = prometheus.NewDesc(
		"queue_it_up",
		"Was talking to Queue-it successful.",
		nil, nil,
	)
	duration = prometheus.NewDesc(
		"queue_it_collector_collect_duration_seconds",
		"Was talking to Queue-it successful.",
		nil, nil,
	)
)

type collector struct {
	logger     *zap.Logger
	queueitAPI *queueitAPI
}

// newCollector returns a queueitAPI connector
func newCollector(logger *zap.Logger, api *queueitAPI) *collector {
	logger.Debug("newCollector()")
	return &collector{
		logger:     logger,
		queueitAPI: api,
	}
}

// Describe implements Collector
func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	c.logger.Debug("collector.Describe")
	// metrics don't change, we can use this helper method
	prometheus.DescribeByCollect(c, ch)
}

// Collect implements Collector
func (c *collector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Debug("collector.Collect()")

	start := time.Now()
	defer func() {
		// track request duration
		now := time.Now()
		ch <- prometheus.MustNewConstMetric(
			duration,
			prometheus.CounterValue,
			float64(now.Sub(start).Seconds()),
		)
	}()

	// Get metrics
	metrics, err := c.queueitAPI.getMetrics()
	if err != nil {
		c.logger.Error("error", zap.Error(err))
		// Queue-it api is unreachable
		ch <- prometheus.MustNewConstMetric(up, prometheus.GaugeValue, 0)
		return
	}

	// Contacted Queue-it api successfully
	ch <- prometheus.MustNewConstMetric(up, prometheus.GaugeValue, 1)

	// Send metrics
	for _, m := range metrics {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				m.exportedMetricName,
				m.description,
				[]string{"waiting_room_id"},
				nil,
			),
			prometheus.GaugeValue,
			m.value,
			m.waitingRoomID,
		)
	}

	c.logger.Debug("collector.Collect(): Finished collecting")
}
