package cmd_toolkit

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func TestCopyFile(t *testing.T) {
	// note: use os filesystem - Create never returns an error in the
	// memory-mapped file system
	fileSystem.Mkdir("sourceDir1", StdDirPermissions)
	fileSystem.Mkdir("sourceDir2", StdDirPermissions)
	fileSystem.Mkdir("sourceDir3", StdDirPermissions)
	fileSystem.Mkdir("sourceDir4", StdDirPermissions)
	fileSystem.Mkdir("destDir1", StdDirPermissions)
	fileSystem.MkdirAll(filepath.Join("destDir2", "file1"), StdDirPermissions)
	fileSystem.Mkdir("destDir3", StdDirPermissions)
	afero.WriteFile(fileSystem, filepath.Join("sourceDir2", "file1"), []byte{1, 2, 3}, StdFilePermissions)
	afero.WriteFile(fileSystem, filepath.Join("sourceDir3", "file1"), []byte{1, 2, 3}, StdFilePermissions)
	afero.WriteFile(fileSystem, filepath.Join("sourceDir4", "file1"), []byte{1, 2, 3}, StdFilePermissions)
	defer func() {
		fileSystem.RemoveAll("sourceDir1")
		fileSystem.RemoveAll("sourceDir2")
		fileSystem.RemoveAll("sourceDir3")
		fileSystem.RemoveAll("sourceDir4")
		fileSystem.RemoveAll("destDir1")
		fileSystem.RemoveAll("destDir2")
		fileSystem.RemoveAll("destDir3")
	}()
	type args struct {
		src  string
		dest string
	}
	tests := map[string]struct {
		args
		wantErr bool
	}{
		"copy file onto itself": {
			args:    args{src: "file1", dest: "file1"},
			wantErr: true,
		},
		"non-existent source": {
			args:    args{src: filepath.Join("sourceDir1", "file1"), dest: filepath.Join("destDir1", "file1")},
			wantErr: true,
		},
		"destination is a directory": {
			args:    args{src: filepath.Join("sourceDir2", "file1"), dest: filepath.Join("destDir2", "file1")},
			wantErr: true,
		},
		"success": {
			args:    args{src: filepath.Join("sourceDir3", "file1"), dest: filepath.Join("destDir3", "file1")},
			wantErr: false,
		},
		"error writing to non-existent directory": {
			args:    args{src: filepath.Join("sourceDir4", "file1"), dest: filepath.Join("destDir4", "file2")},
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotErr := CopyFile(tt.args.src, tt.args.dest); (gotErr != nil) != tt.wantErr {
				t.Errorf("CopyFile() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestCreateFile(t *testing.T) {
	originalFileSystem := fileSystem
	defer func() {
		fileSystem = originalFileSystem
	}()
	fileSystem = afero.NewMemMapFs()
	type args struct {
		fileName string
		content  []byte
	}
	tests := map[string]struct {
		preTest func()
		args
		wantErr bool
	}{
		"file in non-existent directory": {
			preTest: func() {},
			args:    args{fileName: filepath.Join("no such dir", "file1"), content: []byte{1, 2, 3}},
			wantErr: true,
		},
		"pre-existing file": {
			preTest: func() {
				fileSystem.Mkdir("badDir", StdDirPermissions)
				afero.WriteFile(fileSystem, filepath.Join("badDir", "file1"), []byte{2, 4, 6}, StdFilePermissions)
			},
			args:    args{fileName: filepath.Join("badDir", "file1"), content: []byte{1, 2, 3}},
			wantErr: true,
		},
		"good file": {
			preTest: func() {
				fileSystem.Mkdir("goodDir", StdDirPermissions)
			},
			args:    args{fileName: filepath.Join("goodDir", "file1"), content: []byte{1, 2, 3}},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			if gotErr := CreateFile(tt.args.fileName, tt.args.content); (gotErr != nil) != tt.wantErr {
				t.Errorf("CreateFile() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestCreateFileInDirectory(t *testing.T) {
	originalFileSystem := fileSystem
	defer func() {
		fileSystem = originalFileSystem
	}()
	fileSystem = afero.NewMemMapFs()
	type args struct {
		dir     string
		name    string
		content []byte
	}
	tests := map[string]struct {
		preTest func()
		args
		wantErr bool
	}{
		"non-existent directory": {
			preTest: func() {},
			args:    args{dir: "no such directory", name: "who care", content: []byte{0, 1, 2}},
			wantErr: true,
		},
		"file exists": {
			preTest: func() {
				fileSystem.Mkdir("badDir", StdFilePermissions)
				afero.WriteFile(fileSystem, filepath.Join("badDir", "file"), []byte{2, 4, 6}, StdFilePermissions)
			},
			args:    args{dir: "badDir", name: "file", content: []byte{0, 1, 2}},
			wantErr: true,
		},
		"new file": {
			preTest: func() {
				fileSystem.Mkdir("goodDir", StdFilePermissions)
			},
			args:    args{dir: "goodDir", name: "file", content: []byte{0, 1, 2}},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			if gotErr := CreateFileInDirectory(tt.args.dir, tt.args.name, tt.args.content); (gotErr != nil) != tt.wantErr {
				t.Errorf("CreateFileInDirectory() error = %v, wantErr %v", gotErr, tt.wantErr)
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
			if got := DirExists(tt.path); got != tt.want {
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
			LogFileDeletionFailure(o, tt.args.s, tt.args.e)
			o.Report(t, "LogFileDeletionFailure()", tt.WantedRecording)
		})
	}
}

func TestLogUnreadableDirectory(t *testing.T) {
	type args struct {
		s string
		e error
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"basic": {
			args:            args{s: "directory name", e: errors.New("directory is missing")},
			WantedRecording: output.WantedRecording{Log: "level='error' directory='directory name' error='directory is missing' msg='cannot read directory'\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			LogUnreadableDirectory(o, tt.args.s, tt.args.e)
			o.Report(t, "LogUnreadableDirectory()", tt.WantedRecording)
		})
	}
}

func TestMkdir(t *testing.T) {
	originalFileSystem := fileSystem
	defer func() {
		fileSystem = originalFileSystem
	}()
	fileSystem = afero.NewMemMapFs()
	tests := map[string]struct {
		preTest func()
		dir     string
		wantErr bool
	}{
		"subdirectory of non-existent directory": {
			preTest: func() {},
			dir:     filepath.Join("non-existent directory", "subdir"),
			wantErr: true,
		},
		"dir is a plain file": {
			preTest: func() {
				fileSystem.Mkdir("plainfile", StdDirPermissions)
				afero.WriteFile(fileSystem, filepath.Join("plainfile", "subdir"), []byte{0, 1, 2}, StdFilePermissions)
			},
			dir:     filepath.Join("plainfile", "subdir"),
			wantErr: true,
		},
		"successfully create new directory": {
			preTest: func() {
				fileSystem.Mkdir("emptyDir", StdDirPermissions)
			},
			dir:     filepath.Join("emptyDir", "subdir"),
			wantErr: false,
		},
		"directory already exists": {
			preTest: func() {
				fileSystem.MkdirAll(filepath.Join("dirExists", "subdir"), StdDirPermissions)
			},
			dir:     filepath.Join("dirExists", "subdir"),
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			if gotErr := Mkdir(tt.dir); (gotErr != nil) != tt.wantErr {
				t.Errorf("Mkdir() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestPlainFileExists(t *testing.T) {
	originalFileSystem := fileSystem
	defer func() {
		fileSystem = originalFileSystem
	}()
	fileSystem = afero.NewMemMapFs()
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
				fileSystem.Mkdir("file", StdDirPermissions)
			},
			path: "file",
			want: false,
		},
		"real file": {
			preTest: func() {
				fileSystem.Mkdir("dir", StdDirPermissions)
				afero.WriteFile(fileSystem, filepath.Join("dir", "file"), []byte{0, 1, 2}, StdFilePermissions)
			},
			path: filepath.Join("dir", "file"),
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			if got := PlainFileExists(tt.path); got != tt.want {
				t.Errorf("PlainFileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadDirectory(t *testing.T) {
	originalFileSystem := fileSystem
	defer func() {
		fileSystem = originalFileSystem
	}()
	fileSystem = afero.NewMemMapFs()
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
				fileSystem.Mkdir("empty", StdDirPermissions)
			},
			dir:             "empty",
			wantFilesLength: 0,
			wantOk:          true,
		},
		"directory with content": {
			preTest: func() {
				fileSystem.Mkdir("full", StdDirPermissions)
				// make a few files
				for _, filename := range []string{"file1", "file2", "file3"} {
					afero.WriteFile(fileSystem, filepath.Join("full", filename), []byte{}, StdFilePermissions)
				}
				// and a few directories
				for _, subdir := range []string{"sub1", "sub2", "sub3"} {
					fileSystem.Mkdir(filepath.Join("full", subdir), StdDirPermissions)
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
			gotFiles, gotOk := ReadDirectory(o, tt.dir)
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

func TestReportDirectoryCreationFailure(t *testing.T) {
	type args struct {
		cmd string
		dir string
		e   error
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"basic": {
			args: args{cmd: "myCommand", dir: "myPoorDirectory", e: errors.New("system busy")},
			WantedRecording: output.WantedRecording{
				Error: "The directory \"myPoorDirectory\" cannot be created: system busy.\n",
				Log:   "level='error' command='myCommand' directory='myPoorDirectory' error='system busy' msg='cannot create directory'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			ReportDirectoryCreationFailure(o, tt.args.cmd, tt.args.dir, tt.args.e)
			o.Report(t, "ReportDirectoryCreationFailure()", tt.WantedRecording)
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
			ReportFileCreationFailure(o, tt.args.cmd, tt.args.file, tt.args.e)
			o.Report(t, "ReportFileCreationFailure()", tt.WantedRecording)
		})
	}
}

func TestReportFileDeletionFailure(t *testing.T) {
	type args struct {
		file string
		e    error
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"basic": {
			args: args{file: "myPoorFile", e: errors.New("file locked")},
			WantedRecording: output.WantedRecording{
				Error: "The file \"myPoorFile\" cannot be deleted: file locked.\n",
				Log:   "level='error' error='file locked' fileName='myPoorFile' msg='cannot delete file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			ReportFileDeletionFailure(o, tt.args.file, tt.args.e)
			o.Report(t, "ReportFileDeletionFailure()", tt.WantedRecording)
		})
	}
}

func TestSecureAbsolutePath(t *testing.T) {
	goodFilePath, _ := filepath.Abs("goodFile")
	tests := map[string]struct {
		path string
		want string
	}{
		"bad file name":  {path: "badFile\u0000", want: ""},
		"good file name": {path: "goodFile", want: goodFilePath},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := SecureAbsolutePath(tt.path); got != tt.want {
				t.Errorf("SecureAbsolutePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteDirectoryCreationError(t *testing.T) {
	type args struct {
		d string
		e error
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"basic": {
			args:            args{d: "dirname", e: errors.New("parent directory does not exist")},
			WantedRecording: output.WantedRecording{Error: "The directory \"dirname\" cannot be created: parent directory does not exist.\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			WriteDirectoryCreationError(o, tt.args.d, tt.args.e)
			o.Report(t, "WriteDirectoryCreationError()", tt.WantedRecording)
		})
	}
}
