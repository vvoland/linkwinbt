package bt

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Device struct {
	Name string
	Mac  string

	controller *Controller
}

func (c *Controller) Devices() ([]Device, error) {
	root, err := os.OpenRoot(filepath.Join(varLibBluetooth, c.Mac))
	if err != nil {
		return nil, fmt.Errorf("failed to read devices directory for controller %s: %w", c.Mac, err)
	}
	defer func() { _ = root.Close() }()

	files, err := fs.Glob(root.FS(), "*")
	if err != nil {
		return nil, fmt.Errorf("failed to read devices directory for controller %s: %w", c.Mac, err)
	}

	var devices []Device
	for _, filename := range files {
		f, err := root.Open(filename)
		if err != nil {
			continue
		}
		defer func() { _ = f.Close() }()

		stat, err := f.Stat()
		if err != nil {
			continue
		}
		if stat.IsDir() {
			deviceRoot, err := root.OpenRoot(filename)
			if err != nil {
				return nil, fmt.Errorf("failed to open device directory %s: %w", filename, err)
			}

			infoFile, err := deviceRoot.Open("info")
			if err != nil {
				continue
			}
			defer func() { _ = infoFile.Close() }()

			info, err := io.ReadAll(infoFile)
			if err != nil {
				continue
			}

			var name string
			for _, line := range strings.Split(string(info), "\n") {
				k, v, ok := strings.Cut(line, "=")
				if !ok {
					continue
				}
				k = strings.TrimSpace(k)
				v = strings.TrimSpace(v)

				if k == "Name" {
					name = v
					break
				}
			}

			devices = append(devices, Device{
				Name:       name,
				Mac:        filename,
				controller: c,
			})
		}
	}

	return devices, nil
}

func (d *Device) SetLinkKey(linkKeyHex LinkKey) error {
	if d.controller == nil {
		return fmt.Errorf("device not associated with a controller")
	}

	root, err := os.OpenRoot(filepath.Join(varLibBluetooth, d.controller.Mac, d.Mac))
	if err != nil {
		return fmt.Errorf("failed to open device directory: %w", err)
	}
	defer func() { _ = root.Close() }()

	f, err := root.Open("info")
	if err != nil {
		return fmt.Errorf("failed to open device info file: %w", err)
	}
	defer func() { _ = f.Close() }()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get device info file stats: %w", err)
	}

	// Read the current info file
	infoData, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read device info file: %w", err)
	}

	// Parse the info file content
	lines := strings.Split(string(infoData), "\n")

	// Check if LinkKey entry already exists
	linkKeyFound := false
	inLinkSection := false
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "[LinkKey]" {
			inLinkSection = true
			continue
		}
		if !inLinkSection {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inLinkSection = false
			continue
		}

		k, _, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)

		if k == "Key" {
			// Replace the existing link key
			lines[i] = "Key=" + linkKeyHex.String()
			linkKeyFound = true
			break
		}
	}

	if !linkKeyFound {
		return fmt.Errorf("link key not found in device info file")
	}

	// Write the updated info file
	updatedInfo := strings.Join(lines, "\n")
	fw, err := root.OpenFile("info", os.O_WRONLY|os.O_TRUNC, stat.Mode())
	if err != nil {
		return fmt.Errorf("failed to write updated device info file: %w", err)
	}
	defer func() { _ = fw.Close() }()

	_, err = fw.Write([]byte(updatedInfo))
	if err != nil {
		return fmt.Errorf("failed to write updated device info file: %w", err)
	}

	return nil
}
