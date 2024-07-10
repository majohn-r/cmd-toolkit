package cmd_toolkit

import (
	"errors"
	"github.com/majohn-r/output"
	"github.com/spf13/pflag"
	"testing"
)

type testConfigSource struct {
	generateError bool
}

func (tcs testConfigSource) BoolDefault(_ string, defaultValue bool) (bool, error) {
	if tcs.generateError {
		return false, errors.New("boolean error")
	}
	return defaultValue, nil
}

func (tcs testConfigSource) IntDefault(_ string, defaultValue *IntBounds) (int, error) {
	if tcs.generateError {
		return 0, errors.New("int error")
	}
	return defaultValue.DefaultValue, nil
}

func (tcs testConfigSource) StringDefault(_, defaultValue string) (string, error) {
	if tcs.generateError {
		return "", errors.New("string error")
	}
	return defaultValue, nil
}

func TestFlagDetails_addFlag(t *testing.T) {
	type args struct {
		c        configSource
		consumer *pflag.FlagSet
		flag     flagParam
	}
	tests := map[string]struct {
		fD *FlagDetails
		args
		output.WantedRecording
	}{
		"error case": {
			fD: &FlagDetails{
				AbbreviatedName: "",
				Usage:           "",
				ExpectedType:    unspecifiedType,
				DefaultValue:    45,
			},
			args: args{
				c:        nil,
				consumer: nil,
				flag:     flagParam{set: "mySet", name: "myFlag"},
			},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: unspecified flag type; set \"mySet\", flag \"myFlag\".\n",
				Log: "" +
					"level='error'" +
					" default-type='int'" +
					" default='45'" +
					" error='unspecified flag type'" +
					" flag='myFlag'" +
					" set='mySet'" +
					" specified-type='0'" +
					" msg='internal error'\n",
			},
		},
		"bad bool case: badly defined default": {
			fD: &FlagDetails{
				AbbreviatedName: "",
				Usage:           "",
				ExpectedType:    BoolType,
				DefaultValue:    12,
			},
			args: args{
				c:        nil,
				consumer: &pflag.FlagSet{},
				flag:     flagParam{set: "mySet", name: "myFlag"},
			},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: the type of flag \"myFlag\"'s value, '12', is 'int', but 'bool' was expected.\n",
				Log: "" +
					"level='error'" +
					" actual='int'" +
					" error='default value mistyped'" +
					" expected='bool'" +
					" flag='myFlag'" +
					" value='12'" +
					" msg='internal error'\n",
			},
		},
		"bad bool case: badly configured default": {
			fD: &FlagDetails{
				AbbreviatedName: "",
				Usage:           "",
				ExpectedType:    BoolType,
				DefaultValue:    true,
			},
			args: args{
				c:        testConfigSource{generateError: true},
				consumer: &pflag.FlagSet{},
				flag:     flagParam{set: "mySet", name: "myFlag"},
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"mySet\": boolean error.\n",
				Log: "" +
					"level='error'" +
					" error='boolean error'" +
					" section='mySet'" +
					" msg='invalid content in configuration file'\n",
			},
		},
		"good bool case": {
			fD: &FlagDetails{
				AbbreviatedName: "",
				Usage:           "",
				ExpectedType:    BoolType,
				DefaultValue:    true,
			},
			args: args{
				c:        testConfigSource{generateError: false},
				consumer: &pflag.FlagSet{},
				flag:     flagParam{set: "mySet", name: "myFlag"},
			},
			WantedRecording: output.WantedRecording{},
		},
		"good bool case: abbreviated": {
			fD: &FlagDetails{
				AbbreviatedName: "m",
				Usage:           "",
				ExpectedType:    BoolType,
				DefaultValue:    true,
			},
			args: args{
				c:        testConfigSource{generateError: false},
				consumer: &pflag.FlagSet{},
				flag:     flagParam{set: "mySet", name: "myFlag"},
			},
			WantedRecording: output.WantedRecording{},
		},
		"bad int case: badly defined default": {
			fD: &FlagDetails{
				AbbreviatedName: "",
				Usage:           "",
				ExpectedType:    IntType,
				DefaultValue:    false,
			},
			args: args{
				c:        nil,
				consumer: &pflag.FlagSet{},
				flag:     flagParam{set: "mySet", name: "myFlag"},
			},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: the type of flag \"myFlag\"'s value, 'false', is 'bool', but '*cmd_toolkit.IntBounds' was expected.\n",
				Log: "" +
					"level='error'" +
					" actual='bool'" +
					" error='default value mistyped'" +
					" expected='*cmd_toolkit.IntBounds'" +
					" flag='myFlag'" +
					" value='false'" +
					" msg='internal error'\n",
			},
		},
		"bad int case: badly configured default": {
			fD: &FlagDetails{
				AbbreviatedName: "",
				Usage:           "",
				ExpectedType:    IntType,
				DefaultValue:    &IntBounds{0, 1, 2},
			},
			args: args{
				c:        testConfigSource{generateError: true},
				consumer: &pflag.FlagSet{},
				flag:     flagParam{set: "mySet", name: "myFlag"},
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"mySet\": int error.\n",
				Log: "" +
					"level='error'" +
					" error='int error'" +
					" section='mySet'" +
					" msg='invalid content in configuration file'\n",
			},
		},
		"good int case": {
			fD: &FlagDetails{
				AbbreviatedName: "",
				Usage:           "",
				ExpectedType:    IntType,
				DefaultValue:    &IntBounds{0, 1, 2},
			},
			args: args{
				c:        testConfigSource{generateError: false},
				consumer: &pflag.FlagSet{},
				flag:     flagParam{set: "mySet", name: "myFlag"},
			},
			WantedRecording: output.WantedRecording{},
		},
		"good int case: abbreviated": {
			fD: &FlagDetails{
				AbbreviatedName: "m",
				Usage:           "",
				ExpectedType:    IntType,
				DefaultValue:    &IntBounds{0, 1, 2},
			},
			args: args{
				c:        testConfigSource{generateError: false},
				consumer: &pflag.FlagSet{},
				flag:     flagParam{set: "mySet", name: "myFlag"},
			},
			WantedRecording: output.WantedRecording{},
		},
		"bad string case: badly defined default": {
			fD: &FlagDetails{
				AbbreviatedName: "",
				Usage:           "",
				ExpectedType:    StringType,
				DefaultValue:    12,
			},
			args: args{
				c:        nil,
				consumer: &pflag.FlagSet{},
				flag:     flagParam{set: "mySet", name: "myFlag"},
			},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: the type of flag \"myFlag\"'s value, '12', is 'int', but 'string' was expected.\n",
				Log: "" +
					"level='error'" +
					" actual='int'" +
					" error='default value mistyped'" +
					" expected='string'" +
					" flag='myFlag'" +
					" value='12'" +
					" msg='internal error'\n",
			},
		},
		"bad string case: badly configured default": {
			fD: &FlagDetails{
				AbbreviatedName: "",
				Usage:           "",
				ExpectedType:    StringType,
				DefaultValue:    "boo",
			},
			args: args{
				c:        testConfigSource{generateError: true},
				consumer: &pflag.FlagSet{},
				flag:     flagParam{set: "mySet", name: "myFlag"},
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"mySet\": string error.\n",
				Log: "" +
					"level='error'" +
					" error='string error'" +
					" section='mySet'" +
					" msg='invalid content in configuration file'\n",
			},
		},
		"good string case": {
			fD: &FlagDetails{
				AbbreviatedName: "",
				Usage:           "",
				ExpectedType:    StringType,
				DefaultValue:    "boo",
			},
			args: args{
				c:        testConfigSource{generateError: false},
				consumer: &pflag.FlagSet{},
				flag:     flagParam{set: "mySet", name: "myFlag"},
			},
			WantedRecording: output.WantedRecording{},
		},
		"good string case: abbreviated": {
			fD: &FlagDetails{
				AbbreviatedName: "m",
				Usage:           "",
				ExpectedType:    StringType,
				DefaultValue:    "boo",
			},
			args: args{
				c:        testConfigSource{generateError: false},
				consumer: &pflag.FlagSet{},
				flag:     flagParam{set: "mySet", name: "myFlag"},
			},
			WantedRecording: output.WantedRecording{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.fD.addFlag(o, tt.args.c, tt.args.consumer, tt.args.flag)
			o.Report(t, "addFlag()", tt.WantedRecording)
		})
	}
}

func Test_decorateBoolFlagUsage(t *testing.T) {
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
			if got := decorateBoolFlagUsage(tt.args.usage, tt.args.defaultValue); got != tt.want {
				t.Errorf("decorateBoolFlagUsage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_decorateIntFlagUsage(t *testing.T) {
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
			if got := decorateIntFlagUsage(tt.args.usage, tt.args.defaultValue); got != tt.want {
				t.Errorf("decorateIntFlagUsage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_decorateStringFlagUsage(t *testing.T) {
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
			if got := decorateStringFlagUsage(tt.args.usage, tt.args.defaultValue); got != tt.want {
				t.Errorf("decorateStringFlagUsage() = %v, want %v", got, tt.want)
			}
		})
	}
}
