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

func TestOldInitLogging(t *testing.T) {
	tests := map[string]struct {
		oldWriterGetter func(o output.Bus) io.Writer
		want            bool
	}{
		"no writer available": {
			oldWriterGetter: func(o output.Bus) io.Writer { return nil },
			want:            false,
		},
		"success": {
			oldWriterGetter: func(o output.Bus) io.Writer { return &bytes.Buffer{} },
			want:            true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			oldFunc := OldWriterGetter
			defer func() {
				OldWriterGetter = oldFunc
			}()
			OldWriterGetter = tt.oldWriterGetter
			if got := OldInitLogging(nil); got != tt.want {
				t.Errorf("OldInitLogging() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOldInitLoggingWithLevel(t *testing.T) {
	oldoldSavedGetWriterFunc := OldWriterGetter
	defer func() {
		OldWriterGetter = oldoldSavedGetWriterFunc
	}()
	OldWriterGetter = func(_ output.Bus) io.Writer {
		return &bytes.Buffer{}
	}
	// only going to vary the logging level - TestOldInitLogging handles the error
	// cases where a writer cannot be obtained. This ensures that we don't
	// introduce a programming error when the underlying log implementation
	// cannot be initialized with the specified log level
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
			if gotOk := OldInitLoggingWithLevel(nil, tt.l); gotOk != tt.wantOk {
				t.Errorf("OldInitLoggingWithLevel() = %t, want %t", gotOk, tt.wantOk)
			}
		})
	}
}

func TestOldProductionLogger_Debug(t *testing.T) {
	oldoldSavedGetWriterFunc := OldWriterGetter
	defer func() {
		OldWriterGetter = oldoldSavedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		opl OldProductionLogger
		l   output.Level
		args
		want string
	}{
		"panic": {
			opl:  OldProductionLogger{},
			l:    output.Panic,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			opl:  OldProductionLogger{},
			l:    output.Fatal,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "",
		},
		"error": {
			opl:  OldProductionLogger{},
			l:    output.Error,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "",
		},
		"warn": {
			opl:  OldProductionLogger{},
			l:    output.Warning,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "",
		},
		"info": {
			opl:  OldProductionLogger{},
			l:    output.Info,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "",
		},
		"debug": {
			opl: OldProductionLogger{},
			l:   output.Debug,
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
			opl:  OldProductionLogger{},
			l:    output.Trace,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "level=debug msg=debug\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			OldWriterGetter = func(_ output.Bus) io.Writer {
				return buffer
			}
			OldInitLoggingWithLevel(nil, tt.l)
			tt.opl.Debug(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("OldProductionLogger.Debug() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("OldProductionLogger.Debug() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func TestOldProductionLogger_Error(t *testing.T) {
	oldoldSavedGetWriterFunc := OldWriterGetter
	defer func() {
		OldWriterGetter = oldoldSavedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		opl OldProductionLogger
		l   output.Level
		args
		want string
	}{
		"panic": {
			opl:  OldProductionLogger{},
			l:    output.Panic,
			args: args{msg: "error", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			opl:  OldProductionLogger{},
			l:    output.Fatal,
			args: args{msg: "error", fields: map[string]any{}},
			want: "",
		},
		"error": {
			opl:  OldProductionLogger{},
			l:    output.Error,
			args: args{msg: "error", fields: map[string]any{}},
			want: "level=error msg=error\n",
		},
		"warn": {
			opl:  OldProductionLogger{},
			l:    output.Warning,
			args: args{msg: "error", fields: map[string]any{}},
			want: "level=error msg=error\n",
		},
		"info": {
			opl:  OldProductionLogger{},
			l:    output.Info,
			args: args{msg: "error", fields: nil},
			want: "level=error msg=error\n",
		},
		"debug": {
			opl: OldProductionLogger{},
			l:   output.Debug,
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
			opl:  OldProductionLogger{},
			l:    output.Trace,
			args: args{msg: "error", fields: map[string]any{}},
			want: "level=error msg=error\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			OldWriterGetter = func(_ output.Bus) io.Writer {
				return buffer
			}
			OldInitLoggingWithLevel(nil, tt.l)
			tt.opl.Error(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("OldProductionLogger.Error() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("OldProductionLogger.Error() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func TestOldProductionLogger_Fatal(t *testing.T) {
	oldoldSavedGetWriterFunc := OldWriterGetter
	defer func() {
		OldWriterGetter = oldoldSavedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		opl OldProductionLogger
		l   output.Level
		args
		want string
	}{
		"panic": {
			opl:  OldProductionLogger{},
			l:    output.Panic,
			args: args{msg: "fatal", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			opl:  OldProductionLogger{},
			l:    output.Fatal,
			args: args{msg: "fatal", fields: map[string]any{}},
			want: "level=fatal msg=fatal\n",
		},
		"error": {
			opl:  OldProductionLogger{},
			l:    output.Error,
			args: args{msg: "fatal", fields: map[string]any{}},
			want: "level=fatal msg=fatal\n",
		},
		"warn": {
			opl:  OldProductionLogger{},
			l:    output.Warning,
			args: args{msg: "fatal", fields: map[string]any{}},
			want: "level=fatal msg=fatal\n",
		},
		"info": {
			opl:  OldProductionLogger{},
			l:    output.Info,
			args: args{msg: "fatal", fields: nil},
			want: "level=fatal msg=fatal\n",
		},
		"debug": {
			opl: OldProductionLogger{},
			l:   output.Debug,
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
			opl:  OldProductionLogger{},
			l:    output.Trace,
			args: args{msg: "fatal", fields: map[string]any{}},
			want: "level=fatal msg=fatal\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			OldWriterGetter = func(_ output.Bus) io.Writer {
				return buffer
			}
			OldInitLoggingWithLevel(nil, tt.l)
			savedExit := tt.opl.ExitFunc()
			defer func() {
				tt.opl.SetExitFunc(savedExit)
			}()
			exited := false
			tt.opl.SetExitFunc(func(_ int) {
				exited = true
			})
			tt.opl.Fatal(tt.args.msg, tt.args.fields)
			if !exited {
				t.Errorf("OldProductionLogger.Fatal() did not attempt to exit!")
			}
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("OldProductionLogger.Fatal() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("OldProductionLogger.Fatal() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func TestOldProductionLogger_Info(t *testing.T) {
	oldSavedGetWriterFunc := OldWriterGetter
	defer func() {
		OldWriterGetter = oldSavedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		opl OldProductionLogger
		l   output.Level
		args
		want string
	}{
		"panic": {
			opl:  OldProductionLogger{},
			l:    output.Panic,
			args: args{msg: "info", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			opl:  OldProductionLogger{},
			l:    output.Fatal,
			args: args{msg: "info", fields: map[string]any{}},
			want: "",
		},
		"error": {
			opl:  OldProductionLogger{},
			l:    output.Error,
			args: args{msg: "info", fields: map[string]any{}},
			want: "",
		},
		"warn": {
			opl:  OldProductionLogger{},
			l:    output.Warning,
			args: args{msg: "info", fields: map[string]any{}},
			want: "",
		},
		"info": {
			opl:  OldProductionLogger{},
			l:    output.Info,
			args: args{msg: "info", fields: nil},
			want: "level=info msg=info\n",
		},
		"debug": {
			opl: OldProductionLogger{},
			l:   output.Debug,
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
			opl:  OldProductionLogger{},
			l:    output.Trace,
			args: args{msg: "info", fields: map[string]any{}},
			want: "level=info msg=info\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			OldWriterGetter = func(_ output.Bus) io.Writer {
				return buffer
			}
			OldInitLoggingWithLevel(nil, tt.l)
			tt.opl.Info(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("OldProductionLogger.Info() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("OldProductionLogger.Info() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func TestOldProductionLogger_Panic(t *testing.T) {
	oldSavedGetWriterFunc := OldWriterGetter
	defer func() {
		OldWriterGetter = oldSavedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl OldProductionLogger
		l  output.Level
		args
		want string
	}{
		"panic": {
			pl:   OldProductionLogger{},
			l:    output.Panic,
			args: args{msg: "panic", fields: map[string]any{}},
			want: "level=panic msg=panic\n",
		},
		"fatal": {
			pl:   OldProductionLogger{},
			l:    output.Fatal,
			args: args{msg: "panic", fields: map[string]any{}},
			want: "level=panic msg=panic\n",
		},
		"error": {
			pl:   OldProductionLogger{},
			l:    output.Error,
			args: args{msg: "panic", fields: map[string]any{}},
			want: "level=panic msg=panic\n",
		},
		"warn": {
			pl:   OldProductionLogger{},
			l:    output.Warning,
			args: args{msg: "panic", fields: map[string]any{}},
			want: "level=panic msg=panic\n",
		},
		"info": {
			pl:   OldProductionLogger{},
			l:    output.Info,
			args: args{msg: "panic", fields: nil},
			want: "level=panic msg=panic\n",
		},
		"debug": {
			pl: OldProductionLogger{},
			l:  output.Debug,
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
			pl:   OldProductionLogger{},
			l:    output.Trace,
			args: args{msg: "panic", fields: map[string]any{}},
			want: "level=panic msg=panic\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			OldWriterGetter = func(_ output.Bus) io.Writer {
				return buffer
			}
			OldInitLoggingWithLevel(nil, tt.l)
			defer func(t *testing.T) {
				if r := recover(); r == nil {
					t.Errorf("OldProductionLogger.Panic() did not panic")
				}
			}(t)
			tt.pl.Panic(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("OldProductionLogger.Panic() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("OldProductionLogger.Panic() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func TestOldProductionLogger_Trace(t *testing.T) {
	oldSavedGetWriterFunc := OldWriterGetter
	defer func() {
		OldWriterGetter = oldSavedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl OldProductionLogger
		l  output.Level
		args
		want string
	}{
		"panic": {
			pl:   OldProductionLogger{},
			l:    output.Panic,
			args: args{msg: "trace", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			pl:   OldProductionLogger{},
			l:    output.Fatal,
			args: args{msg: "trace", fields: map[string]any{}},
			want: "",
		},
		"error": {
			pl:   OldProductionLogger{},
			l:    output.Error,
			args: args{msg: "trace", fields: map[string]any{}},
			want: "",
		},
		"warn": {
			pl:   OldProductionLogger{},
			l:    output.Warning,
			args: args{msg: "trace", fields: map[string]any{}},
			want: "",
		},
		"info": {
			pl:   OldProductionLogger{},
			l:    output.Info,
			args: args{msg: "trace", fields: nil},
			want: "",
		},
		"debug": {
			pl: OldProductionLogger{},
			l:  output.Debug,
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
			pl:   OldProductionLogger{},
			l:    output.Trace,
			args: args{msg: "trace", fields: map[string]any{}},
			want: "level=trace msg=trace\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			OldWriterGetter = func(_ output.Bus) io.Writer {
				return buffer
			}
			OldInitLoggingWithLevel(nil, tt.l)
			tt.pl.Trace(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("OldProductionLogger.Trace() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("OldProductionLogger.Trace() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func TestOldProductionLogger_Warning(t *testing.T) {
	oldSavedGetWriterFunc := OldWriterGetter
	defer func() {
		OldWriterGetter = oldSavedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl OldProductionLogger
		l  output.Level
		args
		want string
	}{
		"panic": {
			pl:   OldProductionLogger{},
			l:    output.Panic,
			args: args{msg: "warn", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			pl:   OldProductionLogger{},
			l:    output.Fatal,
			args: args{msg: "warn", fields: map[string]any{}},
			want: "",
		},
		"error": {
			pl:   OldProductionLogger{},
			l:    output.Error,
			args: args{msg: "warn", fields: map[string]any{}},
			want: "",
		},
		"warn": {
			pl:   OldProductionLogger{},
			l:    output.Warning,
			args: args{msg: "warn", fields: map[string]any{}},
			want: "level=warning msg=warn\n",
		},
		"info": {
			pl:   OldProductionLogger{},
			l:    output.Info,
			args: args{msg: "warn", fields: nil},
			want: "level=warning msg=warn\n",
		},
		"debug": {
			pl: OldProductionLogger{},
			l:  output.Debug,
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
			pl:   OldProductionLogger{},
			l:    output.Trace,
			args: args{msg: "warn", fields: map[string]any{}},
			want: "level=warning msg=warn\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			OldWriterGetter = func(_ output.Bus) io.Writer {
				return buffer
			}
			OldInitLoggingWithLevel(nil, tt.l)
			tt.pl.Warning(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("OldProductionLogger.Warn() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("OldProductionLogger.Warn() got %q want %q", got, tt.want)
				}
			}
		})
	}
}
