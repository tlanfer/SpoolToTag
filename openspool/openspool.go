package openspool

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	Protocol = "openspool"
	Version  = "1.0"
)

var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

var ValidBrands = []string{"Generic", "Overture", "PolyLite", "eSun", "PolyTerra"}

func NormalizeBrand(brand string) string {
	for _, b := range ValidBrands {
		if strings.EqualFold(brand, b) {
			return b
		}
	}
	return "Generic"
}

type SpoolData struct {
	Protocol string `json:"protocol"`
	Version  string `json:"version"`
	Type     string `json:"type"`
	ColorHex string `json:"color_hex"`
	Brand    string `json:"brand"`
	MinTemp  int    `json:"min_temp"`
	MaxTemp  int    `json:"max_temp"`
}

func New(filamentType, colorHex, brand string, minTemp, maxTemp int) (SpoolData, error) {
	s := SpoolData{
		Protocol: Protocol,
		Version:  Version,
		Type:     filamentType,
		ColorHex: colorHex,
		Brand:    brand,
		MinTemp:  minTemp,
		MaxTemp:  maxTemp,
	}
	if err := s.Validate(); err != nil {
		return SpoolData{}, err
	}
	return s, nil
}

func (s SpoolData) Validate() error {
	if s.Protocol != Protocol {
		return fmt.Errorf("invalid protocol: %q", s.Protocol)
	}
	if s.Version != Version {
		return fmt.Errorf("invalid version: %q", s.Version)
	}
	if s.Type == "" {
		return fmt.Errorf("type is required")
	}
	if !hexColorRe.MatchString(s.ColorHex) {
		return fmt.Errorf("invalid color_hex: %q", s.ColorHex)
	}
	if s.Brand == "" {
		return fmt.Errorf("brand is required")
	}
	if s.MinTemp <= 0 {
		return fmt.Errorf("min_temp must be positive")
	}
	if s.MaxTemp <= 0 {
		return fmt.Errorf("max_temp must be positive")
	}
	if s.MinTemp > s.MaxTemp {
		return fmt.Errorf("min_temp (%d) must not exceed max_temp (%d)", s.MinTemp, s.MaxTemp)
	}
	return nil
}

func ParseTemp(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid temperature %q: %w", s, err)
	}
	return n, nil
}
