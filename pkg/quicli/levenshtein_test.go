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

func TestFindClosestSubcommandDistanceTwoBoundary(t *testing.T) {
	// levenshtein("dele", "delete") = 2, exactly at maxDist boundary — must match.
	subs := Subcommands{{Name: "delete"}}
	if got := findClosestSubcommand(subs, "dele"); got != "delete" {
		t.Errorf("distance 2 boundary: expected 'delete', got %q", got)
	}
}

func TestFindClosestSubcommandDistanceThreeRejects(t *testing.T) {
	// levenshtein("del", "delete") = 3, beyond maxDist — must NOT match.
	subs := Subcommands{{Name: "delete"}}
	if got := findClosestSubcommand(subs, "del"); got != "" {
		t.Errorf("distance 3: expected empty, got %q", got)
	}
}

func TestFindClosestSubcommandFirstMatchWins(t *testing.T) {
	// "ab" is distance 1 from both "ax" and "ay". First match (by name) wins.
	subs := Subcommands{{Name: "ax"}, {Name: "ay"}}
	if got := findClosestSubcommand(subs, "ab"); got != "ax" {
		t.Errorf("equidistant names: expected 'ax', got %q", got)
	}

	// Name vs alias equidistant: name match (found first) wins.
	subs2 := Subcommands{
		{Name: "ax"},
		{Name: "other", Aliases: Aliases("az")},
	}
	if got := findClosestSubcommand(subs2, "ab"); got != "ax" {
		t.Errorf("name vs alias equidistant: expected 'ax', got %q", got)
	}
}
