package bt

import (
	"fmt"
	"io/fs"
	"os"
)

type Controller struct {
	Mac string
}

var varLibBluetooth = "/var/lib/bluetooth"

func Controllers() ([]Controller, error) {
	root, err := os.OpenRoot(varLibBluetooth)
	if err != nil {
		return nil, fmt.Errorf("failed to read bluetooth directory: %w", err)
	}
	defer root.Close()

	files, err := fs.Glob(root.FS(), "*")
	if err != nil {
		return nil, fmt.Errorf("failed to list bluetooth directory: %w", err)
	}

	var controllers []Controller
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
			controllers = append(controllers, Controller{
				Mac: filename,
			})
		}
	}

	return controllers, nil

}
