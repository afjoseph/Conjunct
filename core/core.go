package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/afjoseph/conjunct/argsparser"
	"github.com/afjoseph/conjunct/config"
	"github.com/afjoseph/conjunct/sourcefile"
	"github.com/afjoseph/conjunct/util"
	"github.com/go-playground/errors/v5"
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
		return "", errors.Wrapf(err, "while creating temp file")
	}
	bitcodeFilepath, err = util.ExpandPath(bitcodeFile.Name(), true)
	if err != nil {
		return "", errors.Wrapf(err, "while expanding path")
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
	// XXX We don't want our passes to play with address sanitizer (ASAN) code,
	// so it is best to remove it at this step.
	//
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

	logrus.Infof("Emitting bitcode for %s", objectName)
	cmd := exec.Command(clangPath, args...)
	logrus.Debugf("cmd: %s", cmd.String())
	if isDryRun {
		logrus.Debugln("Dry-run: not running above command")
	} else {
		ret, err := cmd.CombinedOutput()
		if err != nil {
			return "", errors.Wrapf(err, "while emitting bitcode: %s", string(ret))
		}
		logrus.Infof("Bitcode generated in %s", bitcodeFilepath)
	}
	return bitcodeFilepath, nil
}

// schedulePasses runs opt on 'inputFilepath' using information from 'cfg'.
func schedulePasses(
	objectName string,
	optPath string,
	optCLIArgs []string,
	optEnvVars map[string]string,
	inputFilepath string,
	tempDir string,
	isDryRun bool,
) (outputFilepath string, err error) {
	logrus.Debugf("SchedulePasses() on %s at %s", objectName, inputFilepath)

	// Create temporary file to be the output of the opt command
	baseName := util.GetBasenameWithoutExtension(inputFilepath)
	outputFile, err := os.Create(
		filepath.Join(tempDir, fmt.Sprintf("%s.opt.bc", baseName)),
	)
	if err != nil {
		return "", errors.Wrapf(err, "while creating temp file")
	}
	outputFilepath, err = util.ExpandPath(outputFile.Name(), true)
	if err != nil {
		return "", errors.Wrapf(err, "while expanding path")
	}

	cliArgs := []string{}
	for _, arg := range optCLIArgs {
		cliArgs = append(cliArgs, arg)
	}
	cliArgs = append(cliArgs, inputFilepath)
	cliArgs = append(cliArgs, "-o", outputFilepath)

	cmd := exec.Command(optPath, cliArgs...)
	cmd.Env = os.Environ()
	for k, v := range optEnvVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	logrus.Debugf(
		"Running opt on %s @ %s -> %s using this command: %s and these env vars: %+v",
		objectName,
		inputFilepath,
		outputFilepath,
		cmd.String(),
		optEnvVars,
	)
	if isDryRun {
		logrus.Debugln("Dry-run: not running above command")
	} else {
		b, err := cmd.CombinedOutput()
		if err != nil {
			return "", errors.Wrapf(err, "while running opt: %s", string(b))
		}
		logrus.Infof("Opt ran successfully: %s", string(b))
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
		return "", errors.New("missing -o argument")
	}

	cmd := exec.Command(clangPath, args...)
	logrus.Debugf(
		"Building bitcode %s to an object file with the following args %+v",
		bitcodeFilepath,
		cmd.String(),
	)
	if isDryRun {
		logrus.Debugln("Dry-run: not running above command")
	} else {
		ret, err := cmd.CombinedOutput()
		if err != nil {
			return "", errors.Wrapf(err, "while building bitcode: %s: %w", string(ret), err)
		}
		logrus.Infof("Successfully built bitcode for %s at %s", bitcodeFilepath, outFilepath)
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
	cfg *config.Config,
	clangPath string,
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
		err, exitCode := RunClang(clangPath, args)
		if err != nil {
			os.Exit(exitCode)
		}
		return nil
	}

	// Create temp dir
	tempDir, err := os.MkdirTemp("", "conjunct")
	if err != nil {
		return errors.Wrapf(err, "while creating temp dir")
	}
	logrus.Debugf("Created temp dir %s", tempDir)

	if !cfg.RetainTempDir {
		defer os.RemoveAll(tempDir)
	}

	// XXX <29-09-2023, afjoseph> This is not perfectly accurate since there's
	// no obligation by the compiler to postfix -c with the objectName, but it's
	// what usually happens
	sourceFileName, _ := sourcefile.GetSourceFileName(args)
	if sourceFileName == "" {
		return errors.Wrapf(err, "while getting source file name")
	}

	bitcodeFilepath, err := emitBitcode(
		sourceFileName,
		clangPath,
		args,
		tempDir,
		dryRun,
	)
	if err != nil {
		return errors.Wrapf(err, "while emitting bitcode")
	}
	afterOptBitcodeFilepath, err := schedulePasses(
		sourceFileName,
		cfg.OptPath,
		cfg.OptCLIArgs,
		cfg.OptEnvVars,
		bitcodeFilepath,
		tempDir,
		dryRun,
	)
	if err != nil {
		return errors.Wrapf(err, "while scheduling passes")
	}
	_, err = buildBitcode(
		clangPath,
		afterOptBitcodeFilepath,
		args,
		dryRun,
	)
	if err != nil {
		return errors.Wrapf(err, "while building bitcode")
	}
	if dryRun {
		// run original clang
		logrus.Debugln("Running original clang during dry-run...")
		err, _ := RunClang(clangPath, args)
		if err != nil {
			return errors.Wrapf(
				err,
				"while running original clang during dry-run",
			)
		}
	}
	logrus.Infof("Conjunct ran successfully on %s", sourceFileName)
	return nil
}

// RunClang runs the clang from 'clangPath' with 'args'
func RunClang(
	clangPath string,
	args []string,
) (err error, exitCode int) {
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
