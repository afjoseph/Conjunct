package config

import (
	stderr "errors"
	"fmt"
	"os"

	"github.com/afjoseph/conjunct/argsparser"
	"github.com/afjoseph/conjunct/util"
	"github.com/go-playground/errors/v5"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	ErrParsingConfig = stderr.New("Failed to parse config")
)

type ConjunctConfig struct {
	// Seed is the seed used for random number generation. Useful for some
	// passes
	Seed int64 `yaml:"seed"`
	// ClangPath is the path to the Clang binary
	ClangPath string `yaml:"clang-path"`
	// OptPath is the path to the Opt binary
	OptPath string `yaml:"opt-path"`
	// OptEnvArgs is a list of environment variables to setup while running Opt
	OptEnvVars map[string]string `yaml:"opt-env-vars"`
	// OptCLIArgs is a list of arguments to pass to Opt
	OptCLIArgs []string `yaml:"opt-cli-args"`
	// If RetainTempDir is true, don't delete the temporary directory
	// conjunct creates. Useful for debugging.
	RetainTempDir bool `yaml:"-"`
}

// ExtractConfigFromArgs extracts conjunct config from 'args' and returns
// it as a ConjunctConfig struct
func ExtractConfigFromArgs(
	args []string,
) (retArgs []string, cfg *ConjunctConfig, err error) {
	logrus.Debugln("Parsing conjunct params...")

	if len(args) == 0 {
		// TODO <02-03-2024, afjoseph> Should this print clang's usage or are
		// we safe here to print our own usage?
		return args, nil, nil
	}

	// Extract config path from --conjunct-config-path
	configFilePath := argsparser.GetArgVal(args, "--conjunct-config-path")
	if len(configFilePath) == 0 {
		logrus.Debugln("Failed to find --conjunct-config-path")
		// Return without errors since we didn't fail: we just don't have a
		// conjunct config file which is allowed. It just means run the
		// original clang in $PATH
		return args, nil, nil
	}
	args = argsparser.RemoveArg(args, "--conjunct-config-path", true)

	// Read and parse
	configFileContent, err := os.ReadFile(configFilePath)
	if err != nil {
		return args, nil, fmt.Errorf("failed to read YAML file: %w", err)
	}
	if len(configFileContent) == 0 {
		return args, nil, fmt.Errorf("Empty YAML file")
	}
	config := ConjunctConfig{}
	err = yaml.Unmarshal(configFileContent, &config)
	if err != nil {
		return args, nil, errors.Wrapf(
			ErrParsingConfig,
			"at %s: %v",
			configFilePath,
			err,
		)
	}
	if config.Seed == 0 {
		return args, nil, fmt.Errorf(
			"missing seed in Conjunct config YAML file",
		)
	}

	// XXX <05-10-2023, afjoseph> Don't expand symlinks here. There **is** a
	// difference between using clang and clang++ (it's not just a symlink).
	// Ref:
	// https://github.com/llvm/llvm-project/issues/54701#issuecomment-1086055306
	// In summary, using `clang++` links libstdc++ by default, while `clang`
	// doesn't, so expanding symlinks would force `clang++` to resolve to
	// `clang`, which will cause libstd++ linking errors.
	config.ClangPath, err = util.ExpandPath(config.ClangPath, false)
	if err != nil {
		return args, nil, fmt.Errorf(
			"Failed to expand clang path %s: %w",
			config.ClangPath,
			err,
		)
	}
	if err := util.CheckIfClangBinary(config.ClangPath); err != nil {
		return args, nil, fmt.Errorf(
			"Failed to find Clang at %s: %w",
			config.ClangPath,
			err,
		)
	}

	// XXX <06-10-2023, afjoseph> Don't expand symlinks here: this fails a few
	// unit tests where symlinks are not expanded
	config.OptPath, err = util.ExpandPath(config.OptPath, false)
	if err != nil {
		return args, nil, fmt.Errorf(
			"Failed to expand opt path %s: %w",
			config.OptPath,
			err,
		)
	}
	if err := util.CheckIfOptBinary(config.OptPath); err != nil {
		return args, nil, fmt.Errorf(
			"Failed to find Opt at %s: %w",
			config.OptPath,
			err,
		)
	}
	if argsparser.HasArg(args, "--conjunct-retain-temp-dir") {
		config.RetainTempDir = true
		args = argsparser.RemoveArg(
			args,
			"--conjunct-retain-temp-dir",
			false,
		)
	}

	logrus.Debugf("Parsed Conjunct config file successfully: %+v\n", config)
	return args, &config, nil
}
