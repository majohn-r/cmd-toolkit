package cmd_toolkit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/majohn-r/output"
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
	savedAppname := appname
	savedApplicationPath := applicationPath
	var appDataWasSet bool
	var savedAppDataValue string
	if value, ok := os.LookupEnv(appDataVar); ok {
		appDataWasSet = true
		savedAppDataValue = value
	}
	defer func() {
		appname = savedAppname
		applicationPath = savedApplicationPath
		if appDataWasSet {
			os.Setenv(appDataVar, savedAppDataValue)
		} else {
			os.Unsetenv(appDataVar)
		}
	}()
	tests := map[string]struct {
		appname         string
		appDataValue    string
		appDataSet      bool
		wantInitialized bool
		preTest         func() // things to do before calling
		postTest        func() // things to after calling
		output.WantedRecording
	}{
		"no appdata": {
			appname:         "beautifulApp",
			appDataSet:      false,
			wantInitialized: false,
			preTest:         func() {},
			postTest:        func() {},
			WantedRecording: output.WantedRecording{Log: "level='error' environmentVariable='APPDATA' msg='not set'\n"},
		},
		"no app name": {
			appname:         "",
			appDataSet:      true,
			appDataValue:    "foo", // doesn't matter ...
			wantInitialized: false,
			preTest:         func() {},
			postTest:        func() {},
			WantedRecording: output.WantedRecording{Log: "level='error' error='app name has not been initialized' msg='program error'\n"},
		},
		"appData not a directory": {
			appname:         "myApp",
			appDataSet:      true,
			appDataValue:    "apppath_test.go",
			wantInitialized: false,
			preTest:         func() {},
			postTest:        func() {},
			WantedRecording: output.WantedRecording{
				Error: "The directory \"apppath_test.go\\\\myApp\" cannot be created: mkdir apppath_test.go\\myApp: The system cannot find the path specified.\n",
				Log:   "level='error' directory='apppath_test.go\\myApp' error='mkdir apppath_test.go\\myApp: The system cannot find the path specified.' msg='cannot create directory'\n",
			},
		},
		"cannot create subdirectory": {
			appname:         "myApp",
			appDataSet:      true,
			appDataValue:    ".",
			wantInitialized: false,
			preTest: func() {
				fileName := filepath.Join(".", "myApp")
				_ = os.WriteFile(fileName, []byte{1, 2, 3}, StdFilePermissions)
			},
			postTest: func() {
				fileName := filepath.Join(".", "myApp")
				_ = os.Remove(fileName)
			},
			WantedRecording: output.WantedRecording{
				Error: "The directory \"myApp\" cannot be created: file exists and is not a directory.\n",
				Log:   "level='error' directory='myApp' error='file exists and is not a directory' msg='cannot create directory'\n",
			},
		},
		"subdirectory already exists": {
			appname:         "myApp",
			appDataSet:      true,
			appDataValue:    ".",
			wantInitialized: true,
			preTest: func() {
				fileName := filepath.Join(".", "myApp")
				_ = os.Mkdir(fileName, StdDirPermissions)
			},
			postTest: func() {
				fileName := filepath.Join(".", "myApp")
				_ = os.Remove(fileName)
			},
		},
		"subdirectory does not yet exist": {
			appname:         "myApp",
			appDataSet:      true,
			appDataValue:    ".",
			wantInitialized: true,
			preTest: func() {
				fileName := filepath.Join(".", "myApp")
				_ = os.Remove(fileName)
			},
			postTest: func() {
				fileName := filepath.Join(".", "myApp")
				_ = os.Remove(fileName)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			appname = tt.appname
			applicationPath = ""
			if tt.appDataSet {
				os.Setenv(appDataVar, tt.appDataValue)
			} else {
				os.Unsetenv(appDataVar)
			}
			tt.preTest()
			defer tt.postTest()
			o := output.NewRecorder()
			if gotInitialized := InitApplicationPath(o); gotInitialized != tt.wantInitialized {
				t.Errorf("InitApplicationPath() = %v, want %v", gotInitialized, tt.wantInitialized)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
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
