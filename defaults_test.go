package cmd_toolkit_test

import (
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/afero"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDefaultConfigFileStatus(t *testing.T) {
	originalFileSystem := cmdtoolkit.FileSystem()
	defer cmdtoolkit.AssignFileSystem(originalFileSystem)
	originalApplicationPath := cmdtoolkit.ApplicationPath()
	defer cmdtoolkit.SetApplicationPath(originalApplicationPath)
	fs := afero.NewMemMapFs()
	cmdtoolkit.AssignFileSystem(fs)
	tests := map[string]struct {
		preTest func()
		want    string
		want1   bool
	}{
		"does not exist": {
			preTest: func() {
				cmdtoolkit.SetApplicationPath("non-existent-path")
			},
			want:  `non-existent-path\defaults.yaml`,
			want1: false,
		},
		"exists": {
			preTest: func() {
				path := "goodPath"
				cmdtoolkit.SetApplicationPath(path)
				_ = cmdtoolkit.Mkdir(path)
				_ = afero.WriteFile(
					fs,
					filepath.Join(path, "defaults.yaml"),
					[]byte("data"),
					cmdtoolkit.StdFilePermissions,
				)
			},
			want:  `goodPath\defaults.yaml`,
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			got, got1 := cmdtoolkit.DefaultConfigFileStatus()
			if got != tt.want {
				t.Errorf("DefaultConfigFileStatus() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("DefaultConfigFileStatus() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestReadDefaultsConfigFile(t *testing.T) {
	originalFileSystem := cmdtoolkit.FileSystem()
	originalApplicationPath := cmdtoolkit.ApplicationPath()
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
		cmdtoolkit.SetApplicationPath(originalApplicationPath)
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
				cmdtoolkit.SetApplicationPath("configFileDir")
				_ = cmdtoolkit.FileSystem().MkdirAll(filepath.Join(cmdtoolkit.ApplicationPath(), "defaults.yaml"), cmdtoolkit.StdDirPermissions)
			},
			wantC: cmdtoolkit.EmptyConfiguration(),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The configuration file \"configFileDir\\\\defaults.yaml\" is a directory.\n" +
					"What to do:\n" +
					"Delete the directory \"defaults.yaml\" from \"configFileDir\" and restart the application.\n",
				Log: "" +
					"level='error'" +
					" directory='configFileDir'" +
					" fileName='defaults.yaml'" +
					" msg='file is a directory'\n",
			},
		},
		"no config file does not exist": {
			preTest: func() {
				cmdtoolkit.SetApplicationPath("non-existent directory")
			},
			wantC: &cmdtoolkit.Configuration{
				BoolMap:          map[string]bool{},
				ConfigurationMap: map[string]*cmdtoolkit.Configuration{},
				IntMap:           map[string]int{},
				StringMap:        map[string]string{},
			},
			wantOk:          true,
			WantedRecording: output.WantedRecording{Log: "level='info' directory='non-existent directory' fileName='defaults.yaml' msg='file does not exist'\n"},
		},
		"config file contains bad data": {
			preTest: func() {
				cmdtoolkit.SetApplicationPath("garbageDir")
				_ = cmdtoolkit.FileSystem().Mkdir(cmdtoolkit.ApplicationPath(), cmdtoolkit.StdDirPermissions)
				_ = afero.WriteFile(cmdtoolkit.FileSystem(), filepath.Join(cmdtoolkit.ApplicationPath(), "defaults.yaml"), []byte{1, 2, 3}, cmdtoolkit.StdFilePermissions)
			},
			wantC: cmdtoolkit.EmptyConfiguration(),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The configuration file \"garbageDir\\\\defaults.yaml\" is not well-formed YAML: yaml: control characters are not allowed.\n" +
					"What to do:\n" +
					"Delete the file \"defaults.yaml\" from \"garbageDir\" and restart the application.\n",
				Log: "" +
					"level='error'" +
					" directory='garbageDir'" +
					" error='yaml: control characters are not allowed'" +
					" fileName='defaults.yaml'" +
					" msg='cannot unmarshal yaml content'\n",
			},
		},
		"config file contains usable data": {
			preTest: func() {
				cmdtoolkit.SetApplicationPath("happyDir")
				_ = cmdtoolkit.FileSystem().Mkdir(cmdtoolkit.ApplicationPath(), cmdtoolkit.StdDirPermissions)
				content := "" +
					"b: true\n" +
					"i: 12\n" +
					"s: hello\n" +
					"command:\n" +
					"  default: about\n"
				_ = afero.WriteFile(cmdtoolkit.FileSystem(), filepath.Join(cmdtoolkit.ApplicationPath(), "defaults.yaml"), []byte(content), cmdtoolkit.StdFilePermissions)
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
				Log: "" +
					"level='info'" +
					" directory='happyDir'" +
					" fileName='defaults.yaml'" +
					" value='map[b:true], map[i:12], map[s:hello], map[command:map[default:about]]'" +
					" msg='read configuration file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			o := output.NewRecorder()
			gotC, gotOk := cmdtoolkit.ReadDefaultsConfigFile(o)
			if !reflect.DeepEqual(gotC, tt.wantC) {
				t.Errorf("ReadDefaultsConfigFile() gotC = %v, want %v", gotC, tt.wantC)
			}
			if gotOk != tt.wantOk {
				t.Errorf("ReadDefaultsConfigFile() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			o.Report(t, "ReadDefaultsConfigFile()", tt.WantedRecording)
		})
	}
}
