package cmd_toolkit_test

import (
	"bytes"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"io"
	"testing"
)

func TestInitLogging(t *testing.T) {
	tests := map[string]struct {
		logWriterInitFn func(output.Bus, string) (io.Writer, string)
		want            bool
		wantLogPath     string
		logsDebug       bool
		logsError       bool
		logsFatal       bool
		logsInfo        bool
		logsPanic       bool
		logsTrace       bool
		logsWarning     bool
	}{
		"no writer available": {
			logWriterInitFn: func(o output.Bus, _ string) (io.Writer, string) {
				return nil, ""
			},
			want:        false,
			wantLogPath: "",
		},
		"success": {
			logWriterInitFn: func(o output.Bus, _ string) (io.Writer, string) {
				return &bytes.Buffer{}, ""
			},
			want:        true,
			logsDebug:   false,
			logsError:   true,
			logsFatal:   true,
			logsInfo:    true,
			logsPanic:   true,
			logsTrace:   false,
			logsWarning: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			originalLogWriterInitFn := cmdtoolkit.LogWriterInitFn
			defer func() {
				cmdtoolkit.LogWriterInitFn = originalLogWriterInitFn
			}()
			cmdtoolkit.LogWriterInitFn = tt.logWriterInitFn
			got := cmdtoolkit.InitLogging(nil, "")
			if got != tt.want {
				t.Errorf("InitLogging() = %v, want %v", got, tt.want)
			}
			if got {
				if truth := cmdtoolkit.ProductionLogger.WillLog(output.Debug); truth != tt.logsDebug {
					t.Errorf("InitLogging() will log at debug, got %t, want %t", truth, tt.logsDebug)
				}
				if truth := cmdtoolkit.ProductionLogger.WillLog(output.Error); truth != tt.logsError {
					t.Errorf("InitLogging() will log at error, got %t, want %t", truth, tt.logsError)
				}
				if truth := cmdtoolkit.ProductionLogger.WillLog(output.Fatal); truth != tt.logsFatal {
					t.Errorf("InitLogging() will log at fatal, got %t, want %t", truth, tt.logsFatal)
				}
				if truth := cmdtoolkit.ProductionLogger.WillLog(output.Info); truth != tt.logsInfo {
					t.Errorf("InitLogging() will log at info, got %t, want %t", truth, tt.logsInfo)
				}
				if truth := cmdtoolkit.ProductionLogger.WillLog(output.Panic); truth != tt.logsPanic {
					t.Errorf("InitLogging() will log at panic, got %t, want %t", truth, tt.logsPanic)
				}
				if truth := cmdtoolkit.ProductionLogger.WillLog(output.Trace); truth != tt.logsTrace {
					t.Errorf("InitLogging() will log at trace, got %t, want %t", truth, tt.logsTrace)
				}
				if truth := cmdtoolkit.ProductionLogger.WillLog(output.Warning); truth != tt.logsWarning {
					t.Errorf("InitLogging() will log at warning, got %t, want %t", truth, tt.logsWarning)
				}
			}
		})
	}
}

func TestInitLoggingWithLevel(t *testing.T) {
	originalLogWriterInitFn := cmdtoolkit.LogWriterInitFn
	defer func() {
		cmdtoolkit.LogWriterInitFn = originalLogWriterInitFn
	}()
	cmdtoolkit.LogWriterInitFn = func(_ output.Bus, _ string) (io.Writer, string) {
		return &bytes.Buffer{}, "testingLogPath"
	}
	// only going to vary the logging level - TestInitLogging handles the error
	// cases where a writer cannot be obtained. This ensures that we don't
	// introduce a programming error when the underlying log implementation
	// cannot be initialized with the specified log level. Also, the various
	// Test_simpleLogger_[Level] tests verify the expected behavior as to what
	// is and is not logged after initialization with each log level.
	tests := map[string]struct {
		l      output.Level
		wantOk bool
	}{
		"panic": {l: output.Panic, wantOk: true},
		"fatal": {l: output.Fatal, wantOk: true},
		"error": {l: output.Error, wantOk: true},
		"warn":  {l: output.Warning, wantOk: true},
		"info":  {l: output.Info, wantOk: true},
		"debug": {l: output.Debug, wantOk: true},
		"trace": {l: output.Trace, wantOk: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotOk := cmdtoolkit.InitLoggingWithLevel(nil, tt.l, ""); gotOk != tt.wantOk {
				t.Errorf("InitLoggingWithLevel() = %t, want %t", gotOk, tt.wantOk)
			}
			if got := cmdtoolkit.LogPath(); got != "testingLogPath" {
				t.Errorf("LogPath() = %v, want %v", got, "testingLogPath")
			}
		})
	}
}
