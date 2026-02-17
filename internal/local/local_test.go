package local

import (
	"testing"
)

func TestParseCoords(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"127.0286,37.5147", true},
		{"127.0286, 37.5147", true},
		{"hakdong station", false},
		{"abc,def", false},
		{"127.0286", false},
		{"", false},
	}
	for _, tt := range tests {
		result := parseCoords(tt.input)
		got := result != nil
		if got != tt.want {
			t.Errorf("parseCoords(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseCoords_Values(t *testing.T) {
	result := parseCoords("127.0286,37.5147")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result[0] != "127.0286" {
		t.Errorf("x = %q, want %q", result[0], "127.0286")
	}
	if result[1] != "37.5147" {
		t.Errorf("y = %q, want %q", result[1], "37.5147")
	}
}

func TestCategoryCodes(t *testing.T) {
	tests := []struct {
		alias string
		code  string
	}{
		{"cafe", "CE7"},
		{"food", "FD6"},
		{"subway", "SW8"},
		{"pharmacy", "PM9"},
		{"convenience", "CS2"},
		{"hospital", "HP8"},
		{"bank", "BK9"},
		{"gas", "OL7"},
		{"parking", "PK6"},
		{"restaurant", "FD6"},
		{"mart", "MT1"},
		{"school", "SC4"},
		{"hotel", "AD5"},
		{"culture", "CT1"},
	}
	for _, tt := range tests {
		code, ok := CategoryCodes[tt.alias]
		if !ok {
			t.Errorf("CategoryCodes[%q] not found", tt.alias)
			continue
		}
		if code != tt.code {
			t.Errorf("CategoryCodes[%q] = %q, want %q", tt.alias, code, tt.code)
		}
	}
}

func TestCategoryCodes_FoodAndRestaurant(t *testing.T) {
	if CategoryCodes["food"] != CategoryCodes["restaurant"] {
		t.Error("food and restaurant should map to same code FD6")
	}
}
