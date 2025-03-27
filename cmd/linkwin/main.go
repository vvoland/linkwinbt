package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"grono.dev/winbt/bt"
	"grono.dev/winbt/winreg"
)

var dry = true

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

	btController, err := pickController(ctx)
	if err != nil {
		return err
	}

	btDevice, err := pickDevice(ctx, btController)
	if err != nil {
		return err
	}

	hivePath := parsePath(os.Args[1])

	reg, err := winreg.Open(hivePath)
	if err != nil {
		return err
	}
	defer reg.Close()

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

	// Assume Windows installation directory was given
	if fi.IsDir() {
		return filepath.Join(path, "System32", "config", "SYSTEM")
	}

	return path
}

func pickController(ctx context.Context) (*bt.Controller, error) {
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
		fmt.Println(renderControllerList(controllers))
		idx, err := getUserSelection("Enter controller number (or q to quit): ", len(controllers))
		if err != nil {
			return nil, err
		}
		return &controllers[idx], nil
	}
}

func pickDevice(ctx context.Context, controller *bt.Controller) (*bt.Device, error) {
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
		fmt.Println(renderDeviceList(devices))
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
	fmt.Scanln(&input)

	if input == "q" || input == "Q" {
		return -1, errors.New("selection canceled")
	}

	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > itemCount {
		return -1, fmt.Errorf("invalid selection: %s", input)
	}

	return idx, nil
}
