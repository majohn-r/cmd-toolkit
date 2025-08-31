package cmd_toolkit

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func Test_reportInvalidConfigurationData(t *testing.T) {
	type args struct {
		s string
		e error
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"simple": {
			args: args{s: "defaults", e: fmt.Errorf("illegal value")},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The configuration file \"defaults.yaml\" contains an invalid value for \"defaults\": " +
					"'illegal value'.\n",
				Log: "" +
					"level='error' " +
					"error='illegal value' " +
					"section='defaults' " +
					"msg='invalid content in configuration file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			reportInvalidConfigurationData(o, tt.args.s, tt.args.e)
			o.Report(t, "reportInvalidConfigurationData()", tt.WantedRecording)
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
			preTest: func() {},
			path:    filepath.Join(".", "non-existent-file.yaml"),
			WantedRecording: output.WantedRecording{
				Log: "level='info' directory='.' fileName='non-existent-file.yaml' msg='file does not exist'\n",
			},
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
