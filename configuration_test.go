package cmd_toolkit

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func TestDefaultConfigFileName(t *testing.T) {
	savedDefaultConfigFileName := defaultConfigFileName
	defer func() {
		defaultConfigFileName = savedDefaultConfigFileName
	}()
	tests := map[string]struct {
		defaultConfigFileName string
		want                  string
	}{"simple": {defaultConfigFileName: "myDefaults.yaml", want: "myDefaults.yaml"}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			defaultConfigFileName = tt.defaultConfigFileName
			if got := DefaultConfigFileName(); got != tt.want {
				t.Errorf("DefaultConfigFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmptyConfiguration(t *testing.T) {
	tests := map[string]struct {
		want *Configuration
	}{
		"simple": {
			want: &Configuration{
				bMap: map[string]bool{},
				cMap: map[string]*Configuration{},
				iMap: map[string]int{},
				sMap: map[string]string{},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := EmptyConfiguration(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EmptyConfiguration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewConfiguration(t *testing.T) {
	tests := map[string]struct {
		data map[string]any
		want *Configuration
		output.WantedRecording
	}{
		"unrecognized type": {
			data: map[string]any{
				"boolean":     true,
				"integer":     12,
				"string":      "hello",
				"problematic": 1.234,
			},
			want: &Configuration{
				bMap: map[string]bool{"boolean": true},
				cMap: map[string]*Configuration{},
				iMap: map[string]int{"integer": 12},
				sMap: map[string]string{"string": "hello", "problematic": "1.234"},
			},
			WantedRecording: output.WantedRecording{
				Error: "The key \"problematic\", with value '1.234', has an unexpected type float64.\n",
				Log:   "level='error' key='problematic' type='float64' value='1.234' msg='unexpected value type'\n",
			},
		},
		"no unrecognized types": {
			data: map[string]any{
				"boolean": true,
				"integer": 12,
				"string":  "hello",
				"complex": map[string]any{
					"another boolean": false,
					"another integer": 13,
					"another string":  "hi!",
				},
			},
			want: &Configuration{
				bMap: map[string]bool{"boolean": true},
				cMap: map[string]*Configuration{
					"complex": {
						bMap: map[string]bool{"another boolean": false},
						cMap: map[string]*Configuration{},
						iMap: map[string]int{"another integer": 13},
						sMap: map[string]string{"another string": "hi!"},
					},
				},
				iMap: map[string]int{"integer": 12},
				sMap: map[string]string{"string": "hello"},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := newConfiguration(o, tt.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newConfiguration() = %v, want %v", got, tt.want)
			}
			o.Report(t, "newConfiguration()", tt.WantedRecording)
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
		want *IntBounds
	}{
		"0,1,2": {args: args{v1: 0, v2: 1, v3: 2}, want: &IntBounds{minValue: 0, defaultValue: 1, maxValue: 2}},
		"0,2,1": {args: args{v1: 0, v2: 2, v3: 1}, want: &IntBounds{minValue: 0, defaultValue: 1, maxValue: 2}},
		"1,0,2": {args: args{v1: 1, v2: 0, v3: 2}, want: &IntBounds{minValue: 0, defaultValue: 1, maxValue: 2}},
		"1,2,0": {args: args{v1: 1, v2: 2, v3: 0}, want: &IntBounds{minValue: 0, defaultValue: 1, maxValue: 2}},
		"2,0,1": {args: args{v1: 2, v2: 0, v3: 1}, want: &IntBounds{minValue: 0, defaultValue: 1, maxValue: 2}},
		"2,1,0": {args: args{v1: 2, v2: 1, v3: 0}, want: &IntBounds{minValue: 0, defaultValue: 1, maxValue: 2}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := NewIntBounds(tt.args.v1, tt.args.v2, tt.args.v3); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIntBounds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadConfigurationFile(t *testing.T) {
	originalFileSystem := fileSystem
	originalApplicationPath := applicationPath
	originalDefaultConfigFileName := defaultConfigFileName
	defer func() {
		fileSystem = originalFileSystem
		applicationPath = originalApplicationPath
		defaultConfigFileName = originalDefaultConfigFileName
	}()
	fileSystem = afero.NewMemMapFs()
	tests := map[string]struct {
		preTest func()
		wantC   *Configuration
		wantOk  bool
		output.WantedRecording
	}{
		"config file is a directory": {
			preTest: func() {
				applicationPath = "configFileDir"
				defaultConfigFileName = "dir.yaml"
				_ = fileSystem.MkdirAll(filepath.Join(applicationPath, defaultConfigFileName), StdDirPermissions)
			},
			wantC: EmptyConfiguration(),
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"configFileDir\\\\dir.yaml\" is a directory.\n",
				Log:   "level='error' directory='configFileDir' fileName='dir.yaml' msg='file is a directory'\n",
			},
		},
		"no config file does not exist": {
			preTest: func() {
				applicationPath = "non-existent directory"
				defaultConfigFileName = "no such file.yaml"
			},
			wantC: &Configuration{
				bMap: map[string]bool{},
				cMap: map[string]*Configuration{},
				iMap: map[string]int{},
				sMap: map[string]string{},
			},
			wantOk:          true,
			WantedRecording: output.WantedRecording{Log: "level='info' directory='non-existent directory' fileName='no such file.yaml' msg='file does not exist'\n"},
		},
		"config file contains bad data": {
			preTest: func() {
				applicationPath = "garbageDir"
				defaultConfigFileName = "trash.yaml"
				_ = fileSystem.Mkdir(applicationPath, StdDirPermissions)
				_ = afero.WriteFile(fileSystem, filepath.Join(applicationPath, defaultConfigFileName), []byte{1, 2, 3}, StdFilePermissions)
			},
			wantC: EmptyConfiguration(),
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"garbageDir\\\\trash.yaml\" is not well-formed YAML: yaml: control characters are not allowed.\n",
				Log:   "level='error' directory='garbageDir' error='yaml: control characters are not allowed' fileName='trash.yaml' msg='cannot unmarshal yaml content'\n",
			},
		},
		"config file contains usable data": {
			preTest: func() {
				applicationPath = "happyDir"
				defaultConfigFileName = "good.yaml"
				_ = fileSystem.Mkdir(applicationPath, StdDirPermissions)
				content := "" +
					"b: true\n" +
					"i: 12\n" +
					"s: hello\n" +
					"command:\n" +
					"  default: about\n"
				_ = afero.WriteFile(fileSystem, filepath.Join(applicationPath, defaultConfigFileName), []byte(content), StdFilePermissions)
			},
			wantC: &Configuration{
				bMap: map[string]bool{"b": true},
				cMap: map[string]*Configuration{
					"command": {
						bMap: map[string]bool{},
						cMap: map[string]*Configuration{},
						iMap: map[string]int{},
						sMap: map[string]string{"default": "about"},
					},
				},
				iMap: map[string]int{"i": 12},
				sMap: map[string]string{"s": "hello"},
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
			gotC, gotOk := ReadConfigurationFile(o)
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

func TestReportInvalidConfigurationData(t *testing.T) {
	savedDefaultConfigFileName := defaultConfigFileName
	defer func() {
		defaultConfigFileName = savedDefaultConfigFileName
	}()
	type args struct {
		s string
		e error
	}
	tests := map[string]struct {
		defaultConfigFileName string
		args
		output.WantedRecording
	}{
		"simple": {
			defaultConfigFileName: "defaultConfig.yaml",
			args:                  args{s: "defaults", e: fmt.Errorf("illegal value")},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaultConfig.yaml\" contains an invalid value for \"defaults\": illegal value.\n",
				Log:   "level='error' error='illegal value' section='defaults' msg='invalid content in configuration file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			defaultConfigFileName = tt.defaultConfigFileName
			o := output.NewRecorder()
			ReportInvalidConfigurationData(o, tt.args.s, tt.args.e)
			o.Report(t, "ReportInvalidConfigurationData()", tt.WantedRecording)
		})
	}
}

func TestSetDefaultConfigFileName(t *testing.T) {
	savedDefaultConfigFileName := defaultConfigFileName
	defer func() {
		defaultConfigFileName = savedDefaultConfigFileName
	}()
	tests := map[string]struct {
		s    string
		want string
	}{"simple": {s: "defaultConfigFileName.yaml", want: "defaultConfigFileName.yaml"}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			SetDefaultConfigFileName(tt.s)
			if got := DefaultConfigFileName(); got != tt.want {
				t.Errorf("SetDefaultConfigFileName() %q want %q", got, tt.want)
			}
		})
	}
}

func Test_verifyDefaultConfigFileExists(t *testing.T) {
	originalFileSystem := fileSystem
	defer func() {
		fileSystem = originalFileSystem
	}()
	fileSystem = afero.NewMemMapFs()
	tests := map[string]struct {
		preTest func()
		path    string
		wantOk  bool
		wantErr bool
		output.WantedRecording
	}{
		"path is a directory": {
			preTest: func() {
				_ = fileSystem.Mkdir("testPath", StdDirPermissions)
			},
			path:    "testPath",
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"testPath\" is a directory.\n",
				Log:   "level='error' directory='.' fileName='testPath' msg='file is a directory'\n",
			},
		},
		"path does not exist": {
			preTest:         func() {},
			path:            filepath.Join(".", "non-existent-file.yaml"),
			WantedRecording: output.WantedRecording{Log: "level='info' directory='.' fileName='non-existent-file.yaml' msg='file does not exist'\n"},
		},
		"path is a valid file": {
			preTest: func() {
				path := "testPath2"
				_ = fileSystem.Mkdir(path, StdDirPermissions)
				_ = afero.WriteFile(fileSystem, filepath.Join(path, "happy.yaml"), []byte("boo"), StdFilePermissions)
			},
			path:   filepath.Join("testPath2", "happy.yaml"),
			wantOk: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			o := output.NewRecorder()
			gotOk, gotErr := verifyDefaultConfigFileExists(o, tt.path)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("verifyDefaultConfigFileExists() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if gotOk != tt.wantOk {
				t.Errorf("verifyDefaultConfigFileExists() = %v, want %v", gotOk, tt.wantOk)
			}
			o.Report(t, "verifyDefaultConfigFileExists()", tt.WantedRecording)
		})
	}
}

func TestConfiguration_String(t *testing.T) {
	tests := map[string]struct {
		c    *Configuration
		want string
	}{
		"empty": {c: EmptyConfiguration()},
		"busy": {
			c: &Configuration{
				bMap: map[string]bool{"a": false, "b": true},
				cMap: map[string]*Configuration{
					"c": {
						bMap: map[string]bool{"e": false, "f": true},
						cMap: map[string]*Configuration{},
						iMap: map[string]int{"g": 1, "h": 2},
						sMap: map[string]string{"i": "abc", "j": "def"},
					},
				},
				iMap: map[string]int{"k": 3, "l": 4},
				sMap: map[string]string{"m": "ghi", "n": "jkl"},
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
	savedValue, savedStatus := os.LookupEnv(envVar)
	defer func() {
		if savedStatus {
			_ = os.Setenv(envVar, savedValue)
		} else {
			_ = os.Unsetenv(envVar)
		}
	}()
	type args struct {
		key          string
		defaultValue bool
	}
	tests := map[string]struct {
		envValue string
		envSet   bool
		c        *Configuration
		args
		wantB   bool
		wantErr bool
	}{
		"no value found": {
			c:     EmptyConfiguration(),
			args:  args{key: "b", defaultValue: true},
			wantB: true,
		},
		"boolean value found": {
			c:     &Configuration{bMap: map[string]bool{"b": true}},
			args:  args{key: "b", defaultValue: false},
			wantB: true,
		},
		"int 0 found": {
			c:     &Configuration{bMap: map[string]bool{}, iMap: map[string]int{"b": 0}},
			args:  args{key: "b", defaultValue: true},
			wantB: false,
		},
		"int 1 found": {
			c:     &Configuration{bMap: map[string]bool{}, iMap: map[string]int{"b": 1}},
			args:  args{key: "b", defaultValue: false},
			wantB: true,
		},
		"bad int found": {
			c:       &Configuration{bMap: map[string]bool{}, iMap: map[string]int{"b": 2}},
			args:    args{key: "b", defaultValue: true},
			wantB:   true,
			wantErr: true,
		},
		"string 't' found": {
			c:     &Configuration{bMap: map[string]bool{}, iMap: map[string]int{}, sMap: map[string]string{"b": "t"}},
			args:  args{key: "b", defaultValue: false},
			wantB: true,
		},
		"string 'T' found": {
			c:     &Configuration{bMap: map[string]bool{}, iMap: map[string]int{}, sMap: map[string]string{"b": "T"}},
			args:  args{key: "b", defaultValue: false},
			wantB: true,
		},
		"string 'true' found": {
			c:     &Configuration{bMap: map[string]bool{}, iMap: map[string]int{}, sMap: map[string]string{"b": "true"}},
			args:  args{key: "b", defaultValue: false},
			wantB: true,
		},
		"string 'TRUE' found": {
			c:     &Configuration{bMap: map[string]bool{}, iMap: map[string]int{}, sMap: map[string]string{"b": "TRUE"}},
			args:  args{key: "b", defaultValue: false},
			wantB: true,
		},
		"string 'True' found": {
			c:     &Configuration{bMap: map[string]bool{}, iMap: map[string]int{}, sMap: map[string]string{"b": "True"}},
			args:  args{key: "b", defaultValue: false},
			wantB: true,
		},
		"string 'f' found": {
			c:     &Configuration{bMap: map[string]bool{}, iMap: map[string]int{}, sMap: map[string]string{"b": "f"}},
			args:  args{key: "b", defaultValue: true},
			wantB: false,
		},
		"string 'F' found": {
			c:     &Configuration{bMap: map[string]bool{}, iMap: map[string]int{}, sMap: map[string]string{"b": "F"}},
			args:  args{key: "b", defaultValue: true},
			wantB: false,
		},
		"string 'false' found": {
			c:     &Configuration{bMap: map[string]bool{}, iMap: map[string]int{}, sMap: map[string]string{"b": "false"}},
			args:  args{key: "b", defaultValue: true},
			wantB: false,
		},
		"string 'FALSE' found": {
			c:     &Configuration{bMap: map[string]bool{}, iMap: map[string]int{}, sMap: map[string]string{"b": "FALSE"}},
			args:  args{key: "b", defaultValue: true},
			wantB: false,
		},
		"string 'False' found": {
			c:     &Configuration{bMap: map[string]bool{}, iMap: map[string]int{}, sMap: map[string]string{"b": "False"}},
			args:  args{key: "b", defaultValue: true},
			wantB: false,
		},
		"bad string found": {
			c:       &Configuration{bMap: map[string]bool{}, iMap: map[string]int{}, sMap: map[string]string{"b": "nope"}},
			args:    args{key: "b", defaultValue: true},
			wantB:   true,
			wantErr: true,
		},
		"use dereferenced value": {
			envValue: "false",
			envSet:   true,
			c:        &Configuration{bMap: map[string]bool{}, iMap: map[string]int{}, sMap: map[string]string{"b": "$" + envVar}},
			args:     args{key: "b", defaultValue: true},
			wantB:    false,
		},
		"use bad dereferenced value": {
			c:       &Configuration{bMap: map[string]bool{}, iMap: map[string]int{}, sMap: map[string]string{"b": "$" + envVar}},
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
	savedValue, savedStatus := os.LookupEnv(envVar)
	defer func() {
		if savedStatus {
			_ = os.Setenv(envVar, savedValue)
		} else {
			_ = os.Unsetenv(envVar)
		}
	}()
	type args struct {
		key string
		b   *IntBounds
	}
	tests := map[string]struct {
		envValue string
		envSet   bool
		c        *Configuration
		args
		wantI   int
		wantErr bool
	}{
		"empty": {
			c:     EmptyConfiguration(),
			args:  args{key: "i", b: NewIntBounds(1, 2, 3)},
			wantI: 2,
		},
		"too low": {
			c:     &Configuration{iMap: map[string]int{"i": -2}},
			args:  args{key: "i", b: NewIntBounds(1, 2, 3)},
			wantI: 1,
		},
		"too high": {
			c:     &Configuration{iMap: map[string]int{"i": 47}},
			args:  args{key: "i", b: NewIntBounds(1, 2, 3)},
			wantI: 3,
		},
		"string too low": {
			c:     &Configuration{sMap: map[string]string{"i": "-100"}},
			args:  args{key: "i", b: NewIntBounds(1, 2, 3)},
			wantI: 1,
		},
		"string too high": {
			c:     &Configuration{sMap: map[string]string{"i": "100"}},
			args:  args{key: "i", b: NewIntBounds(1, 2, 3)},
			wantI: 3,
		},
		"dereferenced string": {
			envValue: "7",
			envSet:   true,
			c:        &Configuration{sMap: map[string]string{"i": "%" + envVar + "%"}},
			args:     args{key: "i", b: NewIntBounds(-1, 2, 300)},
			wantI:    7,
		},
		"bad dereferenced string": {
			c:       &Configuration{sMap: map[string]string{"i": "%" + envVar + "%"}},
			args:    args{key: "i", b: NewIntBounds(-1, 2, 300)},
			wantI:   2,
			wantErr: true,
		},
		"bad string": {
			c:       &Configuration{sMap: map[string]string{"i": "nine"}},
			args:    args{key: "i", b: NewIntBounds(-1, 20, 300)},
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
	savedValue1, savedStatus1 := os.LookupEnv(envVar1)
	envVar2 := "TEST_VAR2"
	savedValue2, savedStatus2 := os.LookupEnv(envVar2)
	defer func() {
		if savedStatus1 {
			_ = os.Setenv(envVar1, savedValue1)
		} else {
			_ = os.Unsetenv(envVar1)
		}
		if savedStatus2 {
			_ = os.Setenv(envVar2, savedValue2)
		} else {
			_ = os.Unsetenv(envVar2)
		}
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
		c         *Configuration
		args
		wantS   string
		wantErr bool
	}{
		"simple default, no configuration": {
			c:     EmptyConfiguration(),
			args:  args{key: "s", defaultValue: "defaultValue"},
			wantS: "defaultValue",
		},
		"simple config override": {
			c:     &Configuration{sMap: map[string]string{"s": "override"}},
			args:  args{key: "s", defaultValue: "defaultValue"},
			wantS: "override",
		},
		"dereferenced default, no configuration": {
			envValue1: "user",
			envSet1:   true,
			c:         EmptyConfiguration(),
			args:      args{key: "s", defaultValue: "hello $" + envVar1},
			wantS:     "hello user",
		},
		"dereferenced default, dereferenced  configuration": {
			envValue1: "user",
			envSet1:   true,
			envValue2: "other user",
			envSet2:   true,
			c:         &Configuration{sMap: map[string]string{"s": "goodbye %" + envVar2 + "%"}},
			args:      args{key: "s", defaultValue: "hello $" + envVar1},
			wantS:     "goodbye other user",
		},
		"bad reference in default": {
			c:       EmptyConfiguration(),
			args:    args{key: "s", defaultValue: "hello $" + envVar1},
			wantErr: true,
		},
		"bad reference in configuration": {
			c:       &Configuration{sMap: map[string]string{"s": "goodbye %" + envVar2 + "%"}},
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

func TestConfiguration_StringValue(t *testing.T) {
	tests := map[string]struct {
		c         *Configuration
		key       string
		wantValue string
		wantOk    bool
	}{
		"missing": {
			c:   EmptyConfiguration(),
			key: "s",
		},
		"found": {
			c: &Configuration{
				sMap: map[string]string{"s": "hello"},
			},
			key:       "s",
			wantValue: "hello",
			wantOk:    true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotValue, gotOk := tt.c.StringValue(tt.key)
			if gotValue != tt.wantValue {
				t.Errorf("Configuration.StringValue() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Configuration.StringValue() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestConfiguration_SubConfiguration(t *testing.T) {
	tests := map[string]struct {
		c    *Configuration
		key  string
		want *Configuration
	}{
		"no match": {
			c:    EmptyConfiguration(),
			key:  "c",
			want: EmptyConfiguration(),
		},
		"match": {
			c: &Configuration{
				cMap: map[string]*Configuration{
					"c": {
						bMap: map[string]bool{"b": true},
						iMap: map[string]int{"i": 45000},
						sMap: map[string]string{"s": "hey!"},
					},
				},
			},
			key: "c",
			want: &Configuration{
				bMap: map[string]bool{"b": true},
				iMap: map[string]int{"i": 45000},
				sMap: map[string]string{"s": "hey!"},
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

func TestIntBounds_constrainedValue(t *testing.T) {
	tests := map[string]struct {
		b     *IntBounds
		value int
		wantI int
	}{
		"low": {
			b:     NewIntBounds(1, 10, 100),
			value: -500,
			wantI: 1,
		},
		"high": {
			b:     NewIntBounds(1, 10, 100),
			value: 500,
			wantI: 100,
		},
		"middle": {
			b:     NewIntBounds(1, 10, 100),
			value: 50,
			wantI: 50,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotI := tt.b.constrainedValue(tt.value); gotI != tt.wantI {
				t.Errorf("IntBounds.constrainedValue() = %v, want %v", gotI, tt.wantI)
			}
		})
	}
}

func TestSetFlagIndicator(t *testing.T) {
	originalIndicator := flagPrefix
	defer func() {
		flagPrefix = originalIndicator
	}()
	tests := map[string]struct {
		val string
	}{
		"-":  {val: "-"},
		"--": {val: "--"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			SetFlagIndicator(tt.val)
			if got := flagIndicator(); got != tt.val {
				t.Errorf("SetFlagIndicator got %q want %q", got, tt.val)
			}
		})
	}
}

func TestAssignFileSystem(t *testing.T) {
	originalFileSystem := FileSystem()
	defer func() {
		fileSystem = originalFileSystem
	}()
	tests := map[string]struct {
		fs   afero.Fs
		want afero.Fs
	}{
		"simple": {fs: afero.NewMemMapFs(), want: FileSystem()},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := AssignFileSystem(tt.fs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AssignFileSystem() = %v, want %v", got, tt.want)
			}
		})
	}
}
