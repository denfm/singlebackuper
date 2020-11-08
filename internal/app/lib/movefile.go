package lib

import (
	"fmt"
	"io"
	"os"
)

/*
https://gist.github.com/var23rav/23ae5d0d4d830aff886c3c970b8f6c6b
*/

func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)

	if err != nil {
		_ = inputFile.Close()
		return fmt.Errorf("couldn't open dest file: %s", err)
	}

	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()

	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}

	err = os.Remove(sourcePath)

	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}

	return nil
}
