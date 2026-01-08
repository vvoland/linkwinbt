package winreg

import (
	"bytes"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestGetBluetoothLinkKey(t *testing.T) {
	mockData := `
[\ControlSet001\Services\BTHPORT\Parameters\Keys\001122334455]
"112233445566"=hex:aa,bb,cc,dd,ee,ff,00,11,22,33,44,55,66,77,88,99
"aabbccddeeff"=hex:de,ad,be,ef,00,11,22,33,44,55,66,77,88,99,aa,bb
`

	tests := []struct {
		name          string
		controllerMAC string
		deviceMAC     string
		expectedKey   string
		expectedErr   string
	}{
		{
			name:          "missing controller",
			controllerMAC: "ff:ee:dd:cc:bb:aa",
			deviceMAC:     "aa:bb:cc:dd:ee:ff",
			expectedErr:   "controller (ff:ee:dd:cc:bb:aa) not found in the Windows registry",
		},
		{
			name:          "missing device key",
			controllerMAC: "00:11:22:33:44:55",
			deviceMAC:     "ff:ff:ff:ff:ff:ff",
			expectedErr:   "device (ff:ff:ff:ff:ff:ff) not found in the Windows registry",
		},
		{
			name:          "valid key",
			controllerMAC: "00:11:22:33:44:55",
			deviceMAC:     "aa:bb:cc:dd:ee:ff",
			expectedKey:   "hex:de,ad,be,ef,00,11,22,33,44,55,66,77,88,99,aa,bb",
		},
	}

	reg := &Registry{dump: *bytes.NewBufferString(mockData)}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := reg.GetBluetoothLinkKey(tt.controllerMAC, tt.deviceMAC)
			if tt.expectedErr != "" {
				assert.Check(t, cmp.Error(err, tt.expectedErr))
				return
			}
			assert.NilError(t, err)

			assert.Equal(t, key, tt.expectedKey)
		})
	}
}
