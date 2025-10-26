package cmd_toolkit

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func unsafeOpenRoot(filename string) *os.Root {
	root, _ := os.OpenRoot(filename)
	return root
}

func Test_stateFile_Read(t *testing.T) {
	testDir := "stateFileTestRead"
	if err := os.Mkdir(testDir, 0o777); err != nil {
		t.Errorf("error creating test directory %q: %v", testDir, err)
	}
	defer func() {
		_ = os.RemoveAll(testDir)
	}()
	goodFile := "good.txt"
	goodData := []byte("hello fencepost")
	if err := os.WriteFile(filepath.Join(testDir, goodFile), goodData, 0o666); err != nil {
		t.Errorf("error creating good file %q: %v", goodFile, err)
	}
	tests := map[string]struct {
		sf       *stateFile
		filename string
		want     []byte
		wantErr  bool
	}{
		"nil":       {sf: nil, filename: goodFile, want: nil, wantErr: true},
		"closed":    {sf: &stateFile{dir: nil}, filename: goodFile, want: nil, wantErr: true},
		"no file":   {sf: &stateFile{dir: unsafeOpenRoot(testDir)}, filename: "no file", want: nil, wantErr: true},
		"good file": {sf: &stateFile{dir: unsafeOpenRoot(testDir)}, filename: goodFile, want: goodData, wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tt.sf.Read(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Read() got = %v, want %v", got, tt.want)
			}
			tt.sf.Close()
		})
	}
}

func Test_stateFile_Write(t *testing.T) {
	testDir := "stateFileTestWrite"
	if err := os.Mkdir(testDir, 0o777); err != nil {
		t.Errorf("error creating test directory %q: %v", testDir, err)
	}
	defer func() {
		_ = os.RemoveAll(testDir)
	}()
	type args struct {
		filename string
		data     []byte
	}
	tests := map[string]struct {
		sf      *stateFile
		args    args
		wantErr bool
	}{
		"nil": {
			sf:      nil,
			args:    args{},
			wantErr: true,
		},
		"closed": {
			sf:      &stateFile{dir: nil},
			args:    args{},
			wantErr: true,
		},
		"bad filename": {
			sf:      &stateFile{dir: unsafeOpenRoot(testDir)},
			args:    args{filename: "", data: []byte{}},
			wantErr: true,
		},
		"ok": {
			sf:      &stateFile{dir: unsafeOpenRoot(testDir)},
			args:    args{filename: "test", data: []byte("test")},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := tt.sf.Write(tt.args.filename, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.sf.Close()
		})
	}
}

func Test_stateFile_Exists(t *testing.T) {
	testDir := "stateFileTestExists"
	if err := os.Mkdir(testDir, 0o777); err != nil {
		t.Errorf("error creating test directory %q: %v", testDir, err)
	}
	defer func() {
		_ = os.RemoveAll(testDir)
	}()
	realFile := "test file"
	if f, err := os.Create(filepath.Join(testDir, realFile)); err != nil {
		t.Errorf("error creating file %q: %v", realFile, err)
	} else {
		_ = f.Close()
	}
	tests := map[string]struct {
		sf       *stateFile
		filename string
		want     bool
	}{
		"nil":               {sf: nil, filename: "", want: false},
		"closed":            {sf: &stateFile{dir: nil}, filename: "", want: false},
		"non-existent file": {sf: &stateFile{dir: unsafeOpenRoot(testDir)}, filename: "no such file", want: false},
		"existent file":     {sf: &stateFile{dir: unsafeOpenRoot(testDir)}, filename: realFile, want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.sf.Exists(tt.filename); got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
			tt.sf.Close()
		})
	}
}

func Test_stateFile_Create(t *testing.T) {
	testDir := "stateFileTestCreate"
	if err := os.Mkdir(testDir, 0o777); err != nil {
		t.Errorf("error creating test directory %q: %v", testDir, err)
	}
	defer func() {
		_ = os.RemoveAll(testDir)
	}()
	tests := map[string]struct {
		filename string
		sf       *stateFile
		wantErr  bool
	}{
		"nil":           {filename: "", sf: nil, wantErr: true},
		"closed":        {filename: "", sf: &stateFile{dir: nil}, wantErr: true},
		"bad filename":  {filename: "", sf: &stateFile{dir: unsafeOpenRoot(testDir)}, wantErr: true},
		"good filename": {filename: "ok", sf: &stateFile{dir: unsafeOpenRoot(testDir)}, wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := tt.sf.Create(tt.filename); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.sf.Close()
		})
	}
}

