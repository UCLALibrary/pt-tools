package testutils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mholt/archiver"
	"github.com/otiai10/copy"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	root = "--pairtree="
)

// Path to the test pairtree directory
var (
	TestPairtree = filepath.Join("..", "..", "test-dir", "test-pairtree")
)

// MemorySink implements zap.Sink by writing all messages to a buffer.
type MemorySink struct {
	*bytes.Buffer
}

type RunFunc func(args []string, writer io.Writer) error

// Implement Close and Sync as no-ops to satisfy the interface. The Write
// method is provided by the embedded buffer.
func (s *MemorySink) Close() error { return nil }
func (s *MemorySink) Sync() error  { return nil }

// CreateLogger creates a test logger to be used
func CreateLogger() (Logger *zap.Logger, sink *MemorySink) {
	// Create a sink instance, and register it with zap for the "memory"
	// protocol.
	sink = &MemorySink{new(bytes.Buffer)}
	_ = zap.RegisterSink("memory", func(*url.URL) (zap.Sink, error) {
		return sink, nil
	})

	// Create a logger instance using the registered sink.
	Logger = zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(sink),
		zapcore.DebugLevel,
	))
	return Logger, sink
}

func SetupLogger(logFile string) (*zap.Logger, func()) {
	logger, _ := CreateLogger()

	// Create a cleanup function to be deferred
	cleanup := func() {
		if err := logger.Sync(); err != nil {
			// handle the error
			fmt.Printf("Failed to sync logger: %v\n", err)
		}
		CleanupFiles(logFile)
	}

	return logger, cleanup
}

// RedirectStdoutToBuffer redirects stdout to a buffer and returns it.
func RedirectStdoutToBuffer(t *testing.T) *bytes.Buffer {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	returnedBuffer := new(bytes.Buffer)

	go func() {
		_, _ = io.Copy(returnedBuffer, r)
		_ = w.Close() // Ensure the pipe is closed properly
	}()

	t.Cleanup(func() {
		_ = w.Close() // Ensure the write end is closed
		os.Stdout = oldStdout
	})

	return returnedBuffer
}

// CaptureStdout when do is called. Restore stdout as test cleanup
func CaptureStdout(t *testing.T, do func(t *testing.T)) string {
	t.Helper()
	orig := os.Stdout
	t.Cleanup(func() {
		os.Stdout = orig
	})

	r, w, _ := os.Pipe()
	os.Stdout = w
	do(t)
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()
	w.Close()
	return <-outC
}

// CreateTempDir uses afero to create a temporary directory that testing can be done on
func CreateTempDir(t *testing.T, fs afero.Fs) string {
	tempDir, err := afero.TempDir(fs, "", "test-dir-")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}

	// Use t.Cleanup to automatically remove the directory after the test completes
	t.Cleanup(func() {
		_ = fs.RemoveAll(tempDir)
	})

	return tempDir
}

// CreateTempFile creates a temporary file with the specified content
func CreateTempFile(t *testing.T, fs afero.Fs, content []byte) string {
	// Create a temporary file in the "temp" directory with a specific prefix
	tempFile, err := afero.TempFile(fs, "", "test-file")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer tempFile.Close()

	// Get the full path to the temporary file
	tempFilePath := tempFile.Name()

	// Write some data to the temporary file
	if _, err := tempFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	return tempFilePath
}

// CreateFileInDir creates a file with the given name in the specified directory and returns the path location
func CreateFileInDir(t *testing.T, dir string, filename string) string {
	// Join the directory and filename to get the full path of the new file.
	filePath := filepath.Join(dir, filename)

	// Create the file.
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	return filePath
}

// CreateDirInDir creates a directory in a directory
func CreateDirInDir(t *testing.T, fs afero.Fs, dir, newDir string) string {
	dirDest := filepath.Join(dir, newDir)
	err := fs.MkdirAll(dirDest, 0755) // Creates "subfolder" inside dirSrc
	if err != nil {
		t.Fatalf("Failed to create subfolder in dirSrc: %v", err)
	}
	return dirDest
}

// CopyTestDirectory copies all contents of a directory into a src directory
func CopyTestDirectory(t *testing.T, src, dst string) {
	err := copy.Copy(src, dst)
	if err != nil {
		t.Fatalf("Error copying directory from %s to %s: %v", src, dst, err)
	}
}

