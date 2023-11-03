package cmd_toolkit

import (
	"bytes"
	"io"
	"testing"

	"github.com/majohn-r/output"
	"github.com/sirupsen/logrus"
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
