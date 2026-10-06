package module

import (
	"errors"
	"os/exec"
	"strings"
)

func asExitError(err error, dst **exec.ExitError) bool { return errors.As(err, dst) }

// trimStderr keeps the tail of a Nix error, which is where its message sits.
func trimStderr(b []byte) string {
	s := strings.TrimSpace(string(b))
	lines := strings.Split(s, "\n")
	const keep = 8
	if len(lines) > keep {
		lines = lines[len(lines)-keep:]
	}
	return strings.Join(lines, "\n")
}
