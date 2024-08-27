package ptls

// The test-dir that is copied and used throughout this test. Both the pairtree_version0_1
// and the pairtree_prefix are populated. The pairtree_prefix is populated with the prefix ark:/
// unless the test removes or changes that.
import (
	"bytes"
	"os"
	"testing"

	error_msgs "github.com/UCLALibrary/pt-tools/pkg/error-msgs"
	"github.com/UCLALibrary/pt-tools/pkg/testutils"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const (
	testPairtree = "../../test-dir/test-pairtree"
	testDir      = "test-pairtree"
	root         = "--pairtree="
)

// runTestWithArgs
func runTestWithArgs(t *testing.T, args, expected []string) {
	var buf bytes.Buffer

	err := Run(args, &buf)

	assert.NoError(t, err, "There was an error running ptls")

	// Get the output
	output := buf.String()

	// Check if the output contains the expected strings
	for _, expect := range expected {
		assert.Contains(t, output, expect)
	}
}

// TestNonRecursive tests only if nonrecursive files and directores are outputted, hidden directories and folders will not be included
func TestNonRecursive(t *testing.T) {
	tests := []struct {
		id       string
		expected []string
	}{
		{id: "ark:/a5388", expected: []string{"a5388.txt"}},
		{id: "ark:/a54892", expected: []string{"a54892.txt"}},
		{id: "ark:/b5488", expected: []string{"outerb5488.txt", "folder/"}},
	}

	// Create a logger instance using the registered sink.
	logger, cleanup := testutils.SetupLogger(logFile)
	defer cleanup()
	Logger = logger

	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			// var buf bytes.Buffer

			fs := afero.NewOsFs()
			tempDir := testutils.CreateTempDir(t, fs)
			testutils.CopyTestDirectory(t, testPairtree, tempDir)

			args := []string{root + tempDir, test.id}
			runTestWithArgs(t, args, test.expected)
		})
	}

}

// TestRecursive tests if recursive files and directores are outputted, hidden directories and folders will not be included
func TestRecursive(t *testing.T) {
	tests := []struct {
		id       string
		expected []string
	}{
		{id: "ark:/a5488", expected: []string{"a5488.txt"}},
		{id: "ark:/a54892", expected: []string{"a54892.txt"}},
		{id: "ark:/b5488", expected: []string{"outerb5488.txt", "folder/", "innerb5488.txt"}},
	}

	// Create a logger instance using the registered sink.
	logger, cleanup := testutils.SetupLogger(logFile)
	defer cleanup()

	Logger = logger

	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			fs := afero.NewOsFs()
			tempDir := testutils.CreateTempDir(t, fs)
			testutils.CopyTestDirectory(t, testPairtree, tempDir)

			// Backup original os.Args
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }() // Restore os.Args after test

			args := []string{root + tempDir, "-r", test.id}
			runTestWithArgs(t, args, test.expected)
		})
	}

}

// TestDirOnly tests if only directores are outputted, hidden directories and folders will not be included
func TestDirOnly(t *testing.T) {
	tests := []struct {
		id       string
		expected []string
	}{
		{id: "ark:/a54892", expected: []string{}},
		{id: "ark:/b5488", expected: []string{"folder/"}},
	}

	// Create a logger instance using the registered sink.
	logger, cleanup := testutils.SetupLogger(logFile)
	defer cleanup()
	Logger = logger

	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			fs := afero.NewOsFs()
			tempDir := testutils.CreateTempDir(t, fs)
			testutils.CopyTestDirectory(t, testPairtree, tempDir)

			// Backup original os.Args
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }() // Restore os.Args after test

			args := []string{root + tempDir, "-d", test.id}
			runTestWithArgs(t, args, test.expected)
		})
	}

}

