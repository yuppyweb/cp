package main_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Helper type to manage test scenarios.
type e2eTestEnv struct {
	t       *testing.T
	tempDir string
	binPath string
}

// newE2EEnv creates a new E2E test environment.
func newE2EEnv(t *testing.T) *e2eTestEnv {
	t.Helper()

	tempDir := t.TempDir()

	// Determine binary path based on OS
	binName := "cp"
	if os.Getenv("GOOS") == "windows" || os.PathListSeparator == ';' {
		binName = "cp.exe"
	}

	binPath := filepath.Join(tempDir, binName)

	// Build the binary for testing in the temp directory
	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", binPath, ".")

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Dir = "." // Build from current directory

	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v\nstderr: %s", err, stderr.String())
	}

	// Verify binary was created
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("binary not found at %s: %v", binPath, err)
	}

	return &e2eTestEnv{
		t:       t,
		tempDir: tempDir,
		binPath: binPath,
	}
}

// runCmd executes the cp command with given arguments.
func (env *e2eTestEnv) runCmd(args ...string) (string, string, int) {
	env.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, env.binPath, args...) //nolint:gosec

	var outBuf, errBuf bytes.Buffer

	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode := 0

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			// For other errors, set exit code to 1
			exitCode = 1
		}
	}

	stdout := outBuf.String()
	stderr := errBuf.String()

	return stdout, stderr, exitCode
}

// createFile creates a test file with given content.
func (env *e2eTestEnv) createFile(path string, content string) {
	env.t.Helper()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		env.t.Fatalf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		env.t.Fatalf("failed to create file: %v", err)
	}
}

// readFile reads a file and returns its content.
func (env *e2eTestEnv) readFile(path string) string {
	env.t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		env.t.Fatalf("failed to read file: %v", err)
	}

	return string(content)
}

// fileExists checks if a file exists.
func (env *e2eTestEnv) fileExists(path string) bool {
	env.t.Helper()

	_, err := os.Stat(path)

	return err == nil
}

// TestE2E_SimpleFileCopy tests basic file copy operation.
func TestE2E_SimpleFileCopy(t *testing.T) {
	t.Parallel()

	env := newE2EEnv(t)
	defer os.RemoveAll(env.tempDir)

	sourceFile := filepath.Join(env.tempDir, "source.txt")
	destFile := filepath.Join(env.tempDir, "dest.txt")
	expectedContent := "Hello, E2E World!"

	// Arrange
	env.createFile(sourceFile, expectedContent)

	// Act
	stdout, stderr, exitCode := env.runCmd(sourceFile, destFile)

	// Assert
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr)
	}

	if !env.fileExists(destFile) {
		t.Error("destination file was not created")
	}

	content := env.readFile(destFile)
	if content != expectedContent {
		t.Errorf("content mismatch: got %q, want %q", content, expectedContent)
	}

	if !strings.Contains(stdout, "successfully") {
		t.Errorf("expected success message in stdout, got: %q", stdout)
	}
}

// TestE2E_CopyMultilineFile tests copying a file with multiple lines.
func TestE2E_CopyMultilineFile(t *testing.T) {
	t.Parallel()

	env := newE2EEnv(t)
	defer os.RemoveAll(env.tempDir)

	sourceFile := filepath.Join(env.tempDir, "source.txt")
	destFile := filepath.Join(env.tempDir, "dest.txt")
	expectedContent := "Line 1\nLine 2\nLine 3\n"

	// Arrange
	env.createFile(sourceFile, expectedContent)

	// Act
	_, stderr, exitCode := env.runCmd(sourceFile, destFile)

	// Assert
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr)
	}

	content := env.readFile(destFile)
	if content != expectedContent {
		t.Errorf("content mismatch: got %q, want %q", content, expectedContent)
	}
}

