package medusa_collector

import (
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	statusComplete   = 0
	statusIncomplete = 1
	statusMissing    = 2
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
		Name: "medusa_node_backup_size_bytes",
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
//   - medusa_node_backup_size_bytes
//   - medusa_node_backup_objects
func getBackupMetrics(backupData backup, prefix string, setUpMetricValueFun setUpMetricValueFunType, logger *slog.Logger) {
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
	setUpMetric(
		medusaBackupStatusMetric,
		"medusa_backup_status",
		getBackupStatusCode(backupData.Finished),
		setUpMetricValueFun,
		logger,
		backupData.Name,
		backupData.BackupType,
	)
	// Backup duration.
	backupDuration, backupStopTime := calculateDuration(backupData.Started, backupData.Finished)
	setUpMetric(
		medusaBackupDurationMetric,
		"medusa_backup_duration_seconds",
		backupDuration,
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
	// In this case, checking for a Finished field is unnecessary,
	// because according to Medusa, only completed nodes should be in Nodes.
	// The rest of the nodes should be IncompleteNodesList or MissingNodesList.
	// However, there may be bugs in Medusa, so an additional check would not hurt.
	for _, node := range backupData.Nodes {
		setNodeMetrics(
			node,
			backupData.Name,
			backupData.BackupType,
			prefix,
			getBackupStatusCode(node.Finished),
			setUpMetricValueFun,
			logger,
		)
	}
	// Node backup metrics for incomplete nodes.
	// If node is IncompleteNodesList, status should be incomplete.
	for _, node := range backupData.IncompleteNodesList {
		setNodeMetrics(
			node,
			backupData.Name,
			backupData.BackupType,
			prefix,
			statusIncomplete,
			setUpMetricValueFun,
			logger,
		)
	}
	// Node backup metrics for missing nodes.
	for _, nodeFQDN := range backupData.MissingNodesList {
		setNodeMetrics(
			node{
				Finished:       0,
				FQDN:           nodeFQDN,
				NumObjects:     0,
				ReleaseVersion: noneLabel,
				ServerType:     noneLabel,
				Size:           0,
				Started:        0,
			},
			backupData.Name,
			backupData.BackupType,
			prefix,
			statusMissing,
			setUpMetricValueFun,
			logger,
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

func getBackupStatusCode(finished int64) float64 {
	if finished > 0 {
		return statusComplete
	}
	return statusIncomplete
}

func setNodeMetrics(node node, backupName, backupType, prefix string, status float64, setUpMetricValueFun setUpMetricValueFunType, logger *slog.Logger) {
	nodeStartTime := noneLabel
	if node.Started > 0 {
		nodeStartTime = time.Unix(node.Started, 0).Format(layout)
	}
	// Node backup info.
	//  1 - info about node backup is exist.
	setUpMetric(
		medusaNodeBackupsInfosMetric,
		"medusa_node_backup_info",
		1,
		setUpMetricValueFun,
		logger,
		backupName,
		backupType,
		node.FQDN,
		prefix,
		node.ReleaseVersion,
		node.ServerType,
		nodeStartTime,
	)
	// Node backup status.
	setUpMetric(
		medusaNodeBackupsStatusMetric,
		"medusa_node_backup_status",
		status,
		setUpMetricValueFun,
		logger,
		backupName,
		backupType,
		node.FQDN,
	)
	// Node backup duration.
	nodeDuration, nodeStopTime := calculateDuration(node.Started, node.Finished)
	setUpMetric(
		medusaNodeBackupDurationMetric,
		"medusa_node_backup_duration_seconds",
		nodeDuration,
		setUpMetricValueFun,
		logger,
		backupName,
		backupType,
		node.FQDN,
		nodeStartTime,
		nodeStopTime,
	)
	// Node backup size.
	setUpMetric(
		medusaNodeBackupsSizeMetric,
		"medusa_node_backup_size_bytes",
		float64(node.Size),
		setUpMetricValueFun,
		logger,
		backupName,
		backupType,
		node.FQDN,
	)
	// Node backup objects.
	setUpMetric(
		medusaNodeBackupsObjectsMetric,
		"medusa_node_backup_objects",
		float64(node.NumObjects),
		setUpMetricValueFun,
		logger,
		backupName,
		backupType,
		node.FQDN,
	)
}

func calculateDuration(started, finished int64) (float64, string) {
	var duration float64
	var stopTime string
	if finished == 0 {
		duration = 0
		stopTime = noneLabel
	} else {
		duration = float64(finished - started)
		stopTime = time.Unix(finished, 0).Format(layout)
	}
	return duration, stopTime
}
