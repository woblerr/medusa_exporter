package medusa_collector

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/prometheus/exporter-toolkit/web"
)

type mockStruct struct {
	mockStdout string
	mockStderr string
	mockExit   int
}

var (
	logger   = getLogger()
	mockData = mockStruct{}
)

func TestSetPromPortAndPath(t *testing.T) {
	var (
		testFlagsConfig = web.FlagConfig{
			WebListenAddresses: &([]string{":19500"}),
			WebSystemdSocket:   func(i bool) *bool { return &i }(false),
			WebConfigFile:      func(i string) *string { return &i }(""),
		}
		testEndpoint = "/metrics"
	)
	SetPromPortAndPath(testFlagsConfig, testEndpoint)
	if testFlagsConfig.WebListenAddresses != webFlagsConfig.WebListenAddresses ||
		testFlagsConfig.WebSystemdSocket != webFlagsConfig.WebSystemdSocket ||
		testFlagsConfig.WebConfigFile != webFlagsConfig.WebConfigFile ||
		testEndpoint != webEndpoint {
		t.Errorf("\nVariables do not match,\nlistenAddresses: %v, want: %v;\n"+
			"systemSocket: %v, want: %v;\nwebConfig: %v, want: %v;\nendpoint: %s, want: %s",
			ptrToVal(testFlagsConfig.WebListenAddresses), ptrToVal(webFlagsConfig.WebListenAddresses),
			ptrToVal(testFlagsConfig.WebSystemdSocket), ptrToVal(webFlagsConfig.WebSystemdSocket),
			ptrToVal(testFlagsConfig.WebConfigFile), ptrToVal(webFlagsConfig.WebConfigFile),
			testEndpoint, webEndpoint,
		)
	}
}

func TestGetMedusaInfoCommand(t *testing.T) {
	type args struct {
		config string
		prefix string
	}
	tests := []struct {
		name         string
		args         args
		mockTestData mockStruct
		testText     string
	}{
		{
			"GetMedusaInfoCommandSuccess",
			args{"", ""},
			mockStruct{
				`[{"backup_type":"full","complete":true,"completed_nodes":1,` +
					`"finished":1760340315,"incomplete_nodes":0,"incomplete_nodes_list":[],` +
					`"missing_nodes":0,"missing_nodes_list":[],"name":"202510130725",` +
					`"nodes":[{"finished":1760340315,"fqdn":"592136a8eb9f","num_objects":320,` +
					`"release_version":"5.0.4","server_type":"cassandra","size":1507022,"started":1760340312}],` +
					`"num_objects":320,"size":1507022,"started":1760340312,"total_nodes":1}]`,
				"",
				0,
			},
			"",
		},
		{
			"GetMedusaInfoGoodDataReturnWithWarn",
			args{"", ""},
			mockStruct{
				`[{"backup_type":"full","complete":true,"completed_nodes":1,` +
					`"finished":1760340315,"incomplete_nodes":0,"incomplete_nodes_list":[],` +
					`"missing_nodes":0,"missing_nodes_list":[],"name":"202510130725",` +
					`"nodes":[{"finished":1760340315,"fqdn":"592136a8eb9f","num_objects":320,` +
					`"release_version":"5.0.4","server_type":"cassandra","size":1507022,"started":1760340312}],` +
					`"num_objects":320,"size":1507022,"started":1760340312,"total_nodes":1}]`,
				`WARNING: Some warning message occurred`,
				0,
			},
			`level=INFO msg="Medusa message" msg="WARNING: Some warning message occurred"`},
		{
			"GetMedusaInfoGoodDataReturnWithInfo",
			args{"", ""},
			mockStruct{
				`[{"backup_type":"full","complete":true,"completed_nodes":1,` +
					`"finished":1760340315,"incomplete_nodes":0,"incomplete_nodes_list":[],` +
					`"missing_nodes":0,"missing_nodes_list":[],"name":"202510130725",` +
					`"nodes":[{"finished":1760340315,"fqdn":"592136a8eb9f","num_objects":320,` +
					`"release_version":"5.0.4","server_type":"cassandra","size":1507022,"started":1760340312}],` +
					`"num_objects":320,"size":1507022,"started":1760340312,"total_nodes":1}]`,
				`INFO: Resolving ip address` + "\n" +
					`INFO: Using credentials`,
				0,
			},
			`level=INFO msg="Medusa message" msg="INFO: Resolving ip address`},
		{
			"GetMedusaInfoBadDataReturnWithError",
			args{"", ""},
			mockStruct{
				``,
				`ERROR: Something is wrong`,
				1,
			},
			`level=ERROR msg="Medusa message" msg="ERROR: Something is wrong"`,
		},
		{
			"GetMedusaInfoInvalidJSON",
			args{"", ""},
			mockStruct{
				`{invalid json}`,
				"",
				0,
			},
			`level=ERROR msg="Parse JSON failed"`,
		},
		{
			"GetMedusaInfoEmptyBackupList",
			args{"", ""},
			mockStruct{
				`[]`,
				"",
				0,
			},
			`level=WARN msg="No backup data returned"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMetrics()
			mockData = tt.mockTestData
			execCommand = fakeExecCommand
			defer func() { execCommand = exec.Command }()
			out := &bytes.Buffer{}
			lc := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelInfo}))
			GetMedusaInfo(
				tt.args.config,
				tt.args.prefix,
				lc,
			)
			if !strings.Contains(out.String(), tt.testText) {
				t.Errorf("\nVariable do not match:\n%s\nwant:\n%s", tt.testText, out.String())
			}
		})
	}
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestExecCommandHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	es := strconv.Itoa(mockData.mockExit)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + mockData.mockStdout,
		"STDERR=" + mockData.mockStderr,
		"EXIT_STATUS=" + es}
	return cmd
}

func TestExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, "%s", os.Getenv("STDOUT"))
	fmt.Fprintf(os.Stderr, "%s", os.Getenv("STDERR"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}

// Set logger for tests.
func getLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// Helper for displaying web.FlagConfig values test messages.
func ptrToVal[T any](v *T) T {
	return *v
}
