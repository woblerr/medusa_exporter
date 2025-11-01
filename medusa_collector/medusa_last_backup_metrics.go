package medusa_collector

import (
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	medusaBackupSinceLastCompletionSecondsMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_backup_since_last_completion_seconds",
		Help: "Time since last full or differential backup completion.",
	},
		[]string{
			"backup_type"})
	medusaBackupLastDurationMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_backup_last_duration_seconds",
		Help: "Backup duration for the last full or differential backup.",
	},
		[]string{
			"backup_type"})
	medusaBackupLastDatabaseSizeMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_backup_last_size_bytes",
		Help: "Backup size for the last full or differential backup.",
	},
		[]string{
			"backup_type"})
	medusaBackupLastObjectsMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_backup_last_objects",
		Help: "Number of objects in backup for the last full or differential backup.",
	},
		[]string{
			"backup_type"})
)

// Set backup metrics:
//   - medusa_backup_last_completion_seconds
//   - medusa_backup_last_duration_seconds
//   - medusa_backup_last_size_bytes
//   - medusa_backup_last_objects
func getBackupLastMetrics(lastBackups lastBackupsStruct, currentUnixTime int64, setUpMetricValueFun setUpMetricValueFunType, logger *slog.Logger) {
	for _, backup := range []backupStruct{lastBackups.full, lastBackups.differential} {
		// Seconds since the last completed backups.
		setUpMetric(
			medusaBackupSinceLastCompletionSecondsMetric,
			"medusa_backup_since_last_completion_seconds",
			time.Unix(currentUnixTime, 0).Sub(time.Unix(backup.finished, 0)).Seconds(),
			setUpMetricValueFun,
			logger,
			backup.backupType,
		)
		// Last backup duration.
		backupDuration := float64(backup.finished - backup.started)
		setUpMetric(
			medusaBackupLastDurationMetric,
			"medusa_backup_last_duration_seconds",
			backupDuration,
			setUpMetricValueFun,
			logger,
			backup.backupType,
		)
		// Last backup size.
		setUpMetric(
			medusaBackupLastDatabaseSizeMetric,
			"medusa_backup_last_size_bytes",
			float64(backup.size),
			setUpMetricValueFun,
			logger,
			backup.backupType,
		)
		// Last backup objects.
		setUpMetric(
			medusaBackupLastObjectsMetric,
			"medusa_backup_last_objects",
			float64(backup.numObjects),
			setUpMetricValueFun,
			logger,
			backup.backupType,
		)
	}
}

func resetBackupLastMetrics() {
	medusaBackupSinceLastCompletionSecondsMetric.Reset()
	medusaBackupLastDurationMetric.Reset()
	medusaBackupLastDatabaseSizeMetric.Reset()
	medusaBackupLastObjectsMetric.Reset()
}
