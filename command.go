package cmd_toolkit

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/majohn-r/output"
)

type CommandDescription struct {
	IsDefault   bool
	Initializer func(output.Bus, *Configuration, *flag.FlagSet) (CommandProcessor, bool)
}

var descriptions = map[string]*CommandDescription{}

// CommandProcessor defines the functions needed to run a command
type CommandProcessor interface {
	Exec(output.Bus, []string) bool
}

func AddCommandData(name string, d *CommandDescription) {
	descriptions[name] = d
}

// LogCommandStart logs the beginning of command execution; if the provided map
// contains a "command" value, then the name parameter is ignored
func LogCommandStart(o output.Bus, name string, m map[string]any) {
	if m == nil {
		m = map[string]any{}
	}
	if _, ok := m["command"]; !ok {
		m["command"] = name
	}
	o.Log(output.Info, "executing command", m)
}

// ProcessCommand selects which command to be run and returns the relevant
// CommandProcessor, command line arguments and ok status
func ProcessCommand(o output.Bus, args []string) (cmd CommandProcessor, cmdArgs []string, ok bool) {
	var c *Configuration
	if c, ok = ReadConfigurationFile(o); !ok {
		return
	}
	var defaultCmd string
	if defaultCmd, ok = determineDefaultCommand(o, c.SubConfiguration("command")); ok {
		cmd, cmdArgs, ok = selectCommand(o, defaultCmd, c, args)
	}
	return
}

// ReportNothingToDo reports a user error in which a command's parameter values
// prevent the command from doing any work; the report is made to error output
// and to the log
func ReportNothingToDo(o output.Bus, cmd string, fields map[string]any) {
	o.WriteCanonicalError("You disabled all functionality for the command %q", cmd)
	o.Log(output.Error, "the user disabled all functionality", fields)
}

func determineDefaultCommand(o output.Bus, c *Configuration) (defaultCommand string, ok bool) {
	// get the default command name from configuration, if it's defined
	defaultCommand, ok = c.StringValue("default")
	if ok { // there's a default command defined
		_, ok = descriptions[defaultCommand]
		if !ok {
			o.Log(output.Error, "invalid default command", map[string]any{"command": defaultCommand})
			o.WriteCanonicalError("The configuration file specifies %q as the default command. There is no such command", defaultCommand)
			defaultCommand = ""
		}
		return
	}
	switch len(descriptions) {
	case 0:
		o.Log(output.Error, "no commands registered", nil)
		o.WriteCanonicalError("A programming error has occurred - there are no commands registered!")
		return
	case 1:
		for name := range descriptions {
			defaultCommand = name
			ok = true
			return
		}
	default:
		// common case: there is more than 1 command defined
		var defaultCmds []string
		for name, d := range descriptions {
			if d.IsDefault {
				defaultCmds = append(defaultCmds, name)
			}
		}
		switch len(defaultCmds) {
		case 0:
			o.Log(output.Error, "No default command", map[string]any{"commands": describedCommandNames("")})
			o.WriteCanonicalError("A programming error has occurred - none of the defined commands is defined as the default command.")
			ok = false
		case 1:
			defaultCommand = defaultCmds[0]
			ok = true
		default:
			sort.Strings(defaultCmds)
			o.WriteCanonicalError("A programming error has occurred - multiple commands (%v) are defined as default commands.", defaultCmds)
			o.Log(output.Error, "multiple default commands", map[string]any{"commands": defaultCmds})
			ok = false
		}
	}
	return
}

func describedCommandNames(defaultCommand string) []string {
	var names []string
	for name := range descriptions {
		if name == defaultCommand {
			names = append(names, fmt.Sprintf("%s (default)", name))
		} else {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func selectCommand(o output.Bus, defaultCmd string, c *Configuration, args []string) (cmd CommandProcessor, cmdArgs []string, ok bool) {
	m := make(map[string]CommandProcessor)
	allCmdsOk := true
	for name, description := range descriptions {
		fSet := flag.NewFlagSet(name, flag.ContinueOnError)
		cmd, cOk := description.Initializer(o, c, fSet)
		if cOk {
			m[name] = cmd
		} else {
			allCmdsOk = false
		}
	}
	if !allCmdsOk {
		return
	}
	if len(args) < 2 {
		// no arguments at all
		cmd = m[defaultCmd]
		cmdArgs = []string{}
		ok = true
		return
	}
	firstArg := args[1]
	if strings.HasPrefix(firstArg, "-") {
		// first argument is a flag
		cmd = m[defaultCmd]
		cmdArgs = args[1:]
		ok = true
		return
	}
	var found bool
	cmd, found = m[firstArg]
	if !found {
		cmd = nil
		cmdArgs = nil
		names := describedCommandNames(defaultCmd)
		o.Log(output.Error, "unrecognized command", map[string]any{"command": firstArg, "commands": names})
		o.WriteCanonicalError("There is no command named %q; valid commands include [%s]", firstArg, strings.Join(names, ", "))
		return
	}
	cmdArgs = args[2:]
	ok = true
	return
}
