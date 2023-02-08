package cmd_toolkit

import (
	"time"

	"github.com/majohn-r/output"
)

var (
	logInitializer func(output.Bus) bool = InitLogging
)

func Execute(o output.Bus, firstYear int, appName, appVersion, buildTimestamp string, cmdLine []string) (exitCode int) {
	start := time.Now()
	setFirstYear(firstYear)
	exitCode = 1
	if err := SetAppName(appName); err != nil {
		o.WriteCanonicalError("A programming error has occurred - %v", err)
		return
	}
	if logInitializer(o) && InitApplicationPath(o) {
		// parse build data
		InitBuildData(appVersion, buildTimestamp)
		o.Log(output.Info, "execution starts", map[string]any{
			"version":      appVersion,
			"timeStamp":    buildTimestamp,
			"goVersion":    GoVersion(),
			"dependencies": BuildDependencies(),
			"args":         cmdLine,
		})
		if cmd, args, ok := ProcessCommand(o, cmdLine); ok {
			if cmd.Exec(o, args) {
				exitCode = 0
			}
		}
		o.Log(output.Info, "execution ends", map[string]any{
			"duration": time.Since(start),
			"exitCode": exitCode,
		})
	}
	if exitCode != 0 {
		o.WriteCanonicalError("%q version %s, created at %s, failed", appName, appVersion, buildTimestamp)
	}
	return
}