func Test_stateFile_Remove(t *testing.T) {
	testDir := "stateFileTestRemove"
	if err := os.Mkdir(testDir, 0o777); err != nil {
		t.Errorf("error creating test directory %q: %v", testDir, err)
	}
	defer func() {
		_ = os.RemoveAll(testDir)
	}()
	// create a file that can be removed
	removable := "remove.me"
	if f, err := os.Create(filepath.Join(testDir, removable)); err != nil {
		t.Errorf("error creating removable file %q: %v", removable, err)
	} else {
		_ = f.Close()
	}
	// create file that cannot be removed
	unremovable := "unremovable"
	if err := os.Mkdir(filepath.Join(testDir, unremovable), 0o777); err != nil {
		t.Errorf("error creating removable file %q: %v", unremovable, err)
	}
	if f, err := os.Create(filepath.Join(testDir, unremovable, "dirty")); err != nil {
		t.Errorf("error creating removable file in %q: %v", unremovable, err)
	} else {
		_ = f.Close()
	}
	tests := map[string]struct {
		sf       *stateFile
		filename string
		wantErr  bool
	}{
		"nil":         {sf: nil, filename: removable, wantErr: true},
		"closed":      {sf: &stateFile{dir: nil}, filename: removable, wantErr: true},
		"unremovable": {sf: &stateFile{dir: unsafeOpenRoot(testDir)}, filename: unremovable, wantErr: true},
		"removable":   {sf: &stateFile{dir: unsafeOpenRoot(testDir)}, filename: removable, wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := tt.sf.Remove(tt.filename); (err != nil) != tt.wantErr {
				t.Errorf("Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.sf.Close()
		})
	}
}

func Test_stateFile_Close(t *testing.T) {
	savedSf := _sf
	defer func() {
		_sf = savedSf
	}()
	testDir := "./stateFileTestClose"
	if err := os.Mkdir(testDir, 0o777); err != nil {
		t.Errorf("Mkdir() error: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(testDir)
	}()
	tests := map[string]struct {
		instance *stateFile
	}{
		"nil instance":  {instance: nil},
		"no directory":  {instance: &stateFile{dir: nil}},
		"has directory": {instance: &stateFile{dir: unsafeOpenRoot(testDir)}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_sf = &stateFile{}
			tt.instance.Close()
			if tt.instance != nil {
				if tt.instance.dir != nil {
					t.Errorf("stateFile.Close() dir = %v, want nil", tt.instance.dir)
				}
				if _sf != nil {
					t.Errorf("stateFile.Close() _sf = %v, want nil", _sf)
				}
			}
		})
	}
}

func Test_validateStateHome(t *testing.T) {
	tests := map[string]struct {
		home    string
		wantErr bool
	}{
		"no home":         {home: "", wantErr: true},
		"not a directory": {home: "go.mod", wantErr: true},
		"is a directory":  {home: ".", wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := validateStateHome(tt.home); (err != nil) != tt.wantErr {
				t.Errorf("validateStateHome() error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}

func Test_validateProposedStateDir(t *testing.T) {
	testDir := "./stateFileTestValidateProposedStateDir"
	if err := os.Mkdir(testDir, 0o700); err != nil {
		t.Errorf("failed to create test directory %q: %v", testDir, err)
	}
	defer func() {
		_ = os.RemoveAll(testDir)
	}()
	existingDir := filepath.Join(testDir, "existing")
	if err := os.Mkdir(existingDir, 0o700); err != nil {
		t.Errorf("failed to create test directory %q: %v", existingDir, err)
	}
	tests := map[string]struct {
		proposedDir string
		wantErr     bool
	}{
		"weird error":                        {proposedDir: "\x00foo", wantErr: true},
		"file exists and is not a directory": {proposedDir: "go.mod", wantErr: true},
		"file exists and is a directory":     {proposedDir: existingDir, wantErr: false},
		"directory cannot be created":        {proposedDir: filepath.Join(".", "go.mod", "foo"), wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := validateProposedStateDir(tt.proposedDir); (err != nil) != tt.wantErr {
				t.Errorf("validateProposedStateDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createStateFile(t *testing.T) {
	testDir := "./stateFileTestCreateStateFile"
	if err := os.Mkdir(testDir, 0o700); err != nil {
		t.Errorf("failed to create test directory %q: %v", testDir, err)
	}
	defer func() {
		_ = os.RemoveAll(testDir)
	}()
	tests := map[string]struct {
		proposedDir string
		want        *stateFile
		wantErr     bool
	}{
		"bad file": {proposedDir: "go.mod", want: nil, wantErr: true},
		"goodFile": {proposedDir: testDir, want: &stateFile{dir: unsafeOpenRoot(testDir)}, wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := createStateFile(tt.proposedDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("createStateFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && ((got == nil && tt.want != nil) || (got != nil && tt.want == nil)) {
				t.Errorf("createStateFile() got = %v, want %v", got, tt.want)
			}
			if got != nil {
				_ = got.dir.Close()
			}
			if tt.want != nil {
				_ = tt.want.dir.Close()
			}
		})
	}
}
