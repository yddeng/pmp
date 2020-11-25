package util

import (
	"io"
	"os"
)

func WriteFile(filename string, reader io.Reader) (written int64, err error) {
	newFile, err := os.Create(filename)
	if err != nil {
		return 0, err
	}
	defer newFile.Close()

	return io.Copy(newFile, reader)
}

func CopyFile(src, dest string) (written int64, err error) {
	srcF, err := os.Open(src)
	if err != nil {
		return 0, err
	}

	return WriteFile(dest, srcF)
}
