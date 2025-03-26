package main

import (
	"context"
	"fmt"
	"os"

	"grono.dev/winbt/bt"
)

func main() {
	ctx := context.Background()
	err := run(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	if len(os.Args) < 2 {
		return fmt.Errorf("Usage: go run main.go <path-to-SYSTEM>")
	}
	systemRegPath := os.Args[1]

	btController, err := pickController(ctx)
	if err != nil {
		return err
	}

	btDevice, err := pickDevice(ctx, btController)
	if err != nil {
		return err
	}

	reg, err := winreg.Open(systemRegPath)
	if err != nil {
		return err
	}
	defer reg.Close()

	linkKey, err := reg.GetBluetoothLinkKey(btController.Mac, btDevice.Mac)
	if err != nil {
		return err
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
	}

	// TODO: pick
	return nil, fmt.Errorf("TODO: pick controller")
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
	}

	// TODO: pick
	return nil, fmt.Errorf("TODO: pick device")
}
