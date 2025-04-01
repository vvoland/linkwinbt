package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"grono.dev/linkwinbt/bt"
	"grono.dev/linkwinbt/internal/render"
	"grono.dev/linkwinbt/winreg"
)

var dry = false

func main() {
	ctx := context.Background()
	err := run(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	if err := winreg.Check(); err != nil {
		return err
	}
	if len(os.Args) < 2 {
		return errors.New("usage: go run main.go <windows-dir or SYSTEM file path>")
	}

	flag.BoolVar(&dry, "dry", false, "Dry mode - only print the extracted link key (default: false)")
	flag.Parse()

	hivePath := parsePath(os.Args[1])
	reg, err := winreg.Open(hivePath)
	if err != nil {
		return err
	}

	btController, err := pickController()
	if err != nil {
		return err
	}

	btDevice, err := pickDevice(btController)
	if err != nil {
		return err
	}

	linkKeyString, err := reg.GetBluetoothLinkKey(btController.Mac, btDevice.Mac)
	if err != nil {
		return err
	}

	linkKey, err := bt.ParseLinkKey(linkKeyString)
	if err != nil {
		return err
	}

	if dry {
		fmt.Println("Link key:", linkKey)
		return nil
	}

	err = btDevice.SetLinkKey(linkKey)
	if err != nil {
		return err
	}

	err = bt.Restart(ctx)
	if err != nil {
		return err
	}

	return nil
}

func parsePath(path string) string {
	fi, err := os.Lstat(path)
	if err != nil {
		return ""
	}

	// Assume a path inside a Windows installation directory was given
	if fi.IsDir() {
		// Full path is C:\Windows\System32\config\SYSTEM
		// Handle cases where C:\, C:\Windows, C:\Windows\System32 was provided.
		for _, p := range []string{
			filepath.Join("config", "SYSTEM"),
			filepath.Join("System32", "config", "SYSTEM"),
		} {
			if _, err := os.Lstat(filepath.Join(path, p)); err == nil {
				return filepath.Join(path, p)
			}
		}

		return filepath.Join(path, "Windows", "System32", "config", "SYSTEM")
	}

	return path
}

func pickController() (*bt.Controller, error) {
	controllers, err := bt.Controllers()
	if err != nil {
		return nil, err
	}

	switch len(controllers) {
	case 0:
		return nil, fmt.Errorf("no controllers found")
	case 1:
		return &controllers[0], nil
	default:
		render.ControllerList(controllers)
		idx, err := getUserSelection("Enter controller number (or q to quit): ", len(controllers))
		if err != nil {
			return nil, err
		}
		return &controllers[idx], nil
	}
}

func pickDevice(controller *bt.Controller) (*bt.Device, error) {
	devices, err := controller.Devices()
	if err != nil {
		return nil, err
	}

	switch len(devices) {
	case 0:
		return nil, fmt.Errorf("no devices found")
	case 1:
		return &devices[0], nil
	default:
		render.DeviceList(devices)
		idx, err := getUserSelection("Enter device number (or q to quit): ", len(devices))
		if err != nil {
			return nil, err
		}
		return &devices[idx], nil
	}
}

func getUserSelection(prompt string, itemCount int) (int, error) {
	fmt.Print(prompt)
	var input string

	if _, err := fmt.Scanln(&input); err != nil {
		return -1, err
	}

	if input == "q" || input == "Q" {
		return -1, errors.New("selection canceled")
	}

	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > itemCount {
		return -1, fmt.Errorf("invalid selection: %s", input)
	}

	return idx, nil
}
