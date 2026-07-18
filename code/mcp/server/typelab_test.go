package main

import (
	"math"
	"testing"
)

func TestTypeLabGeometryMatchesPhysicalTrial(t *testing.T) {
	face := typeLabFace{AscentUnits: 885, DescentUnits: -235, LineGapUnits: 0, CapHeightUnits: 700, FBInkPixels: 52}
	currentPitch, _ := lineGeometry(face, 52)
	if currentPitch != 52 {
		t.Fatalf("current pitch = %d, want 52", currentPitch)
	}
	newPixels := 49
	targetPitch := int(math.Round(float64(currentPitch) * 1.10))
	gapUnits := minimumLineGapUnits(targetPitch-unspacedHeight(face, newPixels), newPixels, face)
	face.LineGapUnits = gapUnits
	pitch, gap := lineGeometry(face, newPixels)
	if pitch != 57 || gap != 8 {
		t.Fatalf("candidate = %d px pitch / %d px gap, want 57 / 8", pitch, gap)
	}
}
