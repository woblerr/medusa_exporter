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

func TestGetBackupLastMetrics(t *testing.T) {
	type args struct {
		lastBackups         lastBackupsStruct
		currentUnixTime     int64
		setUpMetricValueFun setUpMetricValueFunType
		testText            string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"GetBackupLastMetricsFullOnly",
			args{
				lastBackupsStruct{
					full: backupStruct{
						backupType: fullLabel,
						started:    1697711900,
						finished:   1697712000,
						size:       1024,
						numObjects: 100,
					},
					differential: backupStruct{
						backupType: differentialLabel,
						started:    1697711900,
						finished:   1697712000,
						size:       1024,
						numObjects: 100,
					},
				},
				1697722000,
				setUpMetricValue,
				`# HELP medusa_backup_last_duration_seconds Backup duration for the last full or differential backup.
# TYPE medusa_backup_last_duration_seconds gauge
medusa_backup_last_duration_seconds{backup_type="differential"} 100
medusa_backup_last_duration_seconds{backup_type="full"} 100
# HELP medusa_backup_last_objects Number of objects in backup for the last full or differential backup.
# TYPE medusa_backup_last_objects gauge
medusa_backup_last_objects{backup_type="differential"} 100
medusa_backup_last_objects{backup_type="full"} 100
# HELP medusa_backup_last_size_bytes Backup size for the last full or differential backup.
# TYPE medusa_backup_last_size_bytes gauge
medusa_backup_last_size_bytes{backup_type="differential"} 1024
medusa_backup_last_size_bytes{backup_type="full"} 1024
# HELP medusa_backup_since_last_completion_seconds Time since last full or differential backup completion.
# TYPE medusa_backup_since_last_completion_seconds gauge
medusa_backup_since_last_completion_seconds{backup_type="differential"} 10000
medusa_backup_since_last_completion_seconds{backup_type="full"} 10000
`,
			},
		},
		{
			"GetBackupLastMetricsOnlyFull",
			args{
				lastBackupsStruct{
					full: backupStruct{
						backupType: fullLabel,
						started:    1697711900,
						finished:   1697712000,
						size:       2048,
						numObjects: 200,
					},
					differential: backupStruct{
						backupType: differentialLabel,
					},
				},
				1697732000,
				setUpMetricValue,
				`# HELP medusa_backup_last_duration_seconds Backup duration for the last full or differential backup.
# TYPE medusa_backup_last_duration_seconds gauge
medusa_backup_last_duration_seconds{backup_type="full"} 100
# HELP medusa_backup_last_objects Number of objects in backup for the last full or differential backup.
# TYPE medusa_backup_last_objects gauge
medusa_backup_last_objects{backup_type="full"} 200
# HELP medusa_backup_last_size_bytes Backup size for the last full or differential backup.
# TYPE medusa_backup_last_size_bytes gauge
medusa_backup_last_size_bytes{backup_type="full"} 2048
# HELP medusa_backup_since_last_completion_seconds Time since last full or differential backup completion.
# TYPE medusa_backup_since_last_completion_seconds gauge
medusa_backup_since_last_completion_seconds{backup_type="full"} 20000
`,
			},
		},
		{
			"GetBackupLastMetricsOnlyDifferential",
			args{
				lastBackupsStruct{
					full: backupStruct{
						backupType: fullLabel,
					},
					differential: backupStruct{
						backupType: differentialLabel,
						started:    1697711900,
						finished:   1697712000,
						size:       512,
						numObjects: 50,
					},
				},
				1697722000,
				setUpMetricValue,
				`# HELP medusa_backup_last_duration_seconds Backup duration for the last full or differential backup.
# TYPE medusa_backup_last_duration_seconds gauge
medusa_backup_last_duration_seconds{backup_type="differential"} 100
# HELP medusa_backup_last_objects Number of objects in backup for the last full or differential backup.
# TYPE medusa_backup_last_objects gauge
medusa_backup_last_objects{backup_type="differential"} 50
# HELP medusa_backup_last_size_bytes Backup size for the last full or differential backup.
# TYPE medusa_backup_last_size_bytes gauge
medusa_backup_last_size_bytes{backup_type="differential"} 512
# HELP medusa_backup_since_last_completion_seconds Time since last full or differential backup completion.
# TYPE medusa_backup_since_last_completion_seconds gauge
medusa_backup_since_last_completion_seconds{backup_type="differential"} 10000
`,
			},
		},
		{
			"GetBackupLastMetricsBothTypes",
			args{
				lastBackupsStruct{
					full: backupStruct{
						backupType: fullLabel,
						started:    1697711900,
						finished:   1697712000,
						size:       2048,
						numObjects: 200,
					},
					differential: backupStruct{
						backupType: differentialLabel,
						started:    1697721900,
						finished:   1697722000,
						size:       512,
						numObjects: 50,
					},
				},
				1697732000,
				setUpMetricValue,
				`# HELP medusa_backup_last_duration_seconds Backup duration for the last full or differential backup.
# TYPE medusa_backup_last_duration_seconds gauge
medusa_backup_last_duration_seconds{backup_type="differential"} 100
medusa_backup_last_duration_seconds{backup_type="full"} 100
# HELP medusa_backup_last_objects Number of objects in backup for the last full or differential backup.
# TYPE medusa_backup_last_objects gauge
medusa_backup_last_objects{backup_type="differential"} 50
medusa_backup_last_objects{backup_type="full"} 200
# HELP medusa_backup_last_size_bytes Backup size for the last full or differential backup.
# TYPE medusa_backup_last_size_bytes gauge
medusa_backup_last_size_bytes{backup_type="differential"} 512
medusa_backup_last_size_bytes{backup_type="full"} 2048
# HELP medusa_backup_since_last_completion_seconds Time since last full or differential backup completion.
# TYPE medusa_backup_since_last_completion_seconds gauge
medusa_backup_since_last_completion_seconds{backup_type="differential"} 10000
medusa_backup_since_last_completion_seconds{backup_type="full"} 20000
`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			lc := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))
			resetBackupLastMetrics()
			getBackupLastMetrics(tt.args.lastBackups, tt.args.currentUnixTime, tt.args.setUpMetricValueFun, lc)
			reg := prometheus.NewRegistry()
			reg.MustRegister(
				medusaBackupSinceLastCompletionSecondsMetric,
				medusaBackupLastDurationMetric,
				medusaBackupLastDatabaseSizeMetric,
				medusaBackupLastObjectsMetric,
			)
			metricFamily, err := reg.Gather()
			if err != nil {
				fmt.Println(err)
			}
			out = &bytes.Buffer{}
			for _, mf := range metricFamily {
				if _, err := expfmt.MetricFamilyToText(out, mf); err != nil {
					panic(err)
				}
			}
			if tt.args.testText != out.String() {
				t.Errorf(
					"\nVariables do not match, metrics:\n%s\nwant:\n%s",
					out.String(), tt.args.testText,
				)
			}
		})
	}
}

