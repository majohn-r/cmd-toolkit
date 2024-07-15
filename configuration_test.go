package cmd_toolkit_test

import (
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"os"
	"reflect"
	"testing"
)

func TestEmptyConfiguration(t *testing.T) {
	tests := map[string]struct {
		want *cmdtoolkit.Configuration
	}{
		"simple": {
			want: &cmdtoolkit.Configuration{
				BoolMap:          map[string]bool{},
				ConfigurationMap: map[string]*cmdtoolkit.Configuration{},
				IntMap:           map[string]int{},
				StringMap:        map[string]string{},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmdtoolkit.EmptyConfiguration(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EmptyConfiguration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfiguration_String(t *testing.T) {
	tests := map[string]struct {
		c    *cmdtoolkit.Configuration
		want string
	}{
		"empty": {c: cmdtoolkit.EmptyConfiguration()},
		"busy": {
			c: &cmdtoolkit.Configuration{
				BoolMap: map[string]bool{"a": false, "b": true},
				ConfigurationMap: map[string]*cmdtoolkit.Configuration{
					"c": {
						BoolMap:          map[string]bool{"e": false, "f": true},
						ConfigurationMap: map[string]*cmdtoolkit.Configuration{},
						IntMap:           map[string]int{"g": 1, "h": 2},
						StringMap:        map[string]string{"i": "abc", "j": "def"},
					},
				},
				IntMap:    map[string]int{"k": 3, "l": 4},
				StringMap: map[string]string{"m": "ghi", "n": "jkl"},
			},
			want: "map[a:false b:true], map[k:3 l:4], map[m:ghi n:jkl], map[c:map[e:false f:true], map[g:1 h:2], map[i:abc j:def]]",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.c.String(); got != tt.want {
				t.Errorf("Configuration.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfiguration_BoolDefault(t *testing.T) {
	envVar := "TEST_VAR"
	envVarMemento := cmdtoolkit.NewEnvVarMemento(envVar)
	defer envVarMemento.Restore()
	type args struct {
		key          string
		defaultValue bool
	}
	tests := map[string]struct {
		envValue string
		envSet   bool
		c        *cmdtoolkit.Configuration
		args
		wantB   bool
		wantErr bool
	}{
		"no value found": {
			c:     cmdtoolkit.EmptyConfiguration(),
			args:  args{key: "b", defaultValue: true},
			wantB: true,
		},
		"boolean value found": {
			c:     &cmdtoolkit.Configuration{BoolMap: map[string]bool{"b": true}},
			args:  args{key: "b", defaultValue: false},
			wantB: true,
		},
		"int 0 found": {
			c:     &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{"b": 0}},
			args:  args{key: "b", defaultValue: true},
			wantB: false,
		},
		"int 1 found": {
			c:     &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{"b": 1}},
			args:  args{key: "b", defaultValue: false},
			wantB: true,
		},
		"bad int found": {
			c:       &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{"b": 2}},
			args:    args{key: "b", defaultValue: true},
			wantB:   true,
			wantErr: true,
		},
		"string 't' found": {
			c:     &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{}, StringMap: map[string]string{"b": "t"}},
			args:  args{key: "b", defaultValue: false},
			wantB: true,
		},
		"string 'T' found": {
			c:     &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{}, StringMap: map[string]string{"b": "T"}},
			args:  args{key: "b", defaultValue: false},
			wantB: true,
		},
		"string 'true' found": {
			c:     &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{}, StringMap: map[string]string{"b": "true"}},
			args:  args{key: "b", defaultValue: false},
			wantB: true,
		},
		"string 'TRUE' found": {
			c:     &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{}, StringMap: map[string]string{"b": "TRUE"}},
			args:  args{key: "b", defaultValue: false},
			wantB: true,
		},
		"string 'True' found": {
			c:     &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{}, StringMap: map[string]string{"b": "True"}},
			args:  args{key: "b", defaultValue: false},
			wantB: true,
		},
		"string 'f' found": {
			c:     &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{}, StringMap: map[string]string{"b": "f"}},
			args:  args{key: "b", defaultValue: true},
			wantB: false,
		},
		"string 'F' found": {
			c:     &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{}, StringMap: map[string]string{"b": "F"}},
			args:  args{key: "b", defaultValue: true},
			wantB: false,
		},
		"string 'false' found": {
			c:     &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{}, StringMap: map[string]string{"b": "false"}},
			args:  args{key: "b", defaultValue: true},
			wantB: false,
		},
		"string 'FALSE' found": {
			c:     &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{}, StringMap: map[string]string{"b": "FALSE"}},
			args:  args{key: "b", defaultValue: true},
			wantB: false,
		},
		"string 'False' found": {
			c:     &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{}, StringMap: map[string]string{"b": "False"}},
			args:  args{key: "b", defaultValue: true},
			wantB: false,
		},
		"bad string found": {
			c:       &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{}, StringMap: map[string]string{"b": "nope"}},
			args:    args{key: "b", defaultValue: true},
			wantB:   true,
			wantErr: true,
		},
		"use dereferenced value": {
			envValue: "false",
			envSet:   true,
			c:        &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{}, StringMap: map[string]string{"b": "$" + envVar}},
			args:     args{key: "b", defaultValue: true},
			wantB:    false,
		},
		"use bad dereferenced value": {
			c:       &cmdtoolkit.Configuration{BoolMap: map[string]bool{}, IntMap: map[string]int{}, StringMap: map[string]string{"b": "$" + envVar}},
			args:    args{key: "b", defaultValue: true},
			wantB:   true,
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.envSet {
				_ = os.Setenv(envVar, tt.envValue)
			} else {
				_ = os.Unsetenv(envVar)
			}
			gotB, gotErr := tt.c.BoolDefault(tt.args.key, tt.args.defaultValue)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Configuration.BoolDefault() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if gotB != tt.wantB {
				t.Errorf("Configuration.BoolDefault() = %v, want %v", gotB, tt.wantB)
			}
		})
	}
}

