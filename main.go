package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/afjoseph/conjunct/argsparser"
	"github.com/afjoseph/conjunct/config"
	"github.com/afjoseph/conjunct/core"
	"github.com/sirupsen/logrus"
)

var Version string

func init() {
	// Set version from VERSION file
	Version = "UNKNOWN"
	t, err := os.ReadFile("VERSION")
	if err == nil {
		Version = string(t)
	}

	// Set logrus output to stdout and log file
	logFilePath := os.Getenv("CONJUNCT_LOG_FILE")
	if len(logFilePath) > 0 {
		logFile, err := os.OpenFile(
			logFilePath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
		if err != nil {
			panic(
				fmt.Errorf("failed to open log file %s: %w", logFilePath, err),
			)
		}
		logrus.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}
}

func main() {
	if err := checkForPathBinaries(); err != nil {
		panic(
			fmt.Errorf(
				"important binary doesn't exist in PATH. Make sure to have this binary visible in $PATH: %w",
				err,
			),
		)
	}

	args := os.Args[1:]
	// Check for version flag
	if argsparser.HasArg(args, "--version") {
		fmt.Printf("Conjunct version %s\n", Version)
		os.Exit(0)
	}
	// Check for verbose flags
	if argsparser.HasArg(args, "--conjunct-verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	}
	args = argsparser.RemoveArg(args, "--conjunct-verbose", false)

	args, cfg, shouldRunConjunct, err := config.ExtractConfigFromArgs(args)
	if err != nil {
		panic(fmt.Errorf("failed to extract conjunct config: %w", err))
	}
	// if shouldRunConjunct is false, we don't need to run conjunct; just run
	// Clang regluarly
	if !shouldRunConjunct {
		err, exitCode := core.RunOriginalClang(cfg.ClangPath, args)
		if err != nil {
			os.Exit(exitCode)
		}
		os.Exit(0)
	}
	// Check for dry run
	if argsparser.HasArg(args, "--conjunct-dry-run") {
		cfg.DryRun = true
	}
	args = argsparser.RemoveArg(args, "--conjunct-dry-run", false)

	if err := core.RunConjunctWithConfig(cfg, args); err != nil {
		panic(fmt.Errorf("failed to run conjunct: %w", err))
	}
}

func checkForPathBinaries() error {
	for _, bin := range []string{
		"realpath",
	} {
		_, err := exec.LookPath(bin)
		if err != nil {
			return fmt.Errorf("missing binary %s", bin)
		}
	}
	return nil
}