// TestShowAll tests if all directories and files are outputted including hidden ones
func TestShowAll(t *testing.T) {
	tests := []struct {
		id       string
		expected []string
	}{
		{id: "ark:/a5388", expected: []string{"a5388.txt"}},
		{id: "ark:/a54892", expected: []string{".hidden/", ".hidden.txt", "a54892.txt"}},
	}

	// Create a logger instance using the registered sink.
	logger, cleanup := testutils.SetupLogger(logFile)
	defer cleanup()
	Logger = logger

	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			fs := afero.NewOsFs()
			tempDir := testutils.CreateTempDir(t, fs)
			testutils.CopyTestDirectory(t, testPairtree, tempDir)

			args := []string{root + tempDir, "-a", test.id}
			runTestWithArgs(t, args, test.expected)
		})
	}

}

// TestShowAllAndDirOnly tests if only directories including hidden ones are outputted
func TestShowAllAndDironly(t *testing.T) {
	tests := []struct {
		id       string
		expected []string
	}{
		{id: "ark:/a5388", expected: []string{}},
		{id: "ark:/a54892", expected: []string{".hidden/"}},
		{id: "ark:/b5488", expected: []string{"folder/"}},
	}

	// Create a logger instance using the registered sink.
	logger, cleanup := testutils.SetupLogger(logFile)
	defer cleanup()
	Logger = logger

	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			fs := afero.NewOsFs()
			tempDir := testutils.CreateTempDir(t, fs)
			testutils.CopyTestDirectory(t, testPairtree, tempDir)
			args := []string{root + tempDir, "-a", "-d", test.id}
			runTestWithArgs(t, args, test.expected)
		})
	}

}

// TestShowAllRecursive tests if all directories and files are outputted including hidden ones recursively
func TestShowAllRecursive(t *testing.T) {
	tests := []struct {
		id       string
		expected []string
	}{
		{id: "ark:/a54892", expected: []string{".hidden/", ".hidden.txt", "a54892.txt", "innerHidden.txt"}},
		{id: "ark:/b5488", expected: []string{"folder/", "outerb5488.txt", ".hidden/", ".hiddenFile.txt"}},
	}

	// Create a logger instance using the registered sink.
	logger, cleanup := testutils.SetupLogger(logFile)
	defer cleanup()
	Logger = logger

	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			fs := afero.NewOsFs()
			tempDir := testutils.CreateTempDir(t, fs)
			testutils.CopyTestDirectory(t, testPairtree, tempDir)
			args := []string{root + tempDir, "-r", "-a", test.id}
			runTestWithArgs(t, args, test.expected)
		})
	}

}

// TestDirOnlyRecursive tests if only directories are outputted recursively
func TestDirOnlyRecursive(t *testing.T) {
	tests := []struct {
		id       string
		expected []string
	}{
		{id: "ark:/a54892", expected: []string{".hidden/"}},
		{id: "ark:/b5488", expected: []string{"folder/", ".hidden/"}},
	}

	// Create a logger instance using the registered sink.
	logger, cleanup := testutils.SetupLogger(logFile)
	defer cleanup()
	Logger = logger

	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			fs := afero.NewOsFs()
			tempDir := testutils.CreateTempDir(t, fs)
			testutils.CopyTestDirectory(t, testPairtree, tempDir)
			args := []string{root + tempDir, "-r", "-a", test.id}
			runTestWithArgs(t, args, test.expected)
		})
	}

}

// TestCLIError tests if an error is thrown when various CLI options are missing
func TestCLIError(t *testing.T) {
	tests := []struct {
		name      string
		args      string
		expectErr error
	}{
		{name: "noID", args: root + "dir", expectErr: error_msgs.Err6},
		{name: "noRoot", args: "ID", expectErr: error_msgs.Err7},
	}

	// Create a logger instance using the registered sink.
	logger, cleanup := testutils.SetupLogger(logFile)
	defer cleanup()
	Logger = logger

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer

			args := []string{root + "dir"}
			err := Run(args, &buf)
			assert.Error(t, err, "Expected an error but got none")
		})
	}

}
