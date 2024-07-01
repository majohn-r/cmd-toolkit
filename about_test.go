package cmd_toolkit_test

import (
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"reflect"
	"runtime/debug"
	"testing"

	"github.com/majohn-r/output"
)

func TestCopyright(t *testing.T) {
	type args struct {
		first     int
		timestamp string
		owner     string
	}
	tests := map[string]struct {
		args
		want string
		output.WantedRecording
	}{
		"bad time": {
			args: args{first: 2020, timestamp: "today", owner: "no one"},
			want: "Copyright © 2020 no one",
			WantedRecording: output.WantedRecording{
				Error: "The build time \"today\" cannot be parsed: parsing time \"today\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"today\" as \"2006\".\n",
				Log: "" +
					"level='error'" +
					" error='parsing time \"today\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"today\" as \"2006\"'" +
					" value='today'" +
					" msg='parse error'\n",
			},
		},
		"current year": {
			args:            args{first: 2024, timestamp: "2024-01-02T15:04:05+05:00", owner: "me, myself, and I"},
			want:            "Copyright © 2024 me, myself, and I",
			WantedRecording: output.WantedRecording{},
		},
		"previous year": {
			args:            args{first: 2024, timestamp: "2023-01-02T15:04:05+05:00", owner: "me, myself, and I"},
			want:            "Copyright © 2024 me, myself, and I",
			WantedRecording: output.WantedRecording{},
		},
		"subsequent year": {
			args:            args{first: 2024, timestamp: "2025-01-02T15:04:05+05:00", owner: "me, myself, and I"},
			want:            "Copyright © 2024-2025 me, myself, and I",
			WantedRecording: output.WantedRecording{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := cmdtoolkit.Copyright(o, tt.args.first, tt.args.timestamp, tt.args.owner); got != tt.want {
				t.Errorf("Copyright() = %v, want %v", got, tt.want)
			}
			o.Report(t, "Copyright()", tt.WantedRecording)
		})
	}
}

func TestDecoratedAppName(t *testing.T) {
	type args struct {
		applicationName    string
		applicationVersion string
		timestamp          string
	}
	tests := map[string]struct {
		args
		want string
	}{
		"bad timestamp": {
			args: args{
				applicationName:    "myApp",
				applicationVersion: "0.4.0",
				timestamp:          "today",
			},
			want: "myApp version 0.4.0, built on today",
		},
		"good timestamp": {
			args: args{
				applicationName:    "goodApp",
				applicationVersion: "1.0.4",
				timestamp:          "2024-02-24T15:40:00-05:00",
			},
			want: "goodApp version 1.0.4, built on Saturday, February 24 2024, 15:40:00 -0500",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmdtoolkit.DecoratedAppName(tt.args.applicationName, tt.args.applicationVersion, tt.args.timestamp); got != tt.want {
				t.Errorf("DecoratedAppName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInterpretBuildData(t *testing.T) {
	tests := map[string]struct {
		buildInfoReader  func() (*debug.BuildInfo, bool)
		wantGoVersion    string
		wantDependencies []string
	}{
		"expected": {
			buildInfoReader: func() (*debug.BuildInfo, bool) {
				return &debug.BuildInfo{
					GoVersion: "1.22.22",
					Deps: []*debug.Module{
						{Path: "go.dependency.1", Version: "v1.2.3"},
						{Path: "go.dependency.2", Version: "v1.3.4"},
						{Path: "go.dependency.3", Version: "v0.1.2"},
					},
				}, true
			},
			wantGoVersion: "1.22.22",
			wantDependencies: []string{
				"go.dependency.1 v1.2.3",
				"go.dependency.2 v1.3.4",
				"go.dependency.3 v0.1.2",
			},
		},
		"unhappy": {
			buildInfoReader:  func() (*debug.BuildInfo, bool) { return nil, false },
			wantGoVersion:    "unknown",
			wantDependencies: nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotGoVersion, gotDependencies := cmdtoolkit.InterpretBuildData(tt.buildInfoReader)
			if gotGoVersion != tt.wantGoVersion {
				t.Errorf("InterpretBuildData() gotGoVersion = %v, want %v", gotGoVersion, tt.wantGoVersion)
			}
			if !reflect.DeepEqual(gotDependencies, tt.wantDependencies) {
				t.Errorf("InterpretBuildData() gotDependencies = %v, want %v", gotDependencies, tt.wantDependencies)
			}
		})
	}
}

func TestFormatBuildDependencies(t *testing.T) {
	tests := map[string]struct {
		dependencies []string
		want         []string
	}{
		"no dependencies": {
			dependencies: nil,
			want:         []string{},
		},
		"one dependency": {
			dependencies: []string{"go.dependency.1 v1.2.3"},
			want:         []string{" - Dependency: go.dependency.1 v1.2.3"},
		},
		"multiple dependencies": {
			dependencies: []string{
				"go.dependency.1 v1.2.3",
				"go.dependency.2 v1.3.4",
				"go.dependency.3 v0.1.2",
			},
			want: []string{
				" - Dependency: go.dependency.1 v1.2.3",
				" - Dependency: go.dependency.2 v1.3.4",
				" - Dependency: go.dependency.3 v0.1.2",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmdtoolkit.FormatBuildDependencies(tt.dependencies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FormatBuildDependencies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatGoVersion(t *testing.T) {
	tests := map[string]struct {
		version string
		want    string
	}{
		"unhappy": {version: "unknown", want: " - Go version: unknown"},
		"happy":   {version: "1.22.22", want: " - Go version: 1.22.22"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmdtoolkit.FormatGoVersion(tt.version); got != tt.want {
				t.Errorf("FormatGoVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlowerBox(t *testing.T) {
	tests := map[string]struct {
		lines []string
		want  []string
	}{
		"empty": {
			lines: nil,
			want: []string{
				"+--+",
				"+--+",
				"",
			},
		},
		"one line": {
			lines: []string{"line1"},
			want: []string{
				"+-------+",
				"| line1 |",
				"+-------+",
				"",
			},
		},
		"multiple lines": {
			lines: []string{"line1", "line2", "", "line 4"},
			want: []string{
				"+--------+",
				"| line1  |",
				"| line2  |",
				"|        |",
				"| line 4 |",
				"+--------+",
				"",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmdtoolkit.FlowerBox(tt.lines); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FlowerBox() = %v, want %v", got, tt.want)
			}
		})
	}
}
