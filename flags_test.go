package cmd_toolkit

import (
	"flag"
	"os"
	"testing"

	"github.com/majohn-r/output"
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
			if got := DecorateBoolFlagUsage(tt.args.usage, tt.args.defaultValue); got != tt.want {
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
			if got := DecorateIntFlagUsage(tt.args.usage, tt.args.defaultValue); got != tt.want {
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
			if got := DecorateStringFlagUsage(tt.args.usage, tt.args.defaultValue); got != tt.want {
				t.Errorf("DecorateStringFlagUsage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessArgs(t *testing.T) {
	// environment for malformed value case
	const varName = "TESTVAR"
	varValue, varSet := os.LookupEnv(varName)
	defer func() {
		if varSet {
			os.Setenv(varName, varValue)
		} else {
			os.Unsetenv(varName)
		}
	}()
	os.Unsetenv(varName)
	malformedFlags := flag.NewFlagSet("foo", flag.ContinueOnError)
	malformedFlags.String("flag", "defaultValue", "set this for fun")
	// environment for happy value case
	happyFlags := flag.NewFlagSet("foo", flag.ContinueOnError)
	happyFlags.String("flag", "defaultValue", "set this for fun")
	type args struct {
		f       *flag.FlagSet
		rawArgs []string
	}
	tests := map[string]struct {
		args
		wantOk bool
		output.WantedRecording
	}{
		"empty args": {
			args:   args{f: flag.NewFlagSet("foo", flag.ContinueOnError), rawArgs: []string{}},
			wantOk: true,
		},
		"malformed value": {
			args:   args{f: malformedFlags, rawArgs: []string{"flag=$" + varName}},
			wantOk: false,
			WantedRecording: output.WantedRecording{
				Error: "The value for argument \"flag=$TESTVAR\" cannot be used: missing environment variables: [TESTVAR].\n",
				Log:   "level='error' error='missing environment variables: [TESTVAR]' value='flag=$TESTVAR' msg='argument cannot be used'\n",
			},
		},
		"-help used where undefined": {
			args:   args{f: flag.NewFlagSet("foo", flag.ContinueOnError), rawArgs: []string{"-help"}},
			wantOk: false,
			WantedRecording: output.WantedRecording{
				Error: "Usage of foo:\n",
				Log:   "level='error' arguments='[-help]' msg='flag: help requested'\n",
			},
		},
		"happy value": {
			args:   args{f: happyFlags, rawArgs: []string{"flag=hi!"}},
			wantOk: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotOk := ProcessArgs(o, tt.args.f, tt.args.rawArgs); gotOk != tt.wantOk {
				t.Errorf("ProcessArgs() = %v, want %v", gotOk, tt.wantOk)
			}
			if issues, verified := o.Verify(tt.WantedRecording); !verified {
				for _, issue := range issues {
					t.Errorf("ProcessArgs() %s", issue)
				}
			}
		})
	}
}
