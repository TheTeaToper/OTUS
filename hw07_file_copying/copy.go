package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error { //nolint:funlen
	sourceFile, err := os.Open(fromPath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("error of opening source file: %w", err)
	}
	defer func() {
		_ = sourceFile.Close()
	}()

	sourceFileInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("error of getting source file info: %w", err)
	}

	if !sourceFileInfo.Mode().IsRegular() {
		return ErrUnsupportedFile
	}
	sourceFileSize := sourceFileInfo.Size()
	if sourceFileSize < offset {
		return ErrOffsetExceedsFileSize
	}

	if _, err := sourceFile.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("error of setting -offset in source file: %w", err)
	}

	availableBytes := sourceFileSize - offset

	var totalBytesToCopy int64
	if limit <= 0 || limit > availableBytes {
		totalBytesToCopy = availableBytes
	} else {
		totalBytesToCopy = limit
	}
	if totalBytesToCopy == 0 {
		fmt.Printf("Copying progress: %d%%", 100)
		return nil
	}

	destinationFile, err := os.Create(toPath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("error of creating destination file: %w", err)
	}
	defer func() {
		_ = destinationFile.Close()
	}()

	limitedSourceReader := io.LimitReader(sourceFile, totalBytesToCopy)

	buffer := make([]byte, BUFFER_SIZE)
	var copiedBytes int64

	fmt.Printf("Copying progress: %d%%", 0)
	for {
		readBytes, readingErr := limitedSourceReader.Read(buffer)
		if readingErr != nil && readingErr != io.EOF {
			return fmt.Errorf("error of reading from source file to buffer: %w", readingErr)
		}
		if readBytes > 0 {
			_, writingErr := destinationFile.Write(buffer[:readBytes])
			if writingErr != nil {
				return fmt.Errorf("errof of writing buffer to destination file: %w", writingErr)
			}
			copiedBytes += int64(readBytes)
			progress := (copiedBytes * 100) / totalBytesToCopy
			fmt.Printf("\rCopying progress: %d%%", progress)
		}

		if readingErr == io.EOF {
			break
		}
	}
	if err := destinationFile.Sync(); err != nil {
		return err
	}
	return nil
}
