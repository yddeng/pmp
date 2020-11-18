package main

import (
	"fmt"
	"os"
)

func main() {
	err := os.RemoveAll("ff/tt.go.part*")
	fmt.Println(err)
}
