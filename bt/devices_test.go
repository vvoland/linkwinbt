package bt

import (
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestDevices(t *testing.T) {
	tmpDir := t.TempDir()

	origVarLibBluetooth := varLibBluetooth
	varLibBluetooth = tmpDir
	defer func() { varLibBluetooth = origVarLibBluetooth }()
	testMac := "11:22:33:44:55:66"
	controllerDir := filepath.Join(tmpDir, testMac)
	assert.NilError(t, os.MkdirAll(controllerDir, 0755))

	testDevices := []struct {
		mac, name string
	}{
		{"AA:BB:CC:DD:EE:FF", "Device1"},
		{"00:11:22:33:44:55", "Device2"},
	}

	for _, device := range testDevices {
		deviceDir := filepath.Join(controllerDir, device.mac)
		assert.NilError(t, os.MkdirAll(deviceDir, 0755))
		assert.NilError(t, os.WriteFile(filepath.Join(deviceDir, "info"), []byte(`
[General]
Name=`+device.name+`

[LinkKey]
Key=00000000000000000000000000000000
Type=4
PINLength=0
`), 0644))
	}

	f, err := os.Create(filepath.Join(controllerDir, "not-a-device"))
	assert.NilError(t, err)
	f.Close()

	controller := Controller{Mac: testMac}
	devices, err := controller.Devices()
	assert.NilError(t, err)
	assert.Assert(t, cmp.Len(devices, 2))

	deviceMap := make(map[string]Device)
	for _, d := range devices {
		deviceMap[d.Mac] = d
		assert.Check(t, cmp.Equal(d.controller.Mac, testMac))
	}

	for _, expected := range testDevices {
		device, exists := deviceMap[expected.mac]
		if assert.Check(t, exists, "Device %s not found", expected.mac) {
			assert.Check(t, cmp.Equal(device.Name, expected.name))
		}
	}

	t.Run("SetLinkKey", func(t *testing.T) {
		device := deviceMap["AA:BB:CC:DD:EE:FF"]
		err := device.SetLinkKey("1234567890123456")
		assert.NilError(t, err)

		// Check if the link key was set correctly
		infoData, err := os.ReadFile(filepath.Join(controllerDir, device.Mac, "info"))
		assert.NilError(t, err)

		assert.Check(t, cmp.Equal(string(infoData), `
[General]
Name=Device1

[LinkKey]
Key=1234567890123456
Type=4
PINLength=0
`))
	})
}
