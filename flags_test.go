package cmd_toolkit_test

import (
	"errors"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"reflect"
	"testing"
)

func TestFlagDetails_Copy(t *testing.T) {
	tests := map[string]struct {
		fD   *cmdtoolkit.FlagDetails
		want *cmdtoolkit.FlagDetails
	}{
		"bool": {
			fD: &cmdtoolkit.FlagDetails{
				AbbreviatedName: "f",
				Usage:           "a fine flag for playing",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    true,
			},
			want: &cmdtoolkit.FlagDetails{
				AbbreviatedName: "f",
				Usage:           "a fine flag for playing",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    true,
			},
		},
		"int": {
			fD: &cmdtoolkit.FlagDetails{
				AbbreviatedName: "",
				Usage:           "a fine flag for playing",
				ExpectedType:    cmdtoolkit.IntType,
				DefaultValue:    cmdtoolkit.NewIntBounds(0, 1, 10),
			},
			want: &cmdtoolkit.FlagDetails{
				AbbreviatedName: "",
				Usage:           "a fine flag for playing",
				ExpectedType:    cmdtoolkit.IntType,
				DefaultValue:    cmdtoolkit.NewIntBounds(0, 1, 10),
			},
		},
		"string": {
			fD: &cmdtoolkit.FlagDetails{
				AbbreviatedName: "",
				Usage:           "a fine flag for playing",
				ExpectedType:    cmdtoolkit.StringType,
				DefaultValue:    "hello!",
			},
			want: &cmdtoolkit.FlagDetails{
				AbbreviatedName: "",
				Usage:           "a fine flag for playing",
				ExpectedType:    cmdtoolkit.StringType,
				DefaultValue:    "hello!",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.fD.Copy(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Copy() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testFlagConsumer struct{}

func (testFlagConsumer) Bool(_ string, value bool, _ string) *bool     { return &value }
func (testFlagConsumer) Int(_ string, value int, _ string) *int        { return &value }
func (testFlagConsumer) String(_, value, _ string) *string             { return &value }
func (testFlagConsumer) BoolP(_, _ string, value bool, _ string) *bool { return &value }
func (testFlagConsumer) IntP(_, _ string, value int, _ string) *int    { return &value }
func (testFlagConsumer) StringP(_, _, value, _ string) *string {
	return &value
}

func TestAddFlags(t *testing.T) {
	type args struct {
		c     *cmdtoolkit.Configuration
		flags cmdtoolkit.FlagConsumer
		sets  []*cmdtoolkit.FlagSet
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"thorough": {
			args: args{
				c:     cmdtoolkit.EmptyConfiguration(),
				flags: testFlagConsumer{},
				sets: []*cmdtoolkit.FlagSet{
					{
						Name: "mySet",
						Details: map[string]*cmdtoolkit.FlagDetails{
							"good": {
								AbbreviatedName: "g",
								Usage:           "blah blah",
								ExpectedType:    cmdtoolkit.StringType,
								DefaultValue:    "foo",
							},
							"bad": nil,
						},
					},
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: there are no details for flag \"bad\".\n",
				Log: "" +
					"level='error'" +
					" error='no details present'" +
					" flag='bad'" +
					" set='mySet'" +
					" msg='internal error'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmdtoolkit.AddFlags(o, tt.args.c, tt.args.flags, tt.args.sets...)
			o.Report(t, "AddFlags()", tt.WantedRecording)
		})
	}
}

func TestGetBool(t *testing.T) {
	type args struct {
		results  map[string]*cmdtoolkit.CommandFlag[any]
		flagName string
	}
	tests := map[string]struct {
		args
		want    cmdtoolkit.CommandFlag[bool]
		wantErr bool
		output.WantedRecording
	}{
		"missing results": {
			args: args{
				results:  nil,
				flagName: "myFlag",
			},
			want:    cmdtoolkit.CommandFlag[bool]{},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: no flag values exist.\n",
				Log: "" +
					"level='error'" +
					" error='no results to extract flag values from'" +
					" msg='internal error'\n",
			},
		},
		"missing data": {
			args: args{
				results:  map[string]*cmdtoolkit.CommandFlag[any]{},
				flagName: "myFlag",
			},
			want:    cmdtoolkit.CommandFlag[bool]{},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" is not found.\n",
				Log: "" +
					"level='error'" +
					" error='flag not found'" +
					" flag='myFlag'" +
					" msg='internal error'\n",
			},
		},
		"nil data": {
			args: args{
				results: map[string]*cmdtoolkit.CommandFlag[any]{
					"myFlag": nil,
				},
				flagName: "myFlag",
			},
			want:    cmdtoolkit.CommandFlag[bool]{},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" has no data.\n",
				Log: "" +
					"level='error'" +
					" error='no data associated with flag'" +
					" flag='myFlag'" +
					" msg='internal error'\n",
			},
		},
		"bad default": {
			args: args{
				results: map[string]*cmdtoolkit.CommandFlag[any]{
					"myFlag": {Value: 12, UserSet: true},
				},
				flagName: "myFlag",
			},
			want:    cmdtoolkit.CommandFlag[bool]{},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" is not a boolean (12).\n",
				Log: "" +
					"level='error'" +
					" error='flag value is not a boolean'" +
					" flag='myFlag'" +
					" value='12'" +
					" msg='internal error'\n",
			},
		},
		"happy": {
			args: args{
				results: map[string]*cmdtoolkit.CommandFlag[any]{
					"myFlag": {Value: true, UserSet: true},
				},
				flagName: "myFlag",
			},
			want:            cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			wantErr:         false,
			WantedRecording: output.WantedRecording{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, err := cmdtoolkit.GetBool(o, tt.args.results, tt.args.flagName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBool() got = %v, want %v", got, tt.want)
			}
			o.Report(t, "GetBool()", tt.WantedRecording)
		})
	}
}

func TestGetInt(t *testing.T) {
	type args struct {
		results  map[string]*cmdtoolkit.CommandFlag[any]
		flagName string
	}
	tests := map[string]struct {
		args
		want    cmdtoolkit.CommandFlag[int]
		wantErr bool
		output.WantedRecording
	}{
		"missing results": {
			args: args{
				results:  nil,
				flagName: "myFlag",
			},
			want:    cmdtoolkit.CommandFlag[int]{},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: no flag values exist.\n",
				Log: "" +
					"level='error'" +
					" error='no results to extract flag values from'" +
					" msg='internal error'\n",
			},
		},
		"missing data": {
			args: args{
				results:  map[string]*cmdtoolkit.CommandFlag[any]{},
				flagName: "myFlag",
			},
			want:    cmdtoolkit.CommandFlag[int]{},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" is not found.\n",
				Log: "" +
					"level='error'" +
					" error='flag not found'" +
					" flag='myFlag'" +
					" msg='internal error'\n",
			},
		},
		"nil data": {
			args: args{
				results: map[string]*cmdtoolkit.CommandFlag[any]{
					"myFlag": nil,
				},
				flagName: "myFlag",
			},
			want:    cmdtoolkit.CommandFlag[int]{},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" has no data.\n",
				Log: "" +
					"level='error'" +
					" error='no data associated with flag'" +
					" flag='myFlag'" +
					" msg='internal error'\n",
			},
		},
		"bad default": {
			args: args{
				results: map[string]*cmdtoolkit.CommandFlag[any]{
					"myFlag": {Value: true, UserSet: true},
				},
				flagName: "myFlag",
			},
			want:    cmdtoolkit.CommandFlag[int]{},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" is not an integer (true).\n",
				Log: "" +
					"level='error'" +
					" error='flag value is not an integer'" +
					" flag='myFlag'" +
					" value='true'" +
					" msg='internal error'\n",
			},
		},
		"happy": {
			args: args{
				results: map[string]*cmdtoolkit.CommandFlag[any]{
					"myFlag": {Value: 12, UserSet: true},
				},
				flagName: "myFlag",
			},
			want:            cmdtoolkit.CommandFlag[int]{Value: 12, UserSet: true},
			wantErr:         false,
			WantedRecording: output.WantedRecording{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, err := cmdtoolkit.GetInt(o, tt.args.results, tt.args.flagName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInt() got = %v, want %v", got, tt.want)
			}
			o.Report(t, "GetInt()", tt.WantedRecording)
		})
	}
}

