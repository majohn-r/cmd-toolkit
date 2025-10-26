package cmd_toolkit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// code to deal with state files: read, write, create, and delete state files, and determine whether a specified state
// file exists (sometimes that's all an application wants to know - the file's existence or non-existence can be an
// expression of state)

// StateFile defines the functionality for state file
type StateFile interface {
	Read(filename string) ([]byte, error)
	Write(filename string, data []byte) error
	Exists(filename string) bool
	Create(filename string) error
	Remove(filename string) error
	Close()
}

type stateFile struct {
	dir *os.Root
}

var _sf *stateFile

// Read return the contents specified state file
func (sf *stateFile) Read(filename string) ([]byte, error) {
	if sf == nil || sf.dir == nil {
		return nil, fmt.Errorf("state file is nil or closed")
	}
	return sf.dir.ReadFile(filename)
}

func (sf *stateFile) Write(filename string, data []byte) error {
	if sf == nil || sf.dir == nil {
		return fmt.Errorf("state file is nil or closed")
	}
	return sf.dir.WriteFile(filename, data, 0o755)
}

// Exists returns true if the specified state file exists
func (sf *stateFile) Exists(filename string) bool {
	if sf == nil || sf.dir == nil {
		return false
	}
	_, err := sf.dir.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return err == nil
}

// Create creates the specified state file; if the file already exists, it is truncated
func (sf *stateFile) Create(filename string) error {
	if sf == nil || sf.dir == nil {
		return fmt.Errorf("state file is nil or closed")
	}
	f, err := sf.dir.Create(filename)
	defer func() {
		_ = f.Close()
	}()
	return err
}

// Remove deletes the specified filename
func (sf *stateFile) Remove(filename string) error {
	if sf == nil || sf.dir == nil {
		return fmt.Errorf("state file is nil or closed")
	}
	err := sf.dir.Remove(filename)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// Close closes the StateFile implementation; subsequent method calls will likely fail.
func (sf *stateFile) Close() {
	if sf != nil {
		if sf.dir != nil {
			_ = sf.dir.Close()
		}
		sf.dir = nil // force other calls to fail
		_sf = nil    // another call to InitStateFile can create a new instance
	}
}

// InitStateFile instantiates a useful filesystem-based implementation of StateFile
func InitStateFile(appName string) (StateFile, error) {
	var err error
	if _sf == nil {
		if !isLegalApplicationName(appName) {
			err = fmt.Errorf("application name %q is not valid", appName)
		} else {
			home := xdg.StateHome
			// trust nothing!
			if err = validateStateHome(home); err == nil {
				proposedDir := filepath.Join(home, appName)
				if err = validateProposedStateDir(proposedDir); err == nil {
					_sf, err = createStateFile(proposedDir)
				}
			}
		}
	}
	return _sf, err
}

func createStateFile(proposedDir string) (*stateFile, error) {
	root, rootErr := os.OpenRoot(proposedDir)
	if rootErr != nil {
		return nil, fmt.Errorf("state dir error: %w", rootErr)
	}
	return &stateFile{dir: root}, nil
}

// note: if the proposed directory does not exist, an attempt will be made to create it
func validateProposedStateDir(proposedDir string) error {
	f, err := os.Stat(proposedDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if createErr := os.MkdirAll(proposedDir, 0o700); createErr != nil {
				return fmt.Errorf("failed to create proposed dir %q: %w", proposedDir, createErr)
			}
		} else {
			return fmt.Errorf("state dir error: %w", err)
		}
	} else if !f.IsDir() {
		return fmt.Errorf("state dir %q is not a directory", proposedDir)
	}
	return nil
}

func validateStateHome(home string) error {
	f, err := os.Stat(home)
	if err != nil {
		return fmt.Errorf("state home error: %w", err)
	}
	if !f.IsDir() {
		return fmt.Errorf("state home %q is not a directory", home)
	}
	return err
}
