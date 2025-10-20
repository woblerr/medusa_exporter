package medusa_collector

import (
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	medusaBackupInfoMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_backup_info",
		Help: "Backup info.",
	},
		[]string{
			"backup_name",
			"backup_type",
			"prefix",
			"start_time",
		})
	medusaBackupStatusMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_backup_status",
		Help: "Backup status.",
	},
		[]string{
			"backup_name",
			"backup_type"})
	medusaBackupDurationMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_backup_duration_seconds",
		Help: "Backup duration.",
	},
		[]string{
			"backup_name",
			"backup_type",
			"start_time",
			"stop_time"})
	medusaBackupDatabaseSizeMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_backup_size_bytes",
		Help: "Backup size.",
	},
		[]string{
			"backup_name",
			"backup_type"})
	medusaBackupObjectsMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_backup_objects",
		Help: "Number of objects in backup.",
	},
		[]string{
			"backup_name",
			"backup_type"})
	medusaBackupNodesMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_backup_completed_nodes",
		Help: "Number of completed nodes in backup.",
	},
		[]string{
			"backup_name",
			"backup_type"})
	medusaBackupIncompleteNodesMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_backup_incomplete_nodes",
		Help: "Number of incomplete nodes in backup.",
	},
		[]string{
			"backup_name",
			"backup_type"})
	medusaBackupMissingNodesMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_backup_missing_nodes",
		Help: "Number of missing nodes in backup.",
	},
		[]string{
			"backup_name",
			"backup_type"})
	medusaNodeBackupsInfosMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_node_backup_info",
		Help: "Node backup info.",
	},
		[]string{
			"backup_name",
			"backup_type",
			"node_fqdn",
			"prefix",
			"release_version",
			"server_type",
			"start_time"})
	medusaNodeBackupsStatusMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_node_backup_status",
		Help: "Node backup status.",
	},
		[]string{
			"backup_name",
			"backup_type",
			"node_fqdn"})
	medusaNodeBackupDurationMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_node_backup_duration_seconds",
		Help: "Node backup duration.",
	},
		[]string{
			"backup_name",
			"backup_type",
			"node_fqdn",
			"start_time",
			"stop_time"})
	medusaNodeBackupsSizeMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_node_backup_size",
		Help: "Node backup size.",
	},
		[]string{
			"backup_name",
			"backup_type",
			"node_fqdn",
		})
	medusaNodeBackupsObjectsMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "medusa_node_backup_objects",
		Help: "Number of objects in node backup.",
	},
		[]string{
			"backup_name",
			"backup_type",
			"node_fqdn",
		})
)

