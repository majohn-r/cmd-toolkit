package cmd_toolkit

import (
	"flag"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/majohn-r/output"
)

type happyCommand struct{}

func (*happyCommand) Exec(o output.Bus, _ []string) bool {
	o.WriteConsole("yay!\n")
	return true
}

type unhappyCommand struct{}

func (*unhappyCommand) Exec(o output.Bus, _ []string) bool {
	o.WriteError("so sad ...\n")
	return false
}

func TestExecute(t *testing.T) {
	savedAppname := appname
	savedAppDataValue, savedAppDataSet := os.LookupEnv(ApplicationDataEnvVarName)
	savedDescriptions := descriptions
	savedLogInitializer := logInitializer
	savedBuildInfoReader := buildInfoReader
	savedFirstYear := firstYear
	defer func() {
		appname = savedAppname
		if savedAppDataSet {
			os.Setenv(ApplicationDataEnvVarName, savedAppDataValue)
		} else {
			os.Unsetenv(ApplicationDataEnvVarName)
		}
		descriptions = savedDescriptions
		logInitializer = savedLogInitializer
		buildInfoReader = savedBuildInfoReader
		firstYear = savedFirstYear
	}()
	type args struct {
		firstYear      int
		appName        string
		appVersion     string
		buildTimestamp string
		cmdLine        []string
	}
	tests := map[string]struct {
		appname      string
		appDataValue string
		appDataSet   bool
		descriptions map[string]*CommandDescription
		preTest      func()
		postTest     func()
		args
		wantExitCode int
		output.WantedRecording
	}{
		"set app name fails": {
			appname:         "myApp",
			preTest:         func() {},
			postTest:        func() {},
			wantExitCode:    1,
			WantedRecording: output.WantedRecording{Error: "A programming error has occurred - app name has already been initialized: myApp.\n"},
		},
		"logInit fails": {
			preTest: func() {
				logInitializer = func(o output.Bus) bool {
					o.WriteError("log init failed!!\n")
					return false
				}
			},
			postTest: func() {},
			args: args{
				firstYear:      2021,
				appName:        "myNewApp",
				appVersion:     "0.0.1",
				buildTimestamp: "today",
			},
			wantExitCode: 1,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"log init failed!!\n" +
					"\"myNewApp\" version 0.0.1, created at today, failed.\n",
			},
		},
		"InitApplicationPath fails": {
			preTest: func() {
				logInitializer = func(_ output.Bus) bool {
					return true
				}
			},
			postTest: func() {},
			args: args{
				firstYear:      2021,
				appName:        "myNewApp",
				appVersion:     "0.0.1",
				buildTimestamp: "today",
			},
			wantExitCode: 1,
			WantedRecording: output.WantedRecording{
				Error: "\"myNewApp\" version 0.0.1, created at today, failed.\n",
				Log:   "level='error' environmentVariable='APPDATA' msg='not set'\n",
			},
		},
		"ProcessCommand fails": {
			appDataValue: filepath.Join(".", "appdata"),
			appDataSet:   true,
			descriptions: map[string]*CommandDescription{},
			preTest: func() {
				path := filepath.Join(".", "appdata", "myApp")
				_ = os.MkdirAll(path, StdDirPermissions)
				logInitializer = func(_ output.Bus) bool {
					return true
				}
				buildInfoReader = func() (*debug.BuildInfo, bool) {
					return nil, false
				}
			},
			postTest: func() {
				path := filepath.Join(".", "appdata")
				_ = os.RemoveAll(path)
			},
			args: args{
				firstYear:      2021,
				appName:        "myNewApp",
				appVersion:     "0.0.1",
				buildTimestamp: "today",
			},
			wantExitCode: 1,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"A programming error has occurred - there are no commands registered!\n" +
					"\"myNewApp\" version 0.0.1, created at today, failed.\n",
				Log: "" +
					"level='info' args='[]' dependencies='[]' goVersion='unknown' timeStamp='today' version='0.0.1' msg='execution starts'\n" +
					"level='info' directory='appdata\\myNewApp' fileName='defaults.yaml' msg='file does not exist'\n" +
					"level='error'  msg='no commands registered'\n" +
					"level='info' duration='REDACTED' exitCode='1' msg='execution ends'\n",
			},
		},
		"command execution fails": {
			appDataValue: filepath.Join(".", "appdata"),
			appDataSet:   true,
			descriptions: map[string]*CommandDescription{
				"unhappyCommand": {
					IsDefault: true,
					Initializer: func(_ output.Bus, _ *Configuration, _ *flag.FlagSet) (CommandProcessor, bool) {
						return &unhappyCommand{}, true
					},
				}},
			preTest: func() {
				path := filepath.Join(".", "appdata", "myApp")
				_ = os.MkdirAll(path, StdDirPermissions)
				logInitializer = func(_ output.Bus) bool {
					return true
				}
				buildInfoReader = func() (*debug.BuildInfo, bool) {
					return nil, false
				}
			},
			postTest: func() {
				path := filepath.Join(".", "appdata")
				_ = os.RemoveAll(path)
			},
			args: args{
				firstYear:      2021,
				appName:        "myNewApp",
				appVersion:     "0.0.1",
				buildTimestamp: "today",
			},
			wantExitCode: 1,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"so sad ...\n" +
					"\"myNewApp\" version 0.0.1, created at today, failed.\n",
				Log: "" +
					"level='info' args='[]' dependencies='[]' goVersion='unknown' timeStamp='today' version='0.0.1' msg='execution starts'\n" +
					"level='info' directory='appdata\\myNewApp' fileName='defaults.yaml' msg='file does not exist'\n" +
					"level='info' duration='REDACTED' exitCode='1' msg='execution ends'\n",
			},
		},
		"success": {
			appDataValue: filepath.Join(".", "appdata"),
			appDataSet:   true,
			descriptions: map[string]*CommandDescription{
				"happyCommand": {
					IsDefault: true,
					Initializer: func(_ output.Bus, _ *Configuration, _ *flag.FlagSet) (CommandProcessor, bool) {
						return &happyCommand{}, true
					},
				}},
			preTest: func() {
				path := filepath.Join(".", "appdata", "myApp")
				_ = os.MkdirAll(path, StdDirPermissions)
				logInitializer = func(_ output.Bus) bool {
					return true
				}
				buildInfoReader = func() (*debug.BuildInfo, bool) {
					return nil, false
				}
			},
			postTest: func() {
				path := filepath.Join(".", "appdata")
				_ = os.RemoveAll(path)
			},
			args: args{
				firstYear:      2021,
				appName:        "myNewApp",
				appVersion:     "0.0.1",
				buildTimestamp: "today",
			},
			wantExitCode: 0,
			WantedRecording: output.WantedRecording{
				Console: "yay!\n",
				Log: "" +
					"level='info' args='[]' dependencies='[]' goVersion='unknown' timeStamp='today' version='0.0.1' msg='execution starts'\n" +
					"level='info' directory='appdata\\myNewApp' fileName='defaults.yaml' msg='file does not exist'\n" +
					"level='info' duration='REDACTED' exitCode='0' msg='execution ends'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			appname = tt.appname
			if tt.appDataSet {
				os.Setenv(ApplicationDataEnvVarName, tt.appDataValue)
			} else {
				os.Unsetenv(ApplicationDataEnvVarName)
			}
			descriptions = tt.descriptions
			tt.preTest()
			defer tt.postTest()
			o := output.NewRecorder()
			if gotExitCode := Execute(o, tt.args.firstYear, tt.args.appName, tt.args.appVersion, tt.args.buildTimestamp, tt.args.cmdLine); gotExitCode != tt.wantExitCode {
				t.Errorf("Execute() = %v, want %v", gotExitCode, tt.wantExitCode)
			}
			if gotConsole := o.ConsoleOutput(); gotConsole != tt.WantedRecording.Console {
				t.Errorf("Execute() console %q want console %q", gotConsole, tt.WantedRecording.Console)
			}
			if gotError := o.ErrorOutput(); gotError != tt.WantedRecording.Error {
				t.Errorf("Execute() error %q want error %q", gotError, tt.WantedRecording.Error)
			}
			gotLog := o.LogOutput()
			if strings.Contains(gotLog, " duration='") {
				// snip out the time and replace with 'XXXX'
				before := strings.Index(gotLog, " duration='")
				timeIndex := before + len(" duration='")
				postIndex := timeIndex + 1
				for gotLog[postIndex] != '\'' {
					postIndex++
				}
				gotLog = gotLog[:timeIndex] + "REDACTED" + gotLog[postIndex:]
			}
			if gotLog != tt.WantedRecording.Log {
				t.Errorf("Execute() log %q want log %q", gotLog, tt.WantedRecording.Log)
			}
		})
	}
}
