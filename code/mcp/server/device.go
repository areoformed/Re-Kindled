package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	remoteBase        = "/mnt/us/rekindled"
	remoteIncoming    = remoteBase + "/content.next"
	remoteContent     = remoteBase + "/content.txt"
	maxDocumentBytes  = 256 * 1024
	maxPageBytes      = 24 * 1024
	maxDocumentPages  = 64
	maxDocumentTitle  = 120
	documentSeparator = "\n---PAGE---\n"
)

var presetIDPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

type SSHDevice struct {
	Root      string
	SSHConfig string
	Host      string
	Timeout   time.Duration
}

type ShowInput struct {
	Title  string   `json:"title"`
	Pages  []string `json:"pages"`
	Preset string   `json:"preset"`
}

type PresetSummary struct {
	ID                    string  `json:"id"`
	Label                 string  `json:"label"`
	Category              string  `json:"category,omitempty"`
	BodyFBInkPixels       int     `json:"body_fbink_pixels"`
	BodyCapHeightDegrees  float64 `json:"body_cap_height_degrees"`
	ViewingDistanceInches float64 `json:"viewing_distance_inches"`
	AchievedAngleDegrees  float64 `json:"achieved_angle_degrees,omitempty"`
	LineGapPixels         int     `json:"line_gap_pixels,omitempty"`
	LineSpacingMultiplier float64 `json:"line_spacing_multiplier,omitempty"`
	Note                  string  `json:"note,omitempty"`
}

