package argsparser

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/afjoseph/conjunct/util"
)

// RemoveArg Remove 'targetArg' from 'args'.
// If 'removeArgVal' is true, remove the argument value as well.
//
// Example:
//
// args := []string{"-c", "hello.c", "-o", "hello"}
// args = RemoveArg(args, "-c", true)
// // args is now []string{"-o", "hello"}
// args = RemoveArg(args, "-o", false)
// // args is now []string{"hello"}
//
// Returns modified list of arguments
func RemoveArg(args []string, targetArg string, removeArgVal bool) []string {
	if len(targetArg) == 0 {
		return args
	}
	for i, elem := range args {
		if elem == targetArg {
			if removeArgVal {
				args = append(args[:i], args[i+1:]...)
				args = append(args[:i], args[i+1:]...)
			} else {
				args = append(args[:i], args[i+1:]...)
			}
			return args
		}
	}
	return args
}

// RemoveRegexArg removes arguments from 'args' that match 'regex'.
// Returns modified list of arguments
func RemoveRegexArg(args []string, regex string) []string {
	if len(regex) == 0 {
		return args
	}
	var newArgs []string
	re := regexp.MustCompile(regex)
	for _, elem := range args {
		if re.MatchString(elem) {
			continue
		}
		newArgs = append(newArgs, elem)
	}
	return newArgs
}

// Addarg adds 'newArg' to 'args'.
// Returns modified list of arguments
func AddArg(args []string, newArg string, newArgVal string) []string {
	if len(newArg) == 0 {
		return args
	}
	args = append(args, newArg)
	if len(newArgVal) != 0 {
		args = append(args, newArgVal)
	}
	return args
}

// HasArg returns true if 'args' contains 'targetArgName'
func HasArg(args []string, targetArgName string) bool {
	if len(args) == 0 || len(targetArgName) == 0 {
		return false
	}
	for _, elem := range args {
		if elem == targetArgName {
			return true
		}
	}
	return false
}

// GetArgVal returns the value of 'targetArgName' in 'args'.
func GetArgVal(args []string, targetArgName string) string {
	if len(args) == 0 || len(targetArgName) == 0 {
		return ""
	}
	for i, elem := range args {
		if elem == targetArgName && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// GetSourceFileName fetches the source file name from 'args'
// There are two methods:
//   - First one is to get the value of -c argument since most compilers
//     put the source file name there. It's not a guarantee, just a convention,
//     so this can fail
//   - Second one is to get the argument containing a file with a .c or .cpp
//
// XXX <02-03-2024, afjoseph> Both methods are not accurate so I'm waiting for
// the command that breaks this function breaks to make it better
func GetSourceFileName(args []string) (string, util.SourceFileType, error) {
	// First method: get the value of -c argument
	sourceFileName := GetArgVal(args, "-c")
	if len(sourceFileName) != 0 {
		// get the basename
		sourceFileName = filepath.Base(sourceFileName)
		t, err := util.FetchSourceFileType(sourceFileName)
		if err != nil {
			return sourceFileName, util.SourceFileType_Unknown, fmt.Errorf(
				"while checking source file type for %s: %w",
				sourceFileName,
				err,
			)
		}
		return sourceFileName, t, nil
	}

	// Second method: run through all arguments and check if it's a C/CXX file
	for _, arg := range args {
		arg = filepath.Base(arg)
		t, err := util.FetchSourceFileType(arg)
		if err != nil {
			continue
		}
		return arg, t, nil
	}

	return "", util.SourceFileType_Unknown, fmt.Errorf("source file not found")
}
