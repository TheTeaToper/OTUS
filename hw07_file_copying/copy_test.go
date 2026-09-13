package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// nolint:nolintlint,gosec,funlen,errcheck
func TestCopy(t *testing.T) {
	// Place your code here.
	tmpDir, err := os.MkdirTemp("", "copy_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testDataDir := "testdata"
	if err := os.MkdirAll(testDataDir, 0o755); err != nil {
		t.Fatalf("failed to create testdata dir: %v", err)
	}

	srcPath := filepath.Join(testDataDir, "sample.txt")
	content := []byte("abcdefghijklmnopqrstuvwxyz") // 26 байт
	if err := os.WriteFile(srcPath, content, 0o644); err != nil {
		t.Fatalf("failed to create sample file: %v", err)
	}
	defer os.Remove(srcPath)

	tests := []struct {
		name          string
		offset        int64
		limit         int64
		expectedError error
		expectedBytes []byte
	}{
		{
			name:          "Copy full file",
			offset:        0,
			limit:         0,
			expectedError: nil,
			expectedBytes: content,
		},
		{
			name:          "Copy with offset",
			offset:        10, // Начиная с 'k'
			limit:         0,
			expectedError: nil,
			expectedBytes: []byte("klmnopqrstuvwxyz"),
		},
		{
			name:          "Copy with limit less than size",
			offset:        0,
			limit:         5,
			expectedError: nil,
			expectedBytes: []byte("abcde"),
		},
		{
			name:          "Copy with offset and limit",
			offset:        3, // Начиная с 'd'
			limit:         4, // 'defg'
			expectedError: nil,
			expectedBytes: []byte("defg"),
		},
		{
			name:          "Limit exceeds file size (valid case)",
			offset:        20, // Начиная с 'u'
			limit:         100,
			expectedError: nil,
			expectedBytes: []byte("uvwxyz"),
		},
		{
			name:          "Offset exceeds file size (invalid case)",
			offset:        30,
			limit:         5,
			expectedError: ErrOffsetExceedsFileSize,
			expectedBytes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dstPath := filepath.Join(tmpDir, "output.txt")
			_ = os.Remove(dstPath)

			err := Copy(srcPath, dstPath, tt.offset, tt.limit)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error: %v, got: %v", tt.expectedError, err)
			}

			if tt.expectedError == nil {
				gotBytes, err := os.ReadFile(dstPath)
				if err != nil {
					t.Fatalf("failed to read destination file: %v", err)
				}
				if !bytes.Equal(gotBytes, tt.expectedBytes) {
					t.Errorf("expected bytes: %s, got: %s", string(tt.expectedBytes), string(gotBytes))
				}
			}
		})
	}
}
