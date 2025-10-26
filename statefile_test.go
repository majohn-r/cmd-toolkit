package cmd_toolkit_test

import (
	"os"
	"testing"

	"github.com/adrg/xdg"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
)

type testImpl struct{}

func (ti testImpl) Write(_ string, _ []byte) error {
	return nil
}

func (ti testImpl) Exists(_ string) bool {
	return false
}

func (ti testImpl) Create(_ string) error {
	return nil
}

func (ti testImpl) Remove(_ string) error {
	return nil
}

func (ti testImpl) Close() {
}

func (ti testImpl) Read(_ string) ([]byte, error) {
	return nil, nil
}

func TestInitStateFile(t *testing.T) {
	var originalStateHome = xdg.StateHome
	testDir := "./stateFileTestInitStateFile"
	if err := os.Mkdir(testDir, 0o700); err != nil {
		t.Errorf("failed to create test directory %q: %v", testDir, err)
	}
	defer func() {
		_ = os.RemoveAll(testDir)
	}()
	tests := map[string]struct {
		appName  string
		want     cmdtoolkit.StateFile
		wantErr  bool
		preTest  func()
		postTest func()
	}{
		"bad state home": {
			appName: "foo",
			wantErr: true,
			preTest: func() {
				xdg.StateHome = ""
			},
			postTest: func() {
				xdg.StateHome = originalStateHome
			},
		},
		"bad app name": {
			appName: "",
			wantErr: true,
		},
		"unusable app name": {
			appName: "go.mod",
			wantErr: true,
			preTest: func() {
				xdg.StateHome = "."
			},
			postTest: func() {
				xdg.StateHome = originalStateHome
			},
		},
		"happy path": {
			appName: "foo",
			wantErr: false,
			want:    testImpl{},
			preTest: func() {
				xdg.StateHome = testDir
			},
			postTest: func() {
				xdg.StateHome = originalStateHome
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.postTest != nil {
				defer tt.postTest()
			}
			if tt.preTest != nil {
				tt.preTest()
			}
			got, err := cmdtoolkit.InitStateFile(tt.appName)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitStateFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && ((got == nil && tt.want != nil) || (got != nil && tt.want == nil)) {
				t.Errorf("InitStateFile() got = %v, want %v", got, tt.want)
			}
			if got != nil {
				got.Close()
			}
			if tt.want != nil {
				tt.want.Close()
			}
		})
	}
}
