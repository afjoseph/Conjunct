package config

import (
	stderr "errors"
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

type Config struct {
	// Seed is the seed used for random number generation. Useful for some
	// passes
	Seed int64 `yaml:"seed"`
	// ClangDirPath is the path to the Clang binary
	ClangDirPath string `yaml:"clang-dir-path"`
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
) (retArgs []string, cfg *Config, err error) {
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
		return args, nil, errors.Wrapf(err, "failed to read YAML file")
	}
	if len(configFileContent) == 0 {
		return args, nil, errors.New("empty config file")
	}
	config := Config{}
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
		return args, nil, errors.New("missing seed in config")
	}

	// XXX <05-10-2023, afjoseph> Don't expand symlinks here. There **is** a
	// difference between using clang and clang++ (it's not just a symlink).
	// Ref:
	// https://github.com/llvm/llvm-project/issues/54701#issuecomment-1086055306
	// In summary, using `clang++` links libstdc++ by default, while `clang`
	// doesn't, so expanding symlinks would force `clang++` to resolve to
	// `clang`, which will cause libstd++ linking errors.
	config.ClangDirPath, err = util.ExpandPath(config.ClangDirPath, false)
	if err != nil {
		return args, nil, errors.Wrapf(
			err,
			"failed to expand clang dir path: %s",
			config.ClangDirPath,
		)
	}

	// XXX <06-10-2023, afjoseph> Don't expand symlinks here: this fails a few
	// unit tests where symlinks are not expanded
	config.OptPath, err = util.ExpandPath(config.OptPath, false)
	if err != nil {
		return args, nil, errors.Wrapf(
			err,
			"failed to expand opt path: %s",
			config.OptPath,
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

	logrus.Debugf("Parsed Conjunct config file successfully: %+v", config)
	return args, &config, nil
}
