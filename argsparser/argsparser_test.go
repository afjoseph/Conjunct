package argsparser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveArg(t *testing.T) {
	var testcases = []struct {
		name                string
		inputArgs           []string
		targetArg           string
		targetArgVal        string
		inputDoRemoveArgVal bool
	}{
		{
			name: "Remove arg from argsparser that has it, with argval",
			inputArgs: []string{
				"-aaa",
				"aaaArgVal",
				"-bbb",
				"bbbArgVal",
			},
			targetArg:           "-bbb",
			targetArgVal:        "bbbArgVal",
			inputDoRemoveArgVal: true,
		},
		{
			name: "Remove arg from argsparser that has it, without argval",
			inputArgs: []string{
				"-aaa",
				"aaaArgVal",
				"-bbb",
				"bbbArgVal",
			},
			targetArg:           "-bbb",
			targetArgVal:        "bbbArgVal",
			inputDoRemoveArgVal: false,
		},
		{
			name: "Remove arg from argsparser that does not have it #1",
			inputArgs: []string{
				"-aaa",
				"aaaArgVal",
				"-bbb",
				"bbbArgVal",
			},
			targetArg:           "-ccc",
			targetArgVal:        "cccArgVal",
			inputDoRemoveArgVal: false,
		},
		{
			name: "Remove arg from argsparser that does not have it #2",
			inputArgs: []string{
				"-aaa",
				"aaaArgVal",
				"-bbb",
				"bbbArgVal",
			},
			targetArg:           "-ccc",
			targetArgVal:        "cccArgVal",
			inputDoRemoveArgVal: true,
		},
		{
			name:                "Remove arg from empty argsparser #1",
			inputArgs:           []string{},
			targetArg:           "-ccc",
			targetArgVal:        "cccArgVal",
			inputDoRemoveArgVal: true,
		},
		{
			name:                "Remove arg from empty argsparser #2",
			inputArgs:           []string{},
			targetArg:           "-ccc",
			targetArgVal:        "cccArgVal",
			inputDoRemoveArgVal: false,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualArgs := RemoveArg(
				tc.inputArgs,
				tc.targetArg,
				tc.inputDoRemoveArgVal,
			)
			require.NotContains(t, actualArgs, tc.targetArg)
			if tc.inputDoRemoveArgVal {
				require.NotContains(t, actualArgs, tc.targetArgVal)
			}
		})
	}
}

func TestRemoveRegexArg(t *testing.T) {
	var testcases = []struct {
		name                 string
		inputArgs            []string
		targetArg            string
		expectedReturnedArgs []string
	}{
		{
			name: "Remove arg from argsparser that has it",
			inputArgs: []string{
				"-aaa",
				"aaaArgVal",
				"-bbb",
				"bbbArgVal",
			},
			targetArg:            "-bbb",
			expectedReturnedArgs: []string{"-aaa", "aaaArgVal", "bbbArgVal"},
		},
		{
			name:                 "Remove arg from argsparser that doesn't have it",
			inputArgs:            []string{"-aaa", "aaaArgVal", "bbbArgVal"},
			targetArg:            "-bbb",
			expectedReturnedArgs: []string{"-aaa", "aaaArgVal", "bbbArgVal"},
		},
		{
			name: "Remove arg from argsparser that have it (regex)",
			inputArgs: []string{
				"-aaa",
				"aaaArgVal",
				"-bbb",
				"bbbArgVal",
			},
			targetArg:            "-b+",
			expectedReturnedArgs: []string{"-aaa", "aaaArgVal", "bbbArgVal"},
		},
		{
			name:                 "Remove arg from argsparser that doesn't have it (regex)",
			inputArgs:            []string{"-aaa", "aaaArgVal", "bbbArgVal"},
			targetArg:            "-b+",
			expectedReturnedArgs: []string{"-aaa", "aaaArgVal", "bbbArgVal"},
		},
		{
			name: "Remove duplicate arg from argsparser (regex)",
			inputArgs: []string{
				"-bbb",
				"-bbb",
				"aaaArgVal",
				"bbbArgVal",
			},
			targetArg:            "-b+",
			expectedReturnedArgs: []string{"aaaArgVal", "bbbArgVal"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualReturnedArgs := RemoveRegexArg(tc.inputArgs, tc.targetArg)
			require.Equal(t, tc.expectedReturnedArgs, actualReturnedArgs)
		})
	}
}