func TestConfiguration_IntDefault(t *testing.T) {
	envVar := "TEST_VAR"
	envVarMemento := cmdtoolkit.NewEnvVarMemento(envVar)
	defer envVarMemento.Restore()
	type args struct {
		key string
		b   *cmdtoolkit.IntBounds
	}
	tests := map[string]struct {
		envValue string
		envSet   bool
		c        *cmdtoolkit.Configuration
		args
		wantI   int
		wantErr bool
	}{
		"empty": {
			c:     cmdtoolkit.EmptyConfiguration(),
			args:  args{key: "i", b: cmdtoolkit.NewIntBounds(1, 2, 3)},
			wantI: 2,
		},
		"too low": {
			c:     &cmdtoolkit.Configuration{IntMap: map[string]int{"i": -2}},
			args:  args{key: "i", b: cmdtoolkit.NewIntBounds(1, 2, 3)},
			wantI: 1,
		},
		"too high": {
			c:     &cmdtoolkit.Configuration{IntMap: map[string]int{"i": 47}},
			args:  args{key: "i", b: cmdtoolkit.NewIntBounds(1, 2, 3)},
			wantI: 3,
		},
		"string too low": {
			c:     &cmdtoolkit.Configuration{StringMap: map[string]string{"i": "-100"}},
			args:  args{key: "i", b: cmdtoolkit.NewIntBounds(1, 2, 3)},
			wantI: 1,
		},
		"string too high": {
			c:     &cmdtoolkit.Configuration{StringMap: map[string]string{"i": "100"}},
			args:  args{key: "i", b: cmdtoolkit.NewIntBounds(1, 2, 3)},
			wantI: 3,
		},
		"dereferenced string": {
			envValue: "7",
			envSet:   true,
			c:        &cmdtoolkit.Configuration{StringMap: map[string]string{"i": "%" + envVar + "%"}},
			args:     args{key: "i", b: cmdtoolkit.NewIntBounds(-1, 2, 300)},
			wantI:    7,
		},
		"bad dereferenced string": {
			c:       &cmdtoolkit.Configuration{StringMap: map[string]string{"i": "%" + envVar + "%"}},
			args:    args{key: "i", b: cmdtoolkit.NewIntBounds(-1, 2, 300)},
			wantI:   2,
			wantErr: true,
		},
		"bad string": {
			c:       &cmdtoolkit.Configuration{StringMap: map[string]string{"i": "nine"}},
			args:    args{key: "i", b: cmdtoolkit.NewIntBounds(-1, 20, 300)},
			wantI:   20,
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.envSet {
				_ = os.Setenv(envVar, tt.envValue)
			} else {
				_ = os.Unsetenv(envVar)
			}
			gotI, gotErr := tt.c.IntDefault(tt.args.key, tt.args.b)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Configuration.IntDefault() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if gotI != tt.wantI {
				t.Errorf("Configuration.IntDefault() = %v, want %v", gotI, tt.wantI)
			}
		})
	}
}

