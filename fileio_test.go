package cmd_toolkit

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/majohn-r/output"
)

func TestCopyFile(t *testing.T) {
	type args struct {
		src  string
		dest string
	}
	tests := map[string]struct {
		preTest  func()
		postTest func()
		args
		wantErr bool
	}{
		"copy file onto itself": {
			preTest:  func() {},
			postTest: func() {},
			args:     args{src: "file1", dest: "file1"},
			wantErr:  true,
		},
		"non-existent source": {
			preTest: func() {
				_ = os.Mkdir("sourceDir", StdDirPermissions)
				_ = os.Mkdir("destDir", StdDirPermissions)
			},
			postTest: func() {
				_ = os.RemoveAll("sourceDir")
				_ = os.RemoveAll("destDir")
			},
			args:    args{src: filepath.Join("sourceDir", "file1"), dest: filepath.Join("destDir", "file1")},
			wantErr: true,
		},
		"destination is a directory": {
			preTest: func() {
				_ = os.Mkdir("sourceDir", StdDirPermissions)
				_ = os.WriteFile(filepath.Join("sourceDir", "file1"), []byte{1, 2, 3}, StdFilePermissions)
				_ = os.Mkdir(filepath.Join("destDir", "file1"), StdDirPermissions)
			},
			postTest: func() {
				_ = os.RemoveAll("sourceDir")
				_ = os.RemoveAll("destDir")
			},
			args:    args{src: filepath.Join("sourceDir", "file1"), dest: filepath.Join("destDir", "file1")},
			wantErr: true,
		},
		"success": {
			preTest: func() {
				_ = os.Mkdir("sourceDir", StdDirPermissions)
				_ = os.WriteFile(filepath.Join("sourceDir", "file1"), []byte{1, 2, 3}, StdFilePermissions)
				_ = os.Mkdir("destDir", StdDirPermissions)
			},
			postTest: func() {
				_ = os.RemoveAll("sourceDir")
				_ = os.RemoveAll("destDir")
			},
			args:    args{src: filepath.Join("sourceDir", "file1"), dest: filepath.Join("destDir", "file1")},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			defer tt.postTest()
			if err := CopyFile(tt.args.src, tt.args.dest); (err != nil) != tt.wantErr {
				t.Errorf("CopyFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateFile(t *testing.T) {
	type args struct {
		fileName string
		content  []byte
	}
	tests := map[string]struct {
		preTest  func()
		postTest func()
		args
		wantErr bool
	}{
		"corrupted file name": {
			preTest:  func() {},
			postTest: func() {},
			args:     args{fileName: "\u0000", content: []byte{1, 2, 3}},
			wantErr:  true,
		},
		"file in non-existent directory": {
			preTest:  func() {},
			postTest: func() {},
			args:     args{fileName: filepath.Join("no such dir", "file1"), content: []byte{1, 2, 3}},
			wantErr:  true,
		},
		"pre-existing file": {
			preTest: func() {
				_ = os.Mkdir("badDir", StdDirPermissions)
				_ = os.WriteFile(filepath.Join("badDir", "file1"), []byte{2, 4, 6}, StdFilePermissions)
			},
			postTest: func() {
				_ = os.RemoveAll("badDir")
			},
			args:    args{fileName: filepath.Join("badDir", "file1"), content: []byte{1, 2, 3}},
			wantErr: true,
		},
		"good file": {
			preTest: func() {
				_ = os.Mkdir("goodDir", StdDirPermissions)
			},
			postTest: func() {
				_ = os.RemoveAll("goodDir")
			},
			args:    args{fileName: filepath.Join("goodDir", "file1"), content: []byte{1, 2, 3}},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			defer tt.postTest()
			if err := CreateFile(tt.args.fileName, tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("CreateFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateFileInDirectory(t *testing.T) {
	type args struct {
		dir     string
		name    string
		content []byte
	}
	tests := map[string]struct {
		preTest  func()
		postTest func()
		args
		wantErr bool
	}{
		"non-existent directory": {
			preTest:  func() {},
			postTest: func() {},
			args:     args{dir: "no such directory", name: "who care", content: []byte{0, 1, 2}},
			wantErr:  true,
		},
		"file exists": {
			preTest: func() {
				_ = os.Mkdir("badDir", StdFilePermissions)
				_ = os.WriteFile(filepath.Join("badDir", "file"), []byte{2, 4, 6}, StdFilePermissions)
			},
			postTest: func() {
				_ = os.RemoveAll("badDir")
			},
			args:    args{dir: "badDir", name: "file", content: []byte{0, 1, 2}},
			wantErr: true,
		},
		"new file": {
			preTest: func() {
				_ = os.Mkdir("goodDir", StdFilePermissions)
			},
			postTest: func() {
				_ = os.RemoveAll("goodDir")
			},
			args:    args{dir: "goodDir", name: "file", content: []byte{0, 1, 2}},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			defer tt.postTest()
			if err := CreateFileInDirectory(tt.args.dir, tt.args.name, tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("CreateFileInDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDirExists(t *testing.T) {
	type args struct {
		path string
	}
	tests := map[string]struct {
		args
		want bool
	}{
		"dir":               {args: args{path: "."}, want: true},
		"file":              {args: args{path: "fileio_test.go"}, want: false},
		"non-existent file": {args: args{path: "no such dir"}, want: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := DirExists(tt.args.path); got != tt.want {
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
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("LogFileDeletionFailure() %s", issue)
				}
			}
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
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("LogUnreadableDirectory() %s", issue)
				}
			}
		})
	}
}

func TestMkdir(t *testing.T) {
	type args struct {
		dir string
	}
	tests := map[string]struct {
		preTest  func()
		postTest func()
		args
		wantErr bool
	}{
		"bad name": {
			preTest:  func() {},
			postTest: func() {},
			args:     args{dir: "bad directory name\u0000"},
			wantErr:  true,
		},
		"subdirectory of non-existent directory": {
			preTest:  func() {},
			postTest: func() {},
			args:     args{dir: filepath.Join("non-existent directory", "subdir")},
			wantErr:  true,
		},
		"dir is a plain file": {
			preTest: func() {
				_ = os.Mkdir("plainfile", StdDirPermissions)
				_ = os.WriteFile(filepath.Join("plainfile", "subdir"), []byte{0, 1, 2}, StdFilePermissions)
			},
			postTest: func() {
				_ = os.RemoveAll("plainfile")
			},
			args:    args{dir: filepath.Join("plainfile", "subdir")},
			wantErr: true,
		},
		"successfully create new directory": {
			preTest: func() {
				_ = os.Mkdir("emptyDir", StdDirPermissions)
			},
			postTest: func() {
				_ = os.RemoveAll("emptyDir")
			},
			args:    args{dir: filepath.Join("emptyDir", "subdir")},
			wantErr: false,
		},
		"directory already exists": {
			preTest: func() {
				_ = os.MkdirAll(filepath.Join("dirExists", "subdir"), StdDirPermissions)
			},
			postTest: func() {
				_ = os.RemoveAll("dirExists")
			},
			args:    args{dir: filepath.Join("dirExists", "subdir")},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			defer tt.postTest()
			if err := Mkdir(tt.args.dir); (err != nil) != tt.wantErr {
				t.Errorf("Mkdir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlainFileExists(t *testing.T) {
	type args struct {
		path string
	}
	tests := map[string]struct {
		preTest  func()
		postTest func()
		args
		want bool
	}{
		"bad file name": {
			preTest:  func() {},
			postTest: func() {},
			args:     args{path: "file\u0000"},
			want:     false,
		},
		"non-existent file": {
			preTest:  func() {},
			postTest: func() {},
			args:     args{path: "file"},
			want:     false,
		},
		"directory": {
			preTest: func() {
				_ = os.Mkdir("file", StdDirPermissions)
			},
			postTest: func() {
				_ = os.RemoveAll("file")
			},
			args: args{path: "file"},
			want: false,
		},
		"real file": {
			preTest: func() {
				_ = os.Mkdir("dir", StdDirPermissions)
				_ = os.WriteFile(filepath.Join("dir", "file"), []byte{0, 1, 2}, StdFilePermissions)
			},
			postTest: func() {
				_ = os.RemoveAll("dir")
			},
			args: args{path: filepath.Join("dir", "file")},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			defer tt.postTest()
			if got := PlainFileExists(tt.args.path); got != tt.want {
				t.Errorf("PlainFileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadDirectory(t *testing.T) {
	type args struct {
		dir string
	}
	tests := map[string]struct {
		preTest  func()
		postTest func()
		args
		wantFilesLength int
		wantOk          bool
		output.WantedRecording
	}{
		"non-existent directory": {
			preTest:  func() {},
			postTest: func() {},
			args:     args{dir: "no such dir"},
			wantOk:   false,
			WantedRecording: output.WantedRecording{
				Error: "The directory \"no such dir\" cannot be read: open no such dir: The system cannot find the file specified.\n",
				Log:   "level='error' directory='no such dir' error='open no such dir: The system cannot find the file specified.' msg='cannot read directory'\n",
			},
		},
		"empty directory": {
			preTest: func() {
				_ = os.Mkdir("empty", StdDirPermissions)
			},
			postTest: func() {
				_ = os.RemoveAll("empty")
			},
			args:            args{dir: "empty"},
			wantFilesLength: 0,
			wantOk:          true,
		},
		"directory with content": {
			preTest: func() {
				_ = os.Mkdir("full", StdDirPermissions)
				// make a few files
				for _, filename := range []string{"file1", "file2", "file3"} {
					_ = os.WriteFile(filepath.Join("full", filename), []byte{}, StdFilePermissions)
				}
				// and a few directories
				for _, subdir := range []string{"sub1", "sub2", "sub3"} {
					_ = os.Mkdir(filepath.Join("full", subdir), StdDirPermissions)
				}
			},
			postTest: func() {
				_ = os.RemoveAll("full")
			},
			args:            args{dir: "full"},
			wantFilesLength: 6,
			wantOk:          true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			defer tt.postTest()
			o := output.NewRecorder()
			gotFiles, gotOk := ReadDirectory(o, tt.args.dir)
			if len(gotFiles) != tt.wantFilesLength {
				t.Errorf("ReadDirectory() got %d files, want %d", len(gotFiles), tt.wantFilesLength)
			}
			if gotOk != tt.wantOk {
				t.Errorf("ReadDirectory() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ReadDirectory() %s", issue)
				}
			}
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
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ReportDirectoryCreationFailure() %s", issue)
				}
			}
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
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ReportFileCreationFailure() %s", issue)
				}
			}
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
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ReportFileDeletionFailure() %s", issue)
				}
			}
		})
	}
}

func TestSecureAbsolutePath(t *testing.T) {
	goodFilePath, _ := filepath.Abs("goodFile")
	type args struct {
		path string
	}
	tests := map[string]struct {
		args
		want string
	}{
		"bad file name":  {args: args{path: "badFile\u0000"}, want: ""},
		"good file name": {args: args{path: "goodFile"}, want: goodFilePath},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := SecureAbsolutePath(tt.args.path); got != tt.want {
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
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("WriteDirectoryCreationError() %s", issue)
				}
			}
		})
	}
}
