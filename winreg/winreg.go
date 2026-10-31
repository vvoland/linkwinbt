// SPDX-License-Identifier: BSD-3-Clause

package winreg

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Check() error {
	_, err := exec.LookPath("reged")
	if err != nil {
		return fmt.Errorf("reged command not found: %w", err)
	}
	return nil
}

// Registry represents a Windows registry hive file
type Registry struct {
	dump bytes.Buffer
}

// Open opens a Windows registry hive file
func Open(path string) (*Registry, error) {
	// Create a copy of the registry file so that we don't make the `reged`
	// operate on the original file just in case.
	copyFile, err := createTmpCopy(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = copyFile.Close()
		_ = os.Remove(copyFile.Name())
	}()

	reg, err := export(copyFile.Name(), `\Select`)
	if err != nil {
		return nil, err
	}

	inSelect := false

	scanner := bufio.NewScanner(bytes.NewReader(reg.dump.Bytes()))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inSelect = strings.HasSuffix(line, `\Select]`)
			continue
		}

		if !inSelect {
			continue
		}

		name, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(name) != `"Current"` {
			continue
		}

		value, ok = strings.CutPrefix(strings.TrimSpace(value), "dword:")
		if !ok {
			return nil, errors.New("invalid Windows control set: Select\\Current is not a DWORD")
		}

		current, err := strconv.ParseUint(value, 16, 32)
		if err != nil || current == 0 {
			return nil, errors.New("invalid Windows control set: Select\\Current must be a nonzero DWORD")
		}

		key := fmt.Sprintf(`\ControlSet%03d\Services\BTHPORT\Parameters\Keys`, current)

		return export(copyFile.Name(), key)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read Windows control set: %w", err)
	}

	return nil, errors.New("Windows control set not found: missing Select\\Current")
}

func export(path, key string) (*Registry, error) {
	cmd := exec.Command("reged", "-x", path, "HKEY_LOCAL_MACHINE\\SYSTEM", key, "/dev/stdout")

	reg := &Registry{}

	var stderr bytes.Buffer
	cmd.Stdout = &reg.dump
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to dump registry key %q: %w, stderr %s", key, err, stderr.String())
	}

	return reg, nil
}

func createTmpCopy(path string) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open registry file: %w", err)
	}
	defer func() { _ = f.Close() }()

	tmpCopy, err := os.CreateTemp("", "system-hive-*.reg")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	if _, err := io.Copy(tmpCopy, f); err != nil {
		_ = tmpCopy.Close()
		_ = os.Remove(tmpCopy.Name())

		return nil, fmt.Errorf("failed to copy registry file: %w", err)
	}

	return tmpCopy, nil
}

// GetBluetoothLinkKey extracts the Bluetooth link key for a controller-device
// pair from the active control set exported by Open.
//
// The registry export contains controller sections with device link keys.
// With the HKEY_LOCAL_MACHINE\SYSTEM prefix omitted, an example is:
//
//	[\ControlSet002\Services\BTHPORT\Parameters\Keys\9cfce8b88606]
//	"7845ce0af692"=hex:de,ad,be,ef,ba,da,55,12,34,56,78,9a,bc,cd,ef,01
//	"98583d332bda"=hex:fa,ce,db,ad,de,ed,be,ef,ca,fe,ba,be,13,37,42,24
//	"c87b230bc130"=hex:ab,cd,ef,01,23,45,67,89,0a,bc,de,f0,12,34,56,78
//
// The section's final component is the controller MAC, and each value name
// is a device MAC.
// Both MACs are lowercase with colons removed:
//
//	[\ControlSet002\Services\BTHPORT\Parameters\Keys\<controller-mac>]
//	"<device1-mac>"=hex:de,ad,be,ef,ba,da,55,12,34,56,78,9a,bc,cd,ef,01
//	"<device2-mac>"=hex:fa,ce,db,ad,de,ed,be,ef,ca,fe,ba,be,13,37,42,24
//	"<device3-mac>"=hex:ab,cd,ef,01,23,45,67,89,0a,bc,de,f0,12,34,56,78
//
// ControlSet002 is only an example; Open exports the control set selected by
// SYSTEM\Select\Current.
func (r *Registry) GetBluetoothLinkKey(controllerMAC, deviceMAC string) (string, error) {
	// Normalize MAC addresses to Windows format (lowercase, no colons)
	searchSection := fmt.Sprintf(`\Services\BTHPORT\Parameters\Keys\%s]`, normalizeMAC(controllerMAC))
	searchDevice := fmt.Sprintf(`"%s"=hex:`, normalizeMAC(deviceMAC))

	controllerFound := false
	inSection := false

	scanner := bufio.NewScanner(bytes.NewReader(r.dump.Bytes()))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, searchSection) {
			inSection = true
			controllerFound = true
			continue
		}
		if !inSection {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inSection = false
			continue
		}

		if strings.Contains(line, searchDevice) {
			_, v, ok := strings.Cut(line, "=hex:")
			if !ok {
				continue
			}
			return "hex:" + v, nil
		}
	}

	if !controllerFound {
		return "", fmt.Errorf("controller (%s) not found in the Windows registry", controllerMAC)
	}

	return "", fmt.Errorf("device (%s) not found in the Windows registry", deviceMAC)
}

func normalizeMAC(mac string) string {
	// Remove any separators (colons, hyphens, etc.)
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")

	return strings.ToLower(mac)
}
