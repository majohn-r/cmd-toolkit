package cmd_toolkit

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/majohn-r/output"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

var (
	defaultConfigFileName = "defaults.yaml"
	flagIndicator         = "-"
	fileSystem            = afero.NewOsFs()
)

// FileSystem returns the current afero.Fs instance
func FileSystem() afero.Fs {
	return fileSystem
}

// AssignFileSystem sets the current afero.Fs instance and returns the original
// pre-assignment value
func AssignFileSystem(fs afero.Fs) afero.Fs {
	originalFs := fileSystem
	fileSystem = fs
	return originalFs
}

// Configuration defines the data structure for configuration information.
type Configuration struct {
	sMap map[string]string
	bMap map[string]bool
	iMap map[string]int
	cMap map[string]*Configuration
}

// IntBounds holds the bounds for an int value which has a minimum value, a
// maximum value, and a default that lies within those bounds
type IntBounds struct {
	minValue     int
	defaultValue int
	maxValue     int
}

// DefaultConfigFileName retrieves the name of the configuration file that
// contains defaults for the commands
func DefaultConfigFileName() string {
	return defaultConfigFileName
}

// FlagIndicator returns the current flag indicator, typically "-" or "--"
func FlagIndicator() string {
	return flagIndicator
}

// SetFlagIndicator sets the flag indicator to the specified value
func SetFlagIndicator(val string) {
	flagIndicator = val
}

// EmptyConfiguration creates an empty Configuration instance
func EmptyConfiguration() *Configuration {
	return &Configuration{
		bMap: make(map[string]bool),
		iMap: make(map[string]int),
		sMap: make(map[string]string),
		cMap: make(map[string]*Configuration),
	}
}

// NewConfiguration returns a Configuration instance populated as specified by
// the data parameter
func NewConfiguration(o output.Bus, data map[string]any) *Configuration {
	c := EmptyConfiguration()
	for key, v := range data {
		switch t := v.(type) {
		case string:
			c.sMap[key] = t
		case bool:
			c.bMap[key] = t
		case int:
			c.iMap[key] = t
		case map[string]any:
			c.cMap[key] = NewConfiguration(o, t)
		default:
			o.Log(output.Error, "unexpected value type", map[string]any{
				"key":   key,
				"value": v,
				"type":  fmt.Sprintf("%T", v),
			})
			o.WriteCanonicalError("The key %q, with value '%v', has an unexpected type %T", key, v, v)
			c.sMap[key] = fmt.Sprintf("%v", v)
		}
	}
	return c
}

// NewIntBounds creates a instance of IntBounds, sorting the provided value into
// reasonable fields
func NewIntBounds(v1, v2, v3 int) *IntBounds {
	v := []int{v1, v2, v3}
	sort.Ints(v)
	return &IntBounds{
		minValue:     v[0],
		defaultValue: v[1],
		maxValue:     v[2],
	}
}

// ReadConfigurationFile reads defaults.yaml from the specified path and returns
// a pointer to a cooked Configuration instance; if there is no such file, then
// an empty Configuration is returned and ok is true
func ReadConfigurationFile(o output.Bus) (c *Configuration, ok bool) {
	c = EmptyConfiguration()
	path := ApplicationPath()
	file := filepath.Join(path, defaultConfigFileName)
	exists, err := verifyDefaultConfigFileExists(o, file)
	if err != nil {
		return
	}
	if !exists {
		ok = true
		return
	}
	// only probable error circumvented by verifyFileExists failure
	rawYaml, _ := afero.ReadFile(fileSystem, file)
	data := map[string]any{}
	err = yaml.Unmarshal(rawYaml, &data)
	if err != nil {
		o.Log(output.Error, "cannot unmarshal yaml content", map[string]any{
			"directory": path,
			"fileName":  defaultConfigFileName,
			"error":     err,
		})
		o.WriteCanonicalError("The configuration file %q is not well-formed YAML: %v", file, err)
		return
	}
	c = NewConfiguration(o, data)
	o.Log(output.Info, "read configuration file", map[string]any{
		"directory": path,
		"fileName":  defaultConfigFileName,
		"value":     c,
	})
	ok = true
	return
}

// ReportInvalidConfigurationData handles errors found when attempting to parse
// a YAML configuration file, both logging the error and notifying the user of
// the error
func ReportInvalidConfigurationData(o output.Bus, s string, e error) {
	o.WriteCanonicalError("The configuration file %q contains an invalid value for %q: %v", defaultConfigFileName, s, e)
	o.Log(output.Error, "invalid content in configuration file", map[string]any{
		"section": s,
		"error":   e,
	})
}

// SetDefaultConfigFileName sets the name of the configuration file that
// contains defaults for the commands
func SetDefaultConfigFileName(s string) {
	defaultConfigFileName = s
}

