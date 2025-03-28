package winreg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	defer copyFile.Close()

	cmd := exec.Command("reged", "-x", copyFile.Name(), "\\ControlSet001\\Services\\BTHPORT\\Parameters\\Keys", "\\", "/dev/stdout")

	reg := &Registry{}

	var stderr bytes.Buffer
	cmd.Stdout = &reg.dump
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to dump registry: %w, stderr %s", err, stderr.String())
	}

	return reg, nil
}

func createTmpCopy(path string) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open registry file: %w", err)
	}
	defer f.Close()

	tmpCopy, err := os.CreateTemp("", "system-hive-*.reg")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	if _, err := io.Copy(tmpCopy, f); err != nil {
		return nil, fmt.Errorf("failed to copy registry file: %w", err)
	}

	return tmpCopy, nil
}

// GetBluetoothLinkKey extracts the Bluetooth link key for a specific controller-device pair
//
// The dump will be full of crap, but we only care about contain something like:
//
// ```
// [\ControlSet001\Services\BTHPORT\Parameters\Keys\ControlSet001\Services\BTHPORT\Parameters\Keys\9cfce8b88606]
// "7845ce0af692"=hex:de,ad,be,ef,ba,da,55,12,34,56,78,9a,bc,cd,ef,01
// "98583d332bda"=hex:fa,ce,db,ad,de,ed,be,ef,ca,fe,ba,be,13,37,42,24
// "c87b230bc130"=hex:ab,cd,ef,01,23,45,67,89,0a,bc,de,f0,12,34,56,78
//
// Now, notice that this corresponds to:
//
// ```
// [\ControlSet001\Services\BTHPORT\Parameters\Keys\ControlSet001\Services\BTHPORT\Parameters\Keys\<controller-mac>]
// "<device1-mac>"=hex:de,ad,be,ef,ba,da,55,12,34,56,78,9a,bc,cd,ef,01
// "<device2-mac>"=hex:fa,ce,db,ad,de,ed,be,ef,ca,fe,ba,be,13,37,42,24
// "<device3-mac>"=hex:ab,cd,ef,01,23,45,67,89,0a,bc,de,f0,12,34,56,78
// ...
// ```
func (r *Registry) GetBluetoothLinkKey(controllerMAC, deviceMAC string) (string, error) {
	// Normalize MAC addresses to Windows format (lowercase, no colons)
	controllerMAC = normalizeMAC(controllerMAC)
	deviceMAC = normalizeMAC(deviceMAC)

	scanner := bufio.NewScanner(bytes.NewReader(r.dump.Bytes()))

	controllerFound := false
	searchSection := fmt.Sprintf("%s\\%s", "ControlSet001\\Services\\BTHPORT\\Parameters\\Keys", controllerMAC)
	searchDevice := fmt.Sprintf(`"%s"=hex:`, deviceMAC)

	inSection := false
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.Contains(line, searchSection) {
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
		return "", fmt.Errorf("controller not found in registry")
	}

	return "", fmt.Errorf("device not found in registry")
}

func normalizeMAC(mac string) string {
	// Remove any separators (colons, hyphens, etc.)
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")

	return strings.ToLower(mac)
}
