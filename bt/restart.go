package bt

import (
	"context"
	"fmt"
	"os/exec"
)

func Restart(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "systemctl", "restart", "bluetooth")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart bluetooth: %w, out: %s", err, string(out))
	}

	return nil
}
