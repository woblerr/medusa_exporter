package medusa_collector

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

func TestGetBackupMetrics(t *testing.T) {
	type args struct {
		backupData          backup
		prefix              string
		setUpMetricValueFun setUpMetricValueFunType
		testText            string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"GetBackupMetricsFinished",
			args{
				backup{
					BackupType:          "differential",
					CompletedNodes:      1,
					Finished:            1697712000,
					IncompleteNodes:     0,
					IncompleteNodesList: []node{},
					MissingNodes:        0,
					MissingNodesList:    []string{},
					Name:                "test_backup",
					Nodes: []node{
						{
							Finished:       1697712000,
							FQDN:           "node1.example.com",
							NumObjects:     100,
							ReleaseVersion: "5.0.4",
							ServerType:     "cassandra",
							Size:           1024,
							Started:        1697711900,
						},
					},
					NumObjects: 100,
					Size:       1024,
					Started:    1697711900,
				},
				"no-prefix",
				setUpMetricValue,
				`# HELP medusa_backup_completed_nodes Number of completed nodes in backup.
# TYPE medusa_backup_completed_nodes gauge
medusa_backup_completed_nodes{backup_name="test_backup",backup_type="differential"} 1
# HELP medusa_backup_duration_seconds Backup duration.
# TYPE medusa_backup_duration_seconds gauge
medusa_backup_duration_seconds{backup_name="test_backup",backup_type="differential",start_time="2023-10-19 10:38:20",stop_time="2023-10-19 10:40:00"} 100
# HELP medusa_backup_incomplete_nodes Number of incomplete nodes in backup.
# TYPE medusa_backup_incomplete_nodes gauge
medusa_backup_incomplete_nodes{backup_name="test_backup",backup_type="differential"} 0
# HELP medusa_backup_info Backup info.
# TYPE medusa_backup_info gauge
medusa_backup_info{backup_name="test_backup",backup_type="differential",prefix="no-prefix",start_time="2023-10-19 10:38:20"} 1
# HELP medusa_backup_missing_nodes Number of missing nodes in backup.
# TYPE medusa_backup_missing_nodes gauge
medusa_backup_missing_nodes{backup_name="test_backup",backup_type="differential"} 0
# HELP medusa_backup_objects Number of objects in backup.
# TYPE medusa_backup_objects gauge
medusa_backup_objects{backup_name="test_backup",backup_type="differential"} 100
# HELP medusa_backup_size_bytes Backup size.
# TYPE medusa_backup_size_bytes gauge
medusa_backup_size_bytes{backup_name="test_backup",backup_type="differential"} 1024
# HELP medusa_backup_status Backup status.
# TYPE medusa_backup_status gauge
medusa_backup_status{backup_name="test_backup",backup_type="differential"} 0
# HELP medusa_node_backup_duration_seconds Node backup duration.
# TYPE medusa_node_backup_duration_seconds gauge
medusa_node_backup_duration_seconds{backup_name="test_backup",backup_type="differential",node_fqdn="node1.example.com",start_time="2023-10-19 10:38:20",stop_time="2023-10-19 10:40:00"} 100
# HELP medusa_node_backup_info Node backup info.
# TYPE medusa_node_backup_info gauge
medusa_node_backup_info{backup_name="test_backup",backup_type="differential",node_fqdn="node1.example.com",prefix="no-prefix",release_version="5.0.4",server_type="cassandra",start_time="2023-10-19 10:38:20"} 1
# HELP medusa_node_backup_objects Number of objects in node backup.
# TYPE medusa_node_backup_objects gauge
medusa_node_backup_objects{backup_name="test_backup",backup_type="differential",node_fqdn="node1.example.com"} 100
# HELP medusa_node_backup_size_bytes Node backup size.
# TYPE medusa_node_backup_size_bytes gauge
medusa_node_backup_size_bytes{backup_name="test_backup",backup_type="differential",node_fqdn="node1.example.com"} 1024
# HELP medusa_node_backup_status Node backup status.
# TYPE medusa_node_backup_status gauge
medusa_node_backup_status{backup_name="test_backup",backup_type="differential",node_fqdn="node1.example.com"} 0
`,
			},
		},
		{
			"GetBackupMetricsNotFinished",
			args{
				backup{
					BackupType:      "differential",
					CompletedNodes:  0,
					Finished:        0,
					IncompleteNodes: 1,
					IncompleteNodesList: []node{
						{
							Finished:       0,
							FQDN:           "node1.example.com",
							NumObjects:     0,
							ReleaseVersion: "5.0.4",
							ServerType:     "cassandra",
							Size:           0,
							Started:        1697711900,
						},
					},
					MissingNodes:     0,
					MissingNodesList: []string{},
					Name:             "test_backup_incomplete",
					Nodes:            []node{},
					NumObjects:       0,
					Size:             0,
					Started:          1697711900,
				},
				"prod",
				setUpMetricValue,
				`# HELP medusa_backup_completed_nodes Number of completed nodes in backup.
# TYPE medusa_backup_completed_nodes gauge
medusa_backup_completed_nodes{backup_name="test_backup_incomplete",backup_type="differential"} 0
# HELP medusa_backup_duration_seconds Backup duration.
# TYPE medusa_backup_duration_seconds gauge
medusa_backup_duration_seconds{backup_name="test_backup_incomplete",backup_type="differential",start_time="2023-10-19 10:38:20",stop_time="none"} 0
# HELP medusa_backup_incomplete_nodes Number of incomplete nodes in backup.
# TYPE medusa_backup_incomplete_nodes gauge
medusa_backup_incomplete_nodes{backup_name="test_backup_incomplete",backup_type="differential"} 1
# HELP medusa_backup_info Backup info.
# TYPE medusa_backup_info gauge
medusa_backup_info{backup_name="test_backup_incomplete",backup_type="differential",prefix="prod",start_time="2023-10-19 10:38:20"} 1
# HELP medusa_backup_missing_nodes Number of missing nodes in backup.
# TYPE medusa_backup_missing_nodes gauge
medusa_backup_missing_nodes{backup_name="test_backup_incomplete",backup_type="differential"} 0
# HELP medusa_backup_objects Number of objects in backup.
# TYPE medusa_backup_objects gauge
medusa_backup_objects{backup_name="test_backup_incomplete",backup_type="differential"} 0
# HELP medusa_backup_size_bytes Backup size.
# TYPE medusa_backup_size_bytes gauge
medusa_backup_size_bytes{backup_name="test_backup_incomplete",backup_type="differential"} 0
# HELP medusa_backup_status Backup status.
# TYPE medusa_backup_status gauge
medusa_backup_status{backup_name="test_backup_incomplete",backup_type="differential"} 1
# HELP medusa_node_backup_duration_seconds Node backup duration.
# TYPE medusa_node_backup_duration_seconds gauge
medusa_node_backup_duration_seconds{backup_name="test_backup_incomplete",backup_type="differential",node_fqdn="node1.example.com",start_time="2023-10-19 10:38:20",stop_time="none"} 0
# HELP medusa_node_backup_info Node backup info.
# TYPE medusa_node_backup_info gauge
medusa_node_backup_info{backup_name="test_backup_incomplete",backup_type="differential",node_fqdn="node1.example.com",prefix="prod",release_version="5.0.4",server_type="cassandra",start_time="2023-10-19 10:38:20"} 1
# HELP medusa_node_backup_objects Number of objects in node backup.
# TYPE medusa_node_backup_objects gauge
medusa_node_backup_objects{backup_name="test_backup_incomplete",backup_type="differential",node_fqdn="node1.example.com"} 0
# HELP medusa_node_backup_size_bytes Node backup size.
# TYPE medusa_node_backup_size_bytes gauge
medusa_node_backup_size_bytes{backup_name="test_backup_incomplete",backup_type="differential",node_fqdn="node1.example.com"} 0
# HELP medusa_node_backup_status Node backup status.
# TYPE medusa_node_backup_status gauge
medusa_node_backup_status{backup_name="test_backup_incomplete",backup_type="differential",node_fqdn="node1.example.com"} 1
`,
			},
		},
		{
			"GetBackupMetricsCombinedScenario",
			args{
				backup{
					BackupType:      "full",
					CompletedNodes:  2,
					Finished:        0,
					IncompleteNodes: 2,
					IncompleteNodesList: []node{
						{
							Finished:       0,
							FQDN:           "node3.example.com",
							NumObjects:     0,
							ReleaseVersion: "5.0.4",
							ServerType:     "cassandra",
							Size:           0,
							Started:        1697711950,
						},
						{
							Finished:       0,
							FQDN:           "node4.example.com",
							NumObjects:     0,
							ReleaseVersion: "5.0.4",
							ServerType:     "cassandra",
							Size:           0,
							Started:        1697711960,
						},
					},
					MissingNodes:     2,
					MissingNodesList: []string{"node5.example.com", "node6.example.com"},
					Name:             "test_backup_combined",
					Nodes: []node{
						{
							Finished:       1697712000,
							FQDN:           "node1.example.com",
							NumObjects:     100,
							ReleaseVersion: "5.0.4",
							ServerType:     "cassandra",
							Size:           1024,
							Started:        1697711900,
						},
						{
							Finished:       1697712050,
							FQDN:           "node2.example.com",
							NumObjects:     100,
							ReleaseVersion: "5.0.4",
							ServerType:     "cassandra",
							Size:           1024,
							Started:        1697711920,
						},
					},
					NumObjects: 200,
					Size:       2048,
					Started:    1697711900,
				},
				"prod",
				setUpMetricValue,
				`# HELP medusa_backup_completed_nodes Number of completed nodes in backup.
# TYPE medusa_backup_completed_nodes gauge
medusa_backup_completed_nodes{backup_name="test_backup_combined",backup_type="full"} 2
# HELP medusa_backup_duration_seconds Backup duration.
# TYPE medusa_backup_duration_seconds gauge
medusa_backup_duration_seconds{backup_name="test_backup_combined",backup_type="full",start_time="2023-10-19 10:38:20",stop_time="none"} 0
# HELP medusa_backup_incomplete_nodes Number of incomplete nodes in backup.
# TYPE medusa_backup_incomplete_nodes gauge
medusa_backup_incomplete_nodes{backup_name="test_backup_combined",backup_type="full"} 2
# HELP medusa_backup_info Backup info.
# TYPE medusa_backup_info gauge
medusa_backup_info{backup_name="test_backup_combined",backup_type="full",prefix="prod",start_time="2023-10-19 10:38:20"} 1
# HELP medusa_backup_missing_nodes Number of missing nodes in backup.
# TYPE medusa_backup_missing_nodes gauge
medusa_backup_missing_nodes{backup_name="test_backup_combined",backup_type="full"} 2
# HELP medusa_backup_objects Number of objects in backup.
# TYPE medusa_backup_objects gauge
medusa_backup_objects{backup_name="test_backup_combined",backup_type="full"} 200
# HELP medusa_backup_size_bytes Backup size.
# TYPE medusa_backup_size_bytes gauge
medusa_backup_size_bytes{backup_name="test_backup_combined",backup_type="full"} 2048
# HELP medusa_backup_status Backup status.
# TYPE medusa_backup_status gauge
medusa_backup_status{backup_name="test_backup_combined",backup_type="full"} 1
# HELP medusa_node_backup_duration_seconds Node backup duration.
# TYPE medusa_node_backup_duration_seconds gauge
medusa_node_backup_duration_seconds{backup_name="test_backup_combined",backup_type="full",node_fqdn="node1.example.com",start_time="2023-10-19 10:38:20",stop_time="2023-10-19 10:40:00"} 100
medusa_node_backup_duration_seconds{backup_name="test_backup_combined",backup_type="full",node_fqdn="node2.example.com",start_time="2023-10-19 10:38:40",stop_time="2023-10-19 10:40:50"} 130
medusa_node_backup_duration_seconds{backup_name="test_backup_combined",backup_type="full",node_fqdn="node3.example.com",start_time="2023-10-19 10:39:10",stop_time="none"} 0
medusa_node_backup_duration_seconds{backup_name="test_backup_combined",backup_type="full",node_fqdn="node4.example.com",start_time="2023-10-19 10:39:20",stop_time="none"} 0
# HELP medusa_node_backup_info Node backup info.
# TYPE medusa_node_backup_info gauge
medusa_node_backup_info{backup_name="test_backup_combined",backup_type="full",node_fqdn="node1.example.com",prefix="prod",release_version="5.0.4",server_type="cassandra",start_time="2023-10-19 10:38:20"} 1
medusa_node_backup_info{backup_name="test_backup_combined",backup_type="full",node_fqdn="node2.example.com",prefix="prod",release_version="5.0.4",server_type="cassandra",start_time="2023-10-19 10:38:40"} 1
medusa_node_backup_info{backup_name="test_backup_combined",backup_type="full",node_fqdn="node3.example.com",prefix="prod",release_version="5.0.4",server_type="cassandra",start_time="2023-10-19 10:39:10"} 1
medusa_node_backup_info{backup_name="test_backup_combined",backup_type="full",node_fqdn="node4.example.com",prefix="prod",release_version="5.0.4",server_type="cassandra",start_time="2023-10-19 10:39:20"} 1
# HELP medusa_node_backup_objects Number of objects in node backup.
# TYPE medusa_node_backup_objects gauge
medusa_node_backup_objects{backup_name="test_backup_combined",backup_type="full",node_fqdn="node1.example.com"} 100
medusa_node_backup_objects{backup_name="test_backup_combined",backup_type="full",node_fqdn="node2.example.com"} 100
medusa_node_backup_objects{backup_name="test_backup_combined",backup_type="full",node_fqdn="node3.example.com"} 0
medusa_node_backup_objects{backup_name="test_backup_combined",backup_type="full",node_fqdn="node4.example.com"} 0
# HELP medusa_node_backup_size_bytes Node backup size.
# TYPE medusa_node_backup_size_bytes gauge
medusa_node_backup_size_bytes{backup_name="test_backup_combined",backup_type="full",node_fqdn="node1.example.com"} 1024
medusa_node_backup_size_bytes{backup_name="test_backup_combined",backup_type="full",node_fqdn="node2.example.com"} 1024
medusa_node_backup_size_bytes{backup_name="test_backup_combined",backup_type="full",node_fqdn="node3.example.com"} 0
medusa_node_backup_size_bytes{backup_name="test_backup_combined",backup_type="full",node_fqdn="node4.example.com"} 0
# HELP medusa_node_backup_status Node backup status.
# TYPE medusa_node_backup_status gauge
medusa_node_backup_status{backup_name="test_backup_combined",backup_type="full",node_fqdn="node1.example.com"} 0
medusa_node_backup_status{backup_name="test_backup_combined",backup_type="full",node_fqdn="node2.example.com"} 0
medusa_node_backup_status{backup_name="test_backup_combined",backup_type="full",node_fqdn="node3.example.com"} 1
medusa_node_backup_status{backup_name="test_backup_combined",backup_type="full",node_fqdn="node4.example.com"} 1
medusa_node_backup_status{backup_name="test_backup_combined",backup_type="full",node_fqdn="node5.example.com"} 2
medusa_node_backup_status{backup_name="test_backup_combined",backup_type="full",node_fqdn="node6.example.com"} 2
`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetBackupMetrics()
			getBackupMetrics(tt.args.backupData, tt.args.prefix, tt.args.setUpMetricValueFun, logger)
			reg := prometheus.NewRegistry()
			reg.MustRegister(
				medusaBackupInfoMetric,
				medusaBackupStatusMetric,
				medusaBackupDurationMetric,
				medusaBackupDatabaseSizeMetric,
				medusaBackupObjectsMetric,
				medusaBackupNodesMetric,
				medusaBackupIncompleteNodesMetric,
				medusaBackupMissingNodesMetric,
				medusaNodeBackupsInfosMetric,
				medusaNodeBackupsStatusMetric,
				medusaNodeBackupDurationMetric,
				medusaNodeBackupsSizeMetric,
				medusaNodeBackupsObjectsMetric,
			)
			metricFamily, err := reg.Gather()
			if err != nil {
				fmt.Println(err)
			}
			out := &bytes.Buffer{}
			for _, mf := range metricFamily {
				if _, err := expfmt.MetricFamilyToText(out, mf); err != nil {
					panic(err)
				}
			}
			if tt.args.testText != out.String() {
				t.Errorf(
					"\nVariables do not match, metrics:\n%s\nwant:\n%s",
					tt.args.testText, out.String(),
				)
			}
		})
	}
}

