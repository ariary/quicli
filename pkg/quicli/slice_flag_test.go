package quicli

import "testing"

func TestStringSliceValueString(t *testing.T) {
	val := []string{"a", "b", "c"}
	sv := &stringSliceValue{val: &val}
	if got := sv.String(); got != "a,b,c" {
		t.Errorf("String() = %q, want 'a,b,c'", got)
	}
}

func TestStringSliceValueStringNil(t *testing.T) {
	sv := &stringSliceValue{val: nil}
	if got := sv.String(); got != "" {
		t.Errorf("String() nil = %q, want empty", got)
	}
}

func TestStringSliceValueStringEmpty(t *testing.T) {
	val := []string{}
	sv := &stringSliceValue{val: &val}
	if got := sv.String(); got != "" {
		t.Errorf("String() empty = %q, want empty", got)
	}
}