func (device *SSHDevice) Show(ctx context.Context, input ShowInput) (map[string]any, error) {
	document, err := encodeDocument(input)
	if err != nil {
		return nil, err
	}
	if input.Preset != "" {
		if err := validatePresetID(input.Preset); err != nil {
			return nil, err
		}
	}

	temporary, err := os.CreateTemp("", "rekindled-document-*.txt")
	if err != nil {
		return nil, fmt.Errorf("create temporary document: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath) //nolint:errcheck -- best effort for a private temporary file
	if _, err := temporary.WriteString(document); err != nil {
		temporary.Close() //nolint:errcheck -- the write error is primary
		return nil, fmt.Errorf("write temporary document: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return nil, fmt.Errorf("close temporary document: %w", err)
	}

	if _, err := device.scp(ctx, temporaryPath, device.Host+":"+remoteIncoming); err != nil {
		return nil, err
	}

	presetSelection := `PRESET=$(cat "$BASE/active-preset")`
	activate := `mv "$BASE/content.next" "$BASE/content.txt"; if [ -f "$BASE/reader.pid" ] && kill -0 "$(cat "$BASE/reader.pid")" 2>/dev/null; then kill -HUP "$(cat "$BASE/reader.pid")"; else "$BASE/launch-reader.sh"; fi`
	if input.Preset != "" {
		presetPath := remoteBase + "/presets/mac/" + input.Preset + ".json"
		presetSelection = `PRESET="` + presetPath + `"; test -f "$PRESET"`
		activate = `printf '%s\n' "$PRESET" > "$BASE/active-preset"; mv "$BASE/content.next" "$BASE/content.txt"; "$BASE/stop-reader.sh"; "$BASE/launch-reader.sh"`
	}
	command := `set -eu; BASE=` + remoteBase + `; ` + presetSelection + `; test -n "$PRESET"; "$BASE/bin/rekindled-reader" -check-layout -content "$BASE/content.next" -preset "$PRESET"; ` + activate
	if _, err := device.ssh(ctx, command); err != nil {
		return nil, fmt.Errorf("display rejected before visible swap: %w", err)
	}

	result := map[string]any{
		"ok":             true,
		"title":          strings.TrimSpace(input.Title),
		"pages":          len(input.Pages),
		"document_bytes": len(document),
	}
	if input.Preset != "" {
		result["preset"] = input.Preset
	}
	return result, nil
}

func (device *SSHDevice) Navigate(ctx context.Context, direction string) (map[string]any, error) {
	var signal string
	switch direction {
	case "next":
		signal = "USR1"
	case "previous":
		signal = "USR2"
	default:
		return nil, errors.New(`direction must be "next" or "previous"`)
	}
	command := `set -eu; BASE=` + remoteBase + `; test -f "$BASE/reader.pid"; PID=$(cat "$BASE/reader.pid"); kill -0 "$PID"; kill -` + signal + ` "$PID"`
	if _, err := device.ssh(ctx, command); err != nil {
		return nil, fmt.Errorf("navigate %s: %w", direction, err)
	}
	return map[string]any{"ok": true, "direction": direction}, nil
}

func (device *SSHDevice) SetType(ctx context.Context, preset string) (map[string]any, error) {
	if err := validatePresetID(preset); err != nil {
		return nil, err
	}
	presetPath := remoteBase + "/presets/mac/" + preset + ".json"
	command := `set -eu; BASE=` + remoteBase + `; PRESET="` + presetPath + `"; test -f "$PRESET"; "$BASE/bin/rekindled-reader" -check-layout -content "$BASE/content.txt" -preset "$PRESET"; printf '%s\n' "$PRESET" > "$BASE/active-preset"; "$BASE/stop-reader.sh"; "$BASE/launch-reader.sh"`
	if _, err := device.ssh(ctx, command); err != nil {
		return nil, fmt.Errorf("select preset %s: %w", preset, err)
	}
	return map[string]any{"ok": true, "preset": preset}, nil
}

func (device *SSHDevice) Status(ctx context.Context) (map[string]any, error) {
	command := `BASE=` + remoteBase + `
PID=""
if [ -f "$BASE/reader.pid" ]; then PID=$(cat "$BASE/reader.pid"); fi
printf 'pid\t%s\n' "$PID"
if [ -n "$PID" ] && kill -0 "$PID" 2>/dev/null; then printf 'reader_running\ttrue\n'; else printf 'reader_running\tfalse\n'; fi
printf 'active_preset\t'
if [ -f "$BASE/active-preset" ]; then cat "$BASE/active-preset"; else printf '\n'; fi
printf 'reader_binary_bytes\t'
if [ -f "$BASE/bin/rekindled-reader" ]; then wc -c < "$BASE/bin/rekindled-reader"; else printf '0\n'; fi
TOTAL=0
for FILE in "$BASE"/fonts/*; do
  if [ -f "$FILE" ]; then SIZE=$(wc -c < "$FILE"); TOTAL=$((TOTAL + SIZE)); fi
done
printf 'font_library_bytes\t%s\n' "$TOTAL"
printf 'log_begin\n'
if [ -f "$BASE/reader.log" ]; then tail -n 8 "$BASE/reader.log"; fi`
	output, err := device.ssh(ctx, command)
	status := map[string]any{
		"connected": false,
		"host":      device.Host,
	}
	if executable, statErr := os.Executable(); statErr == nil {
		if info, statErr := os.Stat(executable); statErr == nil {
			status["mcp_binary_bytes"] = info.Size()
		}
	}
	if err != nil {
		status["error"] = err.Error()
		return status, nil
	}
	status["connected"] = true
	parseStatusOutput(status, output)
	return status, nil
}

func (device *SSHDevice) Presets() ([]PresetSummary, error) {
	paths, err := filepath.Glob(filepath.Join(device.Root, "kindle", "reader", "presets", "mac", "*.json"))
	if err != nil {
		return nil, fmt.Errorf("find preset files: %w", err)
	}
	presets := make([]PresetSummary, 0, len(paths))
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		var source struct {
			Label       string `json:"label"`
			Calibration struct {
				BodyAngle float64 `json:"body_cap_height_degrees"`
				Distance  float64 `json:"viewing_distance_inches"`
			} `json:"calibration"`
			Fonts struct {
				Narrative struct {
					Pixels int `json:"fbink_pixels"`
				} `json:"narrative"`
			} `json:"fonts"`
			Metadata struct {
				Category      string  `json:"category"`
				AchievedAngle float64 `json:"body_achieved_angle_degrees"`
				LineGapPixels int     `json:"body_line_gap_pixels"`
				LineSpacing   float64 `json:"body_line_spacing_multiplier"`
				Note          string  `json:"note"`
			} `json:"local_font_metadata"`
		}
		if err := json.Unmarshal(raw, &source); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		presets = append(presets, PresetSummary{
			ID:                    strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
			Label:                 source.Label,
			Category:              source.Metadata.Category,
			BodyFBInkPixels:       source.Fonts.Narrative.Pixels,
			BodyCapHeightDegrees:  source.Calibration.BodyAngle,
			ViewingDistanceInches: source.Calibration.Distance,
			AchievedAngleDegrees:  source.Metadata.AchievedAngle,
			LineGapPixels:         source.Metadata.LineGapPixels,
			LineSpacingMultiplier: source.Metadata.LineSpacing,
			Note:                  source.Metadata.Note,
		})
	}
	sort.Slice(presets, func(i, j int) bool {
		if presets[i].Category != presets[j].Category {
			return presets[i].Category < presets[j].Category
		}
		return presets[i].ID < presets[j].ID
	})
	return presets, nil
}

func encodeDocument(input ShowInput) (string, error) {
	title := strings.TrimSpace(input.Title)
	titleCharacters := utf8.RuneCountInString(title)
	if titleCharacters > maxDocumentTitle {
		return "", fmt.Errorf("title is %d characters; maximum is %d", titleCharacters, maxDocumentTitle)
	}
	if strings.ContainsAny(title, "\r\n") {
		return "", errors.New("title cannot contain a line break")
	}
	if len(input.Pages) == 0 || len(input.Pages) > maxDocumentPages {
		return "", fmt.Errorf("pages must contain 1 to %d items", maxDocumentPages)
	}
	pages := make([]string, len(input.Pages))
	for index, page := range input.Pages {
		page = strings.TrimSpace(strings.ReplaceAll(page, "\r\n", "\n"))
		if page == "" {
			return "", fmt.Errorf("page %d is empty", index+1)
		}
		if len(page) > maxPageBytes {
			return "", fmt.Errorf("page %d is %d bytes; maximum is %d", index+1, len(page), maxPageBytes)
		}
		if strings.Contains(page, documentSeparator) {
			return "", fmt.Errorf("page %d contains the reserved page separator", index+1)
		}
		pages[index] = page
	}
	var document strings.Builder
	if title != "" {
		document.WriteString("@title: ")
		document.WriteString(title)
		document.WriteByte('\n')
	}
	document.WriteString(strings.Join(pages, documentSeparator))
	document.WriteByte('\n')
	if document.Len() > maxDocumentBytes {
		return "", fmt.Errorf("document is %d bytes; maximum is %d", document.Len(), maxDocumentBytes)
	}
	return document.String(), nil
}

func validatePresetID(preset string) error {
	if !presetIDPattern.MatchString(preset) {
		return errors.New("preset must contain only lowercase letters, numbers, and hyphens")
	}
	return nil
}

func parseStatusOutput(status map[string]any, output string) {
	lines := strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
	logLines := []string{}
	inLog := false
	for _, line := range lines {
		if line == "log_begin" {
			inLog = true
			continue
		}
		if inLog {
			if line != "" {
				logLines = append(logLines, line)
			}
			continue
		}
		key, value, found := strings.Cut(line, "\t")
		if !found {
			continue
		}
		value = strings.TrimSpace(value)
		switch key {
		case "reader_running":
			status[key] = value == "true"
		case "reader_binary_bytes", "font_library_bytes":
			if number, err := strconv.ParseInt(value, 10, 64); err == nil {
				status[key] = number
			}
		case "active_preset":
			status[key] = strings.TrimSuffix(filepath.Base(value), filepath.Ext(value))
		case "pid":
			if number, err := strconv.Atoi(value); err == nil {
				status[key] = number
			}
		}
	}
	status["recent_log"] = logLines
}

func (device *SSHDevice) ssh(parent context.Context, command string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, device.Timeout)
	defer cancel()
	args := []string{"-F", device.SSHConfig, device.Host, command}
	output, err := exec.CommandContext(ctx, "ssh", args...).CombinedOutput()
	if err != nil {
		return "", commandError("ssh", ctx, err, output)
	}
	return strings.TrimSpace(string(output)), nil
}

func (device *SSHDevice) scp(parent context.Context, source, destination string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, device.Timeout)
	defer cancel()
	args := []string{"-q", "-F", device.SSHConfig, source, destination}
	output, err := exec.CommandContext(ctx, "scp", args...).CombinedOutput()
	if err != nil {
		return "", commandError("scp", ctx, err, output)
	}
	return strings.TrimSpace(string(output)), nil
}

func commandError(name string, ctx context.Context, err error, output []byte) error {
	if ctx.Err() != nil {
		return fmt.Errorf("%s timed out: %w", name, ctx.Err())
	}
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return fmt.Errorf("%s: %w", name, err)
	}
	return fmt.Errorf("%s: %w: %s", name, err, detail)
}