func TestGetBackupMetricsErrorsAndDebugs(t *testing.T) {
	type args struct {
		backupData          backup
		prefix              string
		setUpMetricValueFun setUpMetricValueFunType
		errorsCount         int
		debugsCount         int
	}
	tests := []struct {
		name string
		args args
	}{
		// Without backup set size.
		{
			"getBackupMetricsLogError",
			args{
				backup{
					BackupType:          "differential",
					CompletedNodes:      1,
					Finished:            1697712000,
					IncompleteNodes:     0,
					IncompleteNodesList: []node{},
					MissingNodes:        0,
					MissingNodesList:    []string{},
					Name:                "test_backup",
					Nodes: []node{
						{
							Finished:       1697712000,
							FQDN:           "node1.example.com",
							NumObjects:     100,
							ReleaseVersion: "5.0.4",
							ServerType:     "cassandra",
							Size:           1024,
							Started:        1697711900,
						},
					},
					NumObjects: 100,
					Size:       0,
					Started:    1697711900,
				},
				"no-prefix",
				fakeSetUpMetricValue,
				13,
				13,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetBackupMetrics()
			out := &bytes.Buffer{}
			lc := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))
			getBackupMetrics(tt.args.backupData, tt.args.prefix, tt.args.setUpMetricValueFun, lc)
			errorsOutputCount := strings.Count(out.String(), "level=ERROR")
			debugsOutputCount := strings.Count(out.String(), "level=DEBUG")
			if tt.args.errorsCount != errorsOutputCount || tt.args.debugsCount != debugsOutputCount {
				t.Errorf("\nVariables do not match:\nerrors=%d, debugs=%d\nwant:\nerrors=%d, debugs=%d",
					tt.args.errorsCount, tt.args.debugsCount,
					errorsOutputCount, debugsOutputCount)
			}
		})
	}
}

