package main

import (
	"fmt"
	"os"
)

func main() {
	// Place your code here.
	if len(os.Args) < 3 {
		fmt.Println("3 args minimun")
		os.Exit(1)
	}
	directory := os.Args[1]
	environments, err := ReadDir(directory)
	if err != nil {
		fmt.Println("error of reading directory: %w", err)
		os.Exit(1)
	}
	cmd := os.Args[2]
	args := os.Args[3:]

	result := RunCmd(append([]string{cmd}, args...), environments)
	os.Exit(result)
}
