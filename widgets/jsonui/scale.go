//go:build windows

package jsonui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AzureIvory/winui/widgets"
)

func parseScalePolicy(raw json.RawMessage) (widgets.ScalePolicy, error) {
	if len(raw) == 0 {
		return widgets.ScalePolicy{}, nil
	}

	var mode string
	if err := json.Unmarshal(raw, &mode); err == nil {
		parsed, err := parseScaleModeValue(mode)
		if err != nil {
			return widgets.ScalePolicy{}, err
		}
		return widgets.ScalePolicy{Mode: parsed}, nil
	}

	var spec struct {
		Mode    string `json:"mode"`
		Layout  string `json:"layout"`
		Font    string `json:"font"`
		Image   string `json:"image"`
		Padding string `json:"padding"`
		Gap     string `json:"gap"`
		Radius  string `json:"radius"`
	}
	if err := json.Unmarshal(raw, &spec); err != nil {
		return widgets.ScalePolicy{}, err
	}

	modeValue, err := parseScaleModeValue(spec.Mode)
	if err != nil {
		return widgets.ScalePolicy{}, err
	}
	layoutValue, err := parseScaleModeValue(spec.Layout)
	if err != nil {
		return widgets.ScalePolicy{}, err
	}
	fontValue, err := parseScaleModeValue(spec.Font)
	if err != nil {
		return widgets.ScalePolicy{}, err
	}
	imageValue, err := parseScaleModeValue(spec.Image)
	if err != nil {
		return widgets.ScalePolicy{}, err
	}
	paddingValue, err := parseScaleModeValue(spec.Padding)
	if err != nil {
		return widgets.ScalePolicy{}, err
	}
	gapValue, err := parseScaleModeValue(spec.Gap)
	if err != nil {
		return widgets.ScalePolicy{}, err
	}
	radiusValue, err := parseScaleModeValue(spec.Radius)
	if err != nil {
		return widgets.ScalePolicy{}, err
	}

	return widgets.ScalePolicy{
		Mode:    modeValue,
		Layout:  layoutValue,
		Font:    fontValue,
		Image:   imageValue,
		Padding: paddingValue,
		Gap:     gapValue,
		Radius:  radiusValue,
	}, nil
}

func parseScaleModeValue(value string) (widgets.ScaleMode, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "inherit":
		return widgets.ScaleInherit, nil
	case "dp":
		return widgets.ScaleDP, nil
	case "px":
		return widgets.ScalePX, nil
	default:
		return widgets.ScaleInherit, fmt.Errorf("invalid scale mode %q", value)
	}
}
