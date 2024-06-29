package cmd_toolkit

import (
	"os"
	"testing"

	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func TestApplicationPath(t *testing.T) {
	savedApplicationPath := applicationPath
	defer func() {
		applicationPath = savedApplicationPath
	}()
	tests := map[string]struct {
		applicationPath string
		want            string
	}{"dummy": {applicationPath: "foo/bar", want: "foo/bar"}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			applicationPath = tt.applicationPath
			if got := ApplicationPath(); got != tt.want {
				t.Errorf("ApplicationPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitApplicationPath(t *testing.T) {
	originalAppName := appName
	originalApplicationPath := applicationPath
	originalFileSystem := fileSystem
	var appDataWasSet bool
	var savedAppDataValue string
	if value, varDefined := os.LookupEnv(applicationDataEnvVarName); varDefined {
		appDataWasSet = true
		savedAppDataValue = value
	}
	defer func() {
		appName = originalAppName
		applicationPath = originalApplicationPath
		if appDataWasSet {
			_ = os.Setenv(applicationDataEnvVarName, savedAppDataValue)
		} else {
			_ = os.Unsetenv(applicationDataEnvVarName)
		}
		fileSystem = originalFileSystem
	}()
	fileSystem = afero.NewMemMapFs()
	tests := map[string]struct {
		appName         string
		appDataValue    string
		appDataSet      bool
		wantInitialized bool
		preTest         func() // things to do before calling
		output.WantedRecording
	}{
		"no appdata": {
			appName:         "beautifulApp",
			appDataSet:      false,
			wantInitialized: false,
			preTest:         func() {},
			WantedRecording: output.WantedRecording{Log: "level='error' environmentVariable='APPDATA' msg='not set'\n"},
		},
		"no app name": {
			appName:         "",
			appDataSet:      true,
			appDataValue:    "foo", // doesn't matter ...
			wantInitialized: false,
			preTest:         func() {},
			WantedRecording: output.WantedRecording{Log: "level='error' error='app name has not been initialized' msg='program error'\n"},
		},
		"appData not a directory": {
			appName:         "myApp",
			appDataSet:      true,
			appDataValue:    "apppath_test.go",
			wantInitialized: false,
			preTest:         func() {},
			WantedRecording: output.WantedRecording{
				Error: "The directory \"apppath_test.go\\\\myApp\" cannot be created: parent directory is not a directory.\n",
				Log: "" +
					"level='error'" +
					" directory='apppath_test.go\\myApp'" +
					" error='parent directory is not a directory'" +
					" msg='cannot create directory'\n",
			},
		},
		"cannot create subdirectory": {
			appName:         "myApp1",
			appDataSet:      true,
			appDataValue:    ".",
			wantInitialized: false,
			preTest: func() {
				_ = afero.WriteFile(fileSystem, "myApp1", []byte{1, 2, 3}, StdFilePermissions)
			},
			WantedRecording: output.WantedRecording{
				Error: "The directory \"myApp1\" cannot be created: file exists and is not a directory.\n",
				Log: "" +
					"level='error'" +
					" directory='myApp1'" +
					" error='file exists and is not a directory'" +
					" msg='cannot create directory'\n",
			},
		},
		"subdirectory already exists": {
			appName:         "myApp2",
			appDataSet:      true,
			appDataValue:    ".",
			wantInitialized: true,
			preTest: func() {
				_ = fileSystem.Mkdir("myApp2", StdDirPermissions)
			},
		},
		"subdirectory does not yet exist": {
			appName:         "myApp3",
			appDataSet:      true,
			appDataValue:    ".",
			wantInitialized: true,
			preTest:         func() {},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			appName = tt.appName
			applicationPath = ""
			if tt.appDataSet {
				_ = os.Setenv(applicationDataEnvVarName, tt.appDataValue)
			} else {
				_ = os.Unsetenv(applicationDataEnvVarName)
			}
			tt.preTest()
			o := output.NewRecorder()
			if gotInitialized := InitApplicationPath(o); gotInitialized != tt.wantInitialized {
				t.Errorf("InitApplicationPath() = %v, want %v", gotInitialized, tt.wantInitialized)
			}
			o.Report(t, "InitApplicationPath()", tt.WantedRecording)
		})
	}
}

func TestSetApplicationPath(t *testing.T) {
	savedApplicationPath := applicationPath
	defer func() {
		applicationPath = savedApplicationPath
	}()
	tests := map[string]struct {
		applicationPath string
		s               string
		wantPrevious    string
	}{"simple": {applicationPath: "foo", s: "bar", wantPrevious: "foo"}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			applicationPath = tt.applicationPath
			if gotPrevious := SetApplicationPath(tt.s); gotPrevious != tt.wantPrevious {
				t.Errorf("SetApplicationPath() = %v, want %v", gotPrevious, tt.wantPrevious)
			}
			if gotNew := applicationPath; gotNew != tt.s {
				t.Errorf("SetApplicationPath() gotNew = %v, want %v", gotNew, tt.s)
			}
		})
	}
}
