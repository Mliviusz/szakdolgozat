package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	test_results = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "selenium_test_results",
			Help: "0 mean success, 1 means error",
		},
		[]string{"test_name", "namespace"},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(test_results)
}
