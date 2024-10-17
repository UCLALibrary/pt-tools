package ptmv

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	error_msgs "github.com/UCLALibrary/pt-tools/pkg/error-msgs"
	"github.com/UCLALibrary/pt-tools/testutils"
	"github.com/mholt/archiver"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	root    = "--pairtree="
	rootDir = "pairtree_root"
)

// Test the basic functionality of ptmv
func TestPTMV(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		dest      string
		pairpath  string
		expectErr error
	}{
		{
			name:      "src is pairtree",
			src:       "ark:/b5488",
			dest:      "",
			pairpath:  filepath.Join("b5", "48", "8", "b5488"),
			expectErr: nil,
		},
		{
			name:      "dest is pairtree",
			src:       "",
			dest:      "ark:/b5488",
			pairpath:  filepath.Join("b5", "48", "8", "b5488"),
			expectErr: nil,
		},
		{
			name:      "dest is new pairtree",
			src:       "",
			dest:      "ark:/b2345",
			pairpath:  filepath.Join("b2", "34", "5", "b2345"),
			expectErr: nil,
		},
		{
			name:      "src and dest are both not pairtree",
			src:       "source",
			dest:      "",
			pairpath:  "",
			expectErr: error_msgs.Err10,
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
			var args []string
			var finalSrc string
			srcDir := testutils.CreateTempDir(t, fs)
			destDir := testutils.CreateTempDir(t, fs)
			if test.src == "" {
				//pairtree is the dest
				testutils.CopyTestDirectory(t, testutils.TestPairtree, destDir)
				// create file to copy to dest
				fileInSrc := testutils.CreateFileInDir(t, srcDir, "file.txt")
				args = []string{root + destDir, fileInSrc, test.dest}
				finalSrc = fileInSrc
			} else {
				// pairtree is the src
				testutils.CopyTestDirectory(t, testutils.TestPairtree, srcDir)
				args = []string{root + srcDir, test.src, destDir}
				finalSrc = filepath.Join(srcDir, rootDir, test.pairpath)
			}

			err := Run(args, &buf)
			require.ErrorIs(t, err, test.expectErr)

			// check if the src file or directory was deleted
			if test.expectErr == nil {
				_, err = os.Stat(finalSrc)
				assert.True(t, os.IsNotExist(err), "Expected path to not exist, but got: %v", err)
			}
		})
	}
}

// TestTar tests if an object in the pairtree is properly tared outside of it
func TestTar(t *testing.T) {
	var buf bytes.Buffer
	var args []string
	src := "ark:/a5388"
	tgzFile := "ark+=a5388.tgz"
	pairpath := filepath.Join("a5", "38", "8", "a5388")

	// Create a logger instance using the registered sink.
	logger, cleanup := testutils.SetupLogger(logFile)
	defer cleanup()
	Logger = logger

	fs := afero.NewOsFs()

	srcDir := testutils.CreateTempDir(t, fs)
	testutils.CopyTestDirectory(t, testutils.TestPairtree, srcDir)

	destDir := testutils.CreateTempDir(t, fs)
	destDir = filepath.Join(destDir, tgzFile)

	args = []string{root + srcDir, src, destDir, "-a"}
	fullSrc := filepath.Join(srcDir, pairpath)
	err := Run(args, &buf)
	require.ErrorIs(t, err, nil)

	// Check if the destination file exists
	exists, err := afero.Exists(fs, destDir)
	assert.ErrorIs(t, err, nil, "Failed to check if dirSrc was copied: %v", err)
	assert.True(t, exists, "File was not copied to destination")

	// check if the src was deleted
	_, err = os.Stat(fullSrc)
	assert.True(t, os.IsNotExist(err), "Expected path to not exist, but got: %v", err)
}

// TestUnTar tests a .tgz file is properly untarred into the pairtree
func TestUnTar(t *testing.T) {
	var buf bytes.Buffer
	var args []string

	dest := "ark:/a5388"
	pairpath := filepath.Join(rootDir, "a5", "38", "8", "a5388")
	ppBase := "a5388"

	// Create a logger instance using the registered sink.
	logger, cleanup := testutils.SetupLogger(logFile)
	defer cleanup()
	Logger = logger

	fs := afero.NewOsFs()
	srcDir := testutils.CreateTempDir(t, fs)
	destDir := testutils.CreateTempDir(t, fs)
	pairpath = filepath.Join(destDir, pairpath)

	testutils.CopyTestDirectory(t, testutils.TestPairtree, destDir)

	// Add files to src and .tgz file
	dirTGZ := testutils.CreateDirInDir(t, fs, srcDir, ppBase)

	dirSrcTGZ := filepath.Join(srcDir, ppBase+".tgz")

	fileNames := []string{"file.txt", "file1.txt", "file2.txt"}
	for _, fileName := range fileNames {
		_ = testutils.CreateFileInDir(t, dirTGZ, fileName)
	}

	// Archive the source directory
	tgz := archiver.NewTarGz()

	if err := tgz.Archive([]string{dirTGZ}, dirSrcTGZ); err != nil {
		t.Fatalf("There was an error archiving the folder %v", err)
	}

	args = []string{root + destDir, dirSrcTGZ, dest, "-a"}
	err := Run(args, &buf)
	require.ErrorIs(t, err, nil)

	// Check if source files were read properly
	files, err := afero.ReadDir(fs, pairpath)
	assert.ErrorIs(t, err, nil, "Failed to read dirSrc contents: %v", err)

	// Further checks can compare individual file names and contents
	for i, srcFile := range files {
		assert.Equal(t, srcFile.Name(), fileNames[i], "File names do not match")
	}

	// check if the src was deleted
	_, err = os.Stat(dirSrcTGZ)
	assert.True(t, os.IsNotExist(err), "Expected path to not exist, but got: %v", err)

}

// TestCLIError tests if an error is thrown when various CLI options are missing or are wrong
func TestCLIError(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectErr error
	}{
		{
			name:      "No pairtree root",
			args:      []string{"ID"},
			expectErr: error_msgs.Err7,
		},
		{
			name:      "Too many arguments passed in",
			args:      []string{root + "root", "ID", "subpath", "extra arg"},
			expectErr: error_msgs.Err8,
		},
		{
			name:      "Too few arguments passed in",
			args:      []string{root + "root", "ID"},
			expectErr: error_msgs.Err9,
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
			assert.ErrorIs(t, err, test.expectErr)
		})
	}

}
