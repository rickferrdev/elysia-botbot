package funslipper

import "testing"

func TestNormalizeUserID(t *testing.T) {
	cases := map[string]string{
		"123":      "123",
		"<@123>":   "123",
		"<@!123>":  "123",
	}
	for in, want := range cases {
		if got := normalizeUserID(in); got != want {
			t.Errorf("normalizeUserID(%q) = %q, want %q", in, got, want)
		}
	}
}
