package cmd_toolkit_test

import (
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func TestDefaultConfigFileName(t *testing.T) {
	originalDefaultConfigFileName := cmdtoolkit.DefaultConfigFileName()
	defer cmdtoolkit.UnsafeSetDefaultConfigFileName(originalDefaultConfigFileName)
	tests := map[string]struct {
		defaultConfigFileName string
		want                  string
	}{"simple": {defaultConfigFileName: "myDefaults.yaml", want: "myDefaults.yaml"}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmdtoolkit.UnsafeSetDefaultConfigFileName(tt.defaultConfigFileName)
			if got := cmdtoolkit.DefaultConfigFileName(); got != tt.want {
				t.Errorf("DefaultConfigFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

func TestNewIntBounds(t *testing.T) {
	type args struct {
		v1 int
		v2 int
		v3 int
	}
	tests := map[string]struct {
		args
		want *cmdtoolkit.IntBounds
	}{
		"0,1,2": {args: args{v1: 0, v2: 1, v3: 2}, want: &cmdtoolkit.IntBounds{MinValue: 0, DefaultValue: 1, MaxValue: 2}},
		"0,2,1": {args: args{v1: 0, v2: 2, v3: 1}, want: &cmdtoolkit.IntBounds{MinValue: 0, DefaultValue: 1, MaxValue: 2}},
		"1,0,2": {args: args{v1: 1, v2: 0, v3: 2}, want: &cmdtoolkit.IntBounds{MinValue: 0, DefaultValue: 1, MaxValue: 2}},
		"1,2,0": {args: args{v1: 1, v2: 2, v3: 0}, want: &cmdtoolkit.IntBounds{MinValue: 0, DefaultValue: 1, MaxValue: 2}},
		"2,0,1": {args: args{v1: 2, v2: 0, v3: 1}, want: &cmdtoolkit.IntBounds{MinValue: 0, DefaultValue: 1, MaxValue: 2}},
		"2,1,0": {args: args{v1: 2, v2: 1, v3: 0}, want: &cmdtoolkit.IntBounds{MinValue: 0, DefaultValue: 1, MaxValue: 2}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmdtoolkit.NewIntBounds(tt.args.v1, tt.args.v2, tt.args.v3); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIntBounds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadConfigurationFile(t *testing.T) {
	originalFileSystem := cmdtoolkit.FileSystem()
	originalApplicationPath := cmdtoolkit.ApplicationPath()
	originalDefaultConfigFileName := cmdtoolkit.DefaultConfigFileName()
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
		cmdtoolkit.UnsafeSetApplicationPath(originalApplicationPath)
		cmdtoolkit.UnsafeSetDefaultConfigFileName(originalDefaultConfigFileName)
	}()
	cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	tests := map[string]struct {
		preTest func()
		wantC   *cmdtoolkit.Configuration
		wantOk  bool
		output.WantedRecording
	}{
		"config file is a directory": {
			preTest: func() {
				cmdtoolkit.UnsafeSetApplicationPath("configFileDir")
				cmdtoolkit.UnsafeSetDefaultConfigFileName("dir.yaml")
				_ = cmdtoolkit.FileSystem().MkdirAll(filepath.Join(cmdtoolkit.ApplicationPath(), cmdtoolkit.DefaultConfigFileName()), cmdtoolkit.StdDirPermissions)
			},
			wantC: cmdtoolkit.EmptyConfiguration(),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The configuration file \"configFileDir\\\\dir.yaml\" is a directory.\n" +
					"What to do:\n" +
					"Delete the directory \"dir.yaml\" from \"configFileDir\" and restart the application.\n",
				Log: "" +
					"level='error'" +
					" directory='configFileDir'" +
					" fileName='dir.yaml'" +
					" msg='file is a directory'\n",
			},
		},
		"no config file does not exist": {
			preTest: func() {
				cmdtoolkit.UnsafeSetApplicationPath("non-existent directory")
				cmdtoolkit.UnsafeSetDefaultConfigFileName("no such file.yaml")
			},
			wantC: &cmdtoolkit.Configuration{
				BoolMap:          map[string]bool{},
				ConfigurationMap: map[string]*cmdtoolkit.Configuration{},
				IntMap:           map[string]int{},
				StringMap:        map[string]string{},
			},
			wantOk:          true,
			WantedRecording: output.WantedRecording{Log: "level='info' directory='non-existent directory' fileName='no such file.yaml' msg='file does not exist'\n"},
		},
		"config file contains bad data": {
			preTest: func() {
				cmdtoolkit.UnsafeSetApplicationPath("garbageDir")
				cmdtoolkit.UnsafeSetDefaultConfigFileName("trash.yaml")
				_ = cmdtoolkit.FileSystem().Mkdir(cmdtoolkit.ApplicationPath(), cmdtoolkit.StdDirPermissions)
				_ = afero.WriteFile(cmdtoolkit.FileSystem(), filepath.Join(cmdtoolkit.ApplicationPath(), cmdtoolkit.DefaultConfigFileName()), []byte{1, 2, 3}, cmdtoolkit.StdFilePermissions)
			},
			wantC: cmdtoolkit.EmptyConfiguration(),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The configuration file \"garbageDir\\\\trash.yaml\" is not well-formed YAML: yaml: control characters are not allowed.\n" +
					"What to do:\n" +
					"Delete the file \"trash.yaml\" from \"garbageDir\" and restart the application.\n",
				Log: "" +
					"level='error'" +
					" directory='garbageDir'" +
					" error='yaml: control characters are not allowed'" +
					" fileName='trash.yaml'" +
					" msg='cannot unmarshal yaml content'\n",
			},
		},
		"config file contains usable data": {
			preTest: func() {
				cmdtoolkit.UnsafeSetApplicationPath("happyDir")
				cmdtoolkit.UnsafeSetDefaultConfigFileName("good.yaml")
				_ = cmdtoolkit.FileSystem().Mkdir(cmdtoolkit.ApplicationPath(), cmdtoolkit.StdDirPermissions)
				content := "" +
					"b: true\n" +
					"i: 12\n" +
					"s: hello\n" +
					"command:\n" +
					"  default: about\n"
				_ = afero.WriteFile(cmdtoolkit.FileSystem(), filepath.Join(cmdtoolkit.ApplicationPath(), cmdtoolkit.DefaultConfigFileName()), []byte(content), cmdtoolkit.StdFilePermissions)
			},
			wantC: &cmdtoolkit.Configuration{
				BoolMap: map[string]bool{"b": true},
				ConfigurationMap: map[string]*cmdtoolkit.Configuration{
					"command": {
						BoolMap:          map[string]bool{},
						ConfigurationMap: map[string]*cmdtoolkit.Configuration{},
						IntMap:           map[string]int{},
						StringMap:        map[string]string{"default": "about"},
					},
				},
				IntMap:    map[string]int{"i": 12},
				StringMap: map[string]string{"s": "hello"},
			},
			wantOk: true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' directory='happyDir' fileName='good.yaml' value='map[b:true], map[i:12], map[s:hello], map[command:map[default:about]]' msg='read configuration file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			o := output.NewRecorder()
			gotC, gotOk := cmdtoolkit.ReadConfigurationFile(o)
			if !reflect.DeepEqual(gotC, tt.wantC) {
				t.Errorf("ReadConfigurationFile() gotC = %v, want %v", gotC, tt.wantC)
			}
			if gotOk != tt.wantOk {
				t.Errorf("ReadConfigurationFile() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			o.Report(t, "ReadConfigurationFile()", tt.WantedRecording)
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

func TestSetFlagIndicator(t *testing.T) {
	originalIndicator := cmdtoolkit.FlagIndicator()
	defer cmdtoolkit.SetFlagIndicator(originalIndicator)
	tests := map[string]struct {
		val string
	}{
		"-":  {val: "-"},
		"--": {val: "--"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmdtoolkit.SetFlagIndicator(tt.val)
			if got := cmdtoolkit.FlagIndicator(); got != tt.val {
				t.Errorf("SetFlagIndicator got %q want %q", got, tt.val)
			}
		})
	}
}

func TestAssignFileSystem(t *testing.T) {
	originalFileSystem := cmdtoolkit.FileSystem()
	defer cmdtoolkit.AssignFileSystem(originalFileSystem)
	tests := map[string]struct {
		fs   afero.Fs
		want afero.Fs
	}{
		"simple": {fs: afero.NewMemMapFs(), want: cmdtoolkit.FileSystem()},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmdtoolkit.AssignFileSystem(tt.fs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AssignFileSystem() = %v, want %v", got, tt.want)
			}
		})
	}
}
