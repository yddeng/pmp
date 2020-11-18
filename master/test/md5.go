package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func main() {
	file := os.Args[1]

	h := md5.New()
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(h, f)
	if err != nil {
		panic(err)
	}

	fmt.Println(hex.EncodeToString(h.Sum(nil)))

}
