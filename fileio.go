package cmd_toolkit

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

const (
	StdFilePermissions = 0o644 // -rw-r--r--
	StdDirPermissions  = 0o755 // -rwxr-xr-x
)

// CopyFile copies a file. Adapted from
// https://github.com/cleversoap/go-cp/blob/master/cp.go
func CopyFile(src, dest string) (err error) {
	absSrc, _ := filepath.Abs(src)
	absDest, _ := filepath.Abs(dest)
	if absSrc == absDest {
		return fmt.Errorf("cannot copy file %q to itself", absSrc)
	}
	var r afero.File
	r, err = fileSystem.Open(src) // error if source does not exist
	if err == nil {
		defer r.Close()
		ok, _ := afero.IsDir(fileSystem, dest)
		if ok {
			err = fmt.Errorf("cannot overwrite a directory")
			return
		}
		var w afero.File
		w, err = fileSystem.Create(dest)
		if err == nil {
			defer w.Close()
			_, _ = io.Copy(w, r)
		}
	}
	return
}

// CreateFile creates a file; it returns an error if the file already exists
func CreateFile(fileName string, content []byte) error {
	ok, _ := afero.Exists(fileSystem, fileName)
	if ok {
		return fmt.Errorf("file %q already exists", fileName)
	}
	dir := filepath.Dir(fileName)
	ok, _ = afero.DirExists(fileSystem, dir)
	if !ok {
		return fmt.Errorf("%q does not exist or is not a directory", dir)
	}
	return afero.WriteFile(fileSystem, fileName, content, StdFilePermissions) // bad path
}

// CreateFileInDirectory creates a file in a specified directory. It returns an
// error if the file already exists
func CreateFileInDirectory(dir, name string, content []byte) error {
	return CreateFile(filepath.Join(dir, name), content)
}

// DirExists returns whether the specified file exists as a directory
func DirExists(path string) bool {
	ok, _ := afero.IsDir(fileSystem, path)
	return ok
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
func Mkdir(dir string) error {
	ok, err := afero.IsDir(fileSystem, dir)
	if ok || (err != nil && !errors.Is(err, afero.ErrFileNotFound)) {
		return err
	}
	if PlainFileExists(dir) {
		return fmt.Errorf("file exists and is not a directory")
	}
	parentIsDir, _ := afero.IsDir(fileSystem, filepath.Dir(dir))
	if !parentIsDir {
		return fmt.Errorf("parent directory is not a directory")
	}
	return fileSystem.Mkdir(dir, StdDirPermissions)
}

// PlainFileExists returns whether the specified file exists as a plain file
// (i.e., not a directory)
func PlainFileExists(path string) bool {
	f, err := fileSystem.Stat(path)
	if err == nil {
		return !f.IsDir()
	}
	return false
}

// ReadDirectory returns the contents of a specified directory
func ReadDirectory(o output.Bus, dir string) (files []fs.FileInfo, ok bool) {
	var err error
	if files, err = afero.ReadDir(fileSystem, dir); err != nil {
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
