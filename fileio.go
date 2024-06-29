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
func CopyFile(src, destination string) error {
	absSrc, _ := filepath.Abs(src)
	absDestination, _ := filepath.Abs(destination)
	if absSrc == absDestination {
		return fmt.Errorf("cannot copy file %q to itself", absSrc)
	}
	openedSrc, srcOpenErr := fileSystem.Open(src)
	if srcOpenErr != nil {
		return srcOpenErr
	}
	defer func() {
		_ = openedSrc.Close()
	}()
	if destinationIsDir, _ := afero.IsDir(fileSystem, destination); destinationIsDir {
		return fmt.Errorf("cannot overwrite a directory")
	}
	openedDestination, destinationOpenErr := fileSystem.Create(destination)
	if destinationOpenErr != nil {
		return destinationOpenErr
	}
	defer func() {
		_ = openedDestination.Close()
	}()
	_, _ = io.Copy(openedDestination, openedSrc)
	return nil
}

// CreateFile creates a file; it returns an error if the file already exists
func CreateFile(fileName string, content []byte) error {
	if fileExists, _ := afero.Exists(fileSystem, fileName); fileExists {
		return fmt.Errorf("file %q already exists", fileName)
	}
	dir := filepath.Dir(fileName)
	if dirExists, _ := afero.DirExists(fileSystem, dir); !dirExists {
		return fmt.Errorf("%q does not exist or is not a directory", dir)
	}
	return afero.WriteFile(fileSystem, fileName, content, StdFilePermissions) // bad path
}

// DirExists returns whether the specified file exists as a directory
func DirExists(path string) bool {
	pathIsDir, _ := afero.IsDir(fileSystem, path)
	return pathIsDir
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

// logUnreadableDirectory logs errors when a directory cannot be read; does not
// write anything to the error output because that typically needs additional
// context
func logUnreadableDirectory(o output.Bus, s string, e error) {
	o.Log(output.Error, "cannot read directory", map[string]any{
		"directory": s,
		"error":     e,
	})
}

// Mkdir makes the specified directory; succeeds if the directory already
// exists. Fails if a plain file exists with the specified path.
func Mkdir(path string) error {
	pathIsDir, err := afero.IsDir(fileSystem, path)
	if pathIsDir || (err != nil && !errors.Is(err, afero.ErrFileNotFound)) {
		return err
	}
	if PlainFileExists(path) {
		return fmt.Errorf("file exists and is not a directory")
	}
	parentIsDir, _ := afero.IsDir(fileSystem, filepath.Dir(path))
	if !parentIsDir {
		return fmt.Errorf("parent directory is not a directory")
	}
	return fileSystem.Mkdir(path, StdDirPermissions)
}

// PlainFileExists returns whether the specified file exists as a plain file
// (i.e., not a directory)
func PlainFileExists(path string) bool {
	f, err := fileSystem.Stat(path)
	if err != nil {
		return false
	}
	return !f.IsDir()
}

// ReadDirectory returns the contents of a specified directory
func ReadDirectory(o output.Bus, dir string) ([]fs.FileInfo, bool) {
	files, fileErr := afero.ReadDir(fileSystem, dir)
	if fileErr == nil {
		return files, true
	}
	logUnreadableDirectory(o, dir, fileErr)
	o.WriteCanonicalError("The directory %q cannot be read: %v", dir, fileErr)
	return nil, false
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

func writeDirectoryCreationError(o output.Bus, d string, e error) {
	o.WriteCanonicalError("The directory %q cannot be created: %v", d, e)
}
