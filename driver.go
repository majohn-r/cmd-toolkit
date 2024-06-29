package cmd_toolkit

import (
	"time"

	"github.com/majohn-r/output"
)

var (
	logInitializer = InitLogging
)

// Deprecated: goes away when users switch to viper
func Execute(o output.Bus, firstYear int, appName, appVersion, buildTimestamp string, cmdLine []string) (exitCode int) {
	start := time.Now()
	SetFirstYear(firstYear)
	exitCode = 1
	if appNameInitErr := SetAppName(appName); appNameInitErr != nil {
		o.WriteCanonicalError("A programming error has occurred - %v", appNameInitErr)
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
		if cmd, args, processed := processCommand(o, cmdLine); processed && cmd != nil {
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
