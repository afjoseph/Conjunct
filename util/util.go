package util

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-playground/errors/v5"
)

// Expand expands the path to an absolute path.
// It does two more things than `filepath.Abs`:
// - Expands `~`, `.` and `..` symbols
// - Expands environment variables
func ExpandPath(path string, expandSymlinks bool) (string, error) {
	// Check if realpath is there
	_, err := exec.LookPath("realpath")
	if err != nil {
		return "", errors.Wrapf(err, "failed to find realpath")
	}
	path, err = ResolveShellVariables(path)
	if err != nil {
		return "", errors.Wrapf(
			err,
			"failed to resolve shell variables in %s",
			path,
		)
	}
	cmdArgs := []string{}
	if !expandSymlinks {
		cmdArgs = append(cmdArgs, "--no-symlinks")
	}
	cmdArgs = append(cmdArgs, path)
	cmd := exec.Command("realpath", cmdArgs...)
	b, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(err, "failed to expand path %s", path)
	}
	return strings.TrimSpace(string(b)), nil
}

func ResolveShellVariables(path string) (string, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("echo %s", path))
	b, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(
			err,
			"failed to resolve shell variables in %s",
			path,
		)
	}
	return strings.TrimSpace(string(b)), nil
}

// GetBasenameWithoutExtension returns the basename of a path without the
// extension.
func GetBasenameWithoutExtension(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
