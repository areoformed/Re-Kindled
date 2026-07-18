package main

import (
	"math"
	"path/filepath"
	"testing"
)

func amazonEmberRegular() FontFace {
	return FontFace{
		Path:           "/usr/java/lib/fonts/Amazon-Ember-Regular.ttf",
		AscentUnits:    973,
		DescentUnits:   -281,
		CapHeightUnits: 693,
		FBInkPixels:    69,
	}
}

func TestNASA040DegreeCalibrationAt18Inches(t *testing.T) {
	face := amazonEmberRegular()
	if got := requiredFBInkPixels(0.4, 18, 300, face); got != 69 {
		t.Fatalf("requiredFBInkPixels() = %d, want 69", got)
	}
	if got := achievedVisualAngleDegrees(face, 18, 300); got < 0.4 {
		t.Fatalf("achieved angle = %.6f, want at least 0.4", got)
	}
}

func TestNASA025DegreeTargetCapHeight(t *testing.T) {
	got := capHeightPixelsForAngle(0.25, 18, 300)
	if math.Abs(got-23.561982) > 0.000001 {
		t.Fatalf("cap height = %.9f, want 23.561982", got)
	}
}

func TestPresetRejectsUndersizedBody(t *testing.T) {
	face := amazonEmberRegular()
	face.FBInkPixels = 68
	preset := testPreset(face)
	if err := preset.Validate(); err == nil {
		t.Fatal("Validate() accepted a body size below the 0.4 degree target")
	}
}

func TestViewingDistanceOverrideRecalibratesBothRoles(t *testing.T) {
	preset := testPreset(amazonEmberRegular())
	if err := preset.CalibrateViewingDistance(24); err != nil {
		t.Fatalf("CalibrateViewingDistance(): %v", err)
	}
	if got := preset.Fonts.Narrative.FBInkPixels; got != 91 {
		t.Fatalf("narrative size = %d, want 91", got)
	}
	if got := preset.Fonts.Utility.FBInkPixels; got != 55 {
		t.Fatalf("utility size = %d, want 55", got)
	}
}

func TestGeneratedMacPresets(t *testing.T) {
	paths, err := filepath.Glob("presets/mac/*.json")
	if err != nil {
		t.Fatalf("glob generated presets: %v", err)
	}
	if len(paths) == 0 {
		t.Skip("Mac font presets have not been generated")
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			if _, err := loadTypographyPreset(path); err != nil {
				t.Fatalf("loadTypographyPreset(%q): %v", path, err)
			}
		})
	}
}

func testPreset(narrative FontFace) TypographyPreset {
	return TypographyPreset{
		SchemaVersion: 1,
		ID:            "test",
		Label:         "Test",
		Header:        "Test header",
		Calibration: Calibration{
			StandardReference:       "NASA-STD-3001 Volume 2 Appendix F F.5.1.7",
			ViewingDistanceInches:   18,
			PanelPixelsPerInch:      300,
			PanelWidthPixels:        1236,
			PanelHeightPixels:       1648,
			BodyCapHeightDegrees:    0.4,
			UtilityCapHeightDegrees: 0.25,
		},
		Fonts: PresetFonts{
			Narrative: narrative,
			Utility: FontFace{
				Path:           "/usr/java/lib/fonts/Amazon-Ember-Medium.ttf",
				AscentUnits:    973,
				DescentUnits:   -224,
				CapHeightUnits: 693,
				FBInkPixels:    41,
			},
		},
		Layout: PresetLayout{
			HeaderTopPixels:    55,
			HeaderBottomPixels: 1510,
			HeaderSidePixels:   60,
			BodyTopPixels:      170,
			BodyBottomPixels:   165,
			BodyLeftPixels:     120,
			BodyRightPixels:    120,
			FooterTopPixels:    1525,
			FooterBottomPixels: 45,
			FooterSidePixels:   60,
		},
	}
}