func verifyDefaultConfigFileExists(o output.Bus, path string) (ok bool, err error) {
	f, err := fileSystem.Stat(path)
	switch {
	case err == nil:
		ok = !f.IsDir()
		if !ok {
			o.Log(output.Error, "file is a directory", map[string]any{
				"directory": filepath.Dir(path),
				"fileName":  filepath.Base(path),
			})
			o.WriteCanonicalError("The configuration file %q is a directory", path)
			err = fmt.Errorf("file exists but is a directory")
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

func (c *Configuration) String() string {
	s := make([]string, 0, 4)
	if len(c.bMap) != 0 {
		s = append(s, fmt.Sprintf("%v", c.bMap))
	}
	if len(c.iMap) != 0 {
		s = append(s, fmt.Sprintf("%v", c.iMap))
	}
	if len(c.sMap) != 0 {
		s = append(s, fmt.Sprintf("%v", c.sMap))
	}
	if len(c.cMap) != 0 {
		s = append(s, fmt.Sprintf("%v", c.cMap))
	}
	return strings.Join(s, ", ")
}

// BoolDefault returns a boolean value for a specified key
func (c *Configuration) BoolDefault(key string, defaultValue bool) (b bool, err error) {
	if value, ok := c.bMap[key]; ok {
		b = value
		return
	}
	if value, ok := c.iMap[key]; ok {
		switch value {
		case 0:
			b = false
		case 1:
			b = true
		default:
			// note: deliberately imitating flags behavior when parsing an
			// invalid boolean
			b = defaultValue
			err = fmt.Errorf("invalid boolean value \"%d\" for %s%s: parse error", value, FlagIndicator(), key)
		}
		return
	}
	// True values may be specified as "t", "T", "true", "TRUE", or "True"
	// False values may be specified as "f", "F", "false", "FALSE", or "False"
	value, ok := c.sMap[key]
	if !ok {
		b = defaultValue
		return
	}
	rawValue, dereferenceErr := DereferenceEnvVar(value)
	if dereferenceErr != nil {
		b = defaultValue
		err = fmt.Errorf("invalid boolean value %q for %s%s: %v", value, FlagIndicator(), key, dereferenceErr)
		return
	}
	cookedValue, e := strconv.ParseBool(rawValue)
	if e != nil {
		// note: deliberately imitating flags behavior when parsing
		// an invalid boolean
		err = fmt.Errorf("invalid boolean value %q for %s%s: parse error", value, FlagIndicator(), key)
		b = defaultValue
		return
	}
	b = cookedValue
	return
}

// BooleanValue returns a boolean value and whether it exists
func (c *Configuration) BooleanValue(key string) (value, ok bool) {
	value, ok = c.bMap[key]
	return
}

// HasSubConfiguration returns whether the specified subConfiguration exists
func (c *Configuration) HasSubConfiguration(key string) bool {
	_, ok := c.cMap[key]
	return ok
}

// IntDefault returns a default value for a specified key, which may or may not
// be defined in the Configuration instance
func (c *Configuration) IntDefault(key string, b *IntBounds) (i int, err error) {
	if value, ok := c.iMap[key]; ok {
		i = b.constrainedValue(value)
		return
	}
	value, ok := c.sMap[key]
	if !ok {
		i = b.Default()
		return
	}
	rawValue, dereferenceErr := DereferenceEnvVar(value)
	if dereferenceErr != nil {
		err = fmt.Errorf("invalid value %q for flag %s%s: %v", rawValue, FlagIndicator(), key, dereferenceErr)
		i = b.Default()
		return
	}
	cookedValue, e := strconv.Atoi(rawValue)
	if e != nil {
		// note: deliberately imitating flags behavior when parsing an
		// invalid int
		err = fmt.Errorf("invalid value %q for flag %s%s: parse error", rawValue, FlagIndicator(), key)
		i = b.Default()
		return
	}
	i = b.constrainedValue(cookedValue)
	return
}

// IntValue returns an int value and whether it exists
func (c *Configuration) IntValue(key string) (value int, ok bool) {
	value, ok = c.iMap[key]
	return
}

// StringDefault returns a string value for a specified key
func (c *Configuration) StringDefault(key, defaultValue string) (s string, err error) {
	var dereferenceErr error
	s, dereferenceErr = DereferenceEnvVar(defaultValue)
	if dereferenceErr != nil {
		err = fmt.Errorf("invalid value %q for flag %s%s: %v", defaultValue, FlagIndicator(), key, dereferenceErr)
		s = ""
		return
	}
	value, ok := c.sMap[key]
	if !ok {
		return
	}
	s, dereferenceErr = DereferenceEnvVar(value)
	if dereferenceErr != nil {
		err = fmt.Errorf("invalid value %q for flag %s%s: %v", value, FlagIndicator(), key, dereferenceErr)
		s = ""
	}
	return
}

// StringValue returns the definition of the specified key and ok if the value
// is defined
func (c *Configuration) StringValue(key string) (value string, ok bool) {
	value, ok = c.sMap[key]
	return
}

// SubConfiguration returns a specified sub-configuration
func (c *Configuration) SubConfiguration(key string) *Configuration {
	if configuration, ok := c.cMap[key]; ok {
		return configuration
	}
	return EmptyConfiguration()
}

// Default returns the default value for a bounded int
func (b *IntBounds) Default() int {
	return b.defaultValue
}

// Maximum returns the maximum value for a bounded int
func (b *IntBounds) Maximum() int {
	return b.maxValue
}

// Minimum returns the minimum value for a bounded int
func (b *IntBounds) Minimum() int {
	return b.minValue
}

func (b *IntBounds) constrainedValue(value int) (i int) {
	switch {
	case value < b.Minimum():
		i = b.Minimum()
	case value > b.Maximum():
		i = b.Maximum()
	default:
		i = value
	}
	return
}