// TestE2E_CopyBinaryFile tests copying a binary file.
func TestE2E_CopyBinaryFile(t *testing.T) {
	t.Parallel()

	env := newE2EEnv(t)
	defer os.RemoveAll(env.tempDir)

	sourceFile := filepath.Join(env.tempDir, "binary.bin")
	destFile := filepath.Join(env.tempDir, "binary_copy.bin")

	// Arrange: Create binary file
	binaryContent := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}
	if err := os.WriteFile(sourceFile, binaryContent, 0o600); err != nil {
		t.Fatalf("failed to create binary file: %v", err)
	}

	// Act
	_, stderr, exitCode := env.runCmd(sourceFile, destFile)

	// Assert
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr)
	}

	destContent, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if !bytes.Equal(destContent, binaryContent) {
		t.Error("binary content mismatch")
	}
}

// TestE2E_CopyWithSubdirectories tests copying file to subdirectory.
func TestE2E_CopyWithSubdirectories(t *testing.T) {
	t.Parallel()

	env := newE2EEnv(t)
	defer os.RemoveAll(env.tempDir)

	sourceFile := filepath.Join(env.tempDir, "source.txt")
	destDir := filepath.Join(env.tempDir, "subdir", "deep", "dir")
	destFile := filepath.Join(destDir, "dest.txt")
	expectedContent := "Subdirectory test"

	// Arrange
	env.createFile(sourceFile, expectedContent)

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatalf("failed to create subdirectories: %v", err)
	}

	// Act
	_, stderr, exitCode := env.runCmd(sourceFile, destFile)

	// Assert
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr)
	}

	content := env.readFile(destFile)
	if content != expectedContent {
		t.Errorf("content mismatch: got %q, want %q", content, expectedContent)
	}
}

// TestE2E_OverwriteExistingFile tests overwriting an existing file.
func TestE2E_OverwriteExistingFile(t *testing.T) {
	t.Parallel()

	env := newE2EEnv(t)
	defer os.RemoveAll(env.tempDir)

	sourceFile := filepath.Join(env.tempDir, "source.txt")
	destFile := filepath.Join(env.tempDir, "dest.txt")
	oldContent := "Old content"
	newContent := "New content"

	// Arrange
	env.createFile(sourceFile, newContent)
	env.createFile(destFile, oldContent)

	// Act
	_, stderr, exitCode := env.runCmd(sourceFile, destFile)

	// Assert
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr)
	}

	content := env.readFile(destFile)
	if content != newContent {
		t.Errorf("content mismatch: got %q, want %q", content, newContent)
	}
}

// TestE2E_SourceFileNotFound tests error when source doesn't exist.
func TestE2E_SourceFileNotFound(t *testing.T) {
	t.Parallel()

	env := newE2EEnv(t)
	defer os.RemoveAll(env.tempDir)

	sourceFile := filepath.Join(env.tempDir, "nonexistent.txt")
	destFile := filepath.Join(env.tempDir, "dest.txt")

	// Act
	stdout, stderr, exitCode := env.runCmd(sourceFile, destFile)

	// Assert
	if exitCode == 0 {
		t.Error("expected non-zero exit code for missing source file")
	}

	if !strings.Contains(stderr, "Error") && !strings.Contains(stdout, "Error") {
		t.Errorf("expected error message, got stdout: %q, stderr: %q", stdout, stderr)
	}

	if env.fileExists(destFile) {
		t.Error("destination file should not be created when source doesn't exist")
	}
}

// TestE2E_SameSourceAndDest tests error when source and dest are same.
func TestE2E_SameSourceAndDest(t *testing.T) {
	t.Parallel()

	env := newE2EEnv(t)
	defer os.RemoveAll(env.tempDir)

	file := filepath.Join(env.tempDir, "file.txt")
	env.createFile(file, "content")

	// Act
	stdout, stderr, exitCode := env.runCmd(file, file)

	// Assert
	if exitCode == 0 {
		t.Error("expected non-zero exit code when source equals destination")
	}

	output := stdout + stderr
	if !strings.Contains(output, "same") {
		t.Errorf("expected 'same' in error message, got stdout: %q, stderr: %q", stdout, stderr)
	}
}

