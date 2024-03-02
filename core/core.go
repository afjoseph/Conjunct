package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/afjoseph/conjunct/argsparser"
	"github.com/afjoseph/conjunct/config"
	"github.com/afjoseph/conjunct/util"
	"github.com/sirupsen/logrus"
)

// emitBitcode emits bitcode for 'objectName' using 'clangPath' and
// 'originalArgs'
func emitBitcode(
	objectName string,
	clangPath string,
	originalArgs []string,
	tempDir string,
	isDryRun bool,
) (bitcodeFilepath string, err error) {
	logrus.Debugln("emitBitcode()")

	bitcodeFile, err := os.Create(filepath.Join(tempDir, objectName+".bc"))
	if err != nil {
		return "", fmt.Errorf("while creating temp file: %w", err)
	}
	bitcodeFilepath, err = util.ExpandPath(bitcodeFile.Name(), true)
	if err != nil {
		return "", fmt.Errorf("while expanding path: %w", err)
	}

	args := append([]string(nil), originalArgs...) // Copies the slice
	// These cause errors when supplied to opt later
	args = argsparser.RemoveArg(args, "-g", false)
	args = argsparser.RemoveArg(args, "-gmodules", false)
	// -fembed-bitcode usually included twice
	args = argsparser.RemoveArg(args, "-fembed-bitcode", false)
	args = argsparser.RemoveArg(args, "-fembed-bitcode", false)
	args = argsparser.RemoveArg(args, "-fembed-bitcode-marker", false)
	args = argsparser.RemoveArg(args, "-o", true)
	// XXX Now, we don't want our passes to play with sanitizer code, so it is
	// best to remove it at this step.
	// In buildBitcode, the sanitization flags will be KEPT so that the passes
	// are **tested** against sanitizers (but not **scheduled** with it).
	args = argsparser.RemoveRegexArg(args, "-fsanitize=[a-z,]+")
	args = argsparser.AddArg(args, "-emit-llvm", "")
	// args = argsparser.AddArg(args, "-c", "")
	args = argsparser.AddArg(args, "-o", bitcodeFilepath)
	// XXX <02-03-2024, afjoseph> When running Conjunct with different build
	// flags, it's wise to tell the compiler to ignore those flags, else some
	// builds will fail
	args = argsparser.AddArg(
		args,
		"-Wno-unused-command-line-argument",
		"",
	)

	logrus.Infof("Emitting bitcode for %s\n", objectName)
	cmd := exec.Command(clangPath, args...)
	logrus.Debugf("cmd: %s\n", cmd.String())
	if isDryRun {
		logrus.Debugln("Dry-run: not running above command")
	} else {
		ret, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf(
				"while emitting bitcode: %s: %w",
				string(ret),
				err,
			)
		}
		logrus.Infof("Bitcode generated in %s\n", bitcodeFilepath)
	}
	return bitcodeFilepath, nil
}

// schedulePasses runs opt on 'inputFilepath' using information from 'cfg'.
func schedulePasses(
	objectName string,
	cfg *config.ConjunctConfig,
	inputFilepath string,
	tempDir string,
	isDryRun bool,
) (outputFilepath string, err error) {
	logrus.Debugf("SchedulePasses() on %s at %s\n", objectName, inputFilepath)

	// Create temporary file to be the output of the opt command
	baseName := util.GetBasenameWithoutExtension(inputFilepath)
	outputFile, err := os.Create(
		filepath.Join(tempDir, fmt.Sprintf("%s.opt.bc", baseName)),
	)
	if err != nil {
		return "", fmt.Errorf("while creating temp file: %w", err)
	}
	outputFilepath, err = util.ExpandPath(outputFile.Name(), true)
	if err != nil {
		return "", fmt.Errorf("while expanding path: %w", err)
	}

	cmdArgs := []string{
		fmt.Sprintf("-passes=%s", strings.Join(cfg.Passes, ",")),
		inputFilepath,
		"-o",
		outputFilepath,
	}
	if cfg.OptExtraArgs != nil {
		for k, v := range cfg.OptExtraArgs {
			cmdArgs = append(cmdArgs, k+"="+v)
		}
	}

	cmd := exec.Command(cfg.OptPath, cmdArgs...)
	logrus.Debugf(
		"Running opt on %s @ %s -> %s using this command: %s\n",
		objectName,
		inputFilepath,
		outputFilepath,
		cmd.String(),
	)
	if isDryRun {
		logrus.Debugln("Dry-run: not running above command")
	} else {
		b, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("while running opt: %w: %s", err, string(b))
		}
		logrus.Infoln("Opt ran successfully")
	}
	return outputFilepath, nil
}

