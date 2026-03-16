package render

import (
	"fmt"
	"strconv"
	"strings"

	"grono.dev/linkwinbt/bt"
)

const (
	ansiReset   = "\x1b[0m"
	ansiBold    = "\x1b[1m"
	ansiItalic  = "\x1b[3m"
	ansiFaint   = "\x1b[2m"
	ansiFgBlack = "\x1b[30m"
	ansiFgBlue  = "\x1b[34m"
	ansiBgWhite = "\x1b[47m"
)

func style(s string, codes ...string) string {
	if len(codes) == 0 {
		return s
	}

	return strings.Join(codes, "") + s + ansiReset
}

func bold(s string) string {
	return style(s, ansiBold)
}

func italic(s string) string {
	return style(s, ansiItalic)
}

func faint(s string) string {
	return style(s, ansiFaint)
}

func blue(s string) string {
	return style(s, ansiFgBlue)
}

func black(s string) string {
	return style(s, ansiFgBlack)
}

func bgWhite(s string) string {
	return style(s, ansiBgWhite)
}

// DeviceList prints a styled list of devices
func DeviceList(devices []bt.Device) {
	var s strings.Builder

	s.WriteString(" ")
	s.WriteString(bgWhite(black(bold(" Select Bluetooth Device "))))
	s.WriteString(" ")
	s.WriteString("\n\n")

	for i, device := range devices {
		s.WriteString("  ")
		s.WriteString(bold(strconv.Itoa(i)))
		s.WriteString("  ")
		s.WriteString(device.Name)
		s.WriteString(faint(italic("(" + device.Mac + ")")))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(blue(italic("Enter the number of the device to select, or 'q' to cancel")))

	_, _ = fmt.Println(s.String())
}

// ControllerList prints a styled list of controllers
func ControllerList(controllers []bt.Controller) {
	var s strings.Builder

	s.WriteString(" ")
	s.WriteString(bgWhite(black(bold(" Select Bluetooth Controller "))))
	s.WriteString(" ")
	s.WriteString("\n\n")

	for i, controller := range controllers {
		s.WriteString("  ")
		s.WriteString(bold(strconv.Itoa(i)))
		s.WriteString("  ")
		s.WriteString(faint(italic(controller.Mac)))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(blue(italic("Enter the number of the controller to select, or 'q' to cancel")))

	_, _ = fmt.Println(s.String())
}
