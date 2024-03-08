//go:build mage

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/afjoseph/conjunct/projectpath"
	"github.com/afjoseph/conjunct/util"
	"github.com/magefile/mage/sh"
)

var iosDemoConjunctConfigPath = filepath.Join(
	projectpath.Root,
	"testassets/ios/ConjunctDemo/conjunct-config.yaml",
)

var androidDemoConjunctConfigPath = filepath.Join(
	projectpath.Root,
	"testassets/android/ConjunctDemo/conjunct-config.yaml",
)

func RunUnitTests() error {
	return sh.Run("go", "test", "-race", "-v", "./...")
}

func Build() error {
	_, err := buildConjunct("")
	return err
}

// Build the Conjunct binary and set a default clang path
func BuildAndSetDefaultClangDirectory(defaultClangDir string) error {
	_, err := buildConjunct(defaultClangDir)
	return err
}

// Build the ios Demo app without conjunct
func BuildIosDemoWithoutConjunct(verbose bool) error {
	return buildIosDemoWithEnv(nil, verbose)
}

// Build the ios Demo app with conjunct
func BuildIosDemoWithConjunct(
	optBinPath, clangDirPath string,
	verbose bool,
) error {
	// Set conjunct path
	conjunctBinPath, err := buildConjunct("")
	orDie(err, "failed to build conjunct")

	// Get absolute config path
	configPath, err := util.ExpandPath(iosDemoConjunctConfigPath, true)
	orDie(err, "failed to expand conjunct config path")

	// Generate xcconfig file
	xcconfigPath, err := MakeXCConfigFile(map[string]string{
		"OTHER_CFLAGS": fmt.Sprintf(
			"--conjunct-config-path %s --conjunct-verbose --conjunct-retain-temp-dir",
			configPath,
		),
	})
	orDie(err, "failed to create xcconfig file")
	fmt.Printf("Generated xcconfig file at %s\n", xcconfigPath)
	defer os.Remove(xcconfigPath)

	// Print config path
	f, err := os.Open(configPath)
	orDie(err, fmt.Sprintf("failed to open config file: %s", configPath))
	defer f.Close()
	b, err := io.ReadAll(f)
	orDie(err, fmt.Sprintf("failed to read config file: %s", configPath))
	fmt.Printf("Conjunct config path contents:\n---\n%s\n---\n", string(b))

	// Build with xcconfig file and conjunct
	return buildIosDemoWithEnv(map[string]string{
		"XCODE_XCCONFIG_FILE": xcconfigPath,
		"CC":                  conjunctBinPath,
		"OPT_PATH":            optBinPath,
		"CLANG_DIR_PATH":      clangDirPath,
	}, verbose)
}

// Build the android Demo app without conjunct
func BuildAndroidDemoWithoutConjunct(verbose bool) error {
	return buildAndroidDemoWithEnv("", nil, verbose)
}

// Build the ios Demo app with conjunct
func BuildAndroidDemoWithConjunct(
	optBinPath, clangDirPath string,
	verbose bool,
) error {
	// Set conjunct path
	conjunctBinPath, err := buildConjunct("")
	orDie(err, "failed to build conjunct")

	// Get absolute config path
	configPath, err := util.ExpandPath(androidDemoConjunctConfigPath, true)
	orDie(err, "failed to expand conjunct config path")

	// Print config path
	f, err := os.Open(configPath)
	orDie(err, fmt.Sprintf("failed to open config file: %s", configPath))
	defer f.Close()
	b, err := io.ReadAll(f)
	orDie(err, fmt.Sprintf("failed to read config file: %s", configPath))
	fmt.Printf("Conjunct config path contents:\n---\n%s\n---\n", string(b))

	// Build with xcconfig file and conjunct
	return buildAndroidDemoWithEnv(conjunctBinPath, map[string]string{
		"OPT_PATH":       optBinPath,
		"CLANG_DIR_PATH": clangDirPath,
	}, verbose)
}

func buildAndroidDemoWithEnv(
	conjunctBinPath string,
	env map[string]string,
	verbose bool,
) error {
	return chdirAndExec("testassets/android/ConjunctDemo", func() error {
		// If conjunctBinPath is empty, just run a regular build without conjunct
		if conjunctBinPath == "" {
			// Just run a regular build without conjunct
			return sh.Run(
				"./gradlew",
				"assembleDebug",
				"--rerun-tasks",
				"-Ponly_arch=arm64-v8a",
			)
		}

		// Else, run a conjunct build
		// - specify the conjunct binary path
		// - specify the conjunct config path
		// - specify the verbosity level
		// - pass CLANG_DIR_PATH and OPT_PATH as environment variables
		conjunctFlags := ""
		if verbose {
			conjunctFlags = fmt.Sprintf(
				`-Pconjunct_flags=--conjunct-config-path %s --conjunct-verbose --conjunct-retain-temp-dir`,
				androidDemoConjunctConfigPath,
			)
		} else {
			conjunctFlags = fmt.Sprintf(`-Pconjunct_flags=--conjunct-config-path %s`, androidDemoConjunctConfigPath)
		}
		args := []string{
			"clean",
			"assembleDebug",
			"--rerun-tasks",
			"-Ponly_arch=arm64-v8a",
			"-Pconjunct_bin_path=" + conjunctBinPath,
			conjunctFlags,
		}
		return sh.RunWith(env, "./gradlew", args...)
	})
}

func MakeXCConfigFile(params map[string]string) (string, error) {
	tmpFile, err := os.CreateTemp("", "xcconfig")
	if err != nil {
		return "", fmt.Errorf(
			"failed to create temporary xcconfig file: %w",
			err,
		)
	}
	defer tmpFile.Close()
	for k, v := range params {
		if _, err := tmpFile.WriteString(fmt.Sprintf("%s=%s\n", k, v)); err != nil {
			return "", fmt.Errorf(
				"failed to write to temporary xcconfig file: %w",
				err,
			)
		}
	}
	return tmpFile.Name(), nil
}

func buildIosDemoWithEnv(env map[string]string, verbose bool) error {
	return chdirAndExec("testassets/ios/ConjunctDemo", func() error {
		args := []string{
			"build",
			"--no-skip-current",
			"--use-xcframeworks",
		}
		if verbose {
			args = append(args, "--verbose")
		}
		return sh.RunWith(env, "carthage", args...)
	})
}

// buildConjunct builds conjunct and returns the path of the binary
func buildConjunct(defaultClangDir string) (binPath string, err error) {
	args := []string{"build"}
	// If defaultClangDir is provided, add it to the build flags
	ldFlags := ""
	if defaultClangDir != "" {
		ldFlags = fmt.Sprintf(
			"--ldflags=-X main.DefaultClangDir=%s",
			defaultClangDir,
		)
	}

	// Add version information
	commit, err := sh.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to read VERSION file: %w", err)
	}
	ldFlags += fmt.Sprintf(
		" -X main.Version=%s",
		strings.TrimSpace(string(commit)),
	)
	args = append(args, ldFlags)
	args = append(args, "-o", "conjunct", ".")
	err = sh.Run("go", args...)
	if err != nil {
		return "", fmt.Errorf("failed to build conjunct: %w", err)
	}
	p, err := util.ExpandPath("conjunct", true)
	if err != nil {
		return "", fmt.Errorf("failed to expand conjunct path: %w", err)
	}
	return p, err
}

func orDie(err error, msg string) {
	if err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
}

// chdirAndExec changes the working directory to `dir` and executes `fn`.
func chdirAndExec(dir string, fn func() error) error {
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
