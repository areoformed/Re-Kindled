package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func holdAwake() error {
	if err := setLIPCInt("com.lab126.powerd", "preventScreenSaver", "1"); err != nil {
		return err
	}
	state, err := powerState()
	if err != nil {
		releaseAwake() //nolint:errcheck -- preserve the original error
		return err
	}
	if state != "active" {
		if err := setLIPCInt("com.lab126.powerd", "wakeUp", "1"); err != nil {
			// powerd rejects wakeUp when a concurrent transition already made the
			// device active. Recheck before treating that race as a failure.
			if current, stateErr := powerState(); stateErr != nil || current != "active" {
				releaseAwake() //nolint:errcheck -- preserve the original error
				return err
			}
		}
	}
	if err := setLIPCInt("com.lab126.deviced", "enable_touch", "1"); err != nil {
		releaseAwake() //nolint:errcheck -- preserve the original error
		return err
	}
	return nil
}

func releaseAwake() error {
	return setLIPCInt("com.lab126.powerd", "preventScreenSaver", "0")
}

func resumeKindleUI() error {
	output, err := exec.Command("killall", "-CONT", "awesome").CombinedOutput()
	if err != nil {
		return fmt.Errorf("resume awesome: %w: %s", err, output)
	}
	return nil
}

func powerState() (string, error) {
	output, err := exec.Command("lipc-get-prop", "com.lab126.powerd", "state").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("read powerd state: %w: %s", err, output)
	}
	return strings.ToLower(strings.TrimSpace(string(output))), nil
}

func setLIPCInt(component, property, value string) error {
	output, err := exec.Command("lipc-set-prop", "-i", component, property, value).CombinedOutput()
	if err != nil {
		return fmt.Errorf("lipc-set-prop %s %s: %w: %s", component, property, err, output)
	}
	return nil
}