func TestGetString(t *testing.T) {
	type args struct {
		results  map[string]*cmdtoolkit.CommandFlag[any]
		flagName string
	}
	tests := map[string]struct {
		args
		want    cmdtoolkit.CommandFlag[string]
		wantErr bool
		output.WantedRecording
	}{
		"missing results": {
			args: args{
				results:  nil,
				flagName: "myFlag",
			},
			want:    cmdtoolkit.CommandFlag[string]{},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: no flag values exist.\n",
				Log: "" +
					"level='error'" +
					" error='no results to extract flag values from'" +
					" msg='internal error'\n",
			},
		},
		"missing data": {
			args: args{
				results:  map[string]*cmdtoolkit.CommandFlag[any]{},
				flagName: "myFlag",
			},
			want:    cmdtoolkit.CommandFlag[string]{},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" is not found.\n",
				Log: "" +
					"level='error'" +
					" error='flag not found'" +
					" flag='myFlag'" +
					" msg='internal error'\n",
			},
		},
		"nil data": {
			args: args{
				results: map[string]*cmdtoolkit.CommandFlag[any]{
					"myFlag": nil,
				},
				flagName: "myFlag",
			},
			want:    cmdtoolkit.CommandFlag[string]{},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" has no data.\n",
				Log: "" +
					"level='error'" +
					" error='no data associated with flag'" +
					" flag='myFlag'" +
					" msg='internal error'\n",
			},
		},
		"bad default": {
			args: args{
				results: map[string]*cmdtoolkit.CommandFlag[any]{
					"myFlag": {Value: true, UserSet: true},
				},
				flagName: "myFlag",
			},
			want:    cmdtoolkit.CommandFlag[string]{},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" is not a string (true).\n",
				Log: "" +
					"level='error'" +
					" error='flag value is not a string'" +
					" flag='myFlag'" +
					" value='true'" +
					" msg='internal error'\n",
			},
		},
		"happy": {
			args: args{
				results: map[string]*cmdtoolkit.CommandFlag[any]{
					"myFlag": {Value: "boo", UserSet: true},
				},
				flagName: "myFlag",
			},
			want:            cmdtoolkit.CommandFlag[string]{Value: "boo", UserSet: true},
			wantErr:         false,
			WantedRecording: output.WantedRecording{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, err := cmdtoolkit.GetString(o, tt.args.results, tt.args.flagName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetString() got = %v, want %v", got, tt.want)
			}
			o.Report(t, "GetString()", tt.WantedRecording)
		})
	}
}

