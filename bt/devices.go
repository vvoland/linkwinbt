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
	defer root.Close()

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
		defer f.Close()

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
			defer infoFile.Close()

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
