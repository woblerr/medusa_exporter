package medusa_collector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	medusaExporterStatusMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_exporter_status",
		Help: "Medusa exporter get data status.",
	},
		[]string{
			"prefix",
		})
)

// Set exporter metrics:
//   - medusa_exporter_status
func getExporterStatusMetrics(getDataStatus bool, prefix string, setUpMetricValueFun setUpMetricValueFunType, logger *slog.Logger) {
	if prefix == "" {
		prefix = noPrefixLabel
	}
	setUpMetric(
		medusaExporterStatusMetric,
		"medusa_exporter_status",
		convertBoolToFloat64(getDataStatus),
		setUpMetricValueFun,
		logger,
		prefix,
	)
}

func resetExporterMetrics() {
	medusaExporterStatusMetric.Reset()
}
