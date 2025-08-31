package cmd_toolkit_test

import (
	"reflect"
	"runtime/debug"
	"testing"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"

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
				Error: "" +
					"The build time \"today\" cannot be parsed: " +
					"'*time.ParseError: parsing time \"today\" as \"2006-01-02T15:04:05Z07:00\": " +
					"cannot parse \"today\" as \"2006\"'.\n",
				Log: "" +
					"level='error'" +
					" error='parsing time \"today\" as \"2006-01-02T15:04:05Z07:00\": " +
					"cannot parse \"today\" as \"2006\"'" +
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
			if got := cmdtoolkit.DecoratedAppName(
				tt.args.applicationName,
				tt.args.applicationVersion,
				tt.args.timestamp,
			); got != tt.want {
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
			wantDependencies: []string{},
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

func TestStyledFlowerBox(t *testing.T) {
	type args struct {
		lines []string
		style cmdtoolkit.FlowerBoxStyle
	}
	tests := map[string]struct {
		args
		want []string
	}{
		"ASCII": {
			args: args{
				lines: []string{"abc"},
				style: cmdtoolkit.ASCIIFlowerBox,
			},
			want: []string{
				"+-----+",
				"| abc |",
				"+-----+",
				"",
			},
		},
		"curved": {
			args: args{
				lines: []string{"abc"},
				style: cmdtoolkit.CurvedFlowerBox,
			},
			want: []string{
				"╭─────╮",
				"│ abc │",
				"╰─────╯",
				"",
			},
		},
		"light lined": {
			args: args{
				lines: []string{"abc"},
				style: cmdtoolkit.LightLinedFlowerBox,
			},
			want: []string{
				"┌─────┐",
				"│ abc │",
				"└─────┘",
				"",
			},
		},
		"Double lined": {
			args: args{
				lines: []string{"abc"},
				style: cmdtoolkit.DoubleLinedFlowerBox,
			},
			want: []string{
				"╔═════╗",
				"║ abc ║",
				"╚═════╝",
				"",
			},
		},
		"Heavy lined": {
			args: args{
				lines: []string{"abc"},
				style: cmdtoolkit.HeavyLinedFlowerBox,
			},
			want: []string{
				"┏━━━━━┓",
				"┃ abc ┃",
				"┗━━━━━┛",
				"",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmdtoolkit.StyledFlowerBox(tt.args.lines, tt.args.style); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StyledFlowerBox() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBuildData(t *testing.T) {
	tests := map[string]struct {
		reader           func() (*debug.BuildInfo, bool)
		wantGoVersion    string
		wantMainVersion  string
		wantDependencies []string
		wantSettings     []string
	}{
		"nil func": {
			reader:           nil,
			wantGoVersion:    "unknown",
			wantMainVersion:  "unknown",
			wantDependencies: []string{},
			wantSettings:     []string{},
		},
		"func returns no data": {
			reader: func() (*debug.BuildInfo, bool) {
				return nil, false
			},
			wantGoVersion:    "unknown",
			wantMainVersion:  "unknown",
			wantDependencies: []string{},
			wantSettings:     []string{},
		},
		"func returns data": {
			reader: func() (*debug.BuildInfo, bool) {
				return &debug.BuildInfo{
					GoVersion: "1.2.3",
					Main:      debug.Module{Version: "v0.1.2"},
					Deps: []*debug.Module{
						{Path: "mod3", Version: "1.2"},
						{Path: "mod2", Version: "1.3"},
						{Path: "mod1", Version: "1.4"},
					},
					Settings: []debug.BuildSetting{
						{Key: "git", Value: "2.3.4"},
						{Key: "-ldflags", Value: "-X main.version=2.3.4"},
						{Key: "cmd", Value: "gcc"},
					},
				}, true
			},
			wantGoVersion:    "1.2.3",
			wantMainVersion:  "v0.1.2",
			wantDependencies: []string{"mod1 1.4", "mod2 1.3", "mod3 1.2"},
			wantSettings:     []string{"-ldflags: -X main.version=2.3.4", "cmd: gcc", "git: 2.3.4"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotBuildData := cmdtoolkit.GetBuildData(tt.reader)
			if got := gotBuildData.GoVersion(); got != tt.wantGoVersion {
				t.Errorf("GetBuildData() gotGoVersion = %v, want %v", got, tt.wantGoVersion)
			}
			if got := gotBuildData.MainVersion(); got != tt.wantMainVersion {
				t.Errorf("GetBuildData() gotMainVersion = %v, want %v", got, tt.wantMainVersion)
			}
			if got := gotBuildData.Dependencies(); !reflect.DeepEqual(got, tt.wantDependencies) {
				t.Errorf("GetBuildData() gotDependencies = %v, want %v", got, tt.wantDependencies)
			}
			if got := gotBuildData.Settings(); !reflect.DeepEqual(got, tt.wantSettings) {
				t.Errorf("GetBuildData() gotSettings = %v, want %v", got, tt.wantSettings)
			}
		})
	}
}
