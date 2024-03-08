package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/afjoseph/conjunct/projectpath"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestExtractConfig(t *testing.T) {
	// Set up environment variables since our test YAML config files use these
	// paths to find clang and opt on your development machine
	clangPath, err := exec.LookPath("clang")
	require.NoError(t, err)
	os.Setenv("CLANG_DIR_PATH", filepath.Dir(clangPath))
	optPath, err := exec.LookPath("opt")
	require.NoError(t, err)
	os.Setenv("OPT_PATH", optPath)

	var testcases = []struct {
		name           string
		inputArgs      []string
		expectedConfig *Config
		expectedError  error
	}{
		{
			name: "Good #1: object compilation",
			inputArgs: []string{
				"--conjunct-config-path",
				filepath.Join(
					projectpath.Root,
					"testassets/unit",
					"example_config_1.yaml"),
				"-c",
				"whatever.c"},
			expectedConfig: &Config{
				Seed:         123456789,
				ClangDirPath: clangPath,
				OptPath:      optPath,
				OptCLIArgs:   []string{"--lowerswitch"},
			},
		},
		{
			name: "Good #2: linking step (no '-c' param)",
			inputArgs: []string{
				"--conjunct-config-path",
				filepath.Join(
					projectpath.Root,
					"testassets/unit",
					"example_config_1.yaml")},
			expectedConfig: &Config{
				Seed:         123456789,
				ClangDirPath: clangPath,
				OptPath:      optPath,
				OptCLIArgs:   []string{"--lowerswitch"},
			},
		},
		{
			name:           "Params not found",
			inputArgs:      []string{},
			expectedConfig: nil,
			expectedError:  nil,
		},
		{
			name: "Param found but bad yaml",
			inputArgs: []string{
				"--conjunct-config-path",
				filepath.Join(
					projectpath.Root,
					"main.go")},
			expectedConfig: nil,
			expectedError:  ErrParsingConfig,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualArgs, actualConfig, err := ExtractConfigFromArgs(
				tc.inputArgs,
			)
			if err != nil {
				require.NotNil(t, tc.expectedError, err)
				require.Contains(
					t,
					err.Error(),
					tc.expectedError.Error(),
				)
			}
			require.Equal(t, tc.expectedConfig, actualConfig)
			require.NotContains(t, actualArgs, "-conjunct-config-path")
		})
	}
}
