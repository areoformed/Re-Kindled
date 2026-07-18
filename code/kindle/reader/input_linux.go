//go:build linux

package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"syscall"
)

// Linux _IOW('E', 0x90, int), stable across the 32-bit ARM ABI used here.
const eviocgrab = uintptr(0x40044590)

type TouchDevice struct {
	file *os.File
}

func openTouch(path string) (*TouchDevice, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &TouchDevice{file: file}, nil
}

func (device *TouchDevice) Grab(enabled bool) error {
	value := uintptr(0)
	if enabled {
		value = 1
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, device.file.Fd(), eviocgrab, value)
	if errno != 0 {
		return fmt.Errorf("EVIOCGRAB: %w", errno)
	}
	return nil
}

func (device *TouchDevice) Read(events chan<- InputEvent) error {
	// input_event uses two 32-bit timeval fields under the Kindle ARMv7 ABI.
	var buffer [16]byte
	for {
		if _, err := io.ReadFull(device.file, buffer[:]); err != nil {
			return err
		}
		events <- InputEvent{
			Sec:   int32(binary.LittleEndian.Uint32(buffer[0:4])),
			Usec:  int32(binary.LittleEndian.Uint32(buffer[4:8])),
			Type:  binary.LittleEndian.Uint16(buffer[8:10]),
			Code:  binary.LittleEndian.Uint16(buffer[10:12]),
			Value: int32(binary.LittleEndian.Uint32(buffer[12:16])),
		}
	}
}

func (device *TouchDevice) Close() error {
	return device.file.Close()
}
