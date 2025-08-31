package cmd_toolkit

import (
	"strings"
	"testing"

	"github.com/majohn-r/output"
)

func Test_finalYear(t *testing.T) {
	tests := map[string]struct {
		firstYear int
		timestamp string
		want      int
		output.WantedRecording
	}{
		"bad timestamp": {
			firstYear: 1900,
			timestamp: "today",
			want:      1900,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The build time \"today\" cannot be parsed: " +
					"'*time.ParseError: parsing time \"today\" as \"2006-01-02T15:04:05Z07:00\": " +
					"cannot parse \"today\" as \"2006\"'.\n",
				Log: "" +
					"level='error' " +
					"error='parsing time \"today\" as \"2006-01-02T15:04:05Z07:00\": " +
					"cannot parse \"today\" as \"2006\"' " +
					"value='today' " +
					"msg='parse error'\n",
			},
		},
		"good timestamp": {
			firstYear: 1999,
			timestamp: "2022-08-10T13:29:57-04:00",
			want:      2022,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := finalYear(o, tt.timestamp, tt.firstYear); got != tt.want {
				t.Errorf("finalYear() = %v, want %v", got, tt.want)
			}
			o.Report(t, "finalYear()", tt.WantedRecording)
		})
	}
}

func Test_formatCopyright(t *testing.T) {
	type args struct {
		firstYear int
		lastYear  int
		author    string
	}
	tests := map[string]struct {
		args
		want string
	}{
		"bad last year": {
			args: args{author: "me", firstYear: 2022, lastYear: 2020},
			want: "Copyright © 2022 me",
		},
		"same year": {
			args: args{author: "myself", firstYear: 2022, lastYear: 2022},
			want: "Copyright © 2022 myself",
		},
		"later year": {
			args: args{author: "I", firstYear: 2022, lastYear: 2023},
			want: "Copyright © 2022-2023 I",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := formatCopyright(tt.args.firstYear, tt.args.lastYear, tt.args.author); got != tt.want {
				t.Errorf("formatCopyright() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_translateTimestamp(t *testing.T) {
	tests := map[string]struct {
		// note, this is what the output should start with; for reasons I do not
		// understand, there is some variation as to how the timezone is written
		// at the end
		s           string
		wantMinusTZ string
	}{
		"bad input": {
			s:           "today",
			wantMinusTZ: "today",
		},
		"good input": {
			s:           "2022-08-10T13:29:57-04:00",
			wantMinusTZ: "Wednesday, August 10 2022, 13:29:57",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := translateTimestamp(tt.s); !strings.HasPrefix(got, tt.wantMinusTZ) {
				t.Errorf("translateTimestamp() = %v, want %v", got, tt.wantMinusTZ)
			}
		})
	}
}
