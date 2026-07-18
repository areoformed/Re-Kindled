package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

type TypeLabInput struct {
	Preset                 string  `json:"preset"`
	BodySizeChangePercent  float64 `json:"body_size_change_percent"`
	TargetCapHeightDegrees float64 `json:"target_cap_height_degrees"`
	LinePitchChangePercent float64 `json:"line_pitch_change_percent"`
	ViewingDistanceInches  float64 `json:"viewing_distance_inches"`
}

type typeLabFace struct {
	AscentUnits    int `json:"ascent_units"`
	DescentUnits   int `json:"descent_units"`
	LineGapUnits   int `json:"line_gap_units"`
	CapHeightUnits int `json:"cap_height_units"`
	FBInkPixels    int `json:"fbink_pixels"`
}

func (device *SSHDevice) TypeLab(input TypeLabInput) (map[string]any, error) {
	if err := validatePresetID(input.Preset); err != nil {
		return nil, err
	}
	if input.BodySizeChangePercent < -40 || input.BodySizeChangePercent > 100 {
		return nil, errors.New("body_size_change_percent must be between -40 and 100")
	}
	if input.LinePitchChangePercent < -20 || input.LinePitchChangePercent > 100 {
		return nil, errors.New("line_pitch_change_percent must be between -20 and 100")
	}
	if input.TargetCapHeightDegrees != 0 && input.BodySizeChangePercent != 0 {
		return nil, errors.New("use target_cap_height_degrees or body_size_change_percent, not both")
	}

	path := filepath.Join(device.Root, "kindle", "reader", "presets", "mac", input.Preset+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read preset %s: %w", input.Preset, err)
	}
	var preset struct {
		Calibration struct {
			ViewingDistance float64 `json:"viewing_distance_inches"`
			PanelPPI        float64 `json:"panel_pixels_per_inch"`
		} `json:"calibration"`
		Fonts struct {
			Narrative typeLabFace `json:"narrative"`
		} `json:"fonts"`
	}
	if err := json.Unmarshal(raw, &preset); err != nil {
		return nil, fmt.Errorf("parse preset %s: %w", input.Preset, err)
	}
	face := preset.Fonts.Narrative
	if err := validateTypeLabFace(face); err != nil {
		return nil, fmt.Errorf("preset %s: %w", input.Preset, err)
	}
	distance := input.ViewingDistanceInches
	if distance == 0 {
		distance = preset.Calibration.ViewingDistance
	}
	if distance < 6 || distance > 60 {
		return nil, errors.New("viewing_distance_inches must be between 6 and 60")
	}
	if preset.Calibration.PanelPPI <= 0 {
		return nil, errors.New("preset panel PPI must be positive")
	}

	currentPitch, currentGap := lineGeometry(face, face.FBInkPixels)
	candidatePixels := int(math.Round(float64(face.FBInkPixels) * (1 + input.BodySizeChangePercent/100)))
	if input.TargetCapHeightDegrees != 0 {
		if input.TargetCapHeightDegrees < 0.15 || input.TargetCapHeightDegrees > 1.5 {
			return nil, errors.New("target_cap_height_degrees must be between 0.15 and 1.5")
		}
		candidatePixels = pixelsForAngle(input.TargetCapHeightDegrees, distance, preset.Calibration.PanelPPI, face)
	}
	if candidatePixels < 8 {
		return nil, errors.New("candidate body size is below 8 pixels")
	}

	targetPitch := int(math.Round(float64(currentPitch) * (1 + input.LinePitchChangePercent/100)))
	baseHeight := unspacedHeight(face, candidatePixels)
	requestedGap := targetPitch - baseHeight
	if requestedGap < 0 {
		requestedGap = 0
	}
	lineGapUnits := minimumLineGapUnits(requestedGap, candidatePixels, face)
	candidateFace := face
	candidateFace.LineGapUnits = lineGapUnits
	candidatePitch, candidateGap := lineGeometry(candidateFace, candidatePixels)
	angle := achievedTypeAngle(candidatePixels, distance, preset.Calibration.PanelPPI, face)

	return map[string]any{
		"ok":     true,
		"preset": input.Preset,
		"current": map[string]any{
			"body_fbink_pixels":     face.FBInkPixels,
			"line_gap_pixels":       currentGap,
			"baseline_pitch_pixels": currentPitch,
		},
		"candidate": map[string]any{
			"body_fbink_pixels":             candidatePixels,
			"line_gap_pixels":               candidateGap,
			"baseline_pitch_pixels":         candidatePitch,
			"achieved_cap_height_degrees":   angle,
			"viewing_distance_inches":       distance,
			"above_nasa_025_degree_minimum": angle >= 0.25,
		},
		"actual_change_percent": map[string]any{
			"body_size":  100 * (float64(candidatePixels)/float64(face.FBInkPixels) - 1),
			"line_pitch": 100 * (float64(candidatePitch)/float64(currentPitch) - 1),
		},
		"recipe": map[string]any{
			"narrative_fbink_pixels": candidatePixels,
			"hhea_line_gap_units":    lineGapUnits,
			"source_preset":          input.Preset,
		},
	}, nil
}

func validateTypeLabFace(face typeLabFace) error {
	if face.AscentUnits <= face.DescentUnits || face.CapHeightUnits <= 0 || face.FBInkPixels <= 0 {
		return errors.New("narrative font metrics are incomplete")
	}
	return nil
}

func unspacedHeight(face typeLabFace, pixels int) int {
	vertical := float64(face.AscentUnits - face.DescentUnits)
	baseline := int(math.Ceil(float64(pixels) * float64(face.AscentUnits) / vertical))
	descent := int(math.Ceil(float64(pixels) * float64(face.DescentUnits) / vertical))
	return baseline + int(math.Abs(float64(descent)))
}

func lineGeometry(face typeLabFace, pixels int) (pitch, gap int) {
	vertical := float64(face.AscentUnits - face.DescentUnits)
	gap = int(math.Ceil(float64(pixels) * float64(face.LineGapUnits) / vertical))
	return unspacedHeight(face, pixels) + gap, gap
}

func minimumLineGapUnits(gapPixels, pixels int, face typeLabFace) int {
	if gapPixels <= 0 {
		return 0
	}
	vertical := face.AscentUnits - face.DescentUnits
	// Smallest integer whose scaled ceiling reaches gapPixels.
	return int(math.Floor(float64((gapPixels-1)*vertical)/float64(pixels))) + 1
}

func pixelsForAngle(angle, distance, ppi float64, face typeLabFace) int {
	capPixels := 2 * distance * ppi * math.Tan(angle*math.Pi/360)
	vertical := float64(face.AscentUnits - face.DescentUnits)
	return int(math.Ceil(capPixels * vertical / float64(face.CapHeightUnits)))
}

func achievedTypeAngle(pixels int, distance, ppi float64, face typeLabFace) float64 {
	vertical := float64(face.AscentUnits - face.DescentUnits)
	capInches := (float64(pixels) * float64(face.CapHeightUnits) / vertical) / ppi
	return 2 * math.Atan(capInches/(2*distance)) * 180 / math.Pi
}
