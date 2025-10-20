package medusa_collector

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

func fakeSetUpMetricValue(metric *prometheus.GaugeVec, value float64, labels ...string) error {
	return errors.New("fake error")
}

func TestGetExporterStatusMetrics(t *testing.T) {
	type args struct {
		getDataStatus       bool
		prefix              string
		testText            string
		setUpMetricValueFun setUpMetricValueFunType
	}
	tests := []struct {
		name string
		args args
	}{
		{"GetExporterStatusGood",
			args{
				true,
				"",
				`# HELP medusa_exporter_status Medusa exporter get data status.
# TYPE medusa_exporter_status gauge
medusa_exporter_status{prefix="no-prefix"} 1
`,
				setUpMetricValue,
			},
		},
		{"GetExporterStatusBad",
			args{
				false,
				"",
				`# HELP medusa_exporter_status Medusa exporter get data status.
# TYPE medusa_exporter_status gauge
medusa_exporter_status{prefix="no-prefix"} 0
`,
				setUpMetricValue,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logOut := &bytes.Buffer{}
			lc := slog.New(slog.NewTextHandler(logOut, &slog.HandlerOptions{Level: slog.LevelDebug}))
			resetExporterMetrics()
			getExporterStatusMetrics(tt.args.getDataStatus, tt.args.prefix, tt.args.setUpMetricValueFun, lc)
			reg := prometheus.NewRegistry()
			reg.MustRegister(medusaExporterStatusMetric)
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
				t.Errorf("\nVariables do not match:\n%s\nwant:\n%s", tt.args.testText, out.String())
			}
		})
	}
}

func TestGetExporterStatusErrorsAndDebugs(t *testing.T) {
	type args struct {
		getDataStatus       bool
		prefix              string
		setUpMetricValueFun setUpMetricValueFunType
		errorsCount         int
		debugsCount         int
	}
	tests := []struct {
		name string
		args args
	}{
		{"GetExporterInfoLogError",
			args{
				true,
				"",
				fakeSetUpMetricValue,
				1,
				1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			lc := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))
			getExporterStatusMetrics(tt.args.getDataStatus, tt.args.prefix, tt.args.setUpMetricValueFun, lc)
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
