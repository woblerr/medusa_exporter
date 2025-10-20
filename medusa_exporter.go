package main

import (
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	version_collector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"github.com/woblerr/medusa_exporter/medusa_collector"
)

const exporterName = "medusa_exporter"

func main() {
	var (
		webPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		webAdditionalToolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, ":19500")
		collectionInterval        = kingpin.Flag(
			"collect.interval",
			"Collecting metrics interval in seconds.",
		).Default("600").Int()
		medusaCustomConfig = kingpin.Flag(
			"medusa.config-file",
			"Full path to Medusa configuration file.",
		).Default("").String()
		medusaPrefix = kingpin.Flag(
			"medusa.prefix",
			"Prefix for shared storage.",
		).Default("").String()
	)
	// Set logger config.
	promslogConfig := &promslog.Config{}
	// Add flags log.level and log.format from promlog package.
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print(exporterName))
	// Add short help flag.
	kingpin.HelpFlag.Short('h')
	// Load command line arguments.
	kingpin.Parse()
	// Setup signal catching.
	sigs := make(chan os.Signal, 1)
	// Catch  listed signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// Set logger.
	logger := promslog.New(promslogConfig)
	// Method invoked upon seeing signal.
	go func(logger *slog.Logger) {
		s := <-sigs
		logger.Warn(
			"Stopping exporter",
			"name", filepath.Base(os.Args[0]),
			"signal", s)
		os.Exit(1)
	}(logger)
	logger.Info(
		"Starting exporter",
		"name", filepath.Base(os.Args[0]),
		"version", version.Info())
	logger.Info("Build context", "build_context", version.BuildContext())
	if *medusaCustomConfig != "" {
		logger.Info(
			"Custom Medusa configuration file",
			"file", *medusaCustomConfig)
	}
	if *medusaPrefix != "" {
		logger.Info(
			"Collecting metrics for specific prefix in shared storage",
			"prefix", *medusaPrefix)
	}
	// Setup parameters for exporter.
	medusa_collector.SetPromPortAndPath(*webAdditionalToolkitFlags, *webPath)
	logger.Info(
		"Use exporter parameters",
		"endpoint", *webPath,
		"config.file", *webAdditionalToolkitFlags.WebConfigFile,
	)
	// Exporter build info metric
	prometheus.MustRegister(version_collector.NewCollector(exporterName))
	// Start web server.
	medusa_collector.StartPromEndpoint(version.Info(), logger)
	for {
		// Get information form Medusa and set metrics.
		medusa_collector.GetMedusaInfo(
			*medusaCustomConfig,
			*medusaPrefix,
			logger,
		)
		// Sleep for 'collection.interval' seconds.
		time.Sleep(time.Duration(*collectionInterval) * time.Second)
	}
}
