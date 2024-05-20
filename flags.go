package cmd_toolkit

import (
	"flag"
	"fmt"

	"github.com/majohn-r/output"
)

// DecorateBoolFlagUsage appends a default value to the provided usage if the
// default value is false. This is a work-around for the flag package's
// defaultUsage function, which displays each flag's usage, along with its
// default value - but it only includes the default value if the default value
// is not the zero value for the flag's type.
func DecorateBoolFlagUsage(usage string, defaultValue bool) string {
	if defaultValue {
		return usage
	}
	return fmt.Sprintf("%s (default false)", usage)
}

// DecorateIntFlagUsage appends a default value to the provided usage if the
// default value is 0. This is a work-around for the flag package's defaultUsage
// function, which displays each flag's usage, along with its default value -
// but it only includes the default value if the default value is not the zero
// value for the flag's type.
func DecorateIntFlagUsage(usage string, defaultValue int) string {
	if defaultValue != 0 {
		return usage
	}
	return fmt.Sprintf("%s (default 0)", usage)
}

// DecorateStringFlagUsage appends a default value to the provided usage if the
// default value is the empty string. This is a work-around for the flag
// package's defaultUsage function, which displays each flag's usage, along with
// its default value - but it only includes the default value if the default
// value is not the zero value for the flag's type.
func DecorateStringFlagUsage(usage, defaultValue string) string {
	if defaultValue != "" {
		return usage
	}
	return fmt.Sprintf("%s (default \"\")", usage)
}

// ProcessArgs processes a slice of command line arguments and handles common
// errors therein
func ProcessArgs(o output.Bus, f *flag.FlagSet, rawArgs []string) (processed bool) {
	args := make([]string, len(rawArgs))
	processed = true
	for i, arg := range rawArgs {
		var dereferenceErr error
		args[i], dereferenceErr = DereferenceEnvVar(arg)
		if dereferenceErr != nil {
			o.WriteCanonicalError("The value for argument %q cannot be used: %v", arg, dereferenceErr)
			o.Log(output.Error, "argument cannot be used", map[string]any{
				"value": arg,
				"error": dereferenceErr,
			})
			processed = false
		}
	}
	if processed {
		f.SetOutput(o.ErrorWriter())
		// note: Parse outputs errors to o.ErrorWriter*()
		if parseErr := f.Parse(args); parseErr != nil {
			o.Log(output.Error, parseErr.Error(), map[string]any{"arguments": args})
			processed = false
		}
	}
	return
}
