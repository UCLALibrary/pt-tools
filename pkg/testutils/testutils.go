package testutils

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"testing"

	"github.com/otiai10/copy"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// MemorySink implements zap.Sink by writing all messages to a buffer.
type MemorySink struct {
	*bytes.Buffer
}

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

// CopyTestDirectory copies all contaents of a directory into a src directory
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
