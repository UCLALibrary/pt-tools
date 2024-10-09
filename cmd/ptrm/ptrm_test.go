package ptrm

// The test-dir that is copied and used throughout this test. Both the pairtree_version0_1
// and the pairtree_prefix are populated. The pairtree_prefix is populated with the prefix ark:/
// unless the test removes or changes that.
import (
	"bytes"
	"os"
	"testing"

	error_msgs "github.com/UCLALibrary/pt-tools/pkg/error-msgs"
	"github.com/UCLALibrary/pt-tools/testutils"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const (
	testDir = "test-pairtree"
	root    = "--pairtree="
)

// TestDelete tests if objects, files, and directories are deleted by ptrm
func TestDelete(t *testing.T) {
	tests := []struct {
		id            string
		path          []string
		expectedError error
	}{
		{id: "object", path: []string{"ark:/a54892"}, expectedError: nil},
		{id: "directory", path: []string{"ark:/b5488", "folder"}, expectedError: nil},
		{id: "file", path: []string{"ark:/a5388", "a5388.txt"}, expectedError: nil},
		{id: "notExist", path: []string{"ark:/idNotExist"}, expectedError: os.ErrNotExist},
		{id: "tooManyArgs", path: []string{"ark:/idNotExist", "folder", "toomanyargs"}, expectedError: error_msgs.Err8},
	}

	// Create a logger instance using the registered sink.
	logger, cleanup := testutils.SetupLogger(logFile)
	defer cleanup()
	Logger = logger

	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			fs := afero.NewOsFs()
			tempDir := testutils.CreateTempDir(t, fs)
			testutils.CopyTestDirectory(t, testutils.TestPairtree, tempDir)

			args := append([]string{root + tempDir}, test.path...)
			var buf bytes.Buffer

			err := Run(args, &buf)
			assert.ErrorIs(t, err, test.expectedError)
		})
	}

}

// TestCLIError tests if an error is thrown when various CLI options are missing
func TestCLIError(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectErr error
	}{
		{name: "No ID provided", args: []string{root + "root"}, expectErr: error_msgs.Err6},
		{name: "No pairtree root provided", args: []string{"ID"}, expectErr: error_msgs.Err7},
	}

	// Create a logger instance using the registered sink.
	logger, cleanup := testutils.SetupLogger(logFile)
	defer cleanup()
	Logger = logger

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := Run(test.args, &buf)
			assert.ErrorIs(t, err, test.expectErr, "Expected an error but got none")
		})
	}

}
