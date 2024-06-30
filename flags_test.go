package cmd_toolkit_test

import (
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"testing"
)

func TestDecorateBoolFlagUsage(t *testing.T) {
	type args struct {
		usage        string
		defaultValue bool
	}
	tests := map[string]struct {
		args
		want string
	}{
		"default false": {args: args{usage: "set magic flag"}, want: "set magic flag (default false)"},
		"default true":  {args: args{usage: "set magic flag", defaultValue: true}, want: "set magic flag"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmdtoolkit.DecorateBoolFlagUsage(tt.args.usage, tt.args.defaultValue); got != tt.want {
				t.Errorf("DecorateBoolFlagUsage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecorateIntFlagUsage(t *testing.T) {
	type args struct {
		usage        string
		defaultValue int
	}
	tests := map[string]struct {
		args
		want string
	}{
		"default zero":     {args: args{usage: "set magic flag"}, want: "set magic flag (default 0)"},
		"default non-zero": {args: args{usage: "set magic flag", defaultValue: 24}, want: "set magic flag"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmdtoolkit.DecorateIntFlagUsage(tt.args.usage, tt.args.defaultValue); got != tt.want {
				t.Errorf("DecorateIntFlagUsage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecorateStringFlagUsage(t *testing.T) {
	type args struct {
		usage        string
		defaultValue string
	}
	tests := map[string]struct {
		args
		want string
	}{
		"default empty string":     {args: args{usage: "set magic flag"}, want: "set magic flag (default \"\")"},
		"default non-empty string": {args: args{usage: "set magic flag", defaultValue: "foo"}, want: "set magic flag"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmdtoolkit.DecorateStringFlagUsage(tt.args.usage, tt.args.defaultValue); got != tt.want {
				t.Errorf("DecorateStringFlagUsage() = %v, want %v", got, tt.want)
			}
		})
	}
}
