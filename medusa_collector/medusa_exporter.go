package medusa_collector

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/exporter-toolkit/web"
)

var (
	webFlagsConfig web.FlagConfig
	webEndpoint    string
)

// SetPromPortAndPath sets HTTP endpoint parameters
// from command line arguments:
// 'web.telemetry-path',
// 'web.listen-address',
// 'web.config.file',
// 'web.systemd-socket' (Linux only)
func SetPromPortAndPath(flagsConfig web.FlagConfig, endpoint string) {
	webFlagsConfig = flagsConfig
	webEndpoint = endpoint
}

// StartPromEndpoint run HTTP endpoint
func StartPromEndpoint(version string, logger *slog.Logger) {
	go func(logger *slog.Logger) {
		if webEndpoint == "" {
			logger.Error("Metric endpoint is empty", "endpoint", webEndpoint)
		}
		http.Handle(webEndpoint, promhttp.Handler())
		if webEndpoint != "/" {
			landingConfig := web.LandingConfig{
				Name:        "Medusa exporter",
				Description: "Prometheus exporter for Medusa for Apache Cassandra",
				HeaderColor: "#476b6b",
				Version:     version,
				Profiling:   "false",
				Links: []web.LandingLinks{
					{
						Address: webEndpoint,
						Text:    "Metrics",
					},
				},
			}
			landingPage, err := web.NewLandingPage(landingConfig)
			if err != nil {
				logger.Error("Error creating landing page", "err", err)
				os.Exit(1)
			}
			http.Handle("/", landingPage)
		}
		server := &http.Server{
			ReadHeaderTimeout: 5 * time.Second,
		}
		if err := web.ListenAndServe(server, &webFlagsConfig, logger); err != nil {
			logger.Error("Run web endpoint failed", "err", err)
			os.Exit(1)
		}
	}(logger)
}

// GetMedusaInfo get and parse Medusa info and set metrics
func GetMedusaInfo(config, prefix string, logger *slog.Logger) {
	// To calculate the time elapsed since the last completed full or differential backup.
	currentUnixTime := time.Now().Unix()
	lastBackups := initLastBackupStruct()
	// The flag indicates whether it was possible to get data from the Medusa.
	// By default, it's set to true.
	getDataSuccessStatus := true
	backupData, err := getInfoData(config, prefix, logger)
	if err != nil {
		getDataSuccessStatus = false
		logger.Error("Get data from Medusa failed", "err", err)
	}
	parseBackupData, err := parseResult(backupData)
	if err != nil {
		getDataSuccessStatus = false
		logger.Error("Parse JSON failed", "err", err)
	}
	if len(parseBackupData) == 0 {
		logger.Warn("No backup data returned")
	}
	// Reset metrics.
	resetMetrics()
	getExporterStatusMetrics(getDataSuccessStatus, prefix, setUpMetricValue, logger)
	for _, singleBackup := range parseBackupData {
		getBackupMetrics(singleBackup, prefix, setUpMetricValue, logger)
		// Only completed backups are considered.
		if singleBackup.Finished > 0 {
			compareLastBackups(&lastBackups, singleBackup)
		}
	}
	// If full backup exists, the values of metrics for differential backups also will be set.
	// If not - metrics won't be set.
	if lastBackups.full.started > 0 {
		getBackupLastMetrics(lastBackups, currentUnixTime, setUpMetricValue, logger)
	}
}
