package cmd_toolkit_test

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"golang.org/x/sys/windows"
)

func TestNewElevationControl(t *testing.T) {
	originalGetCurrentProcessToken := cmdtoolkit.GetCurrentProcessToken
	originalIsElevated := cmdtoolkit.IsElevated
	originalIsTerminal := cmdtoolkit.IsTerminal
	originalIsCygwinTerminal := cmdtoolkit.IsCygwinTerminal
	defer func() {
		cmdtoolkit.GetCurrentProcessToken = originalGetCurrentProcessToken
		cmdtoolkit.IsElevated = originalIsElevated
		cmdtoolkit.IsTerminal = originalIsTerminal
		cmdtoolkit.IsCygwinTerminal = originalIsCygwinTerminal
	}()
	cmdtoolkit.GetCurrentProcessToken = func() (t windows.Token) {
		return
	}
	tests := map[string]struct {
		assertElevated   bool
		assertRedirected bool
		wantStatus       []string
		output.WantedRecording
	}{
		"neither elevated nor redirected": {
			assertElevated:   false,
			assertRedirected: false,
			wantStatus: []string{
				"myApp is not running with elevated privileges",
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='false'" +
					" environment_variable=''" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"not elevated, but redirected": {
			assertElevated:   false,
			assertRedirected: true,
			wantStatus: []string{
				"myApp is not running with elevated privileges",
				"stderr, stdin, and stdout have been redirected",
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='false'" +
					" environment_variable=''" +
					" stderr_redirected='true'" +
					" stdin_redirected='true'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"elevated, but not redirected": {
			assertElevated:   true,
			assertRedirected: false,
			wantStatus: []string{
				"myApp is running with elevated privileges",
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='true'" +
					" environment_variable=''" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"elevated and redirected": {
			assertElevated:   true,
			assertRedirected: true,
			wantStatus: []string{
				"myApp is running with elevated privileges",
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='true'" +
					" environment_variable=''" +
					" stderr_redirected='true'" +
					" stdin_redirected='true'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmdtoolkit.IsElevated = func(_ windows.Token) bool { return tt.assertElevated }
			cmdtoolkit.IsTerminal = func(_ uintptr) bool { return !tt.assertRedirected }
			cmdtoolkit.IsCygwinTerminal = cmdtoolkit.IsTerminal
			ec := cmdtoolkit.NewElevationControl()
			o := output.NewRecorder()
			ec.Log(o, output.Info)
			if got := ec.Status("myApp"); !reflect.DeepEqual(got, tt.wantStatus) {
				t.Errorf("NewElevationControl().Status() got: %v\nwant: %v", got, tt.wantStatus)
			}
			o.Report(t, "NewElevationControl()", tt.WantedRecording)
		})
	}
}

func TestNewElevationControlWithEnvVar(t *testing.T) {
	const envVarName = "MYAPP_RUNS_AS_ADMIN"
	envVarMemento := cmdtoolkit.NewEnvVarMemento(envVarName)
	originalGetCurrentProcessToken := cmdtoolkit.GetCurrentProcessToken
	originalIsElevated := cmdtoolkit.IsElevated
	originalIsTerminal := cmdtoolkit.IsTerminal
	originalIsCygwinTerminal := cmdtoolkit.IsCygwinTerminal
	defer func() {
		envVarMemento.Restore()
		cmdtoolkit.GetCurrentProcessToken = originalGetCurrentProcessToken
		cmdtoolkit.IsElevated = originalIsElevated
		cmdtoolkit.IsTerminal = originalIsTerminal
		cmdtoolkit.IsCygwinTerminal = originalIsCygwinTerminal
	}()
	_ = os.Unsetenv(envVarName)
	cmdtoolkit.GetCurrentProcessToken = func() (t windows.Token) {
		return
	}
	tests := map[string]struct {
		defaultEnvVarValue bool
		assertElevated     bool
		assertRedirected   bool
		wantStatus         []string
		output.WantedRecording
	}{
		"env var false, neither elevated nor redirected": {
			defaultEnvVarValue: false,
			assertElevated:     false,
			assertRedirected:   false,
			wantStatus: []string{
				"myApp is not running with elevated privileges",
				"The environment variable MYAPP_RUNS_AS_ADMIN evaluates as false",
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='false'" +
					" environment_variable='MYAPP_RUNS_AS_ADMIN'" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"env var false, not elevated, but redirected": {
			defaultEnvVarValue: false,
			assertElevated:     false,
			assertRedirected:   true,
			wantStatus: []string{
				"myApp is not running with elevated privileges",
				"stderr, stdin, and stdout have been redirected",
				"The environment variable MYAPP_RUNS_AS_ADMIN evaluates as false",
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='false'" +
					" environment_variable='MYAPP_RUNS_AS_ADMIN'" +
					" stderr_redirected='true'" +
					" stdin_redirected='true'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"env var false, elevated, but not redirected": {
			defaultEnvVarValue: false,
			assertElevated:     true,
			assertRedirected:   false,
			wantStatus: []string{
				"myApp is running with elevated privileges",
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='true'" +
					" environment_variable='MYAPP_RUNS_AS_ADMIN'" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"env var false, elevated and redirected": {
			defaultEnvVarValue: false,
			assertElevated:     true,
			assertRedirected:   true,
			wantStatus: []string{
				"myApp is running with elevated privileges",
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='true'" +
					" environment_variable='MYAPP_RUNS_AS_ADMIN'" +
					" stderr_redirected='true'" +
					" stdin_redirected='true'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"env var true, neither elevated nor redirected": {
			defaultEnvVarValue: true,
			assertElevated:     false,
			assertRedirected:   false,
			wantStatus: []string{
				"myApp is not running with elevated privileges",
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='false'" +
					" environment_variable='MYAPP_RUNS_AS_ADMIN'" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"env var true, not elevated, but redirected": {
			defaultEnvVarValue: true,
			assertElevated:     false,
			assertRedirected:   true,
			wantStatus: []string{
				"myApp is not running with elevated privileges",
				"stderr, stdin, and stdout have been redirected",
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='false'" +
					" environment_variable='MYAPP_RUNS_AS_ADMIN'" +
					" stderr_redirected='true'" +
					" stdin_redirected='true'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"env var true, elevated, but not redirected": {
			defaultEnvVarValue: true,
			assertElevated:     true,
			assertRedirected:   false,
			wantStatus: []string{
				"myApp is running with elevated privileges",
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='true'" +
					" environment_variable='MYAPP_RUNS_AS_ADMIN'" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"env var true, elevated and redirected": {
			defaultEnvVarValue: true,
			assertElevated:     true,
			assertRedirected:   true,
			wantStatus: []string{
				"myApp is running with elevated privileges",
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='true'" +
					" environment_variable='MYAPP_RUNS_AS_ADMIN'" +
					" stderr_redirected='true'" +
					" stdin_redirected='true'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmdtoolkit.IsElevated = func(_ windows.Token) bool { return tt.assertElevated }
			cmdtoolkit.IsTerminal = func(_ uintptr) bool { return !tt.assertRedirected }
			cmdtoolkit.IsCygwinTerminal = cmdtoolkit.IsTerminal
			o := output.NewRecorder()
			ec := cmdtoolkit.NewElevationControlWithEnvVar(envVarName, tt.defaultEnvVarValue)
			ec.Log(o, output.Info)
			if got := ec.Status("myApp"); !reflect.DeepEqual(got, tt.wantStatus) {
				t.Errorf("NewElevationControlWithEnvVar().Status() got: %v\nwant: %v", got, tt.wantStatus)
			}
			o.Report(t, "NewElevationControlWithEnvVar()", tt.WantedRecording)
		})
	}
}

func TestElevationControl_WillRunElevated(t *testing.T) {
	originalGetCurrentProcessToken := cmdtoolkit.GetCurrentProcessToken
	originalIsElevated := cmdtoolkit.IsElevated
	originalIsTerminal := cmdtoolkit.IsTerminal
	originalIsCygwinTerminal := cmdtoolkit.IsCygwinTerminal
	originalShellExecute := cmdtoolkit.ShellExecute
	defer func() {
		cmdtoolkit.GetCurrentProcessToken = originalGetCurrentProcessToken
		cmdtoolkit.IsElevated = originalIsElevated
		cmdtoolkit.IsTerminal = originalIsTerminal
		cmdtoolkit.IsCygwinTerminal = originalIsCygwinTerminal
		cmdtoolkit.ShellExecute = originalShellExecute
	}()
	cmdtoolkit.GetCurrentProcessToken = func() (t windows.Token) {
		return
	}
	tests := map[string]struct {
		assertCanElevate         bool
		assertShellExecuteResult bool
		want                     bool
	}{
		"cannot elevate": {
			assertCanElevate:         false,
			assertShellExecuteResult: true, // shell shouldn't execute anyway
			want:                     false,
		},
		"can elevate, user declines": {
			assertCanElevate:         true,
			assertShellExecuteResult: false,
			want:                     false,
		},
		"can elevate, user accepts": {
			assertCanElevate:         true,
			assertShellExecuteResult: true,
			want:                     true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmdtoolkit.IsElevated = func(_ windows.Token) bool { return !tt.assertCanElevate }
			cmdtoolkit.IsTerminal = func(_ uintptr) bool { return tt.assertCanElevate }
			cmdtoolkit.IsCygwinTerminal = cmdtoolkit.IsTerminal
			cmdtoolkit.ShellExecute = func(_ windows.Handle, _, _, _, _ *uint16, _ int32) error {
				if tt.assertShellExecuteResult {
					return nil
				}
				return fmt.Errorf("user declined")
			}
			ec := cmdtoolkit.NewElevationControl()
			if got := ec.WillRunElevated(); got != tt.want {
				t.Errorf("WillRunElevated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessIsElevated(t *testing.T) {
	originalGetCurrentProcessToken := cmdtoolkit.GetCurrentProcessToken
	originalIsElevatedFunc := cmdtoolkit.IsElevated
	defer func() {
		cmdtoolkit.GetCurrentProcessToken = originalGetCurrentProcessToken
		cmdtoolkit.IsElevated = originalIsElevatedFunc
	}()
	cmdtoolkit.GetCurrentProcessToken = func() (t windows.Token) {
		return
	}
	tests := map[string]struct {
		want bool
	}{
		"no":  {want: false},
		"yes": {want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmdtoolkit.IsElevated = func(_ windows.Token) bool { return tt.want }
			if got := cmdtoolkit.ProcessIsElevated(); got != tt.want {
				t.Errorf("processIsElevated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_elevationControl_AttemptRunElevated(t *testing.T) {
	originalGetCurrentProcessToken := cmdtoolkit.GetCurrentProcessToken
	originalIsElevated := cmdtoolkit.IsElevated
	originalIsTerminal := cmdtoolkit.IsTerminal
	originalIsCygwinTerminal := cmdtoolkit.IsCygwinTerminal
	originalShellExecute := cmdtoolkit.ShellExecute
	defer func() {
		cmdtoolkit.GetCurrentProcessToken = originalGetCurrentProcessToken
		cmdtoolkit.IsElevated = originalIsElevated
		cmdtoolkit.IsTerminal = originalIsTerminal
		cmdtoolkit.IsCygwinTerminal = originalIsCygwinTerminal
		cmdtoolkit.ShellExecute = originalShellExecute
	}()
	cmdtoolkit.GetCurrentProcessToken = func() (t windows.Token) {
		return
	}
	rejectionError := fmt.Errorf("user declined")
	tests := map[string]struct {
		assertCanElevate         bool
		assertShellExecuteResult bool
		wantStatus               bool
		wantError                error
	}{
		"cannot elevate": {
			assertCanElevate:         false,
			assertShellExecuteResult: true, // shell shouldn't execute anyway
			wantStatus:               false,
			wantError:                &cmdtoolkit.ElevationNotAttempted{},
		},
		"can elevate, user declines": {
			assertCanElevate:         true,
			assertShellExecuteResult: false,
			wantStatus:               false,
			wantError:                rejectionError,
		},
		"can elevate, user accepts": {
			assertCanElevate:         true,
			assertShellExecuteResult: true,
			wantStatus:               true,
			wantError:                nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmdtoolkit.IsElevated = func(_ windows.Token) bool { return !tt.assertCanElevate }
			cmdtoolkit.IsTerminal = func(_ uintptr) bool { return tt.assertCanElevate }
			cmdtoolkit.IsCygwinTerminal = cmdtoolkit.IsTerminal
			cmdtoolkit.ShellExecute = func(_ windows.Handle, _, _, _, _ *uint16, _ int32) error {
				if tt.assertShellExecuteResult {
					return nil
				}
				return rejectionError
			}
			ec := cmdtoolkit.NewElevationControl()
			gotError, gotStatus := ec.AttemptRunElevated()
			if gotStatus != tt.wantStatus || !errors.Is(gotError, tt.wantError) {
				t.Errorf("AttemptRunElevated() = %q:%v, want %q:%v", gotError, gotStatus, tt.wantError,
					tt.wantStatus)
			}
		})
	}
}

func TestElevationNotAttempted_Error(t *testing.T) {
	tests := map[string]struct {
		want string
	}{
		"test case": {want: "elevation is not possible"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ei := &cmdtoolkit.ElevationNotAttempted{}
			if got := ei.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
