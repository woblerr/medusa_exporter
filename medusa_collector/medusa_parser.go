package medusa_collector

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type setUpMetricValueFunType func(metric *prometheus.GaugeVec, value float64, labels ...string) error

var execCommand = exec.Command

const (
	// https://golang.org/pkg/time/#Time.Format
	layout            = "2006-01-02 15:04:05"
	noneLabel         = "none"
	noPrefixLabel     = "no-prefix"
	fullLabel         = "full"
	differentialLabel = "differential"
)

type backupStruct struct {
	backupType string
	finished   int64
	numObjects int
	size       int64
	started    int64
}

type lastBackupsStruct struct {
	full         backupStruct
	differential backupStruct
}

// Medusa specific command options.
func returnDefaultExecArgs() []string {
	// Base exec arguments.
	defaultArgs := []string{"list-backups", "--output", "json"}
	return defaultArgs
}

// Medusa global options.
func returnConfigExecArgs(config string) []string {
	var configArgs []string
	switch config {
	case "":
		configArgs = []string{}
	default:
		// Use specific config.
		configArgs = []string{"--config-file", config}
	}
	return configArgs
}

// Medusa global options.
func returnPrefixExecArgs(prefix string) []string {
	var prefixArgs []string
	switch prefix {
	case "":
		prefixArgs = []string{}
	default:
		prefixArgs = []string{"--prefix", prefix}
	}
	return prefixArgs
}

func concatExecArgs(slices [][]string) []string {
	tmp := []string{}
	for _, s := range slices {
		tmp = append(tmp, s...)
	}
	return tmp
}

func getInfoData(config, prefix string, logger *slog.Logger) ([]byte, error) {
	app := "medusa"
	// Don't change the order of arguments.
	// See medusa help:
	// medusa [OPTIONS] COMMAND [ARGS]...
	args := [][]string{
		returnConfigExecArgs(config),
		returnPrefixExecArgs(prefix),
		returnDefaultExecArgs(),
	}
	// Finally arguments for exec command.
	concatArgs := concatExecArgs(args)
	cmd := execCommand(app, concatArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	// If error occurs, log stderr and return error.
	if err != nil {
		logger.Error(
			"Medusa message",
			"msg", stderr.String(),
		)
		return nil, err
	}
	// If stderr from medusa is not empty,
	// log messages with info level.
	// Medusa by default outputs to stderr.
	if stderr.Len() > 0 {
		logger.Info(
			"Medusa message",
			"msg", stderr.String(),
		)
	}
	return stdout.Bytes(), err
}

func parseResult(output []byte) ([]backup, error) {
	var backups []backup
	err := json.Unmarshal(output, &backups)
	return backups, err
}

func setUpMetricValue(metric *prometheus.GaugeVec, value float64, labels ...string) error {
	metricVec, err := metric.GetMetricWithLabelValues(labels...)
	if err != nil {
		return err
	}
	// The situation should be handled by the prometheus libraries.
	// But, anything is possible.
	if metricVec == nil {
		err := errors.New("metric is nil")
		return err
	}
	metricVec.Set(value)
	return nil
}

func setUpMetric(metric *prometheus.GaugeVec, metricName string, value float64, setUpMetricValueFun setUpMetricValueFunType, logger *slog.Logger, labels ...string) {
	logger.Debug(
		"Set up metric",
		"metric", metricName,
		"value", value,
		"labels", strings.Join(labels, ","),
	)
	err := setUpMetricValueFun(metric, value, labels...)
	if err != nil {
		logger.Error(
			"Metric set up failed",
			"metric", metricName,
			"err", err,
		)
	}
}

func resetMetrics() {
	resetBackupMetrics()
	resetBackupLastMetrics()
	resetExporterMetrics()
}

func initLastBackupStruct() lastBackupsStruct {
	lastBackups := lastBackupsStruct{}
	lastBackups.full.backupType = fullLabel
	lastBackups.differential.backupType = differentialLabel
	return lastBackups
}

func compareLastBackups(lastBackups *lastBackupsStruct, backupData backup) {
	switch backupData.BackupType {
	case fullLabel:
		if backupData.Started > lastBackups.full.started {
			lastBackups.full.started = backupData.Started
			lastBackups.full.finished = backupData.Finished
			lastBackups.full.size = backupData.Size
			lastBackups.full.numObjects = backupData.NumObjects
		}
		if backupData.Started > lastBackups.differential.started {
			lastBackups.differential.started = backupData.Started
			lastBackups.differential.finished = backupData.Finished
			lastBackups.differential.size = backupData.Size
			lastBackups.differential.numObjects = backupData.NumObjects
		}
	case differentialLabel:
		if backupData.Started > lastBackups.differential.started {
			lastBackups.differential.started = backupData.Started
			lastBackups.differential.finished = backupData.Finished
			lastBackups.differential.size = backupData.Size
			lastBackups.differential.numObjects = backupData.NumObjects
		}
		// If no full backup exists yet, use differential backup data for full backup metrics.
		// For Medusa differential backup not depends on full backup existence.
		// If only differential backups exist, full backup metrics also will be set.
		if lastBackups.full.started == 0 {
			lastBackups.full.started = backupData.Started
			lastBackups.full.finished = backupData.Finished
			lastBackups.full.size = backupData.Size
			lastBackups.full.numObjects = backupData.NumObjects
		}
	}
}
