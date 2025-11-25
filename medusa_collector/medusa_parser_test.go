package medusa_collector

import (
	"bytes"
	"log/slog"
	"os/exec"
	"reflect"
	"slices"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestReturnDefaultExecArgs(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "DefaultArgs",
			want: []string{"list-backups", "--output", "json"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := returnDefaultExecArgs()
			if !slices.Equal(got, tt.want) {
				t.Errorf("\nVariables do not match:\ngot: %v\nwant: %v", got, tt.want)
			}
		})
	}
}

func TestReturnConfigExecArgs(t *testing.T) {
	tests := []struct {
		name   string
		config string
		want   []string
	}{
		{
			name:   "EmptyConfig",
			config: "",
			want:   []string{},
		},
		{
			name:   "WithConfig",
			config: "/etc/medusa/medusa.ini",
			want:   []string{"--config-file", "/etc/medusa/medusa.ini"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := returnConfigExecArgs(tt.config)
			if !slices.Equal(got, tt.want) {
				t.Errorf("\nVariables do not match:\ngot: %v\nwant: %v", got, tt.want)
			}
		})
	}
}

func TestReturnPrefixExecArgs(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		want   []string
	}{
		{
			name:   "EmptyPrefix",
			prefix: "",
			want:   []string{},
		},
		{
			name:   "WithPrefix",
			prefix: "prod",
			want:   []string{"--prefix", "prod"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := returnPrefixExecArgs(tt.prefix)
			if !slices.Equal(got, tt.want) {
				t.Errorf("\nVariables do not match:\ngot: %v\nwant: %v", got, tt.want)
			}
		})
	}
}

func TestConcatExecArgs(t *testing.T) {
	tests := []struct {
		name   string
		slices [][]string
		want   []string
	}{
		{
			name:   "EmptySlices",
			slices: [][]string{},
			want:   []string{},
		},
		{
			name:   "SingleSlice",
			slices: [][]string{{"arg1", "arg2"}},
			want:   []string{"arg1", "arg2"},
		},
		{
			name: "MultipleSlices",
			slices: [][]string{
				{"--config-file", "/etc/medusa/medusa.ini"},
				{"--prefix", "prod"},
				{"list-backups", "--output", "json"},
			},
			want: []string{"--config-file", "/etc/medusa/medusa.ini", "--prefix", "prod", "list-backups", "--output", "json"},
		},
		{
			name: "WithEmptySlices",
			slices: [][]string{
				{},
				{"--prefix", "prod"},
				{"list-backups", "--output", "json"},
			},
			want: []string{"--prefix", "prod", "list-backups", "--output", "json"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := concatExecArgs(tt.slices)
			if !slices.Equal(got, tt.want) {
				t.Errorf("\nVariables do not match:\ngot: %v\nwant: %v", got, tt.want)
			}
		})
	}
}

func TestParseResult(t *testing.T) {
	tests := []struct {
		name    string
		output  []byte
		want    []backup
		wantErr bool
	}{
		{
			name:   "ValidJSON",
			output: []byte(`[{"backup_type":"differential","completed_nodes":1,"finished":1697712000,"incomplete_nodes":0,"incomplete_nodes_list":[],"missing_nodes":0,"missing_nodes_list":[],"name":"test_backup","nodes":[{"finished":1697712000,"fqdn":"node1.example.com","num_objects":100,"release_version":"5.0.4","server_type":"cassandra","size":1024,"started":1697711900}],"num_objects":100,"size":1024,"started":1697711900}]`),
			want: []backup{
				{
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
			},
			wantErr: false,
		},
		{
			name:    "EmptyArray",
			output:  []byte(`[]`),
			want:    []backup{},
			wantErr: false,
		},
		{
			name:    "InvalidJSON",
			output:  []byte(`{invalid json}`),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "EmptyInput",
			output:  []byte(``),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseResult(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("\nError expectation does not match:\ngot error: %v\nwant error: %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\nVariables do not match:\ngot: %+v\nwant: %+v", got, tt.want)
			}
		})
	}
}

