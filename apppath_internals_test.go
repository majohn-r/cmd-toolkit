package cmd_toolkit

import (
	"testing"
)

func Test_isLegalApplicationName(t *testing.T) {
	tests := map[string]struct {
		applicationName string
		want            bool
	}{
		"empty": {
			applicationName: "",
			want:            false,
		},
		"illegal": {
			applicationName: "../../../abc",
			want:            false,
		},
		"also illegal": {
			applicationName: "C:\\System",
			want:            false,
		},
		"valid": {
			applicationName: "mp3repair",
			want:            true,
		},
		"also valid": {
			applicationName: "SMF-tool",
			want:            true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := isLegalApplicationName(tt.applicationName); got != tt.want {
				t.Errorf("isLegalApplicationName() = %v, want %v", got, tt.want)
			}
		})
	}
}
