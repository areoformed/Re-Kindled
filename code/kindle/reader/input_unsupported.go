//go:build !linux

package main

import (
	"fmt"
)

type TouchDevice struct{}

func openTouch(path string) (*TouchDevice, error) {
	return nil, fmt.Errorf("evdev input is only available on Linux: %s", path)
}

func (device *TouchDevice) Grab(bool) error              { return nil }
func (device *TouchDevice) Read(chan<- InputEvent) error { return nil }
func (device *TouchDevice) Close() error                 { return nil }