// TestE2E_MissingArguments tests error with missing arguments.
func TestE2E_MissingArguments(t *testing.T) {
	t.Parallel()

	env := newE2EEnv(t)
	t.Cleanup(func() {
		os.RemoveAll(env.tempDir)
	})

	tests := []struct {
		name      string
		args      []string
		wantUsage bool
	}{
		{
			name:      "no arguments",
			args:      []string{},
			wantUsage: true,
		},
		{
			name:      "only source",
			args:      []string{filepath.Join(env.tempDir, "source.txt")},
			wantUsage: true,
		},
		{
			name: "too many arguments",
			args: []string{
				filepath.Join(env.tempDir, "src"),
				filepath.Join(env.tempDir, "dst"),
				"extra",
			},
			wantUsage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, exitCode := env.runCmd(tt.args...)

			if exitCode == 0 {
				t.Error("expected non-zero exit code")
			}
		})
	}
}

// TestE2E_CopyWithRelativePaths tests copying with relative paths.
func TestE2E_CopyWithRelativePaths(t *testing.T) { //nolint:paralleltest
	env := newE2EEnv(t)
	defer os.RemoveAll(env.tempDir)

	// Change to temp directory
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(oldCwd); err != nil { //nolint:usetesting
			t.Logf("failed to restore directory: %v", err)
		}
	})

	t.Chdir(env.tempDir)

	sourceFile := "source.txt"
	destFile := "dest.txt"
	expectedContent := "Relative paths test"

	// Arrange
	env.createFile(sourceFile, expectedContent)

	// Act: Use relative paths
	_, stderr, exitCode := env.runCmd(sourceFile, destFile)

	// Assert
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr)
	}

	content := env.readFile(destFile)
	if content != expectedContent {
		t.Errorf("content mismatch: got %q, want %q", content, expectedContent)
	}
}

// TestE2E_CopyLargeFile tests copying a large file.
func TestE2E_CopyLargeFile(t *testing.T) {
	t.Parallel()

	env := newE2EEnv(t)
	defer os.RemoveAll(env.tempDir)

	sourceFile := filepath.Join(env.tempDir, "large.bin")
	destFile := filepath.Join(env.tempDir, "large_copy.bin")

	// Arrange: Create 50MB file
	largeContent := bytes.Repeat([]byte("x"), 50*1024*1024)
	if err := os.WriteFile(sourceFile, largeContent, 0o600); err != nil {
		t.Fatalf("failed to create large file: %v", err)
	}

	// Act
	_, stderr, exitCode := env.runCmd(sourceFile, destFile)

	// Assert
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr)
	}

	// Verify file sizes match
	sourceInfo, err := os.Stat(sourceFile)
	if err != nil {
		t.Fatalf("failed to stat source: %v", err)
	}

	destInfo, err := os.Stat(destFile)
	if err != nil {
		t.Fatalf("failed to stat destination: %v", err)
	}

	if sourceInfo.Size() != destInfo.Size() {
		t.Errorf("file size mismatch: got %d, want %d", destInfo.Size(), sourceInfo.Size())
	}
}

// TestE2E_CopyEmptyFile tests copying an empty file.
func TestE2E_CopyEmptyFile(t *testing.T) {
	t.Parallel()

	env := newE2EEnv(t)
	defer os.RemoveAll(env.tempDir)

	sourceFile := filepath.Join(env.tempDir, "empty.txt")
	destFile := filepath.Join(env.tempDir, "empty_copy.txt")

	// Arrange: Create empty file
	f, err := os.Create(sourceFile)
	if err != nil {
		t.Fatalf("failed to create empty file: %v", err)
	}

	f.Close() // Close immediately

	// Act
	_, stderr, exitCode := env.runCmd(sourceFile, destFile)

	// Assert
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr)
	}

	content := env.readFile(destFile)
	if len(content) != 0 {
		t.Errorf("expected empty file, got %d bytes", len(content))
	}
}

