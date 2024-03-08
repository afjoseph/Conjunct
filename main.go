package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/afjoseph/conjunct/argsparser"
	"github.com/afjoseph/conjunct/config"
	"github.com/afjoseph/conjunct/core"
	"github.com/afjoseph/conjunct/sourcefile"
	"github.com/go-playground/errors/v5"
	"github.com/sirupsen/logrus"
)

var (
	// This is the version of conjunct. It's set through the build process.
	Version string = "UNKNOWN"
	// DefaultClangDir is the default path to the clang binary.
	// It's used if no config is available (i.e., through
	// "--conjunct-config-path") that would point to a specific clang.
	//
	// This variable is usually set through the build process.
	DefaultClangDir string = ""
)

func init() {
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
				errors.Wrapf(err, "failed to open log file %s", logFilePath),
			)
		}
		logrus.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}
}

func main() {
	args := os.Args[1:]
	// Check for version flag
	if argsparser.HasArg(args, "--version") {
		fmt.Printf(
			"Conjunct version %s | Default clang path: %s\n",
			Version,
			DefaultClangDir,
		)
		os.Exit(0)
	}
	// Check for verbose flags
	if argsparser.HasArg(args, "--conjunct-verbose") {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debugf("Running conjunct in verbose mode")
	}
	args = argsparser.RemoveArg(args, "--conjunct-verbose", false)
	// Extract config
	args, cfg, err := config.ExtractConfigFromArgs(args)
	if err != nil {
		panic(errors.Wrapf(err, "failed to extract config from args"))
	}

	// Find which clang binary to run: clang or clang++
	_, sourceFileType := sourcefile.GetSourceFileName(args)
	clangBinaryName := getClangBinaryName(
		filepath.Base(os.Args[0]),
		sourceFileType,
	)

	// If there's no config provided, find a clang binary to run, run it and
	// exit
	if cfg == nil {
		clangPath := ""
		// If we have a default clang dir, use it
		if DefaultClangDir != "" {
			clangPath = filepath.Join(DefaultClangDir, clangBinaryName)
			// Check if there's a clang binary with a '.original' suffix in the same direcoty.
			// If it exists, use it instead of the clang binary because this
			// means there's a symlink with Conjunct
			// TODO <08-03-2024, afjoseph> Expand on this meaning
			clangOriginalPath := clangPath + ".original"
			if _, err := os.Stat(clangOriginalPath); err == nil {
				clangPath = clangOriginalPath
			}
		} else {
			// Else, find a clang binary in $PATH
			clangPath, err = exec.LookPath(clangBinaryName)
			if err != nil {
				panic(errors.Wrapf(err, "while finding %s in $PATH", clangBinaryName))
			}
		}
		err, exitCode := core.RunClang(clangPath, args)
		if err != nil {
			logrus.Errorf("failed to run original clang: %v", err)
			os.Exit(exitCode)
		}
		os.Exit(0)
	}

	// If we have a config, run conjunct with the supplied ClangDirPath in
	// the config
	clangPath := filepath.Join(cfg.ClangDirPath, clangBinaryName)
	// Check if there's a clang binary with a '.original' suffix in the same direcoty.
	// If it exists, use it instead of the clang binary because this
	// means there's a symlink with Conjunct
	// TODO <08-03-2024, afjoseph> Expand on this meaning
	clangOriginalPath := clangPath + ".original"
	if _, err := os.Stat(clangOriginalPath); err == nil {
		clangPath = clangOriginalPath
	}
	if err := core.RunConjunct(cfg, clangPath, args); err != nil {
		panic(errors.Wrapf(err, "failed to run conjunct"))
	}
}

func getClangBinaryName(
	baseProgramName string,
	sourceFileType sourcefile.Type,
) string {
	// If the binary name is not conjunct and it's a clang binary, use it
	// directly
	if baseProgramName != "conjunct" &&
		strings.HasPrefix(baseProgramName, "clang") {
		return baseProgramName
	}

	// Else, use clang or clang++ from $PATH based on the source file type
	// XXX <08-03-2024, afjoseph> We need to know if this is a C or C++ file
	// mainly to know which clang binary to run since there is a difference
	// between clang++ and clang:
	// https://github.com/llvm/llvm-project/issues/54701#issuecomment-1086055306
	if sourceFileType == sourcefile.Type_C {
		return "clang"
	} else if sourceFileType == sourcefile.Type_CPP ||
		sourceFileType == sourcefile.Type_OBJC {
		return "clang++"
	}
	// If we have no idea what the source file type is, just default to clang++
	return "clang++"
}
