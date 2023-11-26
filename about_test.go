package cmd_toolkit

import (
	"flag"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/majohn-r/output"
)

func TestBuildDependencies(t *testing.T) {
	savedBuildDependencies := buildDependencies
	defer func() {
		buildDependencies = savedBuildDependencies
	}()
	tests := map[string]struct {
		preset []string
		want   []string
	}{
		"empty": {preset: nil, want: nil},
		"reasonably filled": {
			preset: []string{
				"github.com/majohn-r/output v0.1.1",
				"github.com/sirupsen/logrus v1.9.0",
				"github.com/utahta/go-cronowriter v1.2.0",
				"gopkg.in/yaml.v3 v3.0.1",
			},
			want: []string{
				"github.com/majohn-r/output v0.1.1",
				"github.com/sirupsen/logrus v1.9.0",
				"github.com/utahta/go-cronowriter v1.2.0",
				"gopkg.in/yaml.v3 v3.0.1",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buildDependencies = tt.preset
			if got := BuildDependencies(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildDependencies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoVersion(t *testing.T) {
	savedGoVersion := goVersion
	defer func() {
		goVersion = savedGoVersion
	}()
	tests := map[string]struct {
		preset string
		want   string
	}{
		"error":  {preset: "unknown", want: "unknown"},
		"normal": {preset: "go1.19.5", want: "go1.19.5"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			goVersion = tt.preset
			if got := GoVersion(); got != tt.want {
				t.Errorf("GoVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitBuildData(t *testing.T) {
	savedAppVersion := appVersion
	savedBuildDependencies := buildDependencies
	savedBuildTimestamp := buildTimestamp
	savedGoVersion := goVersion
	savedBuildInfoReader := buildInfoReader
	defer func() {
		appVersion = savedAppVersion
		buildDependencies = savedBuildDependencies
		buildTimestamp = savedBuildTimestamp
		goVersion = savedGoVersion
		buildInfoReader = savedBuildInfoReader
	}()
	type args struct {
		version               string
		creation              string
		wantBuildDependencies []string
		wantGoVersion         string
		wantAppVersion        string
		wantBuildTimeStamp    string
	}
	tests := map[string]struct {
		args
		preTest func()
	}{
		"failure": {
			args: args{
				version:               "1.2.3",
				creation:              "today",
				wantBuildDependencies: nil,
				wantGoVersion:         "unknown",
				wantAppVersion:        "1.2.3",
				wantBuildTimeStamp:    "today",
			},
			preTest: func() {
				buildInfoReader = func() (*debug.BuildInfo, bool) { return nil, false }
			},
		},
		"success": {
			args: args{
				version:  "2.3.4",
				creation: "2022-08-10T13:29:57-04:00",
				wantBuildDependencies: []string{
					"github.com/majohn-r/output v0.1.1",
					"github.com/sirupsen/logrus v1.9.0",
					"github.com/utahta/go-cronowriter v1.2.0",
					"gopkg.in/yaml.v3 v3.0.1",
				},
				wantGoVersion:      "go1.9.6",
				wantAppVersion:     "2.3.4",
				wantBuildTimeStamp: "2022-08-10T13:29:57-04:00",
			},
			preTest: func() {
				buildInfoReader = func() (*debug.BuildInfo, bool) {
					return &debug.BuildInfo{
						GoVersion: "go1.9.6",
						Deps: []*debug.Module{
							{Path: "github.com/majohn-r/output", Version: "v0.1.1"},
							{Path: "github.com/sirupsen/logrus", Version: "v1.9.0"},
							{Path: "github.com/utahta/go-cronowriter", Version: "v1.2.0"},
							{Path: "gopkg.in/yaml.v3", Version: "v3.0.1"},
						},
					}, true
				}
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			InitBuildData(tt.args.version, tt.args.creation)
			if gotBuildDependencies := buildDependencies; !reflect.DeepEqual(gotBuildDependencies, tt.wantBuildDependencies) {
				t.Errorf("InitBuildData() gotBuildDependencies %v wantBuildDependencies %v", gotBuildDependencies, tt.wantBuildDependencies)
			}
			if gotGoVersion := goVersion; gotGoVersion != tt.wantGoVersion {
				t.Errorf("InitBuildData() gotGoVersion %q wantGoVersion %q", gotGoVersion, tt.wantGoVersion)
			}
			if gotAppVersion := appVersion; gotAppVersion != tt.wantAppVersion {
				t.Errorf("InitBuildData() gotAppVersion %q wantAppVersion %q", gotAppVersion, tt.wantAppVersion)
			}
			if gotBuildTimestamp := buildTimestamp; gotBuildTimestamp != tt.wantBuildTimeStamp {
				t.Errorf("InitBuildData() gotBuildTimestamp %q wantBuildTimestamp %q", gotBuildTimestamp, tt.wantBuildTimeStamp)
			}
		})
	}
}

func TestSetAuthor(t *testing.T) {
	savedAuthor := author
	defer func() {
		author = savedAuthor
	}()
	type args struct {
		s string
	}
	tests := map[string]struct {
		args
		want string
	}{
		"simple": {args: args{s: "a brilliant author"}, want: "a brilliant author"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			SetAuthor(tt.args.s)
			if got := author; got != tt.want {
				t.Errorf("SetAuthor() got %q want %q", got, tt.want)
			}
		})
	}
}

func Test_setFirstYear(t *testing.T) {
	savedFirstYear := firstYear
	defer func() {
		firstYear = savedFirstYear
	}()
	type args struct {
		i int
	}
	tests := map[string]struct {
		args
		want int
	}{
		"simple": {args: args{i: 2022}, want: 2022},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			SetFirstYear(tt.args.i)
			if got := firstYear; got != tt.want {
				t.Errorf("setFirstYear() got %d want %d", got, tt.want)
			}
		})
	}
}

func Test_finalYear(t *testing.T) {
	savedFirstYear := firstYear
	defer func() {
		firstYear = savedFirstYear
	}()
	type args struct {
		timestamp string
	}
	tests := map[string]struct {
		firstYear int
		args
		want int
		output.WantedRecording
	}{
		"bad timestamp": {
			firstYear: 1900,
			args:      args{timestamp: "today"},
			want:      1900,
			WantedRecording: output.WantedRecording{
				Error: "The build time \"today\" cannot be parsed: parsing time \"today\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"today\" as \"2006\".\n",
				Log:   "level='error' error='parsing time \"today\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"today\" as \"2006\"' value='today' msg='parse error'\n",
			},
		},
		"good timestamp": {firstYear: 1999, args: args{timestamp: "2022-08-10T13:29:57-04:00"}, want: 2022},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			firstYear = tt.firstYear
			o := output.NewRecorder()
			if got := finalYear(o, tt.args.timestamp); got != tt.want {
				t.Errorf("finalYear() = %v, want %v", got, tt.want)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("finalYear() %s", issue)
				}
			}
		})
	}
}

func Test_formatBuildData(t *testing.T) {
	savedGoVersion := goVersion
	savedBuildDependencies := buildDependencies
	defer func() {
		goVersion = savedGoVersion
		buildDependencies = savedBuildDependencies
	}()
	tests := map[string]struct {
		goVersion         string
		buildDependencies []string
		want              []string
	}{
		"sample": {
			goVersion: "go1.9.7",
			buildDependencies: []string{
				"github.com/majohn-r/output v0.1.1",
				"github.com/sirupsen/logrus v1.9.0",
				"github.com/utahta/go-cronowriter v1.2.0",
				"gopkg.in/yaml.v3 v3.0.1",
			},
			want: []string{
				"Build Information",
				" - Go version: go1.9.7",
				" - Dependency: github.com/majohn-r/output v0.1.1",
				" - Dependency: github.com/sirupsen/logrus v1.9.0",
				" - Dependency: github.com/utahta/go-cronowriter v1.2.0",
				" - Dependency: gopkg.in/yaml.v3 v3.0.1",
			},
		},
		"awry": {
			goVersion:         "unknown",
			buildDependencies: nil,
			want: []string{
				"Build Information",
				" - Go version: unknown",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			goVersion = tt.goVersion
			buildDependencies = tt.buildDependencies
			if got := formatBuildData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("formatBuildData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_formatCopyright(t *testing.T) {
	savedAuthor := author
	defer func() {
		author = savedAuthor
	}()
	type args struct {
		firstYear int
		lastYear  int
	}
	tests := map[string]struct {
		author string
		args
		want string
	}{
		"bad last year": {author: "me", args: args{firstYear: 2022, lastYear: 2020}, want: "Copyright © 2022 me"},
		"same year":     {author: "myself", args: args{firstYear: 2022, lastYear: 2022}, want: "Copyright © 2022 myself"},
		"later year":    {author: "I", args: args{firstYear: 2022, lastYear: 2023}, want: "Copyright © 2022-2023 I"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			author = tt.author
			if got := formatCopyright(tt.args.firstYear, tt.args.lastYear); got != tt.want {
				t.Errorf("formatCopyright() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reportAbout(t *testing.T) {
	type args struct {
		lines []string
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"empty": {args: args{lines: nil}, WantedRecording: output.WantedRecording{Console: "+--+\n+--+\n"}},
		"typical": {
			args: args{lines: []string{
				"mp3 version 0.33.4, built on Sunday, January 22 2023, 17:44:40 EST",
				"Copyright © 2021-2023 Marc Johnson",
				"Build Information",
				" - Go version: go1.19.5",
				" - Dependency: github.com/VividCortex/ewma v1.2.0",
				" - Dependency: github.com/bogem/id3v2/v2 v2.1.3",
				" - Dependency: github.com/cheggaaa/pb/v3 v3.1.0",
				" - Dependency: github.com/fatih/color v1.14.0",
				" - Dependency: github.com/lestrrat-go/strftime v1.0.6",
				" - Dependency: github.com/majohn-r/output v0.1.1",
				" - Dependency: github.com/mattn/go-colorable v0.1.13",
				" - Dependency: github.com/mattn/go-isatty v0.0.17",
				" - Dependency: github.com/mattn/go-runewidth v0.0.14",
				" - Dependency: github.com/pkg/errors v0.9.1",
				" - Dependency: github.com/rivo/uniseg v0.4.3",
				" - Dependency: github.com/sirupsen/logrus v1.9.0",
				" - Dependency: github.com/utahta/go-cronowriter v1.2.0",
				" - Dependency: golang.org/x/sys v0.4.0",
				" - Dependency: golang.org/x/text v0.5.0",
				" - Dependency: gopkg.in/yaml.v3 v3.0.1",
			},
			},
			WantedRecording: output.WantedRecording{
				Console: "" +
					"+--------------------------------------------------------------------+\n" +
					"| mp3 version 0.33.4, built on Sunday, January 22 2023, 17:44:40 EST |\n" +
					"| Copyright © 2021-2023 Marc Johnson                                 |\n" +
					"| Build Information                                                  |\n" +
					"|  - Go version: go1.19.5                                            |\n" +
					"|  - Dependency: github.com/VividCortex/ewma v1.2.0                  |\n" +
					"|  - Dependency: github.com/bogem/id3v2/v2 v2.1.3                    |\n" +
					"|  - Dependency: github.com/cheggaaa/pb/v3 v3.1.0                    |\n" +
					"|  - Dependency: github.com/fatih/color v1.14.0                      |\n" +
					"|  - Dependency: github.com/lestrrat-go/strftime v1.0.6              |\n" +
					"|  - Dependency: github.com/majohn-r/output v0.1.1                   |\n" +
					"|  - Dependency: github.com/mattn/go-colorable v0.1.13               |\n" +
					"|  - Dependency: github.com/mattn/go-isatty v0.0.17                  |\n" +
					"|  - Dependency: github.com/mattn/go-runewidth v0.0.14               |\n" +
					"|  - Dependency: github.com/pkg/errors v0.9.1                        |\n" +
					"|  - Dependency: github.com/rivo/uniseg v0.4.3                       |\n" +
					"|  - Dependency: github.com/sirupsen/logrus v1.9.0                   |\n" +
					"|  - Dependency: github.com/utahta/go-cronowriter v1.2.0             |\n" +
					"|  - Dependency: golang.org/x/sys v0.4.0                             |\n" +
					"|  - Dependency: golang.org/x/text v0.5.0                            |\n" +
					"|  - Dependency: gopkg.in/yaml.v3 v3.0.1                             |\n" +
					"+--------------------------------------------------------------------+\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			reportAbout(o, tt.args.lines)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("reportAbout() %s", issue)
				}
			}
		})
	}
}

func Test_translateTimestamp(t *testing.T) {
	type args struct {
		s string
	}
	tests := map[string]struct {
		args
		// note, this is what the output should start with; for reasons I do not
		// understand, there is some variation as to how the timezone is written
		// at the end
		wantMinusTZ string
	}{
		"bad input":  {args: args{s: "today"}, wantMinusTZ: "today"},
		"good input": {args: args{s: "2022-08-10T13:29:57-04:00"}, wantMinusTZ: "Wednesday, August 10 2022, 13:29:57"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := translateTimestamp(tt.args.s); !strings.HasPrefix(got, tt.wantMinusTZ) {
				t.Errorf("translateTimestamp() = %v, want %v", got, tt.wantMinusTZ)
			}
		})
	}
}

func Test_aboutCmd_Exec(t *testing.T) {
	makeAboutCmd := func() *aboutCmd {
		return &aboutCmd{}
	}
	savedAppName := appname
	savedGoVersion := goVersion
	savedBuildDependencies := buildDependencies
	savedAppVersion := appVersion
	savedBuildTimestamp := buildTimestamp
	savedFirstYear := firstYear
	defer func() {
		appname = savedAppName
		goVersion = savedGoVersion
		buildDependencies = savedBuildDependencies
		appVersion = savedAppVersion
		buildTimestamp = savedBuildTimestamp
		firstYear = savedFirstYear
	}()
	type args struct {
		args []string
	}
	tests := map[string]struct {
		appname           string
		goVersion         string
		buildDependencies []string
		appVersion        string
		buildTimestamp    string
		firstYear         int
		args
		wantOk bool
		output.WantedRecording
	}{
		"no appname": {
			appname:           "",
			goVersion:         "unknown",
			buildDependencies: nil,
			appVersion:        "0.0.1beta",
			buildTimestamp:    "whenever",
			firstYear:         2020,
			args:              args{},
			wantOk:            true,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"+---------------------------------------------------------------+\n" +
					"| unknown application name version 0.0.1beta, built on whenever |\n" +
					"| Copyright © 2020 Marc Johnson                                 |\n" +
					"| Build Information                                             |\n" +
					"|  - Go version: unknown                                        |\n" +
					"+---------------------------------------------------------------+\n",
				Error: "The build time \"whenever\" cannot be parsed: parsing time \"whenever\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"whenever\" as \"2006\".\n",
				Log: "" +
					"level='info' command='about' msg='executing command'\n" +
					"level='error' error='app name has not been initialized' msg='program error'\n" +
					"level='error' error='parsing time \"whenever\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"whenever\" as \"2006\"' value='whenever' msg='parse error'\n",
			},
		},
		"with appname": {
			appname:   "BrilliantApp.exe",
			goVersion: "go19.4",
			buildDependencies: []string{
				"github.com/majohn-r/output v0.1.1",
				"github.com/sirupsen/logrus v1.9.0",
				"github.com/utahta/go-cronowriter v1.2.0",
				"gopkg.in/yaml.v3 v3.0.1",
			},
			appVersion:     "1.2.3",
			buildTimestamp: "today",
			firstYear:      2021,
			args:           args{},
			wantOk:         true,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"+--------------------------------------------------------+\n" +
					"| BrilliantApp.exe version 1.2.3, built on today         |\n" +
					"| Copyright © 2021 Marc Johnson                          |\n" +
					"| Build Information                                      |\n" +
					"|  - Go version: go19.4                                  |\n" +
					"|  - Dependency: github.com/majohn-r/output v0.1.1       |\n" +
					"|  - Dependency: github.com/sirupsen/logrus v1.9.0       |\n" +
					"|  - Dependency: github.com/utahta/go-cronowriter v1.2.0 |\n" +
					"|  - Dependency: gopkg.in/yaml.v3 v3.0.1                 |\n" +
					"+--------------------------------------------------------+\n",
				Error: "The build time \"today\" cannot be parsed: parsing time \"today\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"today\" as \"2006\".\n",
				Log: "" +
					"level='info' command='about' msg='executing command'\n" +
					"level='error' error='parsing time \"today\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"today\" as \"2006\"' value='today' msg='parse error'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			appname = tt.appname
			goVersion = tt.goVersion
			buildDependencies = tt.buildDependencies
			appVersion = tt.appVersion
			buildTimestamp = tt.buildTimestamp
			firstYear = tt.firstYear
			a := makeAboutCmd()
			o := output.NewRecorder()
			if gotOk := a.Exec(o, tt.args.args); gotOk != tt.wantOk {
				t.Errorf("aboutCmd.Exec() = %v, want %v", gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("aboutCmd.Exec() %s", issue)
				}
			}
		})
	}
}

func Test_newAboutCmd(t *testing.T) {
	type args struct {
		in1 *Configuration
		in2 *flag.FlagSet
	}
	tests := map[string]struct {
		args
		want  CommandProcessor
		want1 bool
		output.WantedRecording
	}{"nothing interesting": {args: args{}, want: &aboutCmd{}, want1: true}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := newAboutCmd(o, tt.args.in1, tt.args.in2)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newAboutCmd() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("newAboutCmd() got1 = %v, want %v", got1, tt.want1)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("newAboutCmd() %s", issue)
				}
			}
		})
	}
}