// TestE2E_CopyFileWithSpecialCharacters tests copying files with special characters in names.
func TestE2E_CopyFileWithSpecialCharacters(t *testing.T) {
	t.Parallel()

	env := newE2EEnv(t)
	defer os.RemoveAll(env.tempDir)

	sourceFile := filepath.Join(env.tempDir, "file with spaces & special.txt")
	destFile := filepath.Join(env.tempDir, "copy with spaces & special.txt")
	expectedContent := "Special characters test"

	// Arrange
	env.createFile(sourceFile, expectedContent)

	// Act
	_, stderr, exitCode := env.runCmd(sourceFile, destFile)

	// Assert
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", exitCode, stderr)
	}

	content := env.readFile(destFile)
	if content != expectedContent {
		t.Errorf("content mismatch: got %q, want %q", content, expectedContent)
	}
}

// TestE2E_OutputMessage verifies the success message format.
func TestE2E_OutputMessage(t *testing.T) {
	t.Parallel()

	env := newE2EEnv(t)
	defer os.RemoveAll(env.tempDir)

	sourceFile := filepath.Join(env.tempDir, "source.txt")
	destFile := filepath.Join(env.tempDir, "dest.txt")

	// Arrange
	env.createFile(sourceFile, "test content")

	// Act
	stdout, _, exitCode := env.runCmd(sourceFile, destFile)

	// Assert
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	// Check output format
	if !strings.Contains(stdout, "File copied") {
		t.Errorf("expected 'File copied' in output, got: %q", stdout)
	}

	if !strings.Contains(stdout, sourceFile) {
		t.Errorf("expected source file path in output, got: %q", stdout)
	}

	if !strings.Contains(stdout, destFile) {
		t.Errorf("expected destination file path in output, got: %q", stdout)
	}

	if !strings.Contains(stdout, "successfully") {
		t.Errorf("expected 'successfully' in output, got: %q", stdout)
	}
}

// TestE2E_SequentialCopies tests multiple sequential copy operations.
func TestE2E_SequentialCopies(t *testing.T) {
	t.Parallel()

	env := newE2EEnv(t)
	defer os.RemoveAll(env.tempDir)

	// Arrange
	sourceFile := filepath.Join(env.tempDir, "source.txt")
	env.createFile(sourceFile, "content")

	// Act & Assert: Perform multiple copies
	for idx := range 5 {
		destFile := filepath.Join(env.tempDir, fmt.Sprintf("dest_%d.txt", idx))
		_, stderr, exitCode := env.runCmd(sourceFile, destFile)

		if exitCode != 0 {
			t.Errorf("copy %d failed: exit code %d, stderr: %s", idx, exitCode, stderr)
		}

		if !env.fileExists(destFile) {
			t.Errorf("copy %d: destination file not created", idx)
		}
	}
}

// TestE2E_ConcurrentCopies tests multiple concurrent copy operations.
func TestE2E_ConcurrentCopies(t *testing.T) {
	t.Parallel()

	env := newE2EEnv(t)
	defer os.RemoveAll(env.tempDir)

	// Arrange
	sourceFile := filepath.Join(env.tempDir, "source.txt")
	env.createFile(sourceFile, "concurrent test content")

	// Act: Run multiple copies concurrently
	done := make(chan error, 5)

	for idx := range 5 {
		go func(index int) {
			destFile := filepath.Join(env.tempDir, fmt.Sprintf("concurrent_%d.txt", index))

			_, stderr, exitCode := env.runCmd(sourceFile, destFile)
			if exitCode != 0 {
				done <- fmt.Errorf("copy %d failed: %s", index, stderr) //nolint:err113

				return
			}

			done <- nil
		}(idx)
	}

	// Assert: Wait for all goroutines
	for range 5 {
		if err := <-done; err != nil {
			t.Error(err)
		}
	}

	// Verify all files were created
	for idx := range 5 {
		destFile := filepath.Join(env.tempDir, fmt.Sprintf("concurrent_%d.txt", idx))
		if !env.fileExists(destFile) {
			t.Errorf("concurrent copy %d: destination file not created", idx)
		}
	}
}
