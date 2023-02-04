package cmd_toolkit

import (
	"runtime/debug"
	"time"

	"github.com/majohn-r/output"
)

func Execute(o output.Bus, logInit func(output.Bus) bool, readBuildInfo func() (*debug.BuildInfo, bool), appName, appVersion, buildTimestamp string, cmdLine []string) (exitCode int) {
	start := time.Now()
	exitCode = 1
	if err := SetAppName(appName); err != nil {
		o.WriteCanonicalError("A programming error has occurred - %v", err)
		return
	}
	if logInit(o) && InitApplicationPath(o) {
		// parse build data
		InitBuildData(readBuildInfo, appVersion, buildTimestamp)
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