func TestGetBackupLastMetricsErrorsAndDebugs(t *testing.T) {
	type args struct {
		lastBackups         lastBackupsStruct
		currentUnixTime     int64
		setUpMetricValueFun setUpMetricValueFunType
		errorsCount         int
		debugsCount         int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"getBackupLastMetricsLogError",
			args{
				lastBackupsStruct{
					full: backupStruct{
						backupType: fullLabel,
						started:    1697711900,
						finished:   1697712000,
						size:       1024,
						numObjects: 100,
					},
					differential: backupStruct{
						backupType: differentialLabel,
						started:    1697711900,
						finished:   1697712000,
						size:       1024,
						numObjects: 100,
					},
				},
				1697722000,
				fakeSetUpMetricValue,
				8,
				8,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetBackupLastMetrics()
			out := &bytes.Buffer{}
			lc := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))
			getBackupLastMetrics(tt.args.lastBackups, tt.args.currentUnixTime, tt.args.setUpMetricValueFun, lc)
			errorsOutputCount := strings.Count(out.String(), "level=ERROR")
			debugsOutputCount := strings.Count(out.String(), "level=DEBUG")
			if tt.args.errorsCount != errorsOutputCount || tt.args.debugsCount != debugsOutputCount {
				t.Errorf("\nVariables do not match:\nerrors=%d, debugs=%d\nwant:\nerrors=%d, debugs=%d",
					errorsOutputCount, debugsOutputCount,
					tt.args.errorsCount, tt.args.debugsCount)
			}
		})
	}
}
