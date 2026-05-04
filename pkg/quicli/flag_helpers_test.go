package quicli

import (
	"strings"
	"testing"
	"time"
)

func TestFlagEnvVarDisplay(t *testing.T) {
	defer setArgs([]string{"prog"})()

	// Opt-out returns ""
	if got := flagEnvVarDisplay(Flag{Name: "secret", EnvVar: "-"}); got != "" {
		t.Errorf("opt-out: got %q, want empty", got)
	}
	// Explicit env var
	if got := flagEnvVarDisplay(Flag{Name: "count", EnvVar: "MY_COUNT"}); got != "MY_COUNT" {
		t.Errorf("explicit: got %q, want MY_COUNT", got)
	}
	// Auto-derived
	if got := flagEnvVarDisplay(Flag{Name: "count"}); got != "PROG_COUNT" {
		t.Errorf("auto: got %q, want PROG_COUNT", got)
	}
}

func TestGetFlagLineRequired(t *testing.T) {
	line := getFlagLine(Flag{Name: "output", Description: "out", Default: "", Required: true}, "o", "")
	if !strings.Contains(line, "(required)") {
		t.Errorf("required flag should contain '(required)': %s", line)
	}
	if strings.Contains(line, "(default:") {
		t.Errorf("required flag should not show default: %s", line)
	}
}

func TestGetFlagLineNotRequired(t *testing.T) {
	line := getFlagLine(Flag{Name: "count", Description: "count", Default: 5}, "c", "")
	if strings.Contains(line, "(required)") {
		t.Errorf("non-required should not contain '(required)': %s", line)
	}
	if !strings.Contains(line, "(default:") {
		t.Errorf("non-required should show default: %s", line)
	}
}

func TestGetFlagLineChoices(t *testing.T) {
	line := getFlagLine(Flag{Name: "format", Description: "fmt", Default: "json", Choices: []string{"json", "yaml"}}, "f", "")
	if !strings.Contains(line, "(choices: json, yaml)") {
		t.Errorf("choices should appear: %s", line)
	}
}

func TestGetFlagLineNoChoices(t *testing.T) {
	line := getFlagLine(Flag{Name: "name", Description: "name", Default: "world"}, "n", "")
	if strings.Contains(line, "(choices:") {
		t.Errorf("no choices flag should not show choices: %s", line)
	}
}

func TestGetFlagLineEnvVar(t *testing.T) {
	line := getFlagLine(Flag{Name: "name", Description: "name", Default: "world"}, "n", "MY_NAME")
	if !strings.Contains(line, "[env: MY_NAME]") {
		t.Errorf("env var should appear: %s", line)
	}
}

func TestGetFlagLineShortPresent(t *testing.T) {
	line := getFlagLine(Flag{Name: "output", Description: "out", Default: ""}, "o", "")
	if !strings.Contains(line, "-o") {
		t.Errorf("short name should appear: %s", line)
	}
}

func TestGetFlagLineShortAbsent(t *testing.T) {
	line := getFlagLine(Flag{Name: "output", Description: "out", Default: ""}, "", "")
	// Without short, format is "--output\t\t\t..."
	if strings.Contains(line, "\t-o") {
		t.Errorf("no short name should not show -o: %s", line)
	}
}

func TestFormatDefault(t *testing.T) {
	cases := []struct {
		val  any
		want string
	}{
		{0, "0"},
		{42, "42"},
		{"hello", `"hello"`},
		{"", `""`},
		{true, "true"},
		{false, "false"},
		{float64(3.14), "3.14"},
		{float64(0), "0"},
		{[]string{}, "[]"},
		{5 * time.Second, "5s"},
	}
	for _, tc := range cases {
		if got := formatDefault(tc.val); got != tc.want {
			t.Errorf("formatDefault(%v) = %q, want %q", tc.val, got, tc.want)
		}
	}
}
