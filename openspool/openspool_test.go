package openspool

import (
	"testing"
)

func TestNew_Valid(t *testing.T) {
	s, err := New("PLA", "FF5733", "eSun", 190, 220)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Protocol != Protocol {
		t.Errorf("protocol = %q, want %q", s.Protocol, Protocol)
	}
	if s.Version != Version {
		t.Errorf("version = %q, want %q", s.Version, Version)
	}
	if s.Type != "PLA" {
		t.Errorf("type = %q, want %q", s.Type, "PLA")
	}
}

func TestNew_InvalidColorHex(t *testing.T) {
	tests := []string{"#FF5733", "GGG000", "FFF", "red", ""}
	for _, hex := range tests {
		_, err := New("PLA", hex, "eSun", 190, 220)
		if err == nil {
			t.Errorf("expected error for color_hex %q", hex)
		}
	}
}

func TestNew_EmptyType(t *testing.T) {
	_, err := New("", "FF5733", "eSun", 190, 220)
	if err == nil {
		t.Error("expected error for empty type")
	}
}

func TestNew_EmptyBrand(t *testing.T) {
	_, err := New("PLA", "FF5733", "", 190, 220)
	if err == nil {
		t.Error("expected error for empty brand")
	}
}

func TestNew_BadTemps(t *testing.T) {
	tests := []struct {
		name    string
		min     int
		max     int
	}{
		{"zero min", 0, 220},
		{"zero max", 190, 0},
		{"negative min", -10, 220},
		{"min exceeds max", 250, 220},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New("PLA", "FF5733", "eSun", tt.min, tt.max)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestValidate_WrongProtocol(t *testing.T) {
	s := SpoolData{
		Protocol: "wrong",
		Version:  Version,
		Type:     "PLA",
		ColorHex: "FF5733",
		Brand:    "eSun",
		MinTemp:  190,
		MaxTemp:  220,
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for wrong protocol")
	}
}

func TestNormalizeBrand(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"eSun", "eSun"},
		{"esun", "eSun"},
		{"ESUN", "eSun"},
		{"Overture", "Overture"},
		{"overture", "Overture"},
		{"PolyLite", "PolyLite"},
		{"PolyTerra", "PolyTerra"},
		{"Generic", "Generic"},
		{"Bambu", "Generic"},
		{"Unknown Brand", "Generic"},
		{"", "Generic"},
	}
	for _, tt := range tests {
		got := NormalizeBrand(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeBrand(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseTemp(t *testing.T) {
	n, err := ParseTemp("200")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 200 {
		t.Errorf("got %d, want 200", n)
	}

	_, err = ParseTemp("abc")
	if err == nil {
		t.Error("expected error for non-numeric temp")
	}
}
