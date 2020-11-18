package util

import (
	"io"
	"os"
)

func WriteFile(filename string, reader io.Reader) error {
	newFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, reader)
	if err != nil {
		return err
	}
	return nil
}
