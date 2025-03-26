package bt

import (
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestControllers(t *testing.T) {
	// Setup test directory
	tmpDir := t.TempDir()
	origVarLibBluetooth := varLibBluetooth
	varLibBluetooth = tmpDir
	defer func() {
		varLibBluetooth = origVarLibBluetooth
	}()

	// Create test controller directories
	testMacs := []string{"11:22:33:44:55:66", "AA:BB:CC:DD:EE:FF"}
	for _, mac := range testMacs {
		err := os.MkdirAll(filepath.Join(tmpDir, mac), 0755)
		assert.NilError(t, err)
	}

	// Create a file (should be ignored)
	f, err := os.Create(filepath.Join(tmpDir, "not-a-controller"))
	assert.NilError(t, err)
	f.Close()

	// Test Controllers function
	controllers, err := Controllers()
	assert.NilError(t, err)
	assert.Assert(t, cmp.Len(controllers, 2))

	for i, mac := range testMacs {
		assert.Equal(t, controllers[i].Mac, mac)
	}
}