func TestGetInfoData(t *testing.T) {
	tests := []struct {
		name         string
		config       string
		prefix       string
		mockTestData mockStruct
		wantErr      bool
	}{
		{
			name:   "SuccessWithConfigAndPrefix",
			config: "/etc/medusa/medusa.ini",
			prefix: "prod",
			mockTestData: mockStruct{
				mockStdout: `[{"backup_type":"differential","completed_nodes":1,"finished":1697712000,"incomplete_nodes":0,"incomplete_nodes_list":[],"missing_nodes":0,"missing_nodes_list":[],"name":"test_backup","nodes":[],"num_objects":100,"size":1024,"started":1697711900}]`,
				mockStderr: "",
				mockExit:   0,
			},
			wantErr: false,
		},
		{
			name:   "SuccessWithoutConfigAndPrefix",
			config: "",
			prefix: "",
			mockTestData: mockStruct{
				mockStdout: `[{"backup_type":"differential","completed_nodes":1,"finished":1697712000,"incomplete_nodes":0,"incomplete_nodes_list":[],"missing_nodes":0,"missing_nodes_list":[],"name":"test_backup","nodes":[],"num_objects":100,"size":1024,"started":1697711900}]`,
				mockStderr: "",
				mockExit:   0,
			},
			wantErr: false,
		},
		{
			name:   "StderrMessage",
			config: "",
			prefix: "",
			mockTestData: mockStruct{
				mockStdout: `[]`,
				mockStderr: "INFO: Medusa informational message",
				mockExit:   0,
			},
			wantErr: false,
		},
		{
			name:   "CommandError",
			config: "",
			prefix: "",
			mockTestData: mockStruct{
				mockStdout: "",
				mockStderr: "ERROR: Failed to connect to storage",
				mockExit:   1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockData = tt.mockTestData
			execCommand = fakeExecCommand
			defer func() { execCommand = exec.Command }()
			out := &bytes.Buffer{}
			lc := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))
			got, err := getInfoData(tt.config, tt.prefix, lc)
			if (err != nil) != tt.wantErr {
				t.Errorf("\nError expectation does not match:\ngot error: %v\nwant error: %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && got != nil {
				t.Errorf("\nExpected nil output on error, got: %v", got)
			}
		})
	}
}

func TestSetUpMetricValue(t *testing.T) {
	type args struct {
		metric *prometheus.GaugeVec
		value  float64
		labels []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"setUpMetricValueError",
			args{medusaBackupInfoMetric, 0, []string{"backup_name", "bad"}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setUpMetricValue(tt.args.metric, tt.args.value, tt.args.labels...); (err != nil) != tt.wantErr {
				t.Errorf("\nVariables do not match:\n%v\nwant:\n%v", err, tt.wantErr)
			}
		})
	}
}

func TestInitLastBackupStruct(t *testing.T) {
	tests := []struct {
		name string
		want lastBackupsStruct
	}{
		{
			name: "InitializedStruct",
			want: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
				},
				differential: backupStruct{
					backupType: differentialLabel,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := initLastBackupStruct()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\nVariables do not match:\ngot: %+v\nwant: %+v", got, tt.want)
			}
		})
	}
}

func TestHasFinishedBackups(t *testing.T) {
	tests := []struct {
		name        string
		lastBackups lastBackupsStruct
		want        bool
	}{
		{
			name:        "NoFinishedBackups",
			lastBackups: initLastBackupStruct(),
			want:        false,
		},
		{
			name: "OnlyFullFinished",
			lastBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
					started:    1697711900,
					finished:   1697712000,
					size:       1024,
					numObjects: 100,
				},
				differential: backupStruct{
					backupType: differentialLabel,
				},
			},
			want: true,
		},
		{
			name: "OnlyDifferentialFinished",
			lastBackups: lastBackupsStruct{
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
			want: true,
		},
		{
			name: "BothFinished",
			lastBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
					started:    1697711900,
					finished:   1697712000,
					size:       1024,
					numObjects: 100,
				},
				differential: backupStruct{
					backupType: differentialLabel,
					started:    1697722000,
					finished:   1697722100,
					size:       512,
					numObjects: 50,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.lastBackups.hasFinishedBackups()
			if got != tt.want {
				t.Errorf("\nVariables do not match:\ngot: %v\nwant: %v", got, tt.want)
			}
		})
	}
}

