package cmd_toolkit

import (
	"errors"
	"fmt"
	"github.com/majohn-r/output"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"io/fs"
	"path/filepath"
)

const (
	defaultConfigFileName = "defaults.yaml"
)

// DefaultConfigFileStatus returns the path of the defaults config file and whether that file exists
func DefaultConfigFileStatus() (string, bool) {
	path := filepath.Join(ApplicationPath(), defaultConfigFileName)
	exists := PlainFileExists(path)
	return path, exists
}

// ReadDefaultsConfigFile reads defaults.yaml from the specified path and returns
// a pointer to a cooked Configuration instance; if there is no such file, then
// an empty Configuration is returned and ok is true
func ReadDefaultsConfigFile(o output.Bus) (*Configuration, bool) {
	c := EmptyConfiguration()
	path := ApplicationPath()
	file := filepath.Join(path, defaultConfigFileName)
	exists, fileError := verifyDefaultConfigFileExists(o, file)
	if fileError != nil {
		return c, false
	}
	if !exists {
		return c, true
	}
	// only probable error circumvented by verifyFileExists failure
	rawYaml, _ := afero.ReadFile(fileSystem, file)
	data := map[string]any{}
	fileError = yaml.Unmarshal(rawYaml, &data)
	if fileError != nil {
		o.Log(output.Error, "cannot unmarshal yaml content", map[string]any{
			"directory": path,
			"fileName":  defaultConfigFileName,
			"error":     fileError,
		})
		o.WriteCanonicalError("The configuration file %q is not well-formed YAML: %v", file, fileError)
		o.WriteCanonicalError("What to do:\nDelete the file %q from %q and restart the application", defaultConfigFileName, path)
		return c, false
	}
	c = newConfiguration(o, data)
	o.Log(output.Info, "read configuration file", map[string]any{
		"directory": path,
		"fileName":  defaultConfigFileName,
		"value":     c,
	})
	return c, true
}

func reportInvalidConfigurationData(o output.Bus, s string, e error) {
	o.WriteCanonicalError("The configuration file %q contains an invalid value for %q: %v", defaultConfigFileName, s, e)
	o.Log(output.Error, "invalid content in configuration file", map[string]any{
		"section": s,
		"error":   e,
	})
}

func verifyDefaultConfigFileExists(o output.Bus, path string) (exists bool, err error) {
	var f fs.FileInfo
	f, err = fileSystem.Stat(path)
	switch {
	case err == nil:
		if f.IsDir() {
			o.Log(output.Error, "file is a directory", map[string]any{
				"directory": filepath.Dir(path),
				"fileName":  filepath.Base(path),
			})
			o.WriteCanonicalError("The configuration file %q is a directory", path)
			o.WriteCanonicalError("What to do:\nDelete the directory %q from %q and restart the application", filepath.Base(path), filepath.Dir(path))
			err = fmt.Errorf("file exists but is a directory")
		} else {
			exists = true
		}
	case errors.Is(err, afero.ErrFileNotFound):
		o.Log(output.Info, "file does not exist", map[string]any{
			"directory": filepath.Dir(path),
			"fileName":  filepath.Base(path),
		})
		err = nil
	}
	return
}
