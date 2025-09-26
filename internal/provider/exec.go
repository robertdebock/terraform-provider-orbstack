package provider

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ClientConfig holds provider runtime configuration.
type ClientConfig struct {
	OrbPath           string
	DefaultUser       string
	DefaultSSHKeyPath string
	CreateTimeout     string
	DeleteTimeout     string
}

// runOrb runs orb with arguments and returns stdout as string.
func runOrb(ctx context.Context, orbPath string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, orbPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stdout.String(), stderr.String(), fmt.Errorf("orb %s failed: %w\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String(), stderr.String(), nil
}

