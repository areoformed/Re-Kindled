package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type Renderer struct {
	FBInk  string
	Preset TypographyPreset
}

func (renderer Renderer) Page(document Document, index int) error {
	if index < 0 || index >= len(document.Pages) {
		return fmt.Errorf("page index %d outside 0..%d", index, len(document.Pages)-1)
	}

	if err := renderer.preflight(document.Pages[index]); err != nil {
		return err
	}

	preset := renderer.Preset
	layout := preset.Layout
	headerSpec := fontSpec(
		preset.Fonts.Utility,
		layout.HeaderTopPixels,
		layout.HeaderBottomPixels,
		layout.HeaderSidePixels,
		layout.HeaderSidePixels,
	)
	bodySpec := fontSpec(
		preset.Fonts.Narrative,
		layout.BodyTopPixels,
		layout.BodyBottomPixels,
		layout.BodyLeftPixels,
		layout.BodyRightPixels,
	)
	footerSpec := fontSpec(
		preset.Fonts.Utility,
		layout.FooterTopPixels,
		layout.FooterBottomPixels,
		layout.FooterSidePixels,
		layout.FooterSidePixels,
	)
	header := preset.Header
	if document.Title != "" {
		header = document.Title
	}
	footer := fmt.Sprintf("Triple-tap left  |  %d / %d  |  Triple-tap right", index+1, len(document.Pages))

	commands := [][]string{
		{"-k", "-b", "-B", "WHITE"},
		{"-q", "-b", "-m", "-t", headerSpec, header},
		{"-q", "-b", "-t", bodySpec + ",notrunc", document.Pages[index]},
		{"-q", "-b", "-m", "-t", footerSpec, footer},
		{"-s", "-f", "-W", "GC16", "-w"},
	}

	for _, args := range commands {
		if output, err := exec.Command(renderer.FBInk, args...).CombinedOutput(); err != nil {
			return fmt.Errorf("fbink %v: %w: %s", args, err, output)
		}
	}
	return nil
}

func (renderer Renderer) Exit() error {
	if output, err := exec.Command(renderer.FBInk, "-k", "-b", "-B", "WHITE").CombinedOutput(); err != nil {
		return fmt.Errorf("clear exit page: %w: %s", err, output)
	}
	face := renderer.Preset.Fonts.Narrative
	if output, err := exec.Command(renderer.FBInk,
		"-q", "-b", "-m", "-M",
		"-t", fontSpec(face, 0, 0, 120, 120),
		"ReKindled reader exited. Tap once to return to Kindle.",
	).CombinedOutput(); err != nil {
		return fmt.Errorf("draw exit page: %w: %s", err, output)
	}
	if output, err := exec.Command(renderer.FBInk, "-s", "-f", "-W", "GC16", "-w").CombinedOutput(); err != nil {
		return fmt.Errorf("refresh exit page: %w: %s", err, output)
	}
	return nil
}

func (renderer Renderer) preflight(text string) error {
	l := renderer.Preset.Layout
	spec := fontSpec(
		renderer.Preset.Fonts.Narrative,
		l.BodyTopPixels,
		l.BodyBottomPixels,
		l.BodyLeftPixels,
		l.BodyRightPixels,
	) + ",compute"
	output, err := exec.Command(renderer.FBInk, "-q", "-l", "-t", spec, text).CombinedOutput()
	if err != nil {
		return fmt.Errorf("fbink layout preflight: %w: %s", err, output)
	}
	result := string(output)
	if !strings.Contains(result, "truncated=") {
		return fmt.Errorf("fbink layout preflight returned an unknown result: %s", result)
	}
	if strings.Contains(result, "truncated=1") {
		return fmt.Errorf("page exceeds the %s body area; split the content at a paragraph boundary", renderer.Preset.ID)
	}
	return nil
}

func fontSpec(face FontFace, top, bottom, left, right int) string {
	return fmt.Sprintf(
		"regular=%s,px=%d,top=%d,bottom=%d,left=%d,right=%d",
		face.Path,
		face.FBInkPixels,
		top,
		bottom,
		left,
		right,
	)
}
