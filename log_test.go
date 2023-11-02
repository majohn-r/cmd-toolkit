package cmd_toolkit

import (
	"os"
	"testing"

	"github.com/majohn-r/output"
	"github.com/sirupsen/logrus"
)

func TestInitLogging(t *testing.T) {
	savedTmp := NewEnvVarMemento("TMP")
	savedTemp := NewEnvVarMemento("TEMP")
	savedAppname := appname
	defer func() {
		savedTmp.Restore()
		savedTemp.Restore()
		appname = savedAppname
	}()
	tests := map[string]struct {
		preTest  func()
		postTest func()
		want     bool
		output.WantedRecording
	}{
		"no temp folder defined": {
			preTest: func() {
				os.Unsetenv("TMP")
				os.Unsetenv("TEMP")
			},
			postTest:        func() {},
			want:            false,
			WantedRecording: output.WantedRecording{Error: "Neither the TMP nor TEMP environment variables are defined.\n"},
		},
		"uninitialized appname": {
			preTest: func() {
				os.Setenv("TMP", "logs")
				os.Unsetenv("TEMP")
				appname = ""
			},
			postTest:        func() {},
			want:            false,
			WantedRecording: output.WantedRecording{Error: "A programming error has occurred: app name has not been initialized.\n"},
		},
		"bad TMP setting": {
			preTest: func() {
				os.Setenv("TMP", "logs")
				os.Unsetenv("TEMP")
				appname = "myApp"
				_ = os.WriteFile("logs", []byte{}, StdFilePermissions)
			},
			postTest: func() {
				os.Remove("logs")
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "The directory \"logs\\\\myApp\\\\logs\" cannot be created: mkdir logs: The system cannot find the path specified.\n",
			},
		},
		"success": {
			preTest: func() {
				os.Setenv("TMP", "goodLogs")
				os.Unsetenv("TEMP")
				appname = "myApp"
			},
			postTest: func() {
				// critical to close the logger, otherwise, "goodLogs" cannot be
				// removed, as the logger will continue hold the current log
				// file open
				_ = logger.Close()
				_ = os.RemoveAll("goodLogs")
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			defer tt.postTest()
			o := output.NewRecorder()
			if got := InitLogging(o); got != tt.want {
				t.Errorf("InitLogging() = %v, want %v", got, tt.want)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("InitLogging() %s", issue)
				}
			}
		})
	}
}

func TestProductionLogger_Debug(t *testing.T) {
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl ProductionLogger
		args
	}{"basic": {pl: ProductionLogger{}, args: args{msg: "debug", fields: map[string]any{}}}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.pl.Debug(tt.args.msg, tt.args.fields)
		})
	}
}

func TestProductionLogger_Error(t *testing.T) {
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl ProductionLogger
		args
	}{"basic": {pl: ProductionLogger{}, args: args{msg: "error", fields: map[string]any{}}}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.pl.Error(tt.args.msg, tt.args.fields)
		})
	}
}

func TestProductionLogger_Fatal(t *testing.T) {
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl ProductionLogger
		args
	}{"basic": {pl: ProductionLogger{}, args: args{msg: "fatal", fields: map[string]any{}}}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			savedExit := logrus.StandardLogger().ExitFunc
			defer func() {
				logrus.StandardLogger().ExitFunc = savedExit
			}()
			exited := false
			logrus.StandardLogger().ExitFunc = func(_ int) {
				exited = true
			}
			tt.pl.Fatal(tt.args.msg, tt.args.fields)
			if !exited {
				t.Errorf("ProductionLogger.Fatal() did not attempt to exit!")
			}
			logrus.StandardLogger().ExitFunc = savedExit
		})
	}
}

func TestProductionLogger_Info(t *testing.T) {
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl ProductionLogger
		args
	}{"basic": {pl: ProductionLogger{}, args: args{msg: "info", fields: map[string]any{}}}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.pl.Info(tt.args.msg, tt.args.fields)
		})
	}
}

func TestProductionLogger_Panic(t *testing.T) {
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl ProductionLogger
		args
	}{"basic": {pl: ProductionLogger{}, args: args{msg: "panicg", fields: map[string]any{}}}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			defer func(t *testing.T) {
				if r := recover(); r == nil {
					t.Errorf("ProductionLogger.Panic() did not panic")
				}
			}(t)
			tt.pl.Panic(tt.args.msg, tt.args.fields)
		})
	}
}

func TestProductionLogger_Trace(t *testing.T) {
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl ProductionLogger
		args
	}{"basic": {pl: ProductionLogger{}, args: args{msg: "trace", fields: map[string]any{}}}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.pl.Trace(tt.args.msg, tt.args.fields)
		})
	}
}

func TestProductionLogger_Warning(t *testing.T) {
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl ProductionLogger
		args
	}{"basic": {pl: ProductionLogger{}, args: args{msg: "warning", fields: map[string]any{}}}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.pl.Warning(tt.args.msg, tt.args.fields)
		})
	}
}
