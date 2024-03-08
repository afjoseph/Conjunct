package sourcefile

import (
	"path/filepath"

	"github.com/afjoseph/conjunct/argsparser"
)

type Type int

const (
	Type_C       Type = 1
	Type_CPP     Type = 2
	Type_OBJC    Type = 3
	Type_Unknown Type = -1
)

var (
	cppFileExtensions  = []string{".cpp", ".cc", ".cxx", ".c++"}
	objcFileExtensions = []string{".m"}
	cFileExtensions    = []string{".c"}
)

func FetchType(path string) Type {
	ext := filepath.Ext(path)
	for _, e := range cppFileExtensions {
		if e == ext {
			return Type_CPP
		}
	}
	for _, e := range objcFileExtensions {
		if e == ext {
			return Type_OBJC
		}
	}
	for _, e := range cFileExtensions {
		if e == ext {
			return Type_C
		}
	}
	return Type_Unknown
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
func GetSourceFileName(args []string) (string, Type) {
	// First method: get the value of -c argument
	sourceFileName := argsparser.GetArgVal(args, "-c")
	if len(sourceFileName) != 0 {
		// get the basename
		sourceFileName = filepath.Base(sourceFileName)
		return sourceFileName, FetchType(sourceFileName)
	}

	// Second method: run through all arguments and check if it's a C/CXX file
	for _, arg := range args {
		arg = filepath.Base(arg)
		t := FetchType(arg)
		if t == Type_Unknown {
			continue
		}
		return arg, t
	}

	return "", Type_Unknown
}
