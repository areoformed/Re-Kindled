package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
)

const typographyPresetSchemaVersion = 1

type TypographyPreset struct {
	SchemaVersion int          `json:"schema_version"`
	ID            string       `json:"id"`
	Label         string       `json:"label"`
	Header        string       `json:"header"`
	Calibration   Calibration  `json:"calibration"`
	Fonts         PresetFonts  `json:"fonts"`
	Layout        PresetLayout `json:"layout"`
}

type Calibration struct {
	StandardReference       string  `json:"standard_reference"`
	ViewingDistanceInches   float64 `json:"viewing_distance_inches"`
	PanelPixelsPerInch      float64 `json:"panel_pixels_per_inch"`
	PanelWidthPixels        int     `json:"panel_width_pixels"`
	PanelHeightPixels       int     `json:"panel_height_pixels"`
	BodyCapHeightDegrees    float64 `json:"body_cap_height_degrees"`
	UtilityCapHeightDegrees float64 `json:"utility_cap_height_degrees"`
}

type PresetFonts struct {
	Narrative FontFace `json:"narrative"`
	Utility   FontFace `json:"utility"`
}

type FontFace struct {
	Path           string `json:"path"`
	AscentUnits    int    `json:"ascent_units"`
	DescentUnits   int    `json:"descent_units"`
	CapHeightUnits int    `json:"cap_height_units"`
	FBInkPixels    int    `json:"fbink_pixels"`
}

type PresetLayout struct {
	HeaderTopPixels    int `json:"header_top_pixels"`
	HeaderBottomPixels int `json:"header_bottom_pixels"`
	HeaderSidePixels   int `json:"header_side_pixels"`
	BodyTopPixels      int `json:"body_top_pixels"`
	BodyBottomPixels   int `json:"body_bottom_pixels"`
	BodyLeftPixels     int `json:"body_left_pixels"`
	BodyRightPixels    int `json:"body_right_pixels"`
	FooterTopPixels    int `json:"footer_top_pixels"`
	FooterBottomPixels int `json:"footer_bottom_pixels"`
	FooterSidePixels   int `json:"footer_side_pixels"`
}

func loadTypographyPreset(path string) (TypographyPreset, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return TypographyPreset{}, err
	}

	var preset TypographyPreset
	if err := json.Unmarshal(raw, &preset); err != nil {
		return TypographyPreset{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := preset.Validate(); err != nil {
		return TypographyPreset{}, fmt.Errorf("validate %s: %w", path, err)
	}
	return preset, nil
}

func (preset *TypographyPreset) CalibrateViewingDistance(distanceInches float64) error {
	if distanceInches <= 0 {
		return fmt.Errorf("viewing distance must be positive")
	}
	preset.Calibration.ViewingDistanceInches = distanceInches
	preset.Fonts.Narrative.FBInkPixels = requiredFBInkPixels(
		preset.Calibration.BodyCapHeightDegrees,
		distanceInches,
		preset.Calibration.PanelPixelsPerInch,
		preset.Fonts.Narrative,
	)
	preset.Fonts.Utility.FBInkPixels = requiredFBInkPixels(
		preset.Calibration.UtilityCapHeightDegrees,
		distanceInches,
		preset.Calibration.PanelPixelsPerInch,
		preset.Fonts.Utility,
	)
	return preset.Validate()
}

func (preset TypographyPreset) Validate() error {
	if preset.SchemaVersion != typographyPresetSchemaVersion {
		return fmt.Errorf("unsupported schema_version %d", preset.SchemaVersion)
	}
	if preset.ID == "" || preset.Label == "" || preset.Header == "" {
		return fmt.Errorf("id, label, and header are required")
	}
	c := preset.Calibration
	if c.ViewingDistanceInches <= 0 || c.PanelPixelsPerInch <= 0 || c.PanelWidthPixels <= 0 || c.PanelHeightPixels <= 0 {
		return fmt.Errorf("panel geometry and viewing distance must be positive")
	}
	if c.BodyCapHeightDegrees <= 0 || c.UtilityCapHeightDegrees <= 0 {
		return fmt.Errorf("cap-height angles must be positive")
	}
	if err := validateFace("narrative", preset.Fonts.Narrative, c.BodyCapHeightDegrees, c); err != nil {
		return err
	}
	if err := validateFace("utility", preset.Fonts.Utility, c.UtilityCapHeightDegrees, c); err != nil {
		return err
	}

	l := preset.Layout
	values := []int{
		l.HeaderTopPixels, l.HeaderBottomPixels, l.HeaderSidePixels,
		l.BodyTopPixels, l.BodyBottomPixels, l.BodyLeftPixels, l.BodyRightPixels,
		l.FooterTopPixels, l.FooterBottomPixels, l.FooterSidePixels,
	}
	for _, value := range values {
		if value < 0 {
			return fmt.Errorf("layout values cannot be negative")
		}
	}
	if l.BodyLeftPixels+l.BodyRightPixels >= c.PanelWidthPixels {
		return fmt.Errorf("body side margins consume the panel width")
	}
	if l.BodyTopPixels+l.BodyBottomPixels >= c.PanelHeightPixels {
		return fmt.Errorf("body top and bottom margins consume the panel height")
	}
	if l.HeaderTopPixels+l.HeaderBottomPixels >= c.PanelHeightPixels {
		return fmt.Errorf("header margins consume the panel height")
	}
	if l.FooterTopPixels+l.FooterBottomPixels >= c.PanelHeightPixels {
		return fmt.Errorf("footer margins consume the panel height")
	}
	return nil
}

func validateFace(role string, face FontFace, angle float64, calibration Calibration) error {
	if face.Path == "" {
		return fmt.Errorf("%s font path is required", role)
	}
	if face.AscentUnits <= face.DescentUnits || face.CapHeightUnits <= 0 || face.FBInkPixels <= 0 {
		return fmt.Errorf("%s font metrics are invalid", role)
	}
	required := requiredFBInkPixels(angle, calibration.ViewingDistanceInches, calibration.PanelPixelsPerInch, face)
	if face.FBInkPixels < required {
		return fmt.Errorf("%s font is %d px; %d px is required for %.3f degrees at %.1f in", role, face.FBInkPixels, required, angle, calibration.ViewingDistanceInches)
	}
	return nil
}

func requiredFBInkPixels(angleDegrees, distanceInches, pixelsPerInch float64, face FontFace) int {
	targetCapPixels := capHeightPixelsForAngle(angleDegrees, distanceInches, pixelsPerInch)
	fontVerticalUnits := float64(face.AscentUnits - face.DescentUnits)
	return int(math.Ceil(targetCapPixels * fontVerticalUnits / float64(face.CapHeightUnits)))
}

func capHeightPixelsForAngle(angleDegrees, distanceInches, pixelsPerInch float64) float64 {
	return 2 * distanceInches * pixelsPerInch * math.Tan(angleDegrees*math.Pi/360)
}

func achievedCapHeightPixels(face FontFace) float64 {
	return float64(face.FBInkPixels) * float64(face.CapHeightUnits) / float64(face.AscentUnits-face.DescentUnits)
}

func achievedVisualAngleDegrees(face FontFace, distanceInches, pixelsPerInch float64) float64 {
	capInches := achievedCapHeightPixels(face) / pixelsPerInch
	return 2 * math.Atan(capInches/(2*distanceInches)) * 180 / math.Pi
}
