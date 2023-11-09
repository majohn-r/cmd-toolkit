package cmd_toolkit

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/majohn-r/output"
)

func TestInitLogging(t *testing.T) {
	tests := map[string]struct {
		writerGetter func(o output.Bus) io.Writer
		want         bool
		logsDebug    bool
		logsError    bool
		logsFatal    bool
		logsInfo     bool
		logsPanic    bool
		logsTrace    bool
		logsWarning  bool
	}{
		"no writer available": {
			writerGetter: func(o output.Bus) io.Writer { return nil },
			want:         false,
		},
		"success": {
			writerGetter: func(o output.Bus) io.Writer { return &bytes.Buffer{} },
			want:         true,
			logsDebug:    false,
			logsError:    true,
			logsFatal:    true,
			logsInfo:     true,
			logsPanic:    true,
			logsTrace:    false,
			logsWarning:  true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			oldFunc := writerGetter
			defer func() {
				writerGetter = oldFunc
			}()
			writerGetter = tt.writerGetter
			got := InitLogging(nil)
			if got != tt.want {
				t.Errorf("InitLogging() = %v, want %v", got, tt.want)
			}
			if got {
				if truth := ProductionLogger.willLog(output.Debug); truth != tt.logsDebug {
					t.Errorf("InitLogging() will log at debug, got %t, want %t", truth, tt.logsDebug)
				}
				if truth := ProductionLogger.willLog(output.Error); truth != tt.logsError {
					t.Errorf("InitLogging() will log at error, got %t, want %t", truth, tt.logsError)
				}
				if truth := ProductionLogger.willLog(output.Fatal); truth != tt.logsFatal {
					t.Errorf("InitLogging() will log at fatal, got %t, want %t", truth, tt.logsFatal)
				}
				if truth := ProductionLogger.willLog(output.Info); truth != tt.logsInfo {
					t.Errorf("InitLogging() will log at info, got %t, want %t", truth, tt.logsInfo)
				}
				if truth := ProductionLogger.willLog(output.Panic); truth != tt.logsPanic {
					t.Errorf("InitLogging() will log at panic, got %t, want %t", truth, tt.logsPanic)
				}
				if truth := ProductionLogger.willLog(output.Trace); truth != tt.logsTrace {
					t.Errorf("InitLogging() will log at trace, got %t, want %t", truth, tt.logsTrace)
				}
				if truth := ProductionLogger.willLog(output.Warning); truth != tt.logsWarning {
					t.Errorf("InitLogging() will log at warning, got %t, want %t", truth, tt.logsWarning)
				}
			}
		})
	}
}