func TestConfiguration_StringDefault(t *testing.T) {
	envVar1 := "TEST_VAR1"
	envVar1Memento := cmdtoolkit.NewEnvVarMemento(envVar1)
	envVar2 := "TEST_VAR2"
	envVar2Memento := cmdtoolkit.NewEnvVarMemento(envVar2)
	defer func() {
		envVar1Memento.Restore()
		envVar2Memento.Restore()
	}()
	type args struct {
		key          string
		defaultValue string
	}
	tests := map[string]struct {
		envValue1 string
		envSet1   bool
		envValue2 string
		envSet2   bool
		c         *cmdtoolkit.Configuration
		args
		wantS   string
		wantErr bool
	}{
		"simple default, no configuration": {
			c:     cmdtoolkit.EmptyConfiguration(),
			args:  args{key: "s", defaultValue: "defaultValue"},
			wantS: "defaultValue",
		},
		"simple config override": {
			c:     &cmdtoolkit.Configuration{StringMap: map[string]string{"s": "override"}},
			args:  args{key: "s", defaultValue: "defaultValue"},
			wantS: "override",
		},
		"dereferenced default, no configuration": {
			envValue1: "user",
			envSet1:   true,
			c:         cmdtoolkit.EmptyConfiguration(),
			args:      args{key: "s", defaultValue: "hello $" + envVar1},
			wantS:     "hello user",
		},
		"dereferenced default, dereferenced  configuration": {
			envValue1: "user",
			envSet1:   true,
			envValue2: "other user",
			envSet2:   true,
			c:         &cmdtoolkit.Configuration{StringMap: map[string]string{"s": "goodbye %" + envVar2 + "%"}},
			args:      args{key: "s", defaultValue: "hello $" + envVar1},
			wantS:     "goodbye other user",
		},
		"bad reference in default": {
			c:       cmdtoolkit.EmptyConfiguration(),
			args:    args{key: "s", defaultValue: "hello $" + envVar1},
			wantErr: true,
		},
		"bad reference in configuration": {
			c:       &cmdtoolkit.Configuration{StringMap: map[string]string{"s": "goodbye %" + envVar2 + "%"}},
			args:    args{key: "s", defaultValue: "hello"},
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.envSet1 {
				_ = os.Setenv(envVar1, tt.envValue1)
			} else {
				_ = os.Unsetenv(envVar1)
			}
			if tt.envSet2 {
				_ = os.Setenv(envVar2, tt.envValue2)
			} else {
				_ = os.Unsetenv(envVar2)
			}
			gotS, gotErr := tt.c.StringDefault(tt.args.key, tt.args.defaultValue)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Configuration.StringDefault() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if gotS != tt.wantS {
				t.Errorf("Configuration.StringDefault() = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}

func TestConfiguration_SubConfiguration(t *testing.T) {
	tests := map[string]struct {
		c    *cmdtoolkit.Configuration
		key  string
		want *cmdtoolkit.Configuration
	}{
		"no match": {
			c:    cmdtoolkit.EmptyConfiguration(),
			key:  "c",
			want: cmdtoolkit.EmptyConfiguration(),
		},
		"match": {
			c: &cmdtoolkit.Configuration{
				ConfigurationMap: map[string]*cmdtoolkit.Configuration{
					"c": {
						BoolMap:   map[string]bool{"b": true},
						IntMap:    map[string]int{"i": 45000},
						StringMap: map[string]string{"s": "hey!"},
					},
				},
			},
			key: "c",
			want: &cmdtoolkit.Configuration{
				BoolMap:   map[string]bool{"b": true},
				IntMap:    map[string]int{"i": 45000},
				StringMap: map[string]string{"s": "hey!"},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.c.SubConfiguration(tt.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Configuration.SubConfiguration() = %v, want %v", got, tt.want)
			}
		})
	}
}
