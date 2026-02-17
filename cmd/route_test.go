package cmd

import (
	"testing"
)

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{100, "100"},
		{999, "999"},
		{1000, "1,000"},
		{8400, "8,400"},
		{10000, "10,000"},
		{1234567, "1,234,567"},
	}
	for _, tt := range tests {
		got := formatNumber(tt.n)
		if got != tt.want {
			t.Errorf("formatNumber(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestParseDepartTime_HHmm(t *testing.T) {
	got, err := parseDepartTime("08:30")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 12 {
		t.Errorf("expected 12-char timestamp, got %q (len=%d)", got, len(got))
	}
}

func TestParseDepartTime_FullTimestamp(t *testing.T) {
	got, err := parseDepartTime("202602181430")
	if err != nil {
		t.Fatal(err)
	}
	if got != "202602181430" {
		t.Errorf("got %q, want %q", got, "202602181430")
	}
}

func TestParseDepartTime_Invalid(t *testing.T) {
	invalids := []string{"invalid", "25:00", "abc", "12345"}
	for _, input := range invalids {
		_, err := parseDepartTime(input)
		if err == nil {
			t.Errorf("parseDepartTime(%q): expected error, got nil", input)
		}
	}
}

func TestContainsStr(t *testing.T) {
	slice := []string{"a", "b", "c"}
	if !containsStr(slice, "b") {
		t.Error("expected true for 'b'")
	}
	if containsStr(slice, "d") {
		t.Error("expected false for 'd'")
	}
	if containsStr(nil, "a") {
		t.Error("expected false for nil slice")
	}
}

func TestResolveAlias_NilConfig(t *testing.T) {
	oldCfg := cfg
	cfg = nil
	defer func() { cfg = oldCfg }()

	got := resolveAlias("home")
	if got != "home" {
		t.Errorf("resolveAlias with nil config: got %q, want %q", got, "home")
	}
}