func TestGetBackupStatusCode(t *testing.T) {
	type args struct {
		finished int64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			"GetBackupStatusCodeComplete",
			args{1697712000},
			statusComplete,
		},
		{
			"GetBackupStatusCodeIncomplete",
			args{0},
			statusIncomplete,
		},
		{
			"GetBackupStatusCodeCompletePositive",
			args{100},
			statusComplete,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getBackupStatusCode(tt.args.finished); got != tt.want {
				t.Errorf("\nVariables do not match:\n%v\nwant:\n%v", got, tt.want)
			}
		})
	}
}

func TestCalculateDuration(t *testing.T) {
	type args struct {
		started  int64
		finished int64
	}
	tests := []struct {
		name         string
		args         args
		wantDuration float64
		wantStopTime string
	}{
		{
			"CalculateDurationComplete",
			args{1697711900, 1697712000},
			100,
			"2023-10-19 10:40:00",
		},
		{
			"CalculateDurationIncomplete",
			args{1697711900, 0},
			0,
			noneLabel,
		},
		{
			"CalculateDurationZeroStarted",
			args{0, 0},
			0,
			noneLabel,
		},
		{
			"CalculateDurationLongRunning",
			args{1697711900, 1697722800},
			10900,
			"2023-10-19 13:40:00",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDuration, gotStopTime := calculateDuration(tt.args.started, tt.args.finished)
			if gotDuration != tt.wantDuration {
				t.Errorf("\nDuration does not match:\ngot: %v\nwant: %v", gotDuration, tt.wantDuration)
			}
			if gotStopTime != tt.wantStopTime {
				t.Errorf("\nStop time does not match:\ngot: %v\nwant: %v", gotStopTime, tt.wantStopTime)
			}
		})
	}
}
