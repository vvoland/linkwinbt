// SPDX-License-Identifier: BSD-3-Clause

package bt

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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
			for line := range strings.SplitSeq(string(info), "\n") {
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
		return errors.New("device not associated with a controller")
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

	if !stat.Mode().IsRegular() {
		return errors.New("device info file is not a regular file")
	}

	owner, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return errors.New("failed to get device info file ownership")
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
		return errors.New("link key not found in device info file")
	}

	updatedInfo := strings.Join(lines, "\n")
	tmpName := "info.tmp-" + rand.Text()

	fw, err := root.OpenFile(tmpName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create updated device info file: %w", err)
	}

	defer func() {
		_ = fw.Close()
		_ = root.Remove(tmpName)
	}()

	if _, err := fw.WriteString(updatedInfo); err != nil {
		return fmt.Errorf("failed to write updated device info file: %w", err)
	}

	if err := fw.Chown(int(owner.Uid), int(owner.Gid)); err != nil {
		return fmt.Errorf("failed to preserve device info file ownership: %w", err)
	}

	if err := fw.Chmod(stat.Mode()); err != nil {
		return fmt.Errorf("failed to preserve device info file permissions: %w", err)
	}

	if err := fw.Sync(); err != nil {
		return fmt.Errorf("failed to sync updated device info file: %w", err)
	}

	if err := fw.Close(); err != nil {
		return fmt.Errorf("failed to close updated device info file: %w", err)
	}

	current, err := root.Lstat("info")
	if err != nil {
		return fmt.Errorf("failed to check device info file before replacement: %w", err)
	}

	if !os.SameFile(stat, current) {
		return errors.New("device info file changed during update or is a symbolic link")
	}

	if err := root.Rename(tmpName, "info"); err != nil {
		return fmt.Errorf("failed to replace device info file: %w", err)
	}

	dir, err := root.Open(".")
	if err != nil {
		return fmt.Errorf("device info file updated, but failed to open directory for syncing: %w", err)
	}
	defer func() { _ = dir.Close() }()

	if err := dir.Sync(); err != nil {
		return fmt.Errorf("device info file updated, but failed to sync directory: %w", err)
	}

	return nil
}