// Set backup metrics:
//   - medusa_backup_info
//   - medusa_backup_status
//   - medusa_backup_duration_seconds
//   - medusa_backup_size_bytes
//   - medusa_backup_objects
//   - medusa_backup_completed_nodes
//   - medusa_backup_incomplete_nodes
//   - medusa_backup_missing_nodes
//   - medusa_node_backup_info
//   - medusa_node_backup_status
//   - medusa_node_backup_duration_seconds
//   - medusa_node_backup_size
//   - medusa_node_backup_objects
func getBackupMetrics(backupData backup, prefix string, setUpMetricValueFun setUpMetricValueFunType, logger *slog.Logger) {
	var backupDuration, nodeDuration float64
	var backupStopTime, nodeStopTime string

	if prefix == "" {
		prefix = noPrefixLabel
	}
	// Backup info.
	//  1 - info about backup is exist.
	setUpMetric(
		medusaBackupInfoMetric,
		"medusa_backup_info",
		1,
		setUpMetricValueFun,
		logger,
		backupData.Name,
		backupData.BackupType,
		prefix,
		time.Unix(backupData.Started, 0).Format(layout),
	)
	// Backup status.
	//  1 - backup is complete.
	//  0 - backup is not complete.
	setUpMetric(
		medusaBackupStatusMetric,
		"medusa_backup_status",
		convertBoolToFloat64(backupData.Finished > 0),
		setUpMetricValueFun,
		logger,
		backupData.Name,
		backupData.BackupType,
	)
	// Backup duration.
	if backupData.Finished == 0 {
		backupDuration = 0
		backupStopTime = noneLabel
	} else {
		backupDuration = float64(backupData.Finished - backupData.Started)
		backupStopTime = time.Unix(backupData.Finished, 0).Format(layout)
	}
	setUpMetric(
		medusaBackupDurationMetric,
		"medusa_backup_duration_seconds",
		float64(backupDuration),
		setUpMetricValueFun,
		logger,
		backupData.Name,
		backupData.BackupType,
		time.Unix(backupData.Started, 0).Format(layout),
		backupStopTime,
	)
	// Backup size.
	setUpMetric(
		medusaBackupDatabaseSizeMetric,
		"medusa_backup_size_bytes",
		float64(backupData.Size),
		setUpMetricValueFun,
		logger,
		backupData.Name,
		backupData.BackupType,
	)
	// Backup objects.
	setUpMetric(
		medusaBackupObjectsMetric,
		"medusa_backup_objects",
		float64(backupData.NumObjects),
		setUpMetricValueFun,
		logger,
		backupData.Name,
		backupData.BackupType,
	)
	// Backup completed nodes.
	setUpMetric(
		medusaBackupNodesMetric,
		"medusa_backup_completed_nodes",
		float64(backupData.CompletedNodes),
		setUpMetricValueFun,
		logger,
		backupData.Name,
		backupData.BackupType,
	)
	// Backup incomplete nodes.
	setUpMetric(
		medusaBackupIncompleteNodesMetric,
		"medusa_backup_incomplete_nodes",
		float64(backupData.IncompleteNodes),
		setUpMetricValueFun,
		logger,
		backupData.Name,
		backupData.BackupType,
	)
	// Backup missing nodes.
	setUpMetric(
		medusaBackupMissingNodesMetric,
		"medusa_backup_missing_nodes",
		float64(backupData.MissingNodes),
		setUpMetricValueFun,
		logger,
		backupData.Name,
		backupData.BackupType,
	)
	// Node backup metrics.
	for _, node := range backupData.Nodes {
		// Node backup info.
		//  1 - info about node backup is exist.
		setUpMetric(
			medusaNodeBackupsInfosMetric,
			"medusa_node_backup_info",
			1,
			setUpMetricValueFun,
			logger,
			backupData.Name,
			backupData.BackupType,
			node.FQDN,
			prefix,
			node.ReleaseVersion,
			node.ServerType,
			time.Unix(node.Started, 0).Format(layout),
		)
		// Node backup status.
		//  1 - node backup is complete.
		//  0 - node backup is not complete.
		setUpMetric(
			medusaNodeBackupsStatusMetric,
			"medusa_node_backup_status",
			convertBoolToFloat64(node.Finished > 0),
			setUpMetricValueFun,
			logger,
			backupData.Name,
			backupData.BackupType,
			node.FQDN,
		)
		// Node backup duration.
		if node.Finished == 0 {
			nodeDuration = 0
			nodeStopTime = noneLabel
		} else {
			nodeDuration = float64(node.Finished - node.Started)
			nodeStopTime = time.Unix(node.Finished, 0).Format(layout)
		}
		setUpMetric(
			medusaNodeBackupDurationMetric,
			"medusa_node_backup_duration_seconds",
			float64(nodeDuration),
			setUpMetricValueFun,
			logger,
			backupData.Name,
			backupData.BackupType,
			node.FQDN,
			time.Unix(node.Started, 0).Format(layout),
			nodeStopTime,
		)
		// Node backup size.
		setUpMetric(
			medusaNodeBackupsSizeMetric,
			"medusa_node_backup_size",
			float64(node.Size),
			setUpMetricValueFun,
			logger,
			backupData.Name,
			backupData.BackupType,
			node.FQDN,
		)
		// Node backup objects.
		setUpMetric(
			medusaNodeBackupsObjectsMetric,
			"medusa_node_backup_objects",
			float64(node.NumObjects),
			setUpMetricValueFun,
			logger,
			backupData.Name,
			backupData.BackupType,
			node.FQDN,
		)
	}
}

func resetBackupMetrics() {
	medusaBackupInfoMetric.Reset()
	medusaBackupStatusMetric.Reset()
	medusaBackupDurationMetric.Reset()
	medusaBackupDatabaseSizeMetric.Reset()
	medusaBackupObjectsMetric.Reset()
	medusaBackupNodesMetric.Reset()
	medusaBackupIncompleteNodesMetric.Reset()
	medusaBackupMissingNodesMetric.Reset()
	medusaNodeBackupsInfosMetric.Reset()
	medusaNodeBackupsStatusMetric.Reset()
	medusaNodeBackupDurationMetric.Reset()
	medusaNodeBackupsSizeMetric.Reset()
	medusaNodeBackupsObjectsMetric.Reset()
}
