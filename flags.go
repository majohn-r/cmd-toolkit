package cmd_toolkit

import (
	"fmt"
	"github.com/majohn-r/output"
	"github.com/spf13/pflag"
	"reflect"
	"slices"
)

type valueType int32

const (
	unspecifiedType valueType = iota
	// BoolType represents a boolean flag type
	BoolType
	// IntType represents an integer flag type
	IntType
	// StringType represents a string flag type
	StringType
)

type commandFlagValue interface {
	string | int | bool | any
}

// CommandFlag captures a flag value and whether it was set by the user, i.e., on the command line
type CommandFlag[V commandFlagValue] struct {
	// Value is the flag's value
	Value V
	// UserSet is set if the flag value came from the command line
	UserSet bool
}

// FlagDetails captures the data needed by the cobra command code: a flag's abbreviated name
// (which may be empty), a usage string, its expected type, and its default value
type FlagDetails struct {
	// AbbreviatedName is the flag's single character abbreviation, if any; typically empty
	AbbreviatedName string
	// Usage is a brief description of what the flag controls
	Usage string
	// ExpectedType describes whether the flag should be boolean, integer, or string
	ExpectedType valueType
	// DefaultValue gives the default value for the flag
	DefaultValue any
}

// Copy provides a copy of a FlagDetails instance - of primary use to test code.
func (fD *FlagDetails) Copy() *FlagDetails {
	return &FlagDetails{
		AbbreviatedName: fD.AbbreviatedName,
		Usage:           fD.Usage,
		ExpectedType:    fD.ExpectedType,
		DefaultValue:    fD.DefaultValue,
	}
}

type configSource interface {
	// BoolDefault provides a boolean default value
	BoolDefault(string, bool) (bool, error)
	// IntDefault provides an integer default value
	IntDefault(string, *IntBounds) (int, error)
	// StringDefault provides a string default value
	StringDefault(string, string) (string, error)
}

func (fD *FlagDetails) addFlag(o output.Bus, c configSource, consumer *pflag.FlagSet, flag flagParam) {
	switch fD.ExpectedType {
	case StringType:
		statedDefault, _ok := fD.DefaultValue.(string)
		if !_ok {
			reportDefaultTypeError(o, flag.name, "string", fD.DefaultValue)
			return
		}
		newDefault, malformedDefault := c.StringDefault(flag.name, statedDefault)
		if malformedDefault != nil {
			reportInvalidConfigurationData(o, flag.set, malformedDefault)
			return
		}
		usage := decorateStringFlagUsage(fD.Usage, newDefault)
		switch fD.AbbreviatedName {
		case "":
			consumer.String(flag.name, newDefault, usage)
		default:
			consumer.StringP(flag.name, fD.AbbreviatedName, newDefault, usage)
		}
	case BoolType:
		statedDefault, _ok := fD.DefaultValue.(bool)
		if !_ok {
			reportDefaultTypeError(o, flag.name, "bool", fD.DefaultValue)
			return
		}
		newDefault, malformedDefault := c.BoolDefault(flag.name, statedDefault)
		if malformedDefault != nil {
			reportInvalidConfigurationData(o, flag.set, malformedDefault)
			return
		}
		usage := decorateBoolFlagUsage(fD.Usage, newDefault)
		switch fD.AbbreviatedName {
		case "":
			consumer.Bool(flag.name, newDefault, usage)
		default:
			consumer.BoolP(flag.name, fD.AbbreviatedName, newDefault, usage)
		}
	case IntType:
		bounds, _ok := fD.DefaultValue.(*IntBounds)
		if !_ok {
			reportDefaultTypeError(o, flag.name, "*cmd_toolkit.IntBounds", fD.DefaultValue)
			return
		}
		newDefault, malformedDefault := c.IntDefault(flag.name, bounds)
		if malformedDefault != nil {
			reportInvalidConfigurationData(o, flag.set, malformedDefault)
			return
		}
		usage := decorateIntFlagUsage(fD.Usage, newDefault)
		switch fD.AbbreviatedName {
		case "":
			consumer.Int(flag.name, newDefault, usage)
		default:
			consumer.IntP(flag.name, fD.AbbreviatedName, newDefault, usage)
		}
	default:
		o.WriteCanonicalError(
			"An internal error occurred: unspecified flag type; set %q, flag %q",
			flag.set, flag.name)
		o.Log(output.Error, "internal error", map[string]any{
			"set":            flag.set,
			"flag":           flag.name,
			"specified-type": fD.ExpectedType,
			"default":        fD.DefaultValue,
			"default-type":   reflect.TypeOf(fD.DefaultValue),
			"error":          "unspecified flag type",
		})
	}
}

