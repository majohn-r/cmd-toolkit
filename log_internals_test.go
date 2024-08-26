package cmd_toolkit

import (
	"bytes"
	"fmt"
	"github.com/majohn-r/output"
	"io"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

func Test_simpleLogger_Debug(t *testing.T) {
	originalLogWriterInitFn := LogWriterInitFn
	defer func() {
		LogWriterInitFn = originalLogWriterInitFn
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
			want: "" +
				"level=debug " +
				"msg=\"debug message\" " +
				"field 9=false " +
				"field1=value1 " +
				"field2=2 " +
				"field3=true " +
				"field4=\"[val1 val2]\" " +
				"field5=\"this is a mistake\" " +
				"field6=\"map[a:1 b b:2]\" " +
				"field7=\".*\" " +
				"field8=\"a b c\"\n",
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
			LogWriterInitFn = func(_ output.Bus, _ string) (io.Writer, string) {
				return buffer, ""
			}
			InitLoggingWithLevel(nil, tt.l, "myApp")
			ProductionLogger.Debug(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("simpleLogger.Debug() got %q want %q", got, tt.want)
					}
					if strings.HasPrefix(got, " ") {
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
	originalLogWriterInitFn := LogWriterInitFn
	defer func() {
		LogWriterInitFn = originalLogWriterInitFn
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
			want: "" +
				"level=error " +
				"msg=error " +
				"field 9=false " +
				"field1=value1 " +
				"field2=2 " +
				"field3=true " +
				"field4=\"[val1 val2]\" " +
				"field5=\"this is a mistake\" " +
				"field6=\"map[a:1 b b:2]\" " +
				"field7=\".*\" " +
				"field8=\"a b c\"\n",
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
			LogWriterInitFn = func(_ output.Bus, _ string) (io.Writer, string) {
				return buffer, ""
			}
			InitLoggingWithLevel(nil, tt.l, "myApp")
			ProductionLogger.Error(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("simpleLogger.Error() got %q want %q", got, tt.want)
					}
					if strings.HasPrefix(got, " ") {
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
	originalLogWriterInitFn := LogWriterInitFn
	savedExitFunction := ProductionLogger.exitFunction
	defer func() {
		LogWriterInitFn = originalLogWriterInitFn
		ProductionLogger.exitFunction = savedExitFunction
	}()
	var exited = false
	pretendToExit := func(_ int) {
		exited = true
	}
	ProductionLogger.exitFunction = pretendToExit
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
			want: "" +
				"level=fatal " +
				"msg=\"fatal message\" " +
				"field 9=false " +
				"field1=value1 " +
				"field2=2 " +
				"field3=true " +
				"field4=\"[val1 val2]\" " +
				"field5=\"this is a mistake\" " +
				"field6=\"map[a:1 b b:2]\" " +
				"field7=\".*\" " +
				"field8=\"a b c\"\n",
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
			LogWriterInitFn = func(_ output.Bus, _ string) (io.Writer, string) {
				return buffer, ""
			}
			InitLoggingWithLevel(nil, tt.l, "myApp")
			exited = false
			f := ProductionLogger.exitFunction
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
					if strings.HasPrefix(got, " ") {
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
	originalLogWriterInitFn := LogWriterInitFn
	defer func() {
		LogWriterInitFn = originalLogWriterInitFn
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
			want: "" +
				"level=info " +
				"msg=\"info message\" " +
				"field 9=false " +
				"field1=value1 " +
				"field2=2 " +
				"field3=true " +
				"field4=\"[val1 val2]\" " +
				"field5=\"this is a mistake\" " +
				"field6=\"map[a:1 b b:2]\" " +
				"field7=\".*\" " +
				"field8=\"a b c\"\n",
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
			LogWriterInitFn = func(_ output.Bus, _ string) (io.Writer, string) {
				return buffer, ""
			}
			InitLoggingWithLevel(nil, tt.l, "myApp")
			ProductionLogger.Info(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("simpleLogger.Info() got %q want %q", got, tt.want)
					}
					if strings.HasPrefix(got, " ") {
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
	originalLogWriterInitFn := LogWriterInitFn
	defer func() {
		LogWriterInitFn = originalLogWriterInitFn
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
			want: "" +
				"level=panic " +
				"msg=\"panic message\" " +
				"field 9=false " +
				"field1=value1 " +
				"field2=2 " +
				"field3=true " +
				"field4=\"[val1 val2]\" " +
				"field5=\"this is a mistake\" " +
				"field6=\"map[a:1 b b:2]\" " +
				"field7=\".*\" " +
				"field8=\"a b c\"\n",
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
			LogWriterInitFn = func(_ output.Bus, _ string) (io.Writer, string) {
				return buffer, ""
			}
			InitLoggingWithLevel(nil, tt.l, "myApp")
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
					if strings.HasPrefix(got, " ") {
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
	originalLogWriterInitFn := LogWriterInitFn
	defer func() {
		LogWriterInitFn = originalLogWriterInitFn
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
			LogWriterInitFn = func(_ output.Bus, _ string) (io.Writer, string) {
				return buffer, ""
			}
			InitLoggingWithLevel(nil, tt.l, "myApp")
			ProductionLogger.Trace(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("simpleLogger.Trace() got %q want %q", got, tt.want)
					}
					if strings.HasPrefix(got, " ") {
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
	originalLogWriterInitFn := LogWriterInitFn
	defer func() {
		LogWriterInitFn = originalLogWriterInitFn
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
			want: "" +
				"level=warning " +
				"msg=\"warn message\" " +
				"field 9=false " +
				"field1=value1 " +
				"field2=2 " +
				"field3=true " +
				"field4=\"[val1 val2]\" " +
				"field5=\"this is a mistake\" " +
				"field6=\"map[a:1 b b:2]\" " +
				"field7=\".*\" " +
				"field8=\"a b c\"\n",
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
			LogWriterInitFn = func(_ output.Bus, _ string) (io.Writer, string) {
				return buffer, ""
			}
			InitLoggingWithLevel(nil, tt.l, "myApp")
			ProductionLogger.Warning(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("simpleLogger.Warn() got %q want %q", got, tt.want)
					}
					if strings.HasPrefix(got, " ") {
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
	tests := map[string]struct {
		cl   output.Level
		l    output.Level
		want bool
	}{
		"debug/debug": {
			cl:   output.Debug,
			l:    output.Debug,
			want: true,
		},
		"debug/error": {
			cl:   output.Debug,
			l:    output.Error,
			want: true,
		},
		"debug/fatal": {
			cl:   output.Debug,
			l:    output.Fatal,
			want: true,
		},
		"debug/info": {
			cl:   output.Debug,
			l:    output.Info,
			want: true,
		},
		"debug/panic": {
			cl:   output.Debug,
			l:    output.Panic,
			want: true,
		},
		"debug/trace": {
			cl:   output.Debug,
			l:    output.Trace,
			want: false,
		},
		"debug/warning": {
			cl:   output.Debug,
			l:    output.Warning,
			want: true,
		},
		"error/debug": {
			cl:   output.Error,
			l:    output.Debug,
			want: false,
		},
		"error/error": {
			cl:   output.Error,
			l:    output.Error,
			want: true,
		},
		"error/fatal": {
			cl:   output.Error,
			l:    output.Fatal,
			want: true,
		},
		"error/info": {
			cl:   output.Error,
			l:    output.Info,
			want: false,
		},
		"error/panic": {
			cl:   output.Error,
			l:    output.Panic,
			want: true,
		},
		"error/trace": {
			cl:   output.Error,
			l:    output.Trace,
			want: false,
		},
		"error/warning": {
			cl:   output.Error,
			l:    output.Warning,
			want: false,
		},
		"fatal/debug": {
			cl:   output.Fatal,
			l:    output.Debug,
			want: false,
		},
		"fatal/error": {
			cl:   output.Fatal,
			l:    output.Error,
			want: false,
		},
		"fatal/fatal": {
			cl:   output.Fatal,
			l:    output.Fatal,
			want: true,
		},
		"fatal/info": {
			cl:   output.Fatal,
			l:    output.Info,
			want: false,
		},
		"fatal/panic": {
			cl:   output.Fatal,
			l:    output.Panic,
			want: false,
		},
		"fatal/trace": {
			cl:   output.Fatal,
			l:    output.Trace,
			want: false,
		},
		"fatal/warning": {
			cl:   output.Fatal,
			l:    output.Warning,
			want: false,
		},
		"info/debug": {
			cl:   output.Info,
			l:    output.Debug,
			want: false,
		},
		"info/error": {
			cl:   output.Info,
			l:    output.Error,
			want: true,
		},
		"info/fatal": {
			cl:   output.Info,
			l:    output.Fatal,
			want: true,
		},
		"info/info": {
			cl:   output.Info,
			l:    output.Info,
			want: true,
		},
		"info/panic": {
			cl:   output.Info,
			l:    output.Panic,
			want: true,
		},
		"info/trace": {
			cl:   output.Info,
			l:    output.Trace,
			want: false,
		},
		"info/warning": {
			cl:   output.Info,
			l:    output.Warning,
			want: true,
		},
		"panic/debug": {
			cl:   output.Panic,
			l:    output.Debug,
			want: false,
		},
		"panic/error": {
			cl:   output.Panic,
			l:    output.Error,
			want: false,
		},
		"panic/fatal": {
			cl:   output.Panic,
			l:    output.Fatal,
			want: true,
		},
		"panic/info": {
			cl:   output.Panic,
			l:    output.Info,
			want: false,
		},
		"panic/panic": {
			cl:   output.Panic,
			l:    output.Panic,
			want: true,
		},
		"panic/trace": {
			cl:   output.Panic,
			l:    output.Trace,
			want: false,
		},
		"panic/warning": {
			cl:   output.Panic,
			l:    output.Warning,
			want: false,
		},
		"trace/debug": {
			cl:   output.Trace,
			l:    output.Debug,
			want: true,
		},
		"trace/error": {
			cl:   output.Trace,
			l:    output.Error,
			want: true,
		},
		"trace/fatal": {
			cl:   output.Trace,
			l:    output.Fatal,
			want: true,
		},
		"trace/info": {
			cl:   output.Trace,
			l:    output.Info,
			want: true,
		},
		"trace/panic": {
			cl:   output.Trace,
			l:    output.Panic,
			want: true,
		},
		"trace/trace": {
			cl:   output.Trace,
			l:    output.Trace,
			want: true,
		},
		"trace/warning": {
			cl:   output.Trace,
			l:    output.Warning,
			want: true,
		},
		"warning/debug": {
			cl:   output.Warning,
			l:    output.Debug,
			want: false,
		},
		"warning/error": {
			cl:   output.Warning,
			l:    output.Error,
			want: true,
		},
		"warning/fatal": {
			cl:   output.Warning,
			l:    output.Fatal,
			want: true,
		},
		"warning/info": {
			cl:   output.Warning,
			l:    output.Info,
			want: false,
		},
		"warning/panic": {
			cl:   output.Warning,
			l:    output.Panic,
			want: true,
		},
		"warning/trace": {
			cl:   output.Warning,
			l:    output.Trace,
			want: false,
		},
		"warning/warning": {
			cl:   output.Warning,
			l:    output.Warning,
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			pl := simpleLogger{currentLogLevel: tt.cl}
			if got := pl.WillLog(tt.l); got != tt.want {
				t.Errorf("simpleLogger.WillLog() = %v, want %v", got, tt.want)
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

func Test_simpleLogger_doLog(t *testing.T) {
	timestamp := time.Unix(0, 0)
	expectedTime := `time="` + timestamp.Format(time.RFC3339) + `"`
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		args
		want string
	}{
		"empty message, no fields": {
			args: args{
				msg:    "",
				fields: nil,
			},
			want: expectedTime +
				" level=info" +
				" msg=\"\"\n",
		},
		"simple message, one field": {
			args: args{
				msg:    "hello",
				fields: map[string]any{"field1": 45},
			},
			want: expectedTime +
				" level=info" +
				" msg=hello" +
				" field1=45\n",
		},
		"interesting message, multiple fields": {
			args: args{
				msg: "hello fencepost",
				fields: map[string]any{
					"field1": 45,
					"field2": true,
					"field3": []string{"hello", "fence post"},
				},
			},
			want: expectedTime +
				" level=info" +
				" msg=\"hello fencepost\"" +
				" field1=45" +
				" field2=true" +
				" field3=\"[hello fence post]\"\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			sl := &simpleLogger{
				writer: buffer,
				lock:   &sync.RWMutex{},
			}
			sl.doLog(output.Info, timestamp, tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				t.Errorf("simpleLogger.doLog() = %q, want %q", got, tt.want)
			}
		})
	}
}
