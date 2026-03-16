# Changelog

This project uses [semantic versioning](https://semver.org/); be aware that, until the major version becomes non-zero,
[this proviso](https://semver.org/#spec-item-4) applies.

That said, the period of time (late June 2024) when this project was pivoting from its own home-grown command
processing functionality to supporting applications using `github.com/spf13/cobra` instead involved an awful lot of
breaking changes. Oops. Should have handled that differently.

Key to symbols

- ❗ breaking change
- 🐛 bug fix
- ⚠️ change in behavior, may surprise the user
- 😒 change is invisible to the user
- 🆕 new feature

## v0.31.2

_release `2026-03-16`_

- 🐛 marked deprecated `InterpretBuildData` function with `go:fix inline` comment

## v0.31.1

_ release `2026-02-18`_

- 😒 accept changes made by go fix and update to go 1.26; update dependencies

## v0.31.0

_release `2026-02-03`_

- ⚠️ fix for [issue 54](https://github.com/majohn-r/cmd-toolkit/issues/54): add adverse `ZoneId` detection in an
alternate data stream

## v0.30.0

_release `2026-02-02`_

- ⚠️ fix for [issue 53](https://github.com/majohn-r/cmd-toolkit/issues/53): add `AttemptRunElevated` method to
`ElevationControl` interface.

## v0.29.4

_release `2026-01-26`_

- 😒 update dependencies

## v0.29.3

_release `2025-10-30`_

- 🐛⚠️ fix for [issue 52](https://github.com/majohn-r/cmd-toolkit/issues/52): use `xdg`'s `STATE_HOME` as the base
directory for the log file directory instead of either the `TMP` or `TEMP` environment variable values.
- 🐛 cleaned up application directory code and corrected some error output to reflect that the typical value of `xdg`'s
`CONFIG_HOME` is `%HOMEPATH%\AppData\Local`, not `%HOMEPATH%\AppData\Roaming`

## v0.29.2

_release `2025-10-30`_

- 🐛 fix for [issue 52](https://github.com/majohn-r/cmd-toolkit/issues/52): use `xdg`'s `CONFIG_HOME` as the base
directory for the configuration defaults file (aka application directory).

## v0.29.1

_release `2025-10-28`_

- 😒 update build dependencies.

## v0.29.0

_release `2025-10-28`_

- 🆕🐛 fix for [issue 52](https://github.com/majohn-r/cmd-toolkit/issues/52): introducing the `AppName` function.

## v0.28.0

_release `2025-10-26`_

- 🆕🐛 first step in fixing [issue 51](https://github.com/majohn-r/cmd-toolkit/issues/51): introducing the `StateFile`
interface.

## v0.27.2

_release `2025-10-09`_

- 😒 updated dependencies.

## v0.27.1

_release `2025-09-01`_

- 😒 updated dependency: **github.com/spf13/pflag** v1.0.8 -> v1.0.9. v1.0.8 badly broke a lot of consumers.

## v0.27.0

_release `2025-08-31`_

- 🐛 [issue 49](https://github.com/majohn-r/cmd-toolkit/issues/49) fixed: published **(*IntBounds).ConstrainedValue()**.
- 😒 updated dependency: **github.com/spf13/pflag** v1.0.7 -> v1.0.8.
- 😒 fixed IDE complaints (unsorted imports).

## v0.26.3

_release `2025-08-29`_

- 😒 update build and code dependencies, as well as migrating to go 1.25

## v0.26.2

_release `2025-07-20`_

- 😒 update build and code dependencies.

## v0.26.1

_release `2025-03-10`_

- 🐛 remove unreachable `ConfigureExit` function and associated test.

## v0.26.0

_release `2025-03-08`_

- ❗remove `ConfigureExit` from `ElevationControl` interface.

## v0.25.0

_release `2025-02-19`_

- 🆕 add `ModificationTime` function to return a specified file's modification time.

## v0.24.1

_release `2024-10-28`_

- 🐛⚠️ change `BuildInformation` from a `struct` to an `interface` for easier mocking by consumers. Consequently,
`GetBuildInfo` returns an _implementation_ of `BuildInformation` instead of a pointer to an instance of
`BuildInformation`.

## v0.24.0

_release `2024-10-28`_

- 🆕add `GetBuildData(reader func() (*debug.BuildInfo, bool)) *BuildInformation`; this also adds the `BuildInformation`
struct and four methods on `*BuildInformation`:
 1. `GoVersion() string`
 2. `Dependencies() []string`
 3. `MainVersion() string`
 4. `Settings() []string`
- ⚠️deprecate `func InterpretBuildData(buildInfoReader func() (*debug.BuildInfo, bool)) (goVersion string, dependencies
[]string)`

## v0.23.0

_release `2024-08-30`_

- 🆕⚠️add `ErrorToString(e Error) string` function; not a breaking change, exactly, but this project uses the new
function, and that changes error output in ways that may surprise consumers, potentially breaking unit tests.

## v0.22.2

_release `2024-07-17`_

- 🐛fix logging bug where seemingly random amounts of white space was placed at the beginning of each record

## v0.22.1

_release `2024-07-16`_

- ❗rename `AsPayload()` function to `WritableDefaults()`

## v0.22.0

_release `2024-07-15`_

- 🆕add `AddDefaults(sf *FlagSet)` function
- 🆕add `AsPayload() []byte` function

## v0.21.1

_release `2024-07-15`_

## v0.21.0

- ❗un-publish `DefaultConfigFileName` constant
- 🆕add `DefaultConfigFileStatus() (string, bool)` function

_release `2024-07-15`_

- ❗delete `DefaultConfigFileName() string` function
- ❗delete `UnsafeSetApplicationPath(path string)` function
- ❗delete `UnsafeSetDefaultConfigFileName(newConfigFileName string)` function
- 🆕publish `DefaultConfigFileName` constant

## v0.20.0

_release `2024-07-10`_

- ❗delete `FlagIndicator() string` function
- ❗delete `FlowerBox(lines []string) []string` function
- ❗delete `SetFlagIndicator(val string)` function
- ❗delete `FlagConsumer interface`
- ❗un-publish `CreateAppSpecificPath(topDir, applicationName string) (string, error)` function
- ❗change signature from `AddFlags(o output.Bus, c *Configuration, flags FlagConsumer, sets ...*FlagSet)` to
`AddFlags(o output.Bus, c *Configuration, flags *pflag.FlagSet, sets ...*FlagSet)`

## v0.19.0

_release `2024-07-10`_

- ❗un-publish `DecorateBoolFlagUsage(usage string, defaultValue bool) string` function
- ❗un-publish `DecorateIntFlagUsage(usage string, defaultValue int) string` function
- ❗un-publish `DecorateStringFlagUsage(usage, defaultValue string) string` function
- ❗un-publish `ReportInvalidConfigurationData(o output.Bus, s string, e error)` function
- 🆕add constants `BoolType`, `IntType`, and `StringType`
- 🆕add `CommandFlag[V commandFlagValue] struct`
- 🆕add `FlagDetails struct`
- 🆕add `FlagSet struct`
- 🆕add `FlagConsumer interface`
- 🆕add `FlagProducer interface`
- 🆕add `(fD *FlagDetails) Copy() *FlagDetails` method
- 🆕add `AddFlags(o output.Bus, c *Configuration, flags FlagConsumer, sets ...*FlagSet)` function
- 🆕add `GetBool(o output.Bus, results map[string]*CommandFlag[any], flagName string) (CommandFlag[bool], error)`
function
- 🆕add `GetInt(o output.Bus, results map[string]*CommandFlag[any], flagName string) (CommandFlag[int], error)` function
- 🆕add `GetString(o output.Bus, results map[string]*CommandFlag[any], flagName string) (CommandFlag[string], error)`
function
- 🆕add `ProcessFlagErrors(o output.Bus, eSlice []error) bool` function
- 🆕add `ReadFlags(producer FlagProducer, set *FlagSet) (map[string]*CommandFlag[any], []error)` function

## v0.18.0

_release `2024-07-09`_

- 🆕add `ExitError struct`
- 🆕add `NewExitUserError(cmd string) *ExitError` function
- 🆕add `NewExitProgrammingError(cmd string) *ExitError` function
- 🆕add `NewExitSystemError(cmd string) *ExitError` function
- 🆕add `ToErrorInterface(e *ExitError) error` function
- 🆕add `(e *ExitError) Error() string` method
- 🆕add `(e *ExitError) Status() int` method

## v0.17.0

_release `2024-07-07`_

- 🆕add `type FlowerBoxStyle`
- 🆕add constants `ASCIIFlowerBox`, `CurvedFlowerBox`, `DoubleLinedFlowerBox`, `HeavyLinedFlowerBox`, and
`LightLinedFlowerBox`
- 🆕add `StyledFlowerBox(lines []string, style FlowerBoxStyle) []string` function

## v0.16.2

_release `2024-07-06`_

- ❗add `ElevationControl interface`
- ❗change signature from `NewElevationControl() *ElevationControl` to `NewElevationControl() ElevationControl`
- ❗change signature from `NewElevationControlWithEnvVar(envVarName string, defaultEnvVarValue bool) *ElevationControl`
to `NewElevationControlWithEnvVar(envVarName string, defaultEnvVarValue bool) ElevationControl`
- ❗un-publish `ElevationControl struct`

## v0.16.1

_release `2024-07-05`_

- 🆕publish `ProcessIsElevated() bool` function

## v0.16.0

_release `2024-07-05`_

- 🆕add `ElevationControl struct`
- 🆕add `NewElevationControl() *ElevationControl` function
- 🆕add `NewElevationControlWithEnvVar(envVarName string, defaultEnvVarValue bool) *ElevationControl` function
- 🆕add `(ec *ElevationControl) ConfigureExit(oldExitFn func(int)) func(int)` method
- 🆕add `(ec *ElevationControl) Log(o output.Bus, level output.Level)` method
- 🆕add `(ec *ElevationControl) Status(appName string) []string` method
- 🆕add `(ec *ElevationControl) WillRunElevated() bool` method

## v0.15.0

_release `2024-07-03`_

- 🐛improve error reporting in log file initialization
- 🐛improve logging of errors in application path initialization
- 🐛improve logging of errors in reading the application configuration file

## v0.14.0

_release `2024-07-03`_

- ❗change signature from `CreateAppSpecificPath(topDir string) (string, error)` to
`CreateAppSpecificPath(topDir, applicationName string) (string, error)`
- ❗change signature from `InitApplicationPath(o output.Bus) bool` to
`InitApplicationPath(o output.Bus, applicationName string) bool`
- ❗change signature from `InitLogging(o output.Bus) (ok bool)` to
`InitLogging(o output.Bus, applicationName string) (ok bool)`
- ❗change signature from `InitLoggingWithLevel(o output.Bus, l output.Level) (ok bool)` to
`InitLoggingWithLevel(o output.Bus, l output.Level, applicationName string) (ok bool)`
- ❗delete `AppName() (string, error)` function
- ❗delete `SetAppName(s string) error` function
- ❗delete `UnsafeAppName() string` function
- ❗delete `UnsafeSetAppName(name string)` function

## v0.13.1

_release `2004-07-02`_

- 🐛improve log file initialization logic

## v0.13.0

_release `2004-07-01`_

- 🆕add `InterpretBuildData(buildInfoReader func() (*debug.BuildInfo, bool)) (goVersion string, dependencies []string)`
function
- ❗delete `BuildDependencies() []string` function
- ❗delete `BuildInformationHeader() string` function
- ❗delete `GoVersion() string` function
- ❗delete `InitBuildData(version, creation string)` function
- ❗delete `InterpretBuildData() (goVersion string, dependencies []string)` function
- ❗delete `SetFirstYear(i int)` function
- ❗delete `(a *aboutCmd) Exec(o output.Bus, _ []string) (ok bool)` method
- ❗delete `(b *IntBounds) Default() int` method

## v0.12.2

_release `2024-06-30`_

- 🆕add `UnsafeAppName() string` function
- 🆕add `UnsafeSetApplicationPath(path string)` function
- 🆕add `UnsafeSetAppName(name string)` function
- 🆕add `UnsafeSetDefaultConfigFileName(newConfigFileName string)` function
- 🆕add `BoolMap`, `ConfigurationMap`, `IntMap`, and `StringMap` fields to `Configuration struct`
- 🆕add `DefaultValue`, `MaxValue`, and `MinValue` fields to `IntBounds struct`
- ⚠️deprecate `(b *IntBounds) Default() int` method
- 🆕publish `EnvVarMemento struct`
- 🆕publish `NewEnvVarMemento(name string) *EnvVarMemento` function
- 🆕publish `(mem *EnvVarMemento) Restore()` method

## v0.12.1

_release `2024-06-30`_

- ❗delete `(sl *simpleLogger) ExitFunc() exitFunc` method
- ❗delete `(sl *simpleLogger) SetExitFunc(f exitFunc)` method
- 🆕publish `(sl *simpleLogger) WillLog(l output.Level)` method

## v0.12.0

_release `2024-06-29`_

- ❗delete `AddCommandData(name string, d *CommandDescription)` function
- ❗delete `CommandDescription struct`
- ❗delete `CommandProcessor interface`
- ❗delete `CreateFile(fileName string, content []byte) error` function
- ❗delete `Execute(o output.Bus, firstYear int, appName, appVersion, buildTimestamp string, cmdLine []string)
(exitCode int)` function
- ❗delete `ReportNothingToDo(o output.Bus, cmd string, fields map[string]any)` function

## v0.11.7

_release `2024-06-29` ️_

- ⚠️deprecate `CreateFile(fileName string, content []byte) error` function
- ⚠️deprecate `Execute(o output.Bus, firstYear int, appName, appVersion, buildTimestamp string, cmdLine []string) 
(exitCode int)` function
- ⚠️deprecate `ReportNothingToDo(o output.Bus, cmd string, fields map[string]any)` function
- ⚠️deprecate `CommandProcessor interface`

## v0.11.6

_release `2024-06-29`_

- ❗delete `ReportDirectoryCreationFailure(o output.Bus, cmd, dir string, e error)` function
- ❗delete `ReportFileDeletionFailure(o output.Bus, file string, e error)` function
- ❗delete `SecureAbsolutePath(path string) string` function
- ❗delete `SetAuthor(s string)` function
- ❗delete `SetDefaultConfigFileName(s string)` function
- ❗un-publish `WriteDirectoryCreationError(o output.Bus, d string, e error)` function
- ❗un-publish `(mem *envVarMemento) Restore()` method
- ❗un-publish `(c *Configuration) StringValue(key string) (value string, found bool)` method

## v0.11.5

_release `2024-06-29`_

- ❗delete `ProcessArgs(o output.Bus, f *flag.FlagSet, rawArgs []string) (processed bool)` function
- ❗un-publish `NewConfiguration(o output.Bus, data map[string]any) *Configuration` function
- ❗un-publish `NewEnvVarMemento(name string) *envVarMemento` function
- ❗un-publish `ProcessCommand(o output.Bus, args []string) (cmd CommandProcessor, cmdArgs []string, processed bool)`
function

## v0.11.4

_release `2024-06-29`_

- ❗delete `(b *IntBounds) Maximum() int` method
- ❗delete `(b *IntBounds) Minimum() int` method
- ❗delete `LogUnreadableDirectory(o output.Bus, s string, e error)` function

## v0.11.3

_release `2024-06-29`_

- ❗delete `(c *Configuration) IntValue(key string) (value int, exists bool)` method

## v0.11.2

_release `2024-06-29`_

- ❗un-publish `EnvVarMemento` struct
- ❗un-publish `FlagIndicator() string` function
- ❗un-publish `GenerateAboutContent(o output.Bus)` function

## v0.11.1

_release `2024-06-29`_

- ❗un-publish `ApplicationDataEnvVarName` constant
- ❗un-publish `(c *Configuration) BooleanValue(key string) (value, exists bool)` method
- ❗delete `CreateFileInDirectory(dir, name string, content []byte) error` function

## v0.10.0

_release `2024-05-03`_

- 🆕add `AssignFileSystem(fs afero.Fs) afero.Fs` function
- 🆕add `FileSystem() afero.Fs` function

## v0.8.1

_release `2024-02-24`_

- 🐛fix bug in copyright generation

## v0.8.0

_release `2024-02-23`_

- 🆕add `Copyright(o output.Bus, first int, timestamp, owner string) string` function
- 🆕add `DecoratedAppName(applicationName, applicationVersion, timestamp string) string` function
- 🆕add `FlowerBox(lines []string) []string` function
- 🆕add `FormatGoVersion(version string) string` function

## v0.7.0

_release `2024-02-22`_

- 🆕add `LogPath() string` function

## v0.6.0

_release `2023-12-05`_

- 🆕add `IntBounds` accessor methods `Default() int`, `Maximum() int`, and `Minimum() int`

## v0.5.1

_release `2023-11-28`_

- 🐛fix intermittent panic at startup

## v0.5.0

_release `2023-11-26`_

- 🆕publish `GenerateAboutContent(o output.Bus)` function
- 🆕publish `SetFirstYear(i int)` function

## v0.4.0

_release `2023-11-23`_

- 🆕add `FlagIndicator() string` and `SetFlagIndicator(string)` functions

## v0.3.0

_release `2023-11-14`_

- ⚠️update to use output v0.3.0, which may impact consumers

## v0.2.2

_release `2023-11-08`_

- 🐛remove logrus dependency and all associated baggage

## v0.2.0

_release `2023-11-04`_

- 🆕add `InitLoggingWithLevel(o output.Bus, l output.Level) (ok bool)` function

## v0.1.0

_release `2023-02-08`_

- 🆕initial release