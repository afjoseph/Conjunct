package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/afjoseph/conjunct/config"
	"github.com/afjoseph/conjunct/projectpath"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestEmitBitcode(t *testing.T) {
	clangPath, err := exec.LookPath("clang")
	require.NoError(t, err)
	testFilepath := filepath.Join(projectpath.Root, "testassets/unit/hello.c")
	// Run emitBitcode on a boring C file
	outPath, err := emitBitcode(
		"hello",
		clangPath,
		[]string{"-c", testFilepath},
		t.TempDir(),
		false, // isDryRun
	)
	require.NoError(t, err)
	// Check if bitcode is emitted
	cmd := exec.Command("file", outPath)
	ret, err := cmd.CombinedOutput()
	require.NoError(t, err)
	require.Contains(t, string(ret), "LLVM bitcode")
	err = os.Remove(outPath)
	require.NoError(t, err)
}

func TestBuildBitcode(t *testing.T) {
	clangPath, err := exec.LookPath("clang")
	require.NoError(t, err)
	testFilepath := filepath.Join(projectpath.Root, "testassets/unit/hello.bc")
	outPath, err := buildBitcode(
		clangPath,
		testFilepath,
		[]string{"-o", "hello"},
		false, // isDryRun
	)
	require.NoError(t, err)
	cmd := exec.Command("file", outPath)
	ret, err := cmd.CombinedOutput()
	require.NoError(t, err)
	require.Contains(t, string(ret), "Mach-O 64-bit object")
	err = os.Remove(outPath)
	require.NoError(t, err)
}

func TestRunOriginalClang(t *testing.T) {
	// Get clang binary from PATH (it doesn't matter here)
	clangPath, err := exec.LookPath("clang")
	require.NoError(t, err)

	// Target a hello world C file
	testFilepath := filepath.Join(projectpath.Root, "testassets/unit/hello.c")

	// Specify output path for object file
	compiledObjectOutPath := filepath.Join("/tmp/", "test_hello")
	RunOriginalClang(
		clangPath,
		[]string{
			"-o", compiledObjectOutPath,
			"-c", testFilepath,
		})

	// Make sure what we have is a compiled object file
	cmd := exec.Command("file", compiledObjectOutPath)
	ret, err := cmd.CombinedOutput()
	require.NoError(t, err)
	require.Contains(t, string(ret), "Mach-O 64-bit object")

	// Delete tmp files
	err = os.Remove(compiledObjectOutPath)
	require.NoError(t, err)
}

func TestRunConjunct(t *testing.T) {
	clangPath, err := exec.LookPath("clang")
	require.NoError(t, err)
	// Get opt binary from PATH (it doesn't matter here)
	optPath, err := exec.LookPath("opt")
	require.NoError(t, err)

	// Target a hello world C file
	testFilepath := filepath.Join(projectpath.Root, "testassets/unit/hello.c")
	config := &config.ConjunctConfig{
		Seed:      123,
		ClangPath: clangPath,
		OptPath:   optPath,
		Passes:    []string{},
	}

	// Specify output path for object file
	compiledObjectOutPath := filepath.Join("/tmp/", "test_hello")
	err = RunConjunct(
		config,
		[]string{
			"-o", compiledObjectOutPath,
			"-c", testFilepath,
		})
	require.NoError(t, err)

	// Make sure what we have is a compiled object file
	cmd := exec.Command("file", compiledObjectOutPath)
	ret, err := cmd.CombinedOutput()
	require.NoError(t, err)
	require.Contains(t, string(ret), "Mach-O 64-bit object")

	// Delete tmp files
	err = os.Remove(compiledObjectOutPath)
	require.NoError(t, err)
}
