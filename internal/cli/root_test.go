package cli

import (
	"errors"
	"testing"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		err  error
		want int
	}{
		{nil, ExitSuccess},
		{errors.New("validation"), ExitValidation},
		{ErrRuntime{Err: errors.New("internal")}, ExitRuntime},
	}
	for _, tt := range tests {
		got := ExitCode(tt.err)
		if got != tt.want {
			t.Errorf("ExitCode(%v) = %d, want %d", tt.err, got, tt.want)
		}
	}
}

func TestParsePackRef(t *testing.T) {
	name, version, err := parsePackRef("security-core@1.0.0")
	if err != nil {
		t.Fatalf("parsePackRef: %v", err)
	}
	if name != "security-core" || version != "1.0.0" {
		t.Errorf("got name=%q version=%q", name, version)
	}
	_, _, err = parsePackRef("no-at-sign")
	if err == nil {
		t.Error("expected error for invalid ref")
	}
}
