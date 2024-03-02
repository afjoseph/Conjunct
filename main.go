package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/afjoseph/conjunct/argsparser"
	"github.com/afjoseph/conjunct/config"
	"github.com/afjoseph/conjunct/core"
	"github.com/afjoseph/conjunct/util"
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
		logrus.Debugf("Running conjunct in verbose mode")
	}
	args = argsparser.RemoveArg(args, "--conjunct-verbose", false)

	// Paths:
	// - Config is provided
	//   - Do things however the config wants it
	// - Config not provided
	//   - Find clang in $PATH
	//   - Run clang
	args, cfg, err := config.ExtractConfigFromArgs(args)
	if err != nil {
		panic(fmt.Errorf("failed to extract conjunct config: %w", err))
	}
	// If there's no config provided, just run whatever Clang we can find in
	// $PATH
	if cfg == nil {
		// Extract source file extension: we need to know if this is a C or C++
		// file mainly to know which clang binary to run since there is a
		// difference between clang++ and clang:
		// https://github.com/llvm/llvm-project/issues/54701#issuecomment-1086055306
		clangBinaryName := ""
		_, sourceFileType, err := argsparser.GetSourceFileName(
			args,
		)
		if err != nil {
			// If we failed to get the source filename, just use clang++
			clangBinaryName = "clang++"
		} else {
			if sourceFileType == util.SourceFileType_C {
				clangBinaryName = "clang"
			} else if sourceFileType == util.SourceFileType_CPP ||
				sourceFileType == util.SourceFileType_OBJC {
				clangBinaryName = "clang++"
			} else {
				// This shouldn't happen, but it's okay to use clang++ here too
				clangBinaryName = "clang++"
			}
		}
		// Get Clang's path from $PATH
		clangPath, err := exec.LookPath(clangBinaryName)
		if err != nil {
			panic(fmt.Errorf("failed to find clang in $PATH: %w", err))
		}
		// Run it and exit
		err, exitCode := core.RunOriginalClang(clangPath, args)
		if err != nil {
			logrus.Errorf("failed to run original clang: %v", err)
			os.Exit(exitCode)
		}
		os.Exit(0)
	}

	if err := core.RunConjunct(cfg, args); err != nil {
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
