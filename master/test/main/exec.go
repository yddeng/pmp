package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		panic("args failed")
	}
	file := os.Args[1]

	cmd := exec.Command("sh", file)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(out))
}
