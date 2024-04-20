package cmd_toolkit

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/majohn-r/output"
)

const (
	StdFilePermissions = 0o644 // -rw-r--r--
	StdDirPermissions  = 0o755 // -rwxr-xr-x
)

// CopyFile copies a file. Adapted from
// https://github.com/cleversoap/go-cp/blob/master/cp.go
func CopyFile(src, dest string) (err error) {
	src, _ = filepath.Abs(src)
	dest, _ = filepath.Abs(dest)
	if src == dest {
		return fmt.Errorf("cannot copy file %q to itself", src)
	}
	var r *os.File
	r, err = os.Open(src) // error if source does not exist
	if err == nil {
		defer r.Close()
		var w *os.File
		w, err = os.Create(dest) // error if destination is a directory
		if err == nil {
			defer w.Close()
			_, _ = io.Copy(w, r)
		}
	}
	return
}

// CreateFile creates a file; it returns an error if the file already exists
func CreateFile(fileName string, content []byte) error {
	_, err := os.Stat(fileName) // error on illegal name (such as, one containing a nul)
	if err == nil {
		return fmt.Errorf("file %q already exists", fileName)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.WriteFile(fileName, content, StdFilePermissions) // bad path
}

// CreateFileInDirectory creates a file in a specified directory. It returns an
// error if the file already exists
func CreateFileInDirectory(dir, name string, content []byte) error {
	return CreateFile(filepath.Join(dir, name), content)
}

// DirExists returns whether the specified file exists as a directory
func DirExists(path string) bool {
	f, err := os.Stat(path)
	if err == nil {
		return f.IsDir()
	}
	return !errors.Is(err, os.ErrNotExist)
}

// LogFileDeletionFailure logs errors when a file cannot be deleted; does not
// write anything to the error output because that typically needs additional
// context
func LogFileDeletionFailure(o output.Bus, s string, e error) {
	o.Log(output.Error, "cannot delete file", map[string]any{
		"fileName": s,
		"error":    e,
	})
}

// LogUnreadableDirectory logs errors when a directory cannot be read; does not
// write anything to the error output because that typically needs additional
// context
func LogUnreadableDirectory(o output.Bus, s string, e error) {
	o.Log(output.Error, "cannot read directory", map[string]any{
		"directory": s,
		"error":     e,
	})
}

// Mkdir makes the specified directory; succeeds if the directory already
// exists. Fails if a plain file exists with the specified path.
func Mkdir(dir string) (err error) {
	status, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(dir, StdDirPermissions)
		}
		return
	}
	if !status.IsDir() {
		err = fmt.Errorf("file exists and is not a directory")
	}
	return
}

// PlainFileExists returns whether the specified file exists as a plain file
// (i.e., not a directory)
func PlainFileExists(path string) bool {
	f, err := os.Stat(path)
	if err == nil {
		return !f.IsDir()
	}
	return false
}

// ReadDirectory returns the contents of a specified directory
func ReadDirectory(o output.Bus, dir string) (files []fs.DirEntry, ok bool) {
	var err error
	if files, err = os.ReadDir(dir); err != nil {
		files = nil
		LogUnreadableDirectory(o, dir, err)
		o.WriteCanonicalError("The directory %q cannot be read: %v", dir, err)
		return
	}
	ok = true
	return
}

// ReportDirectoryCreationFailure reports an error creating a directory to error
// output and to the log
func ReportDirectoryCreationFailure(o output.Bus, cmd, dir string, e error) {
	WriteDirectoryCreationError(o, dir, e)
	o.Log(output.Error, "cannot create directory", map[string]any{
		"command":   cmd,
		"directory": dir,
		"error":     e,
	})
}

// ReportFileCreationFailure reports an error creating a file to error output
// and to the log
func ReportFileCreationFailure(o output.Bus, cmd, file string, e error) {
	o.WriteCanonicalError("The file %q cannot be created: %v", file, e)
	o.Log(output.Error, "cannot create file", map[string]any{
		"command":  cmd,
		"fileName": file,
		"error":    e,
	})
}

// ReportFileDeletionFailure reports an error deleting a file to error output
// and to the log
func ReportFileDeletionFailure(o output.Bus, file string, e error) {
	o.WriteCanonicalError("The file %q cannot be deleted: %v", file, e)
	LogFileDeletionFailure(o, file, e)
}

// SecureAbsolutePath returns a path's absolute value
func SecureAbsolutePath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	return absPath
}

// WriteDirectoryCreationError writes a suitable error message to the user when
// a directory cannot be created
func WriteDirectoryCreationError(o output.Bus, d string, e error) {
	o.WriteCanonicalError("The directory %q cannot be created: %v", d, e)
}