// FlagSet captures a set of flags (typically, but not necessarily, associated with a cobra
// command) and the details of the flags for that set
type FlagSet struct {
	// Name is the name of the set, typically, of a command
	Name string
	// Details provides a map of FlagDetails keyed by their flag names
	Details map[string]*FlagDetails // keys are flag names
}

// FlagProducer encapsulates critical behavior of the cobra command flags for reading flag values
// and whether those values are changed (i.e., are defined on the command line)
type FlagProducer interface {
	// Changed returns true if the flag was set on the command line
	Changed(name string) bool
	// GetBool returns the boolean value of the named flag
	GetBool(name string) (bool, error)
	// GetInt returns the integer value of the named flag
	GetInt(name string) (int, error)
	// GetString returns the string value of the named flag
	GetString(name string) (string, error)
}

type flagParam struct {
	set  string
	name string
}

// AddFlags adds collections of flags to a flag consumer (typically a cobra command flags
// instance)
func AddFlags(o output.Bus, c *Configuration, flags *pflag.FlagSet, sets ...*FlagSet) {
	for _, set := range sets {
		config := c.SubConfiguration(set.Name)
		// sort names for deterministic test output
		sortedNames := make([]string, 0, len(set.Details))
		for name := range set.Details {
			sortedNames = append(sortedNames, name)
		}
		slices.Sort(sortedNames)
		for _, name := range sortedNames {
			details := set.Details[name]
			switch details {
			case nil:
				o.WriteCanonicalError(
					"an internal error occurred: there are no details for flag %q", name)
				o.Log(output.Error, "internal error", map[string]any{
					"set":   set.Name,
					"flag":  name,
					"error": "no details present",
				})
			default:
				details.addFlag(o, config, flags, flagParam{
					set:  set.Name,
					name: name,
				})
			}
		}
	}
}

// GetBool gets the boolean value of a specific flag, handling common error conditions
func GetBool(o output.Bus, results map[string]*CommandFlag[any], flagName string) (CommandFlag[bool], error) {
	fv, flagNotFound := extractFlagValue(o, results, flagName)
	if flagNotFound != nil {
		return CommandFlag[bool]{}, flagNotFound
	}
	if fv == nil {
		return CommandFlag[bool]{}, reportMissingFlagData(o, flagName)
	}
	v, ok := fv.Value.(bool)
	if !ok {
		return CommandFlag[bool]{}, reportIncorrectlyTypedValue(
			o,
			"a boolean",
			flagName,
			fv,
		)
	}
	return CommandFlag[bool]{
		Value:   v,
		UserSet: fv.UserSet,
	}, nil
}

// GetInt gets the integer value of a specific flag, handling common error conditions
func GetInt(o output.Bus, results map[string]*CommandFlag[any], flagName string) (CommandFlag[int], error) {
	fv, flagNotFound := extractFlagValue(o, results, flagName)
	if flagNotFound != nil {
		return CommandFlag[int]{}, flagNotFound
	}
	if fv == nil {
		return CommandFlag[int]{}, reportMissingFlagData(o, flagName)
	}
	v, ok := fv.Value.(int)
	if !ok {
		return CommandFlag[int]{}, reportIncorrectlyTypedValue(
			o,
			"an integer",
			flagName,
			fv,
		)
	}
	return CommandFlag[int]{Value: v, UserSet: fv.UserSet}, nil
}

// GetString gets the string value of a specific flag, handling common error conditions
func GetString(o output.Bus, results map[string]*CommandFlag[any], flagName string) (CommandFlag[string], error) {
	fv, flagNotFound := extractFlagValue(o, results, flagName)
	if flagNotFound != nil {
		return CommandFlag[string]{}, flagNotFound
	}
	if fv == nil {
		e := reportMissingFlagData(o, flagName)
		return CommandFlag[string]{}, e
	}
	v, ok := fv.Value.(string)
	if !ok {
		return CommandFlag[string]{}, reportIncorrectlyTypedValue(
			o,
			"a string",
			flagName,
			fv,
		)
	}
	return CommandFlag[string]{Value: v, UserSet: fv.UserSet}, nil
}

