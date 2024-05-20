package cmd_toolkit

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAppName(t *testing.T) {
	savedAppname := appname
	defer func() {
		appname = savedAppname
	}()
	tests := map[string]struct {
		appname string
		want    string
		wantErr bool
	}{
		"get empty value":     {wantErr: true},
		"get non-empty value": {appname: "myApp", want: "myApp"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			appname = tt.appname
			got, gotErr := AppName()
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("AppName() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AppName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateAppSpecificPath(t *testing.T) {
	savedAppname := appname
	defer func() {
		appname = savedAppname
	}()
	type args struct {
		topDir string
	}
	tests := map[string]struct {
		appname string
		args
		want    string
		wantErr bool
	}{
		"uninitialized appname": {wantErr: true},
		"initialized appname":   {appname: "myApp", args: args{topDir: "dir"}, want: filepath.Join("dir", "myApp")},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			appname = tt.appname
			got, gotErr := CreateAppSpecificPath(tt.args.topDir)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("CreateAppSpecificPath() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateAppSpecificPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDereferenceEnvVar(t *testing.T) {
	type args struct {
		s string
	}
	tests := map[string]struct {
		varSettings map[string]string
		args
		want    string
		wantErr bool
	}{
		"no references": {args: args{s: "no references here"}, want: "no references here"},
		"many references": {
			varSettings: map[string]string{
				"VAR1":     "firstVar",
				"VAR1USER": "secondVar",
				"VAR2":     "thirdVar",
			},
			args: args{s: "$VAR1 $VAR1USER $VAR2 $VAR2, %VAR1% %VAR1USER% %VAR2%"},
			want: "firstVar secondVar thirdVar thirdVar, firstVar secondVar thirdVar",
		},
		"missing references": {
			varSettings: map[string]string{
				"VAR1":     "firstVar",
				"VAR1USER": "secondVar",
				"VAR2":     "thirdVar",
			},
			args:    args{s: "$VAR1 $VAR1USER $VAR2 $VAR2 $VAR3, %VAR1% %VAR1USER% %VAR2% %VAR3%"},
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mementos := []*EnvVarMemento{}
			for varName, varValue := range tt.varSettings {
				mementos = append(mementos, NewEnvVarMemento(varName))
				if varValue == "" {
					os.Unsetenv(varName)
				} else {
					os.Setenv(varName, varValue)
				}
			}
			defer func() {
				for _, memento := range mementos {
					memento.Restore()
				}
			}()
			got, gotErr := DereferenceEnvVar(tt.args.s)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("DereferenceEnvVar() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DereferenceEnvVar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewEnvVarMemento(t *testing.T) {
	const varName = "VAR1"
	savedValue, savedSet := os.LookupEnv(varName)
	defer func() {
		if savedSet {
			os.Setenv(varName, savedValue)
		} else {
			os.Unsetenv(varName)
		}
	}()
	type args struct {
		name string
	}
	tests := map[string]struct {
		value string
		set   bool
		args
		want *EnvVarMemento
	}{
		"set":   {value: "the value", set: true, args: args{name: varName}, want: &EnvVarMemento{name: varName, value: "the value", set: true}},
		"unset": {args: args{name: varName}, want: &EnvVarMemento{name: varName}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.set {
				os.Setenv(varName, tt.value)
			} else {
				os.Unsetenv(varName)
			}
			if got := NewEnvVarMemento(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEnvVarMemento() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetAppName(t *testing.T) {
	savedAppname := appname
	defer func() {
		appname = savedAppname
	}()
	type args struct {
		s string
	}
	tests := map[string]struct {
		appname string
		args
		wantErr bool
	}{
		"unset, set to empty":         {args: args{}, wantErr: true},
		"unset, set to non-empty":     {args: args{s: "myApp"}},
		"set, set to same value":      {appname: "myApp", args: args{s: "myApp"}},
		"set, set to different value": {appname: "myApp", args: args{s: "myOtherApp"}, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			appname = tt.appname
			if gotErr := SetAppName(tt.args.s); (gotErr != nil) != tt.wantErr {
				t.Errorf("SetAppName() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func Test_findReferences(t *testing.T) {
	type args struct {
		s string
	}
	tests := map[string]struct {
		args
		want []string
	}{
		"no references": {
			args: args{s: "no references here, not even this: %VAR1"},
			want: make([]string, 0),
		},
		"many references": {
			args: args{s: "$VAR1 $VAR11 $VAR111 $VAR1 %VAR2% %VAR22% %VAR222% %VAR222%"},
			want: []string{"$VAR111", "$VAR11", "$VAR1", "%VAR2%", "%VAR22%", "%VAR222%"}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := findReferences(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findReferences() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_byLength_Len(t *testing.T) {
	tests := map[string]struct {
		bl   byLength
		want int
	}{
		"empty":  {want: 0},
		"plenty": {bl: byLength([]string{"a", "b", "c"}), want: 3},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.bl.Len(); got != tt.want {
				t.Errorf("byLength.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_byLength_Less(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := map[string]struct {
		bl byLength
		args
		want bool
	}{
		"same":    {bl: byLength([]string{"$VAR1", "$VAR1"}), args: args{i: 0, j: 1}, want: false},
		"shorter": {bl: byLength([]string{"$VAR1", "$VAR11"}), args: args{i: 0, j: 1}, want: false},
		"longer":  {bl: byLength([]string{"$VAR11", "$VAR1"}), args: args{i: 0, j: 1}, want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.bl.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("byLength.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_byLength_Swap(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := map[string]struct {
		bl byLength
		args
		wantBl byLength
	}{
		"same":      {bl: byLength([]string{"$VAR1", "$VAR1"}), args: args{i: 0, j: 1}, wantBl: byLength([]string{"$VAR1", "$VAR1"})},
		"different": {bl: byLength([]string{"$VAR1", "$VAR11"}), args: args{i: 0, j: 1}, wantBl: byLength([]string{"$VAR11", "$VAR1"})},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.bl.Swap(tt.args.i, tt.args.j)
			if !reflect.DeepEqual(tt.bl, tt.wantBl) {
				t.Errorf("byLength.Swap = %v, want %v", tt.bl, tt.wantBl)
			}
		})
	}
}

func TestEnvVarMemento_Restore(t *testing.T) {
	const varName = "VAR1"
	savedValue, savedSet := os.LookupEnv(varName)
	defer func() {
		if savedSet {
			os.Setenv(varName, savedValue)
		} else {
			os.Unsetenv(varName)
		}
	}()
	tests := map[string]struct {
		preValue  string
		preSet    bool
		mem       *EnvVarMemento
		wantValue string
		wantSet   bool
	}{
		"set":             {mem: &EnvVarMemento{name: varName, value: "the value", set: true}, wantValue: "the value", wantSet: true},
		"unset":           {preValue: "the value", preSet: true, mem: &EnvVarMemento{name: varName}},
		"overwrite":       {preValue: "old value", preSet: true, mem: &EnvVarMemento{name: varName, value: "the value", set: true}, wantValue: "the value", wantSet: true},
		"redundant clear": {mem: &EnvVarMemento{name: varName}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.preSet {
				os.Setenv(varName, tt.preValue)
			} else {
				os.Unsetenv(varName)
			}
			tt.mem.Restore()
			if gotValue, gotSet := os.LookupEnv(varName); gotValue != tt.wantValue || gotSet != tt.wantSet {
				t.Errorf("EnvVarMemento.Restore = (%q, %t) want (%q, %t)", gotValue, gotSet, tt.wantValue, tt.wantSet)
			}
		})
	}
}
