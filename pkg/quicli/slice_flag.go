package quicli

import "strings"

// stringSliceValue implements flag.Value for []string flags.
// Supports repeated flags (-f a -f b) and comma-separated values (-f a,b).
type stringSliceValue struct {
	val *[]string
}

func (s *stringSliceValue) String() string {
	if s.val == nil {
		return ""
	}
	return strings.Join(*s.val, ",")
}

func (s *stringSliceValue) Set(v string) error {
	*s.val = append(*s.val, strings.Split(v, ",")...)
	return nil
}
