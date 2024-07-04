package cmd_toolkit

import (
	"github.com/majohn-r/output"
	"github.com/spf13/afero"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_newConfiguration(t *testing.T) {
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
				BoolMap:          map[string]bool{"boolean": true},
				ConfigurationMap: map[string]*Configuration{},
				IntMap:           map[string]int{"integer": 12},
				StringMap:        map[string]string{"string": "hello", "problematic": "1.234"},
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
				BoolMap: map[string]bool{"boolean": true},
				ConfigurationMap: map[string]*Configuration{
					"complex": {
						BoolMap:          map[string]bool{"another boolean": false},
						ConfigurationMap: map[string]*Configuration{},
						IntMap:           map[string]int{"another integer": 13},
						StringMap:        map[string]string{"another string": "hi!"},
					},
				},
				IntMap:    map[string]int{"integer": 12},
				StringMap: map[string]string{"string": "hello"},
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
				Error: "" +
					"The configuration file \"testPath\" is a directory.\n" +
					"What to do:\n" +
					"Delete the directory \"testPath\" from \".\" and restart the application.\n",
				Log: "level='error' directory='.' fileName='testPath' msg='file is a directory'\n",
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

func TestConfiguration_stringValue(t *testing.T) {
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
				StringMap: map[string]string{"s": "hello"},
			},
			key:       "s",
			wantValue: "hello",
			wantOk:    true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotValue, gotOk := tt.c.stringValue(tt.key)
			if gotValue != tt.wantValue {
				t.Errorf("Configuration.stringValue() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Configuration.stringValue() gotOk = %v, want %v", gotOk, tt.wantOk)
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
