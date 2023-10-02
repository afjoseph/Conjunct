package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/afjoseph/conjunct/projectpath"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestExtractConfig(t *testing.T) {
	clangPath, err := exec.LookPath("clang")
	require.NoError(t, err)
	os.Setenv("CLANG_PATH", clangPath)
	optPath, err := exec.LookPath("opt")
	require.NoError(t, err)
	os.Setenv("OPT_PATH", optPath)

	var testcases = []struct {
		name                      string
		inputArgs                 []string
		expectedShouldRunConjunct bool
		expectedConfig            *ConjunctConfig
		expectedErrorHas          error
	}{
		{
			name: "Good #1: object compilation",
			inputArgs: []string{"--conjunct-config-path",
				filepath.Join(projectpath.Root,
					"testassets/unit", "example_config_1.yaml"),
				"-c"},
			expectedShouldRunConjunct: true,
			expectedConfig: &ConjunctConfig{
				Seed:      123456789,
				ClangPath: clangPath,
				OptPath:   optPath,
				Passes:    []string{"lowerswitch"},
				OptExtraArgs: map[string]string{
					"key1": "val1",
					"key2": "val2",
				},
			},
		},
		{
			name: "Good #2: linking step (no '-c' param)",
			inputArgs: []string{"--conjunct_config_path",
				filepath.Join(projectpath.Root,
					"testassets/unit", "example_config_1.yaml")},
			expectedShouldRunConjunct: false,
			expectedConfig: &ConjunctConfig{
				Seed:      123456789,
				ClangPath: clangPath,
				OptPath:   optPath,
				Passes:    []string{"lowerswitch"},
				OptExtraArgs: map[string]string{
					"key1": "val1",
					"key2": "val2",
				},
			},
		},
		{
			name:                      "Params not found",
			inputArgs:                 []string{},
			expectedShouldRunConjunct: false,
			expectedConfig:            nil,
			expectedErrorHas:          fmt.Errorf("no such file or directory"),
		},
		{
			name: "Param found but bad yaml",
			inputArgs: []string{"--conjunct_config_path",
				filepath.Join(
					projectpath.Root,
					"main.go")},
			expectedShouldRunConjunct: false,
			expectedConfig:            nil,
			expectedErrorHas:          fmt.Errorf("cannot unmarshal"),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualArgs, actualConfig, actualShouldRunConjunct, err := ExtractConfigFromArgs(
				tc.inputArgs,
			)
			if err != nil {
				require.NotNil(t, tc.expectedErrorHas, err)
				require.Contains(t, strings.ToLower(err.Error()),
					strings.ToLower(tc.expectedErrorHas.Error()))
			}
			if tc.expectedShouldRunConjunct {
				require.True(t, actualShouldRunConjunct)
				require.Equal(t, tc.expectedConfig, actualConfig)
				require.NotContains(t, actualArgs, "-conjunct_config")
			} else {
				require.False(t, actualShouldRunConjunct)
			}
		})
	}
}