// ProcessFlagErrors handles a slice of errors; returns true iff the slice is empty
func ProcessFlagErrors(o output.Bus, eSlice []error) bool {
	if len(eSlice) != 0 {
		for _, e := range eSlice {
			o.WriteCanonicalError("an internal error occurred: %v", e)
			o.Log(output.Error, "internal error", map[string]any{"error": e})
		}
		return false
	}
	return true
}

// ReadFlags reads the flags from a producer (typically a cobra commands flag structure)
func ReadFlags(producer FlagProducer, set *FlagSet) (map[string]*CommandFlag[any], []error) {
	m := map[string]*CommandFlag[any]{}
	var e []error
	// sort names for deterministic output in unit tests
	sortedNames := make([]string, 0, len(set.Details))
	for name := range set.Details {
		sortedNames = append(sortedNames, name)
	}
	slices.Sort(sortedNames)
	for _, name := range sortedNames {
		details := set.Details[name]
		if details == nil {
			e = append(e, fmt.Errorf("no details for flag %q", name))
			continue
		}
		val := &CommandFlag[any]{
			UserSet: producer.Changed(name),
		}
		var flagError error
		switch details.ExpectedType {
		case BoolType:
			val.Value, flagError = producer.GetBool(name)
		case StringType:
			val.Value, flagError = producer.GetString(name)
		case IntType:
			val.Value, flagError = producer.GetInt(name)
		default:
			flagError = fmt.Errorf("unknown type for flag --%s", name)
		}
		switch flagError {
		case nil:
			m[name] = val
		default:
			e = append(e, flagError)
		}
	}
	return m, e
}

func decorateBoolFlagUsage(usage string, defaultValue bool) string {
	if defaultValue {
		return usage
	}
	return fmt.Sprintf("%s (default false)", usage)
}

func decorateIntFlagUsage(usage string, defaultValue int) string {
	if defaultValue != 0 {
		return usage
	}
	return fmt.Sprintf("%s (default 0)", usage)
}

func decorateStringFlagUsage(usage, defaultValue string) string {
	if defaultValue != "" {
		return usage
	}
	return fmt.Sprintf("%s (default \"\")", usage)
}

func extractFlagValue(o output.Bus, results map[string]*CommandFlag[any], flagName string) (fv *CommandFlag[any], e error) {
	if results == nil {
		e = fmt.Errorf("nil results")
		o.WriteCanonicalError("an internal error occurred: no flag values exist")
		o.Log(output.Error, "internal error", map[string]any{
			"error": "no results to extract flag values from",
		})
		return
	}
	value, found := results[flagName]
	if !found {
		e = fmt.Errorf("flag not found")
		o.WriteCanonicalError("an internal error occurred: flag %q is not found", flagName)
		o.Log(output.Error, "internal error", map[string]any{
			"flag":  flagName,
			"error": e,
		})
		return
	}
	fv = value
	return
}

func reportDefaultTypeError(o output.Bus, flag, expected string, value any) {
	o.WriteCanonicalError(
		"an internal error occurred: the type of flag %q's value, '%v', is '%T', but '%s' was expected", flag, value, value, expected)
	o.Log(output.Error, "internal error", map[string]any{
		"flag":     flag,
		"value":    value,
		"expected": expected,
		"actual":   reflect.TypeOf(value),
		"error":    "default value mistyped",
	})
}

func reportIncorrectlyTypedValue(o output.Bus, expected, flagName string, fv *CommandFlag[any]) error {
	e := fmt.Errorf("flag value is not %s", expected)
	o.WriteCanonicalError("an internal error occurred: flag %q is not %s (%v)",
		flagName, expected, fv.Value)
	o.Log(output.Error, "internal error", map[string]any{
		"flag":  flagName,
		"value": fv.Value,
		"error": e})
	return e
}

func reportMissingFlagData(o output.Bus, flagName string) error {
	e := fmt.Errorf("no data associated with flag")
	o.WriteCanonicalError("an internal error occurred: flag %q has no data", flagName)
	o.Log(output.Error, "internal error", map[string]any{
		"flag":  flagName,
		"error": e})
	return e
}
