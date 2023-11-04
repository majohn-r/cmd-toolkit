package cmd_toolkit

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/majohn-r/output"
)

func TestInitLogging(t *testing.T) {
	tests := map[string]struct {
		writerGetter func(o output.Bus) io.Writer
		want         bool
	}{
		"no writer available": {
			writerGetter: func(o output.Bus) io.Writer { return nil },
			want:         false,
		},
		"success": {
			writerGetter: func(o output.Bus) io.Writer { return &bytes.Buffer{} },
			want:         true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			oldFunc := writerGetter
			defer func() {
				writerGetter = oldFunc
			}()
			writerGetter = tt.writerGetter
			if got := InitLogging(nil); got != tt.want {
				t.Errorf("InitLogging() = %v, want %v", got, tt.want)
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
	// cannot be initialized with the specified log level
	tests := map[string]struct {
		l      LoggingLevel
		wantOk bool
	}{
		"panic": {l: Panic, wantOk: true},
		"fatal": {l: Fatal, wantOk: true},
		"error": {l: Error, wantOk: true},
		"warn":  {l: Warn, wantOk: true},
		"info":  {l: Info, wantOk: true},
		"debug": {l: Debug, wantOk: true},
		"trace": {l: Trace, wantOk: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotOk := InitLoggingWithLevel(nil, tt.l); gotOk != tt.wantOk {
				t.Errorf("InitLoggingWithLevel() = %t, want %t", gotOk, tt.wantOk)
			}
		})
	}
}

func TestProductionLogger_Debug(t *testing.T) {
	savedGetWriterFunc := writerGetter
	defer func() {
		writerGetter = savedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl ProductionLogger
		l  LoggingLevel
		args
		want string
	}{
		"panic": {
			pl:   ProductionLogger{},
			l:    Panic,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			pl:   ProductionLogger{},
			l:    Fatal,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "",
		},
		"error": {
			pl:   ProductionLogger{},
			l:    Error,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "",
		},
		"warn": {
			pl:   ProductionLogger{},
			l:    Warn,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "",
		},
		"info": {
			pl:   ProductionLogger{},
			l:    Info,
			args: args{msg: "debug", fields: map[string]any{}},
			want: "",
		},
		"debug": {
			pl: ProductionLogger{},
			l:  Debug,
			args: args{msg: "debug message", fields: map[string]any{
				"field1": "value1",
				"field2": 2,
				"field3": true,
				"field4": []string{"val1", "val2"},
			}},
			want: "level=debug msg=\"debug message\" field1=value1 field2=2 field3=true field4=\"[val1 val2]\"\n",
		},
		"trace": {
			pl:   ProductionLogger{},
			l:    Trace,
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
			tt.pl.Debug(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("ProductionLogger.Debug() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("ProductionLogger.Debug() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func TestProductionLogger_Error(t *testing.T) {
	savedGetWriterFunc := writerGetter
	defer func() {
		writerGetter = savedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl ProductionLogger
		l  LoggingLevel
		args
		want string
	}{
		"panic": {
			pl:   ProductionLogger{},
			l:    Panic,
			args: args{msg: "error", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			pl:   ProductionLogger{},
			l:    Fatal,
			args: args{msg: "error", fields: map[string]any{}},
			want: "",
		},
		"error": {
			pl:   ProductionLogger{},
			l:    Error,
			args: args{msg: "error", fields: map[string]any{}},
			want: "level=error msg=error\n",
		},
		"warn": {
			pl:   ProductionLogger{},
			l:    Warn,
			args: args{msg: "error", fields: map[string]any{}},
			want: "level=error msg=error\n",
		},
		"info": {
			pl:   ProductionLogger{},
			l:    Info,
			args: args{msg: "error", fields: nil},
			want: "level=error msg=error\n",
		},
		"debug": {
			pl: ProductionLogger{},
			l:  Debug,
			args: args{msg: "error", fields: map[string]any{
				"field1": "value1",
				"field2": 2,
				"field3": true,
				"field4": []string{"val1", "val2"},
			}},
			want: "level=error msg=error field1=value1 field2=2 field3=true field4=\"[val1 val2]\"\n",
		},
		"trace": {
			pl:   ProductionLogger{},
			l:    Trace,
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
			tt.pl.Error(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("ProductionLogger.Error() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("ProductionLogger.Error() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func TestProductionLogger_Fatal(t *testing.T) {
	savedGetWriterFunc := writerGetter
	defer func() {
		writerGetter = savedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl ProductionLogger
		l  LoggingLevel
		args
		want string
	}{
		"panic": {
			pl:   ProductionLogger{},
			l:    Panic,
			args: args{msg: "fatal", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			pl:   ProductionLogger{},
			l:    Fatal,
			args: args{msg: "fatal", fields: map[string]any{}},
			want: "level=fatal msg=fatal\n",
		},
		"error": {
			pl:   ProductionLogger{},
			l:    Error,
			args: args{msg: "fatal", fields: map[string]any{}},
			want: "level=fatal msg=fatal\n",
		},
		"warn": {
			pl:   ProductionLogger{},
			l:    Warn,
			args: args{msg: "fatal", fields: map[string]any{}},
			want: "level=fatal msg=fatal\n",
		},
		"info": {
			pl:   ProductionLogger{},
			l:    Info,
			args: args{msg: "fatal", fields: nil},
			want: "level=fatal msg=fatal\n",
		},
		"debug": {
			pl: ProductionLogger{},
			l:  Debug,
			args: args{msg: "fatal message", fields: map[string]any{
				"field1": "value1",
				"field2": 2,
				"field3": true,
				"field4": []string{"val1", "val2"},
			}},
			want: "level=fatal msg=\"fatal message\" field1=value1 field2=2 field3=true field4=\"[val1 val2]\"\n",
		},
		"trace": {
			pl:   ProductionLogger{},
			l:    Trace,
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
			savedExit := tt.pl.ExitFunc()
			defer func() {
				tt.pl.SetExitFunc(savedExit)
			}()
			exited := false
			tt.pl.SetExitFunc(func(_ int) {
				exited = true
			})
			tt.pl.Fatal(tt.args.msg, tt.args.fields)
			if !exited {
				t.Errorf("ProductionLogger.Fatal() did not attempt to exit!")
			}
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("ProductionLogger.Fatal() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("ProductionLogger.Fatal() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func TestProductionLogger_Info(t *testing.T) {
	savedGetWriterFunc := writerGetter
	defer func() {
		writerGetter = savedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl ProductionLogger
		l  LoggingLevel
		args
		want string
	}{
		"panic": {
			pl:   ProductionLogger{},
			l:    Panic,
			args: args{msg: "info", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			pl:   ProductionLogger{},
			l:    Fatal,
			args: args{msg: "info", fields: map[string]any{}},
			want: "",
		},
		"error": {
			pl:   ProductionLogger{},
			l:    Error,
			args: args{msg: "info", fields: map[string]any{}},
			want: "",
		},
		"warn": {
			pl:   ProductionLogger{},
			l:    Warn,
			args: args{msg: "info", fields: map[string]any{}},
			want: "",
		},
		"info": {
			pl:   ProductionLogger{},
			l:    Info,
			args: args{msg: "info", fields: nil},
			want: "level=info msg=info\n",
		},
		"debug": {
			pl: ProductionLogger{},
			l:  Debug,
			args: args{msg: "info message", fields: map[string]any{
				"field1": "value1",
				"field2": 2,
				"field3": true,
				"field4": []string{"val1", "val2"},
			}},
			want: "level=info msg=\"info message\" field1=value1 field2=2 field3=true field4=\"[val1 val2]\"\n",
		},
		"trace": {
			pl:   ProductionLogger{},
			l:    Trace,
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
			tt.pl.Info(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("ProductionLogger.Info() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("ProductionLogger.Info() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func TestProductionLogger_Panic(t *testing.T) {
	savedGetWriterFunc := writerGetter
	defer func() {
		writerGetter = savedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl ProductionLogger
		l  LoggingLevel
		args
		want string
	}{
		"panic": {
			pl:   ProductionLogger{},
			l:    Panic,
			args: args{msg: "panic", fields: map[string]any{}},
			want: "level=panic msg=panic\n",
		},
		"fatal": {
			pl:   ProductionLogger{},
			l:    Fatal,
			args: args{msg: "panic", fields: map[string]any{}},
			want: "level=panic msg=panic\n",
		},
		"error": {
			pl:   ProductionLogger{},
			l:    Error,
			args: args{msg: "panic", fields: map[string]any{}},
			want: "level=panic msg=panic\n",
		},
		"warn": {
			pl:   ProductionLogger{},
			l:    Warn,
			args: args{msg: "panic", fields: map[string]any{}},
			want: "level=panic msg=panic\n",
		},
		"info": {
			pl:   ProductionLogger{},
			l:    Info,
			args: args{msg: "panic", fields: nil},
			want: "level=panic msg=panic\n",
		},
		"debug": {
			pl: ProductionLogger{},
			l:  Debug,
			args: args{msg: "panic message", fields: map[string]any{
				"field1": "value1",
				"field2": 2,
				"field3": true,
				"field4": []string{"val1", "val2"},
			}},
			want: "level=panic msg=\"panic message\" field1=value1 field2=2 field3=true field4=\"[val1 val2]\"\n",
		},
		"trace": {
			pl:   ProductionLogger{},
			l:    Trace,
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
					t.Errorf("ProductionLogger.Panic() did not panic")
				}
			}(t)
			tt.pl.Panic(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("ProductionLogger.Panic() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("ProductionLogger.Panic() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func TestProductionLogger_Trace(t *testing.T) {
	savedGetWriterFunc := writerGetter
	defer func() {
		writerGetter = savedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl ProductionLogger
		l  LoggingLevel
		args
		want string
	}{
		"panic": {
			pl:   ProductionLogger{},
			l:    Panic,
			args: args{msg: "trace", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			pl:   ProductionLogger{},
			l:    Fatal,
			args: args{msg: "trace", fields: map[string]any{}},
			want: "",
		},
		"error": {
			pl:   ProductionLogger{},
			l:    Error,
			args: args{msg: "trace", fields: map[string]any{}},
			want: "",
		},
		"warn": {
			pl:   ProductionLogger{},
			l:    Warn,
			args: args{msg: "trace", fields: map[string]any{}},
			want: "",
		},
		"info": {
			pl:   ProductionLogger{},
			l:    Info,
			args: args{msg: "trace", fields: nil},
			want: "",
		},
		"debug": {
			pl: ProductionLogger{},
			l:  Debug,
			args: args{msg: "trace message", fields: map[string]any{
				"field1": "value1",
				"field2": 2,
				"field3": true,
				"field4": []string{"val1", "val2"},
			}},
			want: "",
		},
		"trace": {
			pl:   ProductionLogger{},
			l:    Trace,
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
			tt.pl.Trace(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("ProductionLogger.Trace() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("ProductionLogger.Trace() got %q want %q", got, tt.want)
				}
			}
		})
	}
}

func TestProductionLogger_Warning(t *testing.T) {
	savedGetWriterFunc := writerGetter
	defer func() {
		writerGetter = savedGetWriterFunc
	}()
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := map[string]struct {
		pl ProductionLogger
		l  LoggingLevel
		args
		want string
	}{
		"panic": {
			pl:   ProductionLogger{},
			l:    Panic,
			args: args{msg: "warn", fields: map[string]any{}},
			want: "",
		},
		"fatal": {
			pl:   ProductionLogger{},
			l:    Fatal,
			args: args{msg: "warn", fields: map[string]any{}},
			want: "",
		},
		"error": {
			pl:   ProductionLogger{},
			l:    Error,
			args: args{msg: "warn", fields: map[string]any{}},
			want: "",
		},
		"warn": {
			pl:   ProductionLogger{},
			l:    Warn,
			args: args{msg: "warn", fields: map[string]any{}},
			want: "level=warning msg=warn\n",
		},
		"info": {
			pl:   ProductionLogger{},
			l:    Info,
			args: args{msg: "warn", fields: nil},
			want: "level=warning msg=warn\n",
		},
		"debug": {
			pl: ProductionLogger{},
			l:  Debug,
			args: args{msg: "warn message", fields: map[string]any{
				"field1": "value1",
				"field2": 2,
				"field3": true,
				"field4": []string{"val1", "val2"},
			}},
			want: "level=warning msg=\"warn message\" field1=value1 field2=2 field3=true field4=\"[val1 val2]\"\n",
		},
		"trace": {
			pl:   ProductionLogger{},
			l:    Trace,
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
			tt.pl.Warning(tt.args.msg, tt.args.fields)
			if got := buffer.String(); got != tt.want {
				if got != "" {
					if tt.want == "" || !strings.HasSuffix(got, tt.want) {
						t.Errorf("ProductionLogger.Warn() got %q want %q", got, tt.want)
					}
				} else {
					t.Errorf("ProductionLogger.Warn() got %q want %q", got, tt.want)
				}
			}
		})
	}
}
