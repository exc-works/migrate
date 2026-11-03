package main

import "testing"

func TestConsoleKeyRepeat(t *testing.T) {
	for _, tc := range []struct {
		ch    rune
		count uint16
		want  string
	}{{'a', 3, "aaa"}, {'密', 2, "密密"}, {127, 3, string([]byte{127, 127, 127})}, {'a', 0, ""}} {
		if got := string(repeatConsoleRune(tc.ch, tc.count)); got != tc.want {
			t.Fatalf("repeat %U: %q", tc.ch, got)
		}
	}
}
