package cmd_toolkit_test

import (
	"errors"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func TestCopyFile(t *testing.T) {
	// note: use os filesystem - Create never returns an error in the
	// memory-mapped file system
	_ = cmdtoolkit.FileSystem().Mkdir("sourceDir1", cmdtoolkit.StdDirPermissions)
	_ = cmdtoolkit.FileSystem().Mkdir("sourceDir2", cmdtoolkit.StdDirPermissions)
	_ = cmdtoolkit.FileSystem().Mkdir("sourceDir3", cmdtoolkit.StdDirPermissions)
	_ = cmdtoolkit.FileSystem().Mkdir("sourceDir4", cmdtoolkit.StdDirPermissions)
	_ = cmdtoolkit.FileSystem().Mkdir("destinationDir1", cmdtoolkit.StdDirPermissions)
	_ = cmdtoolkit.FileSystem().MkdirAll(filepath.Join("destinationDir2", "file1"), cmdtoolkit.StdDirPermissions)
	_ = cmdtoolkit.FileSystem().Mkdir("destinationDir3", cmdtoolkit.StdDirPermissions)
	_ = afero.WriteFile(cmdtoolkit.FileSystem(), filepath.Join("sourceDir2", "file1"), []byte{1, 2, 3}, cmdtoolkit.StdFilePermissions)
	_ = afero.WriteFile(cmdtoolkit.FileSystem(), filepath.Join("sourceDir3", "file1"), []byte{1, 2, 3}, cmdtoolkit.StdFilePermissions)
	_ = afero.WriteFile(cmdtoolkit.FileSystem(), filepath.Join("sourceDir4", "file1"), []byte{1, 2, 3}, cmdtoolkit.StdFilePermissions)
	defer func() {
		_ = cmdtoolkit.FileSystem().RemoveAll("sourceDir1")
		_ = cmdtoolkit.FileSystem().RemoveAll("sourceDir2")
		_ = cmdtoolkit.FileSystem().RemoveAll("sourceDir3")
		_ = cmdtoolkit.FileSystem().RemoveAll("sourceDir4")
		_ = cmdtoolkit.FileSystem().RemoveAll("destinationDir1")
		_ = cmdtoolkit.FileSystem().RemoveAll("destinationDir2")
		_ = cmdtoolkit.FileSystem().RemoveAll("destinationDir3")
	}()
	type args struct {
		src         string
		destination string
	}
	tests := map[string]struct {
		args
		wantErr bool
	}{
		"copy file onto itself": {
			args:    args{src: "file1", destination: "file1"},
			wantErr: true,
		},
		"non-existent source": {
			args: args{
				src:         filepath.Join("sourceDir1", "file1"),
				destination: filepath.Join("destinationDir1", "file1"),
			},
			wantErr: true,
		},
		"destination is a directory": {
			args: args{
				src:         filepath.Join("sourceDir2", "file1"),
				destination: filepath.Join("destinationDir2", "file1"),
			},
			wantErr: true,
		},
		"success": {
			args: args{
				src:         filepath.Join("sourceDir3", "file1"),
				destination: filepath.Join("destinationDir3", "file1"),
			},
			wantErr: false,
		},
		"error writing to non-existent directory": {
			args: args{
				src:         filepath.Join("sourceDir4", "file1"),
				destination: filepath.Join("destinationDir4", "file2")},
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotErr := cmdtoolkit.CopyFile(tt.args.src, tt.args.destination); (gotErr != nil) != tt.wantErr {
				t.Errorf("CopyFile() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestDirExists(t *testing.T) {
	tests := map[string]struct {
		path string
		want bool
	}{
		"dir":               {path: ".", want: true},
		"file":              {path: "fileio_test.go", want: false},
		"non-existent file": {path: "no such dir", want: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmdtoolkit.DirExists(tt.path); got != tt.want {
				t.Errorf("DirExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogFileDeletionFailure(t *testing.T) {
	type args struct {
		s string
		e error
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"basic": {
			args:            args{s: "filename", e: errors.New("file is locked")},
			WantedRecording: output.WantedRecording{Log: "level='error' error='file is locked' fileName='filename' msg='cannot delete file'\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmdtoolkit.LogFileDeletionFailure(o, tt.args.s, tt.args.e)
			o.Report(t, "LogFileDeletionFailure()", tt.WantedRecording)
		})
	}
}

func TestMkdir(t *testing.T) {
	originalFileSystem := cmdtoolkit.FileSystem()
	defer cmdtoolkit.AssignFileSystem(originalFileSystem)
	cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	tests := map[string]struct {
		preTest func()
		dir     string
		wantErr bool
	}{
		"subdirectory of non-existent directory": {
			preTest: func() {},
			dir:     filepath.Join("non-existent directory", "subDir"),
			wantErr: true,
		},
		"dir is a plain file": {
			preTest: func() {
				_ = cmdtoolkit.FileSystem().Mkdir("plainFile", cmdtoolkit.StdDirPermissions)
				_ = afero.WriteFile(cmdtoolkit.FileSystem(), filepath.Join("plainFile", "subDir"), []byte{0, 1, 2}, cmdtoolkit.StdFilePermissions)
			},
			dir:     filepath.Join("plainFile", "subDir"),
			wantErr: true,
		},
		"successfully create new directory": {
			preTest: func() {
				_ = cmdtoolkit.FileSystem().Mkdir("emptyDir", cmdtoolkit.StdDirPermissions)
			},
			dir:     filepath.Join("emptyDir", "subDir"),
			wantErr: false,
		},
		"directory already exists": {
			preTest: func() {
				_ = cmdtoolkit.FileSystem().MkdirAll(filepath.Join("dirExists", "subDir"), cmdtoolkit.StdDirPermissions)
			},
			dir:     filepath.Join("dirExists", "subDir"),
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			if gotErr := cmdtoolkit.Mkdir(tt.dir); (gotErr != nil) != tt.wantErr {
				t.Errorf("Mkdir() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestPlainFileExists(t *testing.T) {
	originalFileSystem := cmdtoolkit.FileSystem()
	defer cmdtoolkit.AssignFileSystem(originalFileSystem)
	cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	tests := map[string]struct {
		preTest func()
		path    string
		want    bool
	}{
		"non-existent file": {
			preTest: func() {},
			path:    "file",
			want:    false,
		},
		"directory": {
			preTest: func() {
				_ = cmdtoolkit.FileSystem().Mkdir("file", cmdtoolkit.StdDirPermissions)
			},
			path: "file",
			want: false,
		},
		"real file": {
			preTest: func() {
				_ = cmdtoolkit.FileSystem().Mkdir("dir", cmdtoolkit.StdDirPermissions)
				_ = afero.WriteFile(cmdtoolkit.FileSystem(), filepath.Join("dir", "file"), []byte{0, 1, 2}, cmdtoolkit.StdFilePermissions)
			},
			path: filepath.Join("dir", "file"),
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			if got := cmdtoolkit.PlainFileExists(tt.path); got != tt.want {
				t.Errorf("PlainFileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadDirectory(t *testing.T) {
	originalFileSystem := cmdtoolkit.FileSystem()
	defer cmdtoolkit.AssignFileSystem(originalFileSystem)
	cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	tests := map[string]struct {
		preTest         func()
		dir             string
		wantFilesLength int
		wantOk          bool
		output.WantedRecording
	}{
		"non-existent directory": {
			preTest: func() {},
			dir:     "no such dir",
			wantOk:  false,
			WantedRecording: output.WantedRecording{
				Error: "The directory \"no such dir\" cannot be read: open no such dir: file does not exist.\n",
				Log:   "level='error' directory='no such dir' error='open no such dir: file does not exist' msg='cannot read directory'\n",
			},
		},
		"empty directory": {
			preTest: func() {
				_ = cmdtoolkit.FileSystem().Mkdir("empty", cmdtoolkit.StdDirPermissions)
			},
			dir:             "empty",
			wantFilesLength: 0,
			wantOk:          true,
		},
		"directory with content": {
			preTest: func() {
				_ = cmdtoolkit.FileSystem().Mkdir("full", cmdtoolkit.StdDirPermissions)
				// make a few files
				for _, filename := range []string{"file1", "file2", "file3"} {
					_ = afero.WriteFile(cmdtoolkit.FileSystem(), filepath.Join("full", filename), []byte{}, cmdtoolkit.StdFilePermissions)
				}
				// and a few directories
				for _, subDirectory := range []string{"sub1", "sub2", "sub3"} {
					_ = cmdtoolkit.FileSystem().Mkdir(filepath.Join("full", subDirectory), cmdtoolkit.StdDirPermissions)
				}
			},
			dir:             "full",
			wantFilesLength: 6,
			wantOk:          true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			o := output.NewRecorder()
			gotFiles, gotOk := cmdtoolkit.ReadDirectory(o, tt.dir)
			if len(gotFiles) != tt.wantFilesLength {
				t.Errorf("ReadDirectory() got %d files, want %d", len(gotFiles), tt.wantFilesLength)
			}
			if gotOk != tt.wantOk {
				t.Errorf("ReadDirectory() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			o.Report(t, "ReadDirectory()", tt.WantedRecording)
		})
	}
}

func TestReportFileCreationFailure(t *testing.T) {
	type args struct {
		cmd  string
		file string
		e    error
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"basic": {
			args: args{cmd: "myCommand", file: "myPoorFile", e: errors.New("no disk space")},
			WantedRecording: output.WantedRecording{
				Error: "The file \"myPoorFile\" cannot be created: no disk space.\n",
				Log:   "level='error' command='myCommand' error='no disk space' fileName='myPoorFile' msg='cannot create file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmdtoolkit.ReportFileCreationFailure(o, tt.args.cmd, tt.args.file, tt.args.e)
			o.Report(t, "ReportFileCreationFailure()", tt.WantedRecording)
		})
	}
}

func TestAssignFileSystem(t *testing.T) {
	originalFileSystem := cmdtoolkit.FileSystem()
	defer cmdtoolkit.AssignFileSystem(originalFileSystem)
	tests := map[string]struct {
		fs   afero.Fs
		want afero.Fs
	}{
		"simple": {fs: afero.NewMemMapFs(), want: cmdtoolkit.FileSystem()},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmdtoolkit.AssignFileSystem(tt.fs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AssignFileSystem() = %v, want %v", got, tt.want)
			}
		})
	}
}
