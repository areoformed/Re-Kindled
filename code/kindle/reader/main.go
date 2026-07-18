package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const pageSeparator = "\n---PAGE---\n"

type Document struct {
	Title string
	Pages []string
}

func main() {
	contentPath := flag.String("content", "/mnt/us/rekindled/content.txt", "path to form-paginated UTF-8 text")
	inputPath := flag.String("input", "/dev/input/event1", "evdev touchscreen path")
	fbinkPath := flag.String("fbink", "/mnt/us/libkh/bin/fbink", "FBInk executable path")
	presetPath := flag.String("preset", "/mnt/us/rekindled/presets/mac/rekindled-mono-air.json", "path to a typography preset")
	viewDistance := flag.Float64("view-distance-in", 0, "override preset viewing distance and recalibrate type")
	checkLayout := flag.Bool("check-layout", false, "preflight every page without drawing or opening touch input")
	debugInput := flag.Bool("debug-input", false, "log raw evdev events")
	flag.Parse()

	document, err := loadDocument(*contentPath)
	if err != nil {
		log.Fatalf("load content: %v", err)
	}

	preset, err := loadTypographyPreset(*presetPath)
	if err != nil {
		log.Fatalf("load typography preset: %v", err)
	}
	if *viewDistance > 0 {
		if err := preset.CalibrateViewingDistance(*viewDistance); err != nil {
			log.Fatalf("calibrate typography preset: %v", err)
		}
	}

	renderer := Renderer{FBInk: *fbinkPath, Preset: preset}
	if *checkLayout {
		for index, page := range document.Pages {
			if err := renderer.preflight(page); err != nil {
				log.Fatalf("page %d/%d: %v", index+1, len(document.Pages), err)
			}
		}
		log.Printf("layout valid: %d pages, preset=%s, distance=%.1f in", len(document.Pages), preset.ID, preset.Calibration.ViewingDistanceInches)
		return
	}
	touch, err := openTouch(*inputPath)
	if err != nil {
		log.Fatalf("open touch: %v", err)
	}
	defer touch.Close()

	if err := touch.Grab(true); err != nil {
		log.Fatalf("grab touch: %v", err)
	}
	defer func() {
		if err := resumeKindleUI(); err != nil {
			log.Printf("resume Kindle UI: %v", err)
		}
	}()
	defer touch.Grab(false) //nolint:errcheck -- best-effort emergency cleanup
	if err := holdAwake(); err != nil {
		log.Printf("hold power/touch lease: %v", err)
		return
	}
	defer func() {
		if err := releaseAwake(); err != nil {
			log.Printf("release power lease: %v", err)
		}
	}()

	if err := renderer.Page(document, 0); err != nil {
		log.Printf("render first page: %v", err)
		return
	}

	events := make(chan InputEvent, 128)
	errs := make(chan error, 1)
	go func() {
		errs <- touch.Read(events)
	}()

	signals := make(chan os.Signal, 8)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2)

	page := 0
	gestures := NewGestureDetector()
	log.Printf("reader active: %d pages, input=%s", len(document.Pages), *inputPath)

	renderPage := func(candidate int) {
		if candidate < 0 || candidate >= len(document.Pages) {
			return
		}
		if err := renderer.Page(document, candidate); err != nil {
			log.Printf("render page %d: %v", candidate+1, err)
			return
		}
		page = candidate
	}

	for {
		select {
		case event := <-events:
			if *debugInput {
				log.Printf("input type=0x%02x code=0x%02x value=%d", event.Type, event.Code, event.Value)
			}
			action := gestures.Handle(event)
			switch action {
			case ActionNext:
				if page < len(document.Pages)-1 {
					log.Printf("next page: %d/%d", page+2, len(document.Pages))
					renderPage(page + 1)
				}
			case ActionPrevious:
				if page > 0 {
					log.Printf("previous page: %d/%d", page, len(document.Pages))
					renderPage(page - 1)
				}
			case ActionExit:
				log.Printf("exit gesture received")
				renderer.Exit() //nolint:errcheck -- release input even if drawing fails
				return
			}
		case err := <-errs:
			if err != nil && !errors.Is(err, os.ErrClosed) {
				log.Printf("read touch: %v", err)
			}
			return
		case sig := <-signals:
			log.Printf("received %s", sig)
			switch sig {
			case syscall.SIGHUP:
				candidate, err := loadDocument(*contentPath)
				if err != nil {
					log.Printf("reload content: %v", err)
					continue
				}
				if err := preflightDocument(renderer, candidate); err != nil {
					log.Printf("reload rejected: %v", err)
					continue
				}
				if err := renderer.Page(candidate, 0); err != nil {
					log.Printf("reload render: %v", err)
					continue
				}
				document = candidate
				page = 0
				log.Printf("content reloaded: %d pages", len(document.Pages))
			case syscall.SIGUSR1:
				renderPage(page + 1)
			case syscall.SIGUSR2:
				renderPage(page - 1)
			default:
				renderer.Exit() //nolint:errcheck -- release input even if drawing fails
				return
			}
		}
	}
}

func loadDocument(path string) (Document, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Document{}, err
	}
	return parseDocument(string(raw), path)
}

func parseDocument(raw, source string) (Document, error) {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	title := ""
	if firstLine, rest, found := strings.Cut(normalized, "\n"); found && strings.HasPrefix(firstLine, "@title:") {
		title = strings.TrimSpace(strings.TrimPrefix(firstLine, "@title:"))
		normalized = rest
		if title == "" {
			return Document{}, fmt.Errorf("empty @title in %s", source)
		}
	}

	parts := strings.Split(normalized, pageSeparator)
	pages := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			pages = append(pages, part)
		}
	}
	if len(pages) == 0 {
		return Document{}, fmt.Errorf("no pages found in %s", source)
	}
	return Document{Title: title, Pages: pages}, nil
}

func preflightDocument(renderer Renderer, document Document) error {
	for index, page := range document.Pages {
		if err := renderer.preflight(page); err != nil {
			return fmt.Errorf("page %d/%d: %w", index+1, len(document.Pages), err)
		}
	}
	return nil
}
