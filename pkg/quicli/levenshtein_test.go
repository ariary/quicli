package quicli

import "testing"

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "abc", 0},
		{"abc", "ab", 1},
		{"abc", "axc", 1},
		{"kitten", "sitting", 3},
		{"get", "gett", 1},
	}
	for _, tc := range cases {
		if got := levenshtein(tc.a, tc.b); got != tc.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestFindClosestSubcommand(t *testing.T) {
	subs := Subcommands{
		{Name: "get"},
		{Name: "delete"},
		{Name: "list", Aliases: Aliases("ls")},
	}

	if got := findClosestSubcommand(subs, "gett"); got != "get" {
		t.Errorf("expected 'get', got %q", got)
	}
	if got := findClosestSubcommand(subs, "delet"); got != "delete" {
		t.Errorf("expected 'delete', got %q", got)
	}
	// alias close match should return the canonical name
	if got := findClosestSubcommand(subs, "lss"); got != "list" {
		t.Errorf("expected 'list', got %q", got)
	}
	// too far away — no suggestion
	if got := findClosestSubcommand(subs, "zzzzzzz"); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
