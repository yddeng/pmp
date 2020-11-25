package main

import (
	"fmt"
	"os"
)

func main() {
	_, err := os.Open("file.go")
	fmt.Println(err)
	_, err = os.Open("file.go")
	fmt.Println(err)
	_, err = os.Open("file.go")
	fmt.Println(err)

}
