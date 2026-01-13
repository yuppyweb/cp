package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const requiredNumberArgs = 3

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != requiredNumberArgs {
		return fmt.Errorf("usage: %s <source file> <destination file>", os.Args[0]) //nolint:err113
	}

	source := os.Args[1]
	dest := os.Args[2]

	sourceAbs, err := filepath.Abs(source)
	if err != nil {
		return fmt.Errorf("getting absolute path of source file: %w", err)
	}

	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("getting absolute path of destination file: %w", err)
	}

	if sourceAbs == destAbs {
		return errors.New("source and destination files are the same") //nolint:err113
	}

	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}

	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}

	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("copying file: %w", err)
	}

	fmt.Printf("File copied from %s to %s successfully.\n", source, dest)

	return nil
}