func TestAddArg(t *testing.T) {
	var testcases = []struct {
		name             string
		inputArgs        []string
		targetArg        string
		targetArgVal     string
		expectedArgsSize int
	}{
		{
			name: "Add arg to argsparser that does not have it",
			inputArgs: []string{
				"-aaa",
				"aaaArgVal",
				"-bbb",
				"bbbArgVal",
			},
			targetArg:        "-ccc",
			targetArgVal:     "cccArgVal",
			expectedArgsSize: 6,
		},
		{
			name: "Add arg to argsparser that has it",
			inputArgs: []string{
				"-aaa",
				"aaaArgVal",
				"-ccc",
				"cccArgVal",
			},
			targetArg:        "-ccc",
			targetArgVal:     "cccArgVal",
			expectedArgsSize: 6,
		},
		{
			name:             "add empty arg",
			inputArgs:        []string{},
			targetArg:        "",
			targetArgVal:     "",
			expectedArgsSize: 0,
		},
		{
			name:             "add arg with empty argval",
			inputArgs:        []string{},
			targetArg:        "-aaaaa",
			targetArgVal:     "",
			expectedArgsSize: 1,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualArgs := AddArg(tc.inputArgs, tc.targetArg, tc.targetArgVal)
			if len(tc.targetArg) != 0 {
				require.Contains(t, actualArgs, tc.targetArg)
			}
			if len(tc.targetArgVal) != 0 {
				require.Contains(t, actualArgs, tc.targetArgVal)
			}
			require.Equal(t, len(actualArgs), tc.expectedArgsSize)
		})
	}
}

func TestHasArg(t *testing.T) {
	var testcases = []struct {
		name           string
		inputArgs      []string
		targetArg      string
		expectedRetval bool
	}{
		{
			name:           "Has arg",
			inputArgs:      []string{"-aaa", "aaaArgVal", "-bbb", "bbbArgVal"},
			targetArg:      "-bbb",
			expectedRetval: true,
		},
		{
			name:           "Does not have arg",
			inputArgs:      []string{"-aaa", "aaaArgVal", "-bbb", "bbbArgVal"},
			targetArg:      "-ccc",
			expectedRetval: false,
		},
		{
			name:           "Empty args",
			inputArgs:      []string{},
			targetArg:      "-ccc",
			expectedRetval: false,
		},
		{
			name:           "Empty value",
			inputArgs:      []string{"-aaa", "aaaArgVal", "-bbb", "bbbArgVal"},
			targetArg:      "",
			expectedRetval: false,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(
				t,
				tc.expectedRetval,
				HasArg(tc.inputArgs, tc.targetArg),
			)
		})
	}
}

func TestGetArgVal(t *testing.T) {
	var testcases = []struct {
		name           string
		inputArgs      []string
		inputTargetArg string
		expectedRetval string
	}{
		{
			name:           "Has Arg with argval",
			inputArgs:      []string{"-aaa", "aaaArgVal", "-bbb", "bbbArgVal"},
			inputTargetArg: "-bbb",
			expectedRetval: "bbbArgVal",
		},
		{
			name:           "Has Arg without argval",
			inputArgs:      []string{"-aaa", "aaaArgVal", "-bbb"},
			inputTargetArg: "-bbb",
			expectedRetval: "",
		},
		{
			name:           "Empty args",
			inputArgs:      []string{},
			inputTargetArg: "-bbb",
			expectedRetval: "",
		},
		{
			name:           "Empty targetArg",
			inputArgs:      []string{"-aaa", "aaaArgVal", "-bbb"},
			inputTargetArg: "",
			expectedRetval: "",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(
				t,
				tc.expectedRetval,
				GetArgVal(tc.inputArgs, tc.inputTargetArg),
			)
		})
	}
}