func TestInitLoggingWithLevel(t *testing.T) {
	savedGetWriterFunc := writerGetter
	defer func() {
		writerGetter = savedGetWriterFunc
	}()
	writerGetter = func(_ output.Bus) io.Writer {
		return &bytes.Buffer{}
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
			if gotOk := InitLoggingWithLevel(nil, tt.l); gotOk != tt.wantOk {
				t.Errorf("InitLoggingWithLevel() = %t, want %t", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_simpleLogger_Debug(t *testing.T) {
	savedGetWriterFunc := writerGetter
	defer func() {
		writerGetter = savedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		l output.Level
		args
		want string
	}{
		"panic": {
			l:    output.Panic,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			l:    output.Fatal,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "",
		},
		"error": {
			l:    output.Error,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "",
		},
		"warn": {
			l:    output.Warning,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "",
		},
		"info": {
			l:    output.Info,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "",
		},
		"debug": {
			l: output.Debug,
			args: args{msg: "debug message", fields: map[string]any{
				"field1":  "value1",
				"field2":  2,
				"field3":  true,
				"field4":  []string{"val1", "val2"},
				"field5":  fmt.Errorf("this is a mistake"),
				"field6":  map[string]int{"a": 1, "b b": 2},
				"field7":  regexp.MustCompile(".*"),
				"field8":  "a b c",
				"field 9": false,
			}},
			want: "level=debug msg=\"debug message\" field 9=false field1=value1 field2=2 field3=true field4=\"[val1 val2]\" field5=\"this is a mistake\" field6=\"map[a:1 b b:2]\" field7=\".*\" field8=\"a b c\"\n",
		},
		"trace": {
			l:    output.Trace,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "level=debug msg=debug\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			writerGetter = func(_ output.Bus) io.Writer {
				return buffer
			}
			InitLoggingWithLevel(nil, tt.l)
			ProductionLogger.Debug(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("simpleLogger.Debug() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("simpleLogger.Debug() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func Test_simpleLogger_Error(t *testing.T) {
	savedGetWriterFunc := writerGetter
	defer func() {
		writerGetter = savedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		l output.Level
		args
		want string
	}{
		"panic": {
			l:    output.Panic,
			args: args{msg: "error", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			l:    output.Fatal,
			args: args{msg: "error", fields: map[string]any{}},
			want: "",
		},
		"error": {
			l:    output.Error,
			args: args{msg: "error", fields: map[string]any{}},
			want: "level=error msg=error\n",
		},
		"warn": {
			l:    output.Warning,
			args: args{msg: "error", fields: map[string]any{}},
			want: "level=error msg=error\n",
		},
		"info": {
			l:    output.Info,
			args: args{msg: "error", fields: nil},
			want: "level=error msg=error\n",
		},
		"debug": {
			l: output.Debug,
			args: args{msg: "error", fields: map[string]any{
				"field1":  "value1",
				"field2":  2,
				"field3":  true,
				"field4":  []string{"val1", "val2"},
				"field5":  fmt.Errorf("this is a mistake"),
				"field6":  map[string]int{"a": 1, "b b": 2},
				"field7":  regexp.MustCompile(".*"),
				"field8":  "a b c",
				"field 9": false,
			}},
			want: "level=error msg=error field 9=false field1=value1 field2=2 field3=true field4=\"[val1 val2]\" field5=\"this is a mistake\" field6=\"map[a:1 b b:2]\" field7=\".*\" field8=\"a b c\"\n",
		},
		"trace": {
			l:    output.Trace,
			args: args{msg: "error", fields: map[string]any{}},
			want: "level=error msg=error\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			writerGetter = func(_ output.Bus) io.Writer {
				return buffer
			}
			InitLoggingWithLevel(nil, tt.l)
			ProductionLogger.Error(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("simpleLogger.Error() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("simpleLogger.Error() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func Test_simpleLogger_Fatal(t *testing.T) {
	savedGetWriterFunc := writerGetter
	savedExit := ProductionLogger.ExitFunc()
	defer func() {
		writerGetter = savedGetWriterFunc
		ProductionLogger.SetExitFunc(savedExit)
	}()
	var exited = false
	pretendToExit := func(_ int) {
		exited = true
	}
	ProductionLogger.SetExitFunc(pretendToExit)
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		l output.Level
		args
		want string
	}{
		"panic": {
			l:    output.Panic,
			args: args{msg: "fatal", fields: map[string]any{}},
			want: "level=fatal msg=fatal\n",
		},
		"fatal": {
			l:    output.Fatal,
			args: args{msg: "fatal", fields: map[string]any{}},
			want: "level=fatal msg=fatal\n",
		},
		"error": {
			l:    output.Error,
			args: args{msg: "fatal", fields: map[string]any{}},
			want: "level=fatal msg=fatal\n",
		},
		"warn": {
			l:    output.Warning,
			args: args{msg: "fatal", fields: map[string]any{}},
			want: "level=fatal msg=fatal\n",
		},
		"info": {
			l:    output.Info,
			args: args{msg: "fatal", fields: nil},
			want: "level=fatal msg=fatal\n",
		},
		"debug": {
			l: output.Debug,
			args: args{msg: "fatal message", fields: map[string]any{
				"field1":  "value1",
				"field2":  2,
				"field3":  true,
				"field4":  []string{"val1", "val2"},
				"field5":  fmt.Errorf("this is a mistake"),
				"field6":  map[string]int{"a": 1, "b b": 2},
				"field7":  regexp.MustCompile(".*"),
				"field8":  "a b c",
				"field 9": false,
			}},
			want: "level=fatal msg=\"fatal message\" field 9=false field1=value1 field2=2 field3=true field4=\"[val1 val2]\" field5=\"this is a mistake\" field6=\"map[a:1 b b:2]\" field7=\".*\" field8=\"a b c\"\n",
		},
		"trace": {
			l:    output.Trace,
			args: args{msg: "fatal", fields: map[string]any{}},
			want: "level=fatal msg=fatal\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			writerGetter = func(_ output.Bus) io.Writer {
				return buffer
			}
			InitLoggingWithLevel(nil, tt.l)
			exited = false
			f := ProductionLogger.ExitFunc()
			if f == nil {
				t.Errorf("ProductionLogger has no exit function")
			}
			ProductionLogger.Fatal(tt.args.msg, tt.args.fields)
			if !exited {
				t.Errorf("simpleLogger.Fatal() did not attempt to exit!")
			}
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("simpleLogger.Fatal() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("simpleLogger.Fatal() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func Test_simpleLogger_Info(t *testing.T) {
	savedGetWriterFunc := writerGetter
	defer func() {
		writerGetter = savedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		l output.Level
		args
		want string
	}{
		"panic": {
			l:    output.Panic,
			args: args{msg: "info", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			l:    output.Fatal,
			args: args{msg: "info", fields: map[string]any{}},
			want: "",
		},
		"error": {
			l:    output.Error,
			args: args{msg: "info", fields: map[string]any{}},
			want: "",
		},
		"warn": {
			l:    output.Warning,
			args: args{msg: "info", fields: map[string]any{}},
			want: "",
		},
		"info": {
			l:    output.Info,
			args: args{msg: "info", fields: nil},
			want: "level=info msg=info\n",
		},
		"debug": {
			l: output.Debug,
			args: args{msg: "info message", fields: map[string]any{
				"field1":  "value1",
				"field2":  2,
				"field3":  true,
				"field4":  []string{"val1", "val2"},
				"field5":  fmt.Errorf("this is a mistake"),
				"field6":  map[string]int{"a": 1, "b b": 2},
				"field7":  regexp.MustCompile(".*"),
				"field8":  "a b c",
				"field 9": false,
			}},
			want: "level=info msg=\"info message\" field 9=false field1=value1 field2=2 field3=true field4=\"[val1 val2]\" field5=\"this is a mistake\" field6=\"map[a:1 b b:2]\" field7=\".*\" field8=\"a b c\"\n",
		},
		"trace": {
			l:    output.Trace,
			args: args{msg: "info", fields: map[string]any{}},
			want: "level=info msg=info\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			writerGetter = func(_ output.Bus) io.Writer {
				return buffer
			}
			InitLoggingWithLevel(nil, tt.l)
			ProductionLogger.Info(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("simpleLogger.Info() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("simpleLogger.Info() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func Test_simpleLogger_Panic(t *testing.T) {
	savedGetWriterFunc := writerGetter
	defer func() {
		writerGetter = savedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		l output.Level
		args
		want string
	}{
		"panic": {
			l:    output.Panic,
			args: args{msg: "panic", fields: map[string]any{}},
			want: "level=panic msg=panic\n",
		},
		"fatal": {
			l:    output.Fatal,
			args: args{msg: "panic", fields: map[string]any{}},
			want: "level=panic msg=panic\n",
		},
		"error": {
			l:    output.Error,
			args: args{msg: "panic", fields: map[string]any{}},
			want: "level=panic msg=panic\n",
		},
		"warn": {
			l:    output.Warning,
			args: args{msg: "panic", fields: map[string]any{}},
			want: "level=panic msg=panic\n",
		},
		"info": {
			l:    output.Info,
			args: args{msg: "panic", fields: nil},
			want: "level=panic msg=panic\n",
		},
		"debug": {
			l: output.Debug,
			args: args{msg: "panic message", fields: map[string]any{
				"field1":  "value1",
				"field2":  2,
				"field3":  true,
				"field4":  []string{"val1", "val2"},
				"field5":  fmt.Errorf("this is a mistake"),
				"field6":  map[string]int{"a": 1, "b b": 2},
				"field7":  regexp.MustCompile(".*"),
				"field8":  "a b c",
				"field 9": false,
			}},
			want: "level=panic msg=\"panic message\" field 9=false field1=value1 field2=2 field3=true field4=\"[val1 val2]\" field5=\"this is a mistake\" field6=\"map[a:1 b b:2]\" field7=\".*\" field8=\"a b c\"\n",
		},
		"trace": {
			l:    output.Trace,
			args: args{msg: "panic", fields: map[string]any{}},
			want: "level=panic msg=panic\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			writerGetter = func(_ output.Bus) io.Writer {
				return buffer
			}
			InitLoggingWithLevel(nil, tt.l)
			defer func(t *testing.T) {
				if r := recover(); r == nil {
					t.Errorf("simpleLogger.Panic() did not panic")
				}
			}(t)
			ProductionLogger.Panic(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("simpleLogger.Panic() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("simpleLogger.Panic() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func Test_simpleLogger_Trace(t *testing.T) {
	savedGetWriterFunc := writerGetter
	defer func() {
		writerGetter = savedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		l output.Level
		args
		want string
	}{
		"panic": {
			l:    output.Panic,
			args: args{msg: "trace", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			l:    output.Fatal,
			args: args{msg: "trace", fields: map[string]any{}},
			want: "",
		},
		"error": {
			l:    output.Error,
			args: args{msg: "trace", fields: map[string]any{}},
			want: "",
		},
		"warn": {
			l:    output.Warning,
			args: args{msg: "trace", fields: map[string]any{}},
			want: "",
		},
		"info": {
			l:    output.Info,
			args: args{msg: "trace", fields: nil},
			want: "",
		},
		"debug": {
			l: output.Debug,
			args: args{msg: "trace message", fields: map[string]any{
				"field1":  "value1",
				"field2":  2,
				"field3":  true,
				"field4":  []string{"val1", "val2"},
				"field5":  fmt.Errorf("this is a mistake"),
				"field6":  map[string]int{"a": 1, "b b": 2},
				"field7":  regexp.MustCompile(".*"),
				"field8":  "a b c",
				"field 9": false,
			}},
			want: "",
		},
		"trace": {
			l:    output.Trace,
			args: args{msg: "trace", fields: map[string]any{}},
			want: "level=trace msg=trace\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			writerGetter = func(_ output.Bus) io.Writer {
				return buffer
			}
			InitLoggingWithLevel(nil, tt.l)
			ProductionLogger.Trace(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("simpleLogger.Trace() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("simpleLogger.Trace() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func Test_simpleLogger_Warning(t *testing.T) {
	savedGetWriterFunc := writerGetter
	defer func() {
		writerGetter = savedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		l output.Level
		args
		want string
	}{
		"panic": {
			l:    output.Panic,
			args: args{msg: "warn", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			l:    output.Fatal,
			args: args{msg: "warn", fields: map[string]any{}},
			want: "",
		},
		"error": {
			l:    output.Error,
			args: args{msg: "warn", fields: map[string]any{}},
			want: "",
		},
		"warn": {
			l:    output.Warning,
			args: args{msg: "warn", fields: map[string]any{}},
			want: "level=warning msg=warn\n",
		},
		"info": {
			l:    output.Info,
			args: args{msg: "warn", fields: nil},
			want: "level=warning msg=warn\n",
		},
		"debug": {
			l: output.Debug,
			args: args{msg: "warn message", fields: map[string]any{
				"field1":  "value1",
				"field2":  2,
				"field3":  true,
				"field4":  []string{"val1", "val2"},
				"field5":  fmt.Errorf("this is a mistake"),
				"field6":  map[string]int{"a": 1, "b b": 2},
				"field7":  regexp.MustCompile(".*"),
				"field8":  "a b c",
				"field 9": false,
			}},
			want: "level=warning msg=\"warn message\" field 9=false field1=value1 field2=2 field3=true field4=\"[val1 val2]\" field5=\"this is a mistake\" field6=\"map[a:1 b b:2]\" field7=\".*\" field8=\"a b c\"\n",
		},
		"trace": {
			l:    output.Trace,
			args: args{msg: "warn", fields: map[string]any{}},
			want: "level=warning msg=warn\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			writerGetter = func(_ output.Bus) io.Writer {
				return buffer
			}
			InitLoggingWithLevel(nil, tt.l)
			ProductionLogger.Warning(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("simpleLogger.Warn() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("simpleLogger.Warn() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func Test_simpleLogger_willLog(t *testing.T) {
	type args struct {
		l output.Level
	}
	tests := map[string]struct {
		cl output.Level
		args
		want bool
	}{
		"debug/debug":     {cl: output.Debug, args: args{l: output.Debug}, want: true},
		"debug/error":     {cl: output.Debug, args: args{l: output.Error}, want: true},
		"debug/fatal":     {cl: output.Debug, args: args{l: output.Fatal}, want: true},
		"debug/info":      {cl: output.Debug, args: args{l: output.Info}, want: true},
		"debug/panic":     {cl: output.Debug, args: args{l: output.Panic}, want: true},
		"debug/trace":     {cl: output.Debug, args: args{l: output.Trace}, want: false},
		"debug/warning":   {cl: output.Debug, args: args{l: output.Warning}, want: true},
		"error/debug":     {cl: output.Error, args: args{l: output.Debug}, want: false},
		"error/error":     {cl: output.Error, args: args{l: output.Error}, want: true},
		"error/fatal":     {cl: output.Error, args: args{l: output.Fatal}, want: true},
		"error/info":      {cl: output.Error, args: args{l: output.Info}, want: false},
		"error/panic":     {cl: output.Error, args: args{l: output.Panic}, want: true},
		"error/trace":     {cl: output.Error, args: args{l: output.Trace}, want: false},
		"error/warning":   {cl: output.Error, args: args{l: output.Warning}, want: false},
		"fatal/debug":     {cl: output.Fatal, args: args{l: output.Debug}, want: false},
		"fatal/error":     {cl: output.Fatal, args: args{l: output.Error}, want: false},
		"fatal/fatal":     {cl: output.Fatal, args: args{l: output.Fatal}, want: true},
		"fatal/info":      {cl: output.Fatal, args: args{l: output.Info}, want: false},
		"fatal/panic":     {cl: output.Fatal, args: args{l: output.Panic}, want: false},
		"fatal/trace":     {cl: output.Fatal, args: args{l: output.Trace}, want: false},
		"fatal/warning":   {cl: output.Fatal, args: args{l: output.Warning}, want: false},
		"info/debug":      {cl: output.Info, args: args{l: output.Debug}, want: false},
		"info/error":      {cl: output.Info, args: args{l: output.Error}, want: true},
		"info/fatal":      {cl: output.Info, args: args{l: output.Fatal}, want: true},
		"info/info":       {cl: output.Info, args: args{l: output.Info}, want: true},
		"info/panic":      {cl: output.Info, args: args{l: output.Panic}, want: true},
		"info/trace":      {cl: output.Info, args: args{l: output.Trace}, want: false},
		"info/warning":    {cl: output.Info, args: args{l: output.Warning}, want: true},
		"panic/debug":     {cl: output.Panic, args: args{l: output.Debug}, want: false},
		"panic/error":     {cl: output.Panic, args: args{l: output.Error}, want: false},
		"panic/fatal":     {cl: output.Panic, args: args{l: output.Fatal}, want: true},
		"panic/info":      {cl: output.Panic, args: args{l: output.Info}, want: false},
		"panic/panic":     {cl: output.Panic, args: args{l: output.Panic}, want: true},
		"panic/trace":     {cl: output.Panic, args: args{l: output.Trace}, want: false},
		"panic/warning":   {cl: output.Panic, args: args{l: output.Warning}, want: false},
		"trace/debug":     {cl: output.Trace, args: args{l: output.Debug}, want: true},
		"trace/error":     {cl: output.Trace, args: args{l: output.Error}, want: true},
		"trace/fatal":     {cl: output.Trace, args: args{l: output.Fatal}, want: true},
		"trace/info":      {cl: output.Trace, args: args{l: output.Info}, want: true},
		"trace/panic":     {cl: output.Trace, args: args{l: output.Panic}, want: true},
		"trace/trace":     {cl: output.Trace, args: args{l: output.Trace}, want: true},
		"trace/warning":   {cl: output.Trace, args: args{l: output.Warning}, want: true},
		"warning/debug":   {cl: output.Warning, args: args{l: output.Debug}, want: false},
		"warning/error":   {cl: output.Warning, args: args{l: output.Error}, want: true},
		"warning/fatal":   {cl: output.Warning, args: args{l: output.Fatal}, want: true},
		"warning/info":    {cl: output.Warning, args: args{l: output.Info}, want: false},
		"warning/panic":   {cl: output.Warning, args: args{l: output.Panic}, want: true},
		"warning/trace":   {cl: output.Warning, args: args{l: output.Trace}, want: false},
		"warning/warning": {cl: output.Warning, args: args{l: output.Warning}, want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			pl := simpleLogger{currentLogLevel: tt.cl}
			if got := pl.willLog(tt.args.l); got != tt.want {
				t.Errorf("simpleLogger.willLog() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_requiresQuotes(t *testing.T) {
	tests := map[string]struct {
		s    string
		want bool
	}{
		"empty":         {s: "", want: true},
		"alphanumerics": {s: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", want: false},
		"specials":      {s: "-._/@^+", want: false},
		"others":        {s: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._/@^+ ", want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := requiresQuotes(tt.s); got != tt.want {
				t.Errorf("requiresQuotes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toString(t *testing.T) {
	tests := map[string]struct {
		v    any
		want string
	}{
		"empty":              {v: "", want: `""`},
		"number":             {v: 1, want: "1"},
		"true":               {v: true, want: "true"},
		"false":              {v: false, want: "false"},
		"plain string":       {v: "hello", want: "hello"},
		"string with spaces": {v: "hello fencepost", want: `"hello fencepost"`},
		"array":              {v: []string{"foo", "bar"}, want: `"[foo bar]"`},
		"regexp":             {v: typicalChars, want: fmt.Sprintf("%q", typicalChars)},
		"error":              {v: fmt.Errorf("this is a mistake"), want: `"this is a mistake"`},
		"map":                {v: map[string]int{"a": 1, "b b": 2}, want: `"map[a:1 b b:2]"`},
		/*
			fmt.Errorf("this is a mistake"),
							"field6":  map[string]int{"a": 1, "b b": 2},		*/
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := toString(tt.v); got != tt.want {
				t.Errorf("toString() = %v, want %v", got, tt.want)
			}
		})
	}
}
