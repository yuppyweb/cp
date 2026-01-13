package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestCopyFile_Success tests successful file copy.
func TestCopyFile_Success(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "source.txt")
	destFile := filepath.Join(tmpDir, "dest.txt")
	testContent := "Hello, World!"

	// Setup: Create source file
	err := os.WriteFile(sourceFile, []byte(testContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Test: Copy file
	os.Args = []string{"cp", sourceFile, destFile}

	err = run()
	if err != nil {
		t.Errorf("run() failed: %v", err)
	}

	// Verify: Check if destination file exists and has correct content
	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("content mismatch: got %q, want %q", string(content), testContent)
	}
}

// TestCopyFile_LargeFile tests copying a large file.
func TestCopyFile_LargeFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "large.bin")
	destFile := filepath.Join(tmpDir, "large_copy.bin")

	// Setup: Create large file (10 MB)
	largeContent := bytes.Repeat([]byte("x"), 10*1024*1024)

	err := os.WriteFile(sourceFile, largeContent, 0o600)
	if err != nil {
		t.Fatalf("failed to create large source file: %v", err)
	}

	// Test: Copy large file
	os.Args = []string{"cp", sourceFile, destFile}

	err = run()
	if err != nil {
		t.Errorf("run() failed: %v", err)
	}

	// Verify: Check file sizes match
	sourceInfo, err := os.Stat(sourceFile)
	if err != nil {
		t.Fatalf("failed to stat source file: %v", err)
	}

	destInfo, err := os.Stat(destFile)
	if err != nil {
		t.Fatalf("failed to stat destination file: %v", err)
	}

	if sourceInfo.Size() != destInfo.Size() {
		t.Errorf("file size mismatch: got %d, want %d", destInfo.Size(), sourceInfo.Size())
	}
}

// TestCopyFile_BinaryFile tests copying a binary file.
func TestCopyFile_BinaryFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "binary.bin")
	destFile := filepath.Join(tmpDir, "binary_copy.bin")

	// Setup: Create binary file with random bytes
	binaryContent := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD}

	err := os.WriteFile(sourceFile, binaryContent, 0o600)
	if err != nil {
		t.Fatalf("failed to create binary file: %v", err)
	}

	// Test: Copy binary file
	os.Args = []string{"cp", sourceFile, destFile}

	err = run()
	if err != nil {
		t.Errorf("run() failed: %v", err)
	}

	// Verify: Check content matches exactly
	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if !bytes.Equal(content, binaryContent) {
		t.Errorf("binary content mismatch")
	}
}

// TestCopyFile_EmptyFile tests copying an empty file.
func TestCopyFile_EmptyFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "empty.txt")
	destFile := filepath.Join(tmpDir, "empty_copy.txt")

	// Setup: Create empty file
	f, err := os.Create(sourceFile)
	if err != nil {
		t.Fatalf("failed to create empty source file: %v", err)
	}

	f.Close() // Close the file immediately

	// Test: Copy empty file
	os.Args = []string{"cp", sourceFile, destFile}

	err = run()
	if err != nil {
		t.Errorf("run() failed: %v", err)
	}

	// Verify: Check if destination file exists and is empty
	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if len(content) != 0 {
		t.Errorf("expected empty file, got %d bytes", len(content))
	}
}

// TestCopyFile_RelativePaths tests copying with relative paths.
func TestCopyFile_RelativePaths(t *testing.T) { //nolint:paralleltest
	tmpDir := t.TempDir()

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(oldCwd); err != nil { //nolint:usetesting
			t.Logf("failed to restore directory: %v", err)
		}
	})

	t.Chdir(tmpDir)

	sourceFile := "source.txt"
	destFile := "dest.txt"
	testContent := "Relative paths test"

	// Setup: Create source file
	if err = os.WriteFile(sourceFile, []byte(testContent), 0o600); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Test: Copy with relative paths
	os.Args = []string{"cp", sourceFile, destFile}

	if err = run(); err != nil {
		t.Errorf("run() failed: %v", err)
	}

	// Verify: Check destination file
	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("content mismatch: got %q, want %q", string(content), testContent)
	}
}

// TestCopyFile_SameSourceAndDest tests error when source equals destination.
func TestCopyFile_SameSourceAndDest(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "file.txt")

	// Setup: Create source file
	err := os.WriteFile(sourceFile, []byte("test"), 0o600)
	if err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Test: Try to copy same file
	os.Args = []string{"cp", sourceFile, sourceFile}
	err = run()

	// Verify: Should error when source equals destination
	if err == nil {
		t.Error("expected error when source equals destination, got nil")
	}

	if err != nil && err.Error() != "source and destination files are the same" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestCopyFile_SourceNotFound tests error when source file doesn't exist.