// buildBitcode() builds an object file from the bitcode file located in
// 'bitcodeFilepath', after modifying args from 'originalArgs' array.
//
// To repeat, buildBitcode() does not link, so you will not get an
// executable: you'll get a compiled object file. This is so because both
// Android and iOS build systems have a separate step where they do the
// linking that is separate from the step where object files are built.
// Conjunct only cares about producing modified object files, not the linking
// step.
func buildBitcode(
	clangPath string,
	bitcodeFilepath string,
	originalArgs []string,
	isDryRun bool,
) (string, error) {
	args := append([]string(nil), originalArgs...) // Copies the slice
	args = argsparser.RemoveArg(args, "-x", true)
	args = argsparser.AddArg(args, "-x", "ir")
	args = argsparser.RemoveArg(args, "-c", true)
	args = argsparser.AddArg(args, "-c", bitcodeFilepath)
	// XXX <02-03-2024, afjoseph> When running Conjunct with different build
	// flags, it's wise to tell the compiler to ignore those flags, else some
	// builds will fail
	args = argsparser.AddArg(
		args,
		"-Wno-unused-command-line-argument",
		"",
	)
	outFilepath := argsparser.GetArgVal(args, "-o")
	if len(outFilepath) == 0 {
		return "", fmt.Errorf("missing -o argument")
	}

	cmd := exec.Command(clangPath, args...)
	logrus.Debugf(
		"Building bitcode %s to an object file with the following args %+v\n",
		bitcodeFilepath,
		cmd.String(),
	)
	if isDryRun {
		logrus.Debugln("Dry-run: not running above command")
	} else {
		ret, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf(
				"while building bitcode: %s: %w",
				string(ret),
				err,
			)
		}
		logrus.Infof("Successfully built bitcode for %s at %s\n", bitcodeFilepath, outFilepath)
	}

	return outFilepath, nil
}

// RunConjunct runs the Conjunct core using 'cfg', which looks
// like this:
// - Emit bitcode using emitBitcode()
// - Run the passes using opt, sequentially, on every emitted bitcode using
//   schedulePasses()
// - Build the modified bitcode (without linking) using buildBitcode()
//   - To repeat: buildBitcode() does not link, so you will not get an
//     executable: you'll get a compiled object file. This is so because both
//     Android and iOS build systems have a separate step where they do the
//     linking that is separate from the step where object files are built.
//     Conjunct only cares about producing modified object files, not the
//     linking step.

// It then exits with err if command failed
//
//	else, it'll exit normally
func RunConjunct(
	cfg *config.ConjunctConfig,
	args []string,
) error {
	logrus.Debugf("Config: %+v, args: %+v", cfg, args)

	// Check for dry runs
	dryRun := false
	// If dryRun is true, run Conjunct's pipeline but don't actually run
	// any commands. See README's FAQ for more info.
	if argsparser.HasArg(args, "--conjunct-dry-run") {
		dryRun = true
	}
	args = argsparser.RemoveArg(args, "--conjunct-dry-run", false)

	// Conjunct must run only during object compilation steps.
	// In any other instance, just run original clang
	if !argsparser.HasArg(args, "-c") {
		logrus.Debugln("Not an object compilation step: using Clang instead")
		err, exitCode := RunOriginalClang(cfg.ClangPath, args)
		if err != nil {
			os.Exit(exitCode)
		}
		return nil
	}

	// Create temp dir
	tempDir, err := os.MkdirTemp("", "conjunct")
	if err != nil {
		return fmt.Errorf("while creating temp dir: %w", err)
	}
	logrus.Debugf("Created temp dir %s\n", tempDir)

	if !cfg.RetainTempDir {
		defer os.RemoveAll(tempDir)
	}

	// XXX <29-09-2023, afjoseph> This is not perfectly accurate since there's
	// no obligation by the compiler to postfix -c with the objectName, but it's
	// what usually happens
	sourceFileName, _, err := argsparser.GetSourceFileName(args)
	if err != nil {
		return fmt.Errorf("while getting source file name: %w", err)
	}

	bitcodeFilepath, err := emitBitcode(
		sourceFileName,
		cfg.ClangPath,
		args,
		tempDir,
		dryRun,
	)
	if err != nil {
		return fmt.Errorf("while emitting bitcode: %w", err)
	}
	afterOptBitcodeFilepath, err := schedulePasses(
		sourceFileName,
		cfg,
		bitcodeFilepath,
		tempDir,
		dryRun,
	)
	if err != nil {
		return fmt.Errorf("while scheduling passes: %w", err)
	}
	_, err = buildBitcode(
		cfg.ClangPath,
		afterOptBitcodeFilepath,
		args,
		dryRun,
	)
	if err != nil {
		return fmt.Errorf("while building bitcode: %w", err)
	}
	if dryRun {
		// run original clang
		logrus.Debugln("Running original clang during dry-run...")
		err, _ := RunOriginalClang(cfg.ClangPath, args)
		if err != nil {
			return fmt.Errorf(
				"while running original clang during dry-run: %w",
				err,
			)
		}
	}
	logrus.Infof("Conjunct ran successfully on %s\n", sourceFileName)
	return nil
}

// RunOriginalClang runs the original clang using 'args'
func RunOriginalClang(
	clangPath string,
	args []string,
) (err error, exitCode int) {
	logrus.Debugln("Running original clang...")
	if len(clangPath) == 0 {
		panic("clangPath is nil. Add it in the YAML configuration file")
	}

	cmd := exec.Command(clangPath, args...)
	ret, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Println(string(ret))
			return err, exitError.ExitCode()
		}
	}
	fmt.Printf("%s\n", ret)
	return nil, 0
}