// CleanupFiles removes files that are not necessary
func CleanupFiles(file string) {
	os.Remove(file)
}

// CheckDirCopy checks if the a directory was fully copied to another directory
func CheckDirCopy(fs afero.Fs, srcDir, destDir, expFolderName string) error {
	// Check if the destination directory exists
	exists, err := afero.DirExists(fs, destDir)
	if err != nil {
		return fmt.Errorf("failed to check if directory was copied: %w", err)
	}
	if !exists {
		return fmt.Errorf("directory %s was not copied to destination: %s", srcDir, destDir)
	}

	// Check if source files were read properly
	srcFiles, err := afero.ReadDir(fs, srcDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory contents: %w", err)
	}

	// Ensure destination folder has the expected suffix
	if !strings.HasSuffix(destDir, expFolderName) {
		return fmt.Errorf("destination folder should have suffix %s", expFolderName)
	}

	// Check if destination files were read properly
	destFiles, err := afero.ReadDir(fs, destDir)
	if err != nil {
		return fmt.Errorf("failed to read destination directory contents: %w", err)
	}

	// Compare individual file names
	for _, srcFile := range srcFiles {
		found := false
		for _, destFile := range destFiles {
			if srcFile.Name() == destFile.Name() {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("file %s does not exist in destination directory", srcFile.Name())
		}
	}

	return nil
}

// UntarCLI is the untar testing for ptmv and ptcp untar component
func UntarCLI(t *testing.T, runFunc RunFunc, dest, pairpath, ppBase string, checkSrc bool) error {
	var buf bytes.Buffer

	fs := afero.NewOsFs()
	srcDir := CreateTempDir(t, fs)
	destDir := CreateTempDir(t, fs)
	pairpath = filepath.Join(destDir, pairpath)

	CopyTestDirectory(t, TestPairtree, destDir)

	// Add files to src and .tgz file
	dirTGZ := CreateDirInDir(t, fs, srcDir, ppBase)

	dirSrcTGZ := filepath.Join(srcDir, ppBase+".tgz")

	fileNames := []string{"file.txt", "file1.txt", "file2.txt"}
	for _, fileName := range fileNames {
		_ = CreateFileInDir(t, dirTGZ, fileName)
	}

	// Archive the source directory
	tgz := archiver.NewTarGz()

	if err := tgz.Archive([]string{dirTGZ}, dirSrcTGZ); err != nil {
		return err
	}

	args := []string{root + destDir, dirSrcTGZ, dest, "-a"}
	err := runFunc(args, &buf)
	if err != nil {
		return err
	}

	// Check if source files were read properly
	files, err := afero.ReadDir(fs, pairpath)
	if err != nil {
		return err
	}

	// Further checks can compare individual file names and contents
	for i, srcFile := range files {
		if srcFile.Name() != fileNames[i] {
			return fmt.Errorf("filenames after being untarred do not match: %s,%s", srcFile.Name(), fileNames[i])
		}
	}

	if checkSrc {
		_, err = os.Stat(dirSrcTGZ)
		if !os.IsNotExist(err) {
			return fmt.Errorf("the source directory should not exist, but it does %w", err)
		}
	}

	return nil
}

// TarCLI is the tar testing for ptmv and ptcp tar component
func TarCLI(t *testing.T, runFunc RunFunc, src, tgzFile string) error {
	var buf bytes.Buffer
	var args []string

	fs := afero.NewOsFs()

	srcDir := CreateTempDir(t, fs)
	CopyTestDirectory(t, TestPairtree, srcDir)

	destDir := CreateTempDir(t, fs)
	destDir = filepath.Join(destDir, tgzFile)

	args = []string{root + srcDir, src, destDir, "-a"}

	err := runFunc(args, &buf)
	if err != nil {
		return err
	}

	// Check if the destination file exists
	exists, err := afero.Exists(fs, destDir)
	if err != nil {
		return fmt.Errorf("the directory was not properly copied: %w", err)
	}

	if !exists {
		return errors.New("tar file was not properly copied to the desitnation")
	}
	return nil
}

// OpenFileAndCheck check the contents of a file
func OpenFileAndCheck(fs afero.Fs, filePath string) ([]byte, error) {
	// Open the file
	file, err := fs.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close() // Ensure the file is closed after we're done

	// Read the content of the file
	content, err := afero.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return content, nil
}