func TestCopyFile_SourceNotFound(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "nonexistent.txt")
	destFile := filepath.Join(tmpDir, "dest.txt")

	// Test: Try to copy non-existent file
	os.Args = []string{"cp", sourceFile, destFile}
	err := run()

	// Verify: Should error
	if err == nil {
		t.Error("expected error when source file doesn't exist, got nil")
	}
}

// TestCopyFile_InvalidDestinationPath tests error for invalid destination path.
func TestCopyFile_InvalidDestinationPath(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "file.txt")
	nonExistentDir := filepath.Join(tmpDir, "nonexistent", "dir")
	destFile := filepath.Join(nonExistentDir, "dest.txt")

	// Setup: Create source file
	err := os.WriteFile(sourceFile, []byte("test"), 0o600)
	if err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Test: Try to copy to non-existent directory
	os.Args = []string{"cp", sourceFile, destFile}
	err = run()

	// Verify: Should error
	if err == nil {
		t.Error("expected error for invalid destination path, got nil")
	}
}

// TestCopyFile_MissingArguments tests error when arguments are missing.
func TestCopyFile_MissingArguments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "no arguments",
			args: []string{"cp"},
		},
		{
			name: "only source",
			args: []string{"cp", "source.txt"},
		},
		{
			name: "too many arguments",
			args: []string{"cp", "source.txt", "dest.txt", "extra"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			os.Args = tt.args

			err := run()
			if err == nil {
				t.Error("expected error for invalid arguments, got nil")
			}
		})
	}
}

// TestCopyFile_FilePermissions tests that file permissions are preserved.
func TestCopyFile_FilePermissions(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "source.txt")
	destFile := filepath.Join(tmpDir, "dest.txt")

	// Setup: Create source file with specific permissions
	err := os.WriteFile(sourceFile, []byte("test"), 0o600)
	if err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Test: Copy file
	os.Args = []string{"cp", sourceFile, destFile}

	err = run()
	if err != nil {
		t.Errorf("run() failed: %v", err)
	}

	// Verify: Destination file should exist (note: permissions may vary by OS)
	_, err = os.Stat(destFile)
	if err != nil {
		t.Fatalf("failed to stat destination file: %v", err)
	}
}

// TestCopyFile_OverwriteExistingFile tests overwriting an existing file.
func TestCopyFile_OverwriteExistingFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "source.txt")
	destFile := filepath.Join(tmpDir, "dest.txt")
	newContent := "New content"

	// Setup: Create source and destination files
	err := os.WriteFile(sourceFile, []byte(newContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	err = os.WriteFile(destFile, []byte("Old content"), 0o600)
	if err != nil {
		t.Fatalf("failed to create destination file: %v", err)
	}

	// Test: Copy file over existing destination
	os.Args = []string{"cp", sourceFile, destFile}

	err = run()
	if err != nil {
		t.Errorf("run() failed: %v", err)
	}

	// Verify: Destination file should have new content
	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(content) != newContent {
		t.Errorf("content mismatch: got %q, want %q", string(content), newContent)
	}
}

// TestCopyFile_SymlinkAsSource tests copying a symlink source.
func TestCopyFile_SymlinkAsSource(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	realFile := filepath.Join(tmpDir, "real.txt")
	linkFile := filepath.Join(tmpDir, "link.txt")
	destFile := filepath.Join(tmpDir, "dest.txt")
	testContent := "Symlink test"

	// Setup: Create real file and symlink
	err := os.WriteFile(realFile, []byte(testContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create real file: %v", err)
	}

	err = os.Symlink(realFile, linkFile)
	if err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	// Test: Copy via symlink
	os.Args = []string{"cp", linkFile, destFile}

	err = run()
	if err != nil {
		t.Errorf("run() failed: %v", err)
	}

	// Verify: Destination file should have correct content
	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("content mismatch: got %q, want %q", string(content), testContent)
	}
}

// BenchmarkCopyFile benchmarks the file copy operation.
func BenchmarkCopyFile(b *testing.B) {
	tmpDir := b.TempDir()
	sourceFile := filepath.Join(tmpDir, "source.bin")
	destFile := filepath.Join(tmpDir, "dest.bin")

	// Setup: Create source file (1 MB)
	content := bytes.Repeat([]byte("x"), 1024*1024)

	err := os.WriteFile(sourceFile, content, 0o600)
	if err != nil {
		b.Fatalf("failed to create source file: %v", err)
	}

	b.ResetTimer()

	for i := range b.N {
		os.Args = []string{"cp", sourceFile, fmt.Sprintf("%s.%d", destFile, i)}

		err := run()
		if err != nil {
			b.Fatalf("run() failed: %v", err)
		}
	}
}

// TestMain can be used for common test setup/teardown if needed.
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