func TestCompareLastBackups(t *testing.T) {
	tests := []struct {
		name        string
		lastBackups lastBackupsStruct
		backupData  backup
		wantBackups lastBackupsStruct
	}{
		{
			name:        "FirstFullBackup",
			lastBackups: initLastBackupStruct(),
			backupData: backup{
				BackupType: fullLabel,
				Started:    1697711900,
				Finished:   1697712000,
				Size:       1024,
				NumObjects: 100,
			},
			wantBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
					started:    1697711900,
					finished:   1697712000,
					size:       1024,
					numObjects: 100,
				},
				differential: backupStruct{
					backupType: differentialLabel,
				},
			},
		},
		{
			name:        "FirstDifferentialBackup",
			lastBackups: initLastBackupStruct(),
			backupData: backup{
				BackupType: differentialLabel,
				Started:    1697711900,
				Finished:   1697712000,
				Size:       512,
				NumObjects: 50,
			},
			wantBackups: lastBackupsStruct{
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
		},
		{
			name: "NewerFullBackup",
			lastBackups: lastBackupsStruct{
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
			backupData: backup{
				BackupType: fullLabel,
				Started:    1697722000,
				Finished:   1697722100,
				Size:       2048,
				NumObjects: 200,
			},
			wantBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
					started:    1697722000,
					finished:   1697722100,
					size:       2048,
					numObjects: 200,
				},
				differential: backupStruct{
					backupType: differentialLabel,
					started:    1697711900,
					finished:   1697712000,
					size:       1024,
					numObjects: 100,
				},
			},
		},
		{
			name: "OlderFullBackupIgnored",
			lastBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
					started:    1697722000,
					finished:   1697722100,
					size:       2048,
					numObjects: 200,
				},
				differential: backupStruct{
					backupType: differentialLabel,
					started:    1697732000,
					finished:   1697732100,
					size:       1536,
					numObjects: 150,
				},
			},
			backupData: backup{
				BackupType: fullLabel,
				Started:    1697711900,
				Finished:   1697712000,
				Size:       1024,
				NumObjects: 100,
			},
			wantBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
					started:    1697722000,
					finished:   1697722100,
					size:       2048,
					numObjects: 200,
				},
				differential: backupStruct{
					backupType: differentialLabel,
					started:    1697732000,
					finished:   1697732100,
					size:       1536,
					numObjects: 150,
				},
			},
		},
		{
			name: "DifferentialAfterFull",
			lastBackups: lastBackupsStruct{
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
			backupData: backup{
				BackupType: differentialLabel,
				Started:    1697722000,
				Finished:   1697722100,
				Size:       512,
				NumObjects: 50,
			},
			wantBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
					started:    1697711900,
					finished:   1697712000,
					size:       1024,
					numObjects: 100,
				},
				differential: backupStruct{
					backupType: differentialLabel,
					started:    1697722000,
					finished:   1697722100,
					size:       512,
					numObjects: 50,
				},
			},
		},
		{
			name: "OlderDifferentialBackupIgnored",
			lastBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
					started:    1697711900,
					finished:   1697712000,
					size:       2048,
					numObjects: 200,
				},
				differential: backupStruct{
					backupType: differentialLabel,
					started:    1697722000,
					finished:   1697722100,
					size:       1536,
					numObjects: 150,
				},
			},
			backupData: backup{
				BackupType: differentialLabel,
				Started:    1697711900,
				Finished:   1697712000,
				Size:       512,
				NumObjects: 50,
			},
			wantBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
					started:    1697711900,
					finished:   1697712000,
					size:       2048,
					numObjects: 200,
				},
				differential: backupStruct{
					backupType: differentialLabel,
					started:    1697722000,
					finished:   1697722100,
					size:       1536,
					numObjects: 150,
				},
			},
		},
		{
			name: "FullAfterDifferential",
			lastBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
					started:    1697711900,
					finished:   1697712000,
					size:       512,
					numObjects: 50,
				},
				differential: backupStruct{
					backupType: differentialLabel,
					started:    1697711900,
					finished:   1697712000,
					size:       512,
					numObjects: 50,
				},
			},
			backupData: backup{
				BackupType: fullLabel,
				Started:    1697722000,
				Finished:   1697722100,
				Size:       2048,
				NumObjects: 200,
			},
			wantBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
					started:    1697722000,
					finished:   1697722100,
					size:       2048,
					numObjects: 200,
				},
				differential: backupStruct{
					backupType: differentialLabel,
					started:    1697711900,
					finished:   1697712000,
					size:       512,
					numObjects: 50,
				},
			},
		},
		{
			name: "NewerDifferentialOnly",
			lastBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
					started:    1697711900,
					finished:   1697712000,
					size:       512,
					numObjects: 50,
				},
				differential: backupStruct{
					backupType: differentialLabel,
					started:    1697711900,
					finished:   1697712000,
					size:       512,
					numObjects: 50,
				},
			},
			backupData: backup{
				BackupType: differentialLabel,
				Started:    1697722000,
				Finished:   1697722100,
				Size:       256,
				NumObjects: 25,
			},
			wantBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
					started:    1697711900,
					finished:   1697712000,
					size:       512,
					numObjects: 50,
				},
				differential: backupStruct{
					backupType: differentialLabel,
					started:    1697722000,
					finished:   1697722100,
					size:       256,
					numObjects: 25,
				},
			},
		},
		{
			name: "NotFinishedFullBackup",
			lastBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
					started:    1697700000,
					finished:   1697701000,
					size:       2048,
					numObjects: 200,
				},
				differential: backupStruct{
					backupType: differentialLabel,
				},
			},
			backupData: backup{
				BackupType: fullLabel,
				Started:    1697711900,
				Finished:   0,
				Size:       1024,
				NumObjects: 100,
			},
			wantBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
					started:    1697700000,
					finished:   1697701000,
					size:       2048,
					numObjects: 200,
				},
				differential: backupStruct{
					backupType: differentialLabel,
				},
			},
		},
		{
			name: "NotFinishedDifferentialBackup",
			lastBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
				},
				differential: backupStruct{
					backupType: differentialLabel,
					started:    1697700000,
					finished:   1697701000,
					size:       1024,
					numObjects: 100,
				},
			},
			backupData: backup{
				BackupType: differentialLabel,
				Started:    1697711900,
				Finished:   0,
				Size:       512,
				NumObjects: 50,
			},
			wantBackups: lastBackupsStruct{
				full: backupStruct{
					backupType: fullLabel,
				},
				differential: backupStruct{
					backupType: differentialLabel,
					started:    1697700000,
					finished:   1697701000,
					size:       1024,
					numObjects: 100,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.lastBackups.compareLastBackups(tt.backupData)
			if !reflect.DeepEqual(tt.lastBackups, tt.wantBackups) {
				t.Errorf("\nVariables do not match:\ngot: %+v\nwant: %+v", tt.lastBackups, tt.wantBackups)
			}
		})
	}
}
