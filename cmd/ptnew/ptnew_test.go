package ptnew

// The test-dir that is copied and used throughout this test. Both the pairtree_version0_1
// and the pairtree_prefix are populated. The pairtree_prefix is populated with the prefix ark:/
// unless the test removes or changes that.
import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	error_msgs "github.com/UCLALibrary/pt-tools/pkg/error-msgs"
	"github.com/UCLALibrary/pt-tools/testutils"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const (
	testDir = "test-pairtree"
	root    = "--pairtree="
	pre     = "--prefix="
)

// TestPtnew tests if an error is thrown when various CLI options are missing
func TestPtnew(t *testing.T) {
	tests := []struct {
		name         string
		pairtreeRoot string
		expectErr    error
	}{
		{
			name:         "Too many args with prefix",
			pairtreeRoot: "directory",
			expectErr:    nil,
		},
		{
			name:         "Too many args",
			pairtreeRoot: filepath.Join("directory", "innerDirectory"),
			expectErr:    nil,
		},
		{
			name:         "No pairtree root provided",
			pairtreeRoot: " ",
			expectErr:    error_msgs.Err15,
		},
	}

	// Create a logger instance using the registered sink.
	logger, cleanup := testutils.SetupLogger(logFile)
	defer cleanup()
	Logger = logger

	fs := afero.NewOsFs()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			var rootDir string
			if strings.TrimSpace(test.pairtreeRoot) == "" {
				rootDir = test.pairtreeRoot
			} else {
				rootDir = testutils.CreateTempDir(t, fs)
				rootDir = filepath.Join(rootDir, test.pairtreeRoot)
			}
			args := []string{root + rootDir, pre + "ark:/"}
			err := Run(args, &buf)
			assert.ErrorIs(t, err, test.expectErr)
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
		{
			name:      "Too many args with prefix",
			args:      []string{root + "root", pre + "ark:/", "argument"},
			expectErr: error_msgs.Err8,
		},
		{
			name:      "Too many args",
			args:      []string{root + "root", "argument"},
			expectErr: error_msgs.Err8,
		},
		{
			name:      "No pairtree root provided",
			args:      []string{"ID"},
			expectErr: error_msgs.Err7,
		},
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
