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
	originalAppname := appname
	originalApplicationPath := applicationPath
	originalFileSystem := fileSystem
	var appDataWasSet bool
	var savedAppDataValue string
	if value, varDefined := os.LookupEnv(ApplicationDataEnvVarName); varDefined {
		appDataWasSet = true
		savedAppDataValue = value
	}
	defer func() {
		appname = originalAppname
		applicationPath = originalApplicationPath
		if appDataWasSet {
			os.Setenv(ApplicationDataEnvVarName, savedAppDataValue)
		} else {
			os.Unsetenv(ApplicationDataEnvVarName)
		}
		fileSystem = originalFileSystem
	}()
	fileSystem = afero.NewMemMapFs()
	tests := map[string]struct {
		appname         string
		appDataValue    string
		appDataSet      bool
		wantInitialized bool
		preTest         func() // things to do before calling
		output.WantedRecording
	}{
		"no appdata": {
			appname:         "beautifulApp",
			appDataSet:      false,
			wantInitialized: false,
			preTest:         func() {},
			WantedRecording: output.WantedRecording{Log: "level='error' environmentVariable='APPDATA' msg='not set'\n"},
		},
		"no app name": {
			appname:         "",
			appDataSet:      true,
			appDataValue:    "foo", // doesn't matter ...
			wantInitialized: false,
			preTest:         func() {},
			WantedRecording: output.WantedRecording{Log: "level='error' error='app name has not been initialized' msg='program error'\n"},
		},
		"appData not a directory": {
			appname:         "myApp",
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
			appname:         "myApp1",
			appDataSet:      true,
			appDataValue:    ".",
			wantInitialized: false,
			preTest: func() {
				afero.WriteFile(fileSystem, "myApp1", []byte{1, 2, 3}, StdFilePermissions)
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
			appname:         "myApp2",
			appDataSet:      true,
			appDataValue:    ".",
			wantInitialized: true,
			preTest: func() {
				fileSystem.Mkdir("myApp2", StdDirPermissions)
			},
		},
		"subdirectory does not yet exist": {
			appname:         "myApp3",
			appDataSet:      true,
			appDataValue:    ".",
			wantInitialized: true,
			preTest:         func() {},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			appname = tt.appname
			applicationPath = ""
			if tt.appDataSet {
				os.Setenv(ApplicationDataEnvVarName, tt.appDataValue)
			} else {
				os.Unsetenv(ApplicationDataEnvVarName)
			}
			tt.preTest()
			o := output.NewRecorder()
			if gotInitialized := InitApplicationPath(o); gotInitialized != tt.wantInitialized {
				t.Errorf("InitApplicationPath() = %v, want %v", gotInitialized, tt.wantInitialized)
			}
			if issues, verified := o.Verify(tt.WantedRecording); !verified {
				for _, issue := range issues {
					t.Errorf("InitApplicationPath() %s", issue)
				}
			}
		})
	}
}

func TestSetApplicationPath(t *testing.T) {
	savedApplicationPath := applicationPath
	defer func() {
		applicationPath = savedApplicationPath
	}()
	type args struct {
		s string
	}
	tests := map[string]struct {
		applicationPath string
		args
		wantPrevious string
	}{"simple": {applicationPath: "foo", args: args{s: "bar"}, wantPrevious: "foo"}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			applicationPath = tt.applicationPath
			if gotPrevious := SetApplicationPath(tt.args.s); gotPrevious != tt.wantPrevious {
				t.Errorf("SetApplicationPath() = %v, want %v", gotPrevious, tt.wantPrevious)
			}
			if gotNew := applicationPath; gotNew != tt.args.s {
				t.Errorf("SetApplicationPath() gotNew = %v, want %v", gotNew, tt.args.s)
			}
		})
	}
}
