package cmd_toolkit_test

import (
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAppSpecificPath(t *testing.T) {
	tests := map[string]struct {
		applicationName string
		topDir          string
		want            string
		wantErr         bool
	}{
		"uninitialized applicationName": {
			applicationName: "",
			topDir:          "topDir",
			wantErr:         true,
		},
		"initialized applicationName": {
			applicationName: "myApp",
			topDir:          "dir",
			want:            filepath.Join("dir", "myApp"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotErr := cmdtoolkit.CreateAppSpecificPath(tt.topDir, tt.applicationName)
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
