package cmd_toolkit_test

import (
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"os"
	"path/filepath"
	"testing"
)

func TestAppName(t *testing.T) {
	originalAppName := cmdtoolkit.UnsafeAppName()
	defer func() {
		cmdtoolkit.UnsafeSetAppName(originalAppName)
	}()
	tests := map[string]struct {
		appName string
		want    string
		wantErr bool
	}{
		"get empty value":     {wantErr: true},
		"get non-empty value": {appName: "myApp", want: "myApp"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmdtoolkit.UnsafeSetAppName(tt.appName)
			got, gotErr := cmdtoolkit.AppName()
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("AppName() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AppName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateAppSpecificPath(t *testing.T) {
	originalAppName := cmdtoolkit.UnsafeAppName()
	defer func() {
		cmdtoolkit.UnsafeSetAppName(originalAppName)
	}()
	tests := map[string]struct {
		appName string
		topDir  string
		want    string
		wantErr bool
	}{
		"uninitialized appName": {wantErr: true},
		"initialized appName": {
			appName: "myApp",
			topDir:  "dir",
			want:    filepath.Join("dir", "myApp"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmdtoolkit.UnsafeSetAppName(tt.appName)
			got, gotErr := cmdtoolkit.CreateAppSpecificPath(tt.topDir)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("CreateAppSpecificPath() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateAppSpecificPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDereferenceEnvVar(t *testing.T) {
	tests := map[string]struct {
		varSettings map[string]string
		s           string
		want        string
		wantErr     bool
	}{
		"no references": {s: "no references here", want: "no references here"},
		"many references": {
			varSettings: map[string]string{
				"VAR1":     "firstVar",
				"VAR1USER": "secondVar",
				"VAR2":     "thirdVar",
			},
			s:    "$VAR1 $VAR1USER $VAR2 $VAR2, %VAR1% %VAR1USER% %VAR2%",
			want: "firstVar secondVar thirdVar thirdVar, firstVar secondVar thirdVar",
		},
		"missing references": {
			varSettings: map[string]string{
				"VAR1":     "firstVar",
				"VAR1USER": "secondVar",
				"VAR2":     "thirdVar",
			},
			s:       "$VAR1 $VAR1USER $VAR2 $VAR2 $VAR3, %VAR1% %VAR1USER% %VAR2% %VAR3%",
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mementos := make([]*cmdtoolkit.EnvVarMemento, 0)
			for varName, varValue := range tt.varSettings {
				mementos = append(mementos, cmdtoolkit.NewEnvVarMemento(varName))
				if varValue == "" {
					_ = os.Unsetenv(varName)
				} else {
					_ = os.Setenv(varName, varValue)
				}
			}
			defer func() {
				for _, memento := range mementos {
					memento.Restore()
				}
			}()
			got, gotErr := cmdtoolkit.DereferenceEnvVar(tt.s)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("DereferenceEnvVar() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DereferenceEnvVar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetAppName(t *testing.T) {
	originalAppName := cmdtoolkit.UnsafeAppName()
	defer func() {
		cmdtoolkit.UnsafeSetAppName(originalAppName)
	}()
	tests := map[string]struct {
		appName string
		s       string
		wantErr bool
	}{
		"unset, set to empty":         {wantErr: true},
		"unset, set to non-empty":     {s: "myApp"},
		"set, set to same value":      {appName: "myApp", s: "myApp"},
		"set, set to different value": {appName: "myApp", s: "myOtherApp", wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmdtoolkit.UnsafeSetAppName(tt.appName)
			if gotErr := cmdtoolkit.SetAppName(tt.s); (gotErr != nil) != tt.wantErr {
				t.Errorf("SetAppName() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}
