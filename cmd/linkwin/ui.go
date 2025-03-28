package main

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"grono.dev/winbt/bt"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.ANSIColor(0)).
			Background(lipgloss.ANSIColor(7)).
			Padding(0, 1)
	deviceStyle = lipgloss.NewStyle().
			Padding(0, 2)
	numberStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Bold(true)
	macAddrStyle = lipgloss.NewStyle().
			Italic(true).
			Faint(true)
	infoStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.ANSIColor(4))
)

// renderDeviceList creates a styled list of devices
func renderDeviceList(devices []bt.Device) {
	var s strings.Builder

	s.WriteString(titleStyle.Render(" Select Bluetooth Device "))
	s.WriteString("\n\n")

	for i, device := range devices {
		s.WriteString(numberStyle.Render(strconv.Itoa(i)))
		s.WriteString(deviceStyle.Render(device.Name))
		s.WriteString(macAddrStyle.Render("(" + device.Mac + ")"))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(infoStyle.Render("Enter the number of the device to select, or 'q' to cancel"))

	lipgloss.Println(s.String())
}

func renderControllerList(controllers []bt.Controller) {
	var s strings.Builder

	s.WriteString(titleStyle.Render(" Select Bluetooth Controller "))
	s.WriteString("\n\n")

	for i, controller := range controllers {
		s.WriteString(numberStyle.Render(strconv.Itoa(i)))
		s.WriteString(macAddrStyle.Render(controller.Mac))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(infoStyle.Render("Enter the number of the controller to select, or 'q' to cancel"))

	lipgloss.Println(s.String())
}
