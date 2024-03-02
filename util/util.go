package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Expand expands the path to an absolute path.
// It does two more things than `filepath.Abs`:
// - Expands `~`, `.` and `..` symbols
// - Expands environment variables
func ExpandPath(path string, expandSymlinks bool) (string, error) {
	path, err := ResolveShellVariables(path)
	if err != nil {
		return "", fmt.Errorf(
			"failed to resolve shell variables in %s: %w",
			path,
			err,
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
		return "", fmt.Errorf("failed to expand path %s: %w", path, err)
	}
	return strings.TrimSpace(string(b)), nil
}

func ResolveShellVariables(path string) (string, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("echo %s", path))
	b, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf(
			"failed to resolve shell variables in %s: %w",
			path,
			err,
		)
	}
	return strings.TrimSpace(string(b)), nil
}

// ChdirAndExec changes the working directory to `dir` and executes `fn`.
func ChdirAndExec(dir string, fn func() error) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("failed to change directory to %s: %w", dir, err)
	}
	defer os.Chdir(wd)
	return fn()
}

// GetBasenameWithoutExtension returns the basename of a path without the
// extension.
func GetBasenameWithoutExtension(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// CheckIfClangBinary checks if the binary at `path` is a clang binary.
func CheckIfClangBinary(path string) error {
	cmd := exec.Command(path, "--version")
	b, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run %s: %w", path, err)
	}
	if !strings.Contains(string(b), "clang") {
		return fmt.Errorf("not a clang binary: %s", path)
	}
	return nil
}

// CheckIfOptBinary checks if the binary at `path` is an opt binary.
func CheckIfOptBinary(path string) error {
	cmd := exec.Command(path, "--version")
	b, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run %s: %w", path, err)
	}
	if !strings.Contains(string(b), "LLVM") {
		return fmt.Errorf("not an opt binary: %s", path)
	}
	return nil
}

type SourceFileType int

const (
	SourceFileType_C       SourceFileType = 1
	SourceFileType_CPP     SourceFileType = 2
	SourceFileType_OBJC    SourceFileType = 3
	SourceFileType_Unknown SourceFileType = -1
)

var (
	cppFileExtensions  = []string{".cpp", ".cc", ".cxx", ".c++"}
	objcFileExtensions = []string{".m"}
	cFileExtensions    = []string{".c"}
)

func FetchSourceFileType(path string) (SourceFileType, error) {
	ext := filepath.Ext(path)
	for _, e := range cppFileExtensions {
		if e == ext {
			return SourceFileType_CPP, nil
		}
	}
	for _, e := range objcFileExtensions {
		if e == ext {
			return SourceFileType_OBJC, nil
		}
	}
	for _, e := range cFileExtensions {
		if e == ext {
			return SourceFileType_C, nil
		}
	}
	return SourceFileType_Unknown, fmt.Errorf("unknown file extension: %s", ext)
}