func TestProcessFlagErrors(t *testing.T) {
	tests := map[string]struct {
		eSlice []error
		want   bool
		output.WantedRecording
	}{
		"no errors": {
			eSlice:          []error{},
			want:            true,
			WantedRecording: output.WantedRecording{},
		},
		"errors": {
			eSlice: []error{errors.New("some error")},
			want:   false,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: some error.\n",
				Log:   "level='error' error='some error' msg='internal error'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := cmdtoolkit.ProcessFlagErrors(o, tt.eSlice); got != tt.want {
				t.Errorf("ProcessFlagErrors() = %v, want %v", got, tt.want)
			}
			o.Report(t, "ProgressFlagErrors()", tt.WantedRecording)
		})
	}
}

type testFlagProducer struct{}

func (testFlagProducer) Changed(_ string) bool              { return true }
func (testFlagProducer) GetBool(_ string) (bool, error)     { return true, nil }
func (testFlagProducer) GetInt(_ string) (int, error)       { return 12, nil }
func (testFlagProducer) GetString(_ string) (string, error) { return "foo", nil }

func TestReadFlags(t *testing.T) {
	type args struct {
		producer cmdtoolkit.FlagProducer
		set      *cmdtoolkit.FlagSet
	}
	tests := map[string]struct {
		args
		want  map[string]*cmdtoolkit.CommandFlag[any]
		want1 int // error count
	}{
		"thorough": {
			args: args{
				producer: testFlagProducer{},
				set: &cmdtoolkit.FlagSet{
					Name: "mySet",
					Details: map[string]*cmdtoolkit.FlagDetails{
						"nil": nil,
						"b": {
							AbbreviatedName: "",
							Usage:           "",
							ExpectedType:    cmdtoolkit.BoolType,
							DefaultValue:    false,
						},
						"i": {
							AbbreviatedName: "",
							Usage:           "",
							ExpectedType:    cmdtoolkit.IntType,
							DefaultValue:    &cmdtoolkit.IntBounds{0, 1, 2},
						},
						"s": {
							AbbreviatedName: "",
							Usage:           "",
							ExpectedType:    cmdtoolkit.StringType,
							DefaultValue:    false,
						},
						"ugh": {
							AbbreviatedName: "",
							Usage:           "",
							DefaultValue:    []byte("blah"),
						},
					},
				},
			},
			want: map[string]*cmdtoolkit.CommandFlag[any]{
				"b": {Value: true, UserSet: true},
				"i": {Value: 12, UserSet: true},
				"s": {Value: "foo", UserSet: true},
			},
			want1: 2,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, got1 := cmdtoolkit.ReadFlags(tt.args.producer, tt.args.set)
			if len(got) != len(tt.want) {
				t.Errorf("ReadFlags() got = %d entries, want %d", len(got), len(tt.want))
			} else {
				for k, v := range got {
					if *v != *tt.want[k] {
						t.Errorf("ReadFlags() got[%s] = %v , want %v", k, v, tt.want[k])
					}
				}
			}
			if len(got1) != tt.want1 {
				t.Errorf("ReadFlags() got1 = %v, want %v", len(got1), tt.want1)
			}
		})
	}
}
