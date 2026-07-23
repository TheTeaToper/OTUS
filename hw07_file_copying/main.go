package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	from, to      string
	limit, offset int64
)

const BUFFER_SIZE = 1024

func init() {
	flag.StringVar(&from, "from", "", "file to read from")
	flag.StringVar(&to, "to", "", "file to write to")
	flag.Int64Var(&limit, "limit", 0, "limit of bytes to copy")
	flag.Int64Var(&offset, "offset", 0, "offset in input file")
}

func main() {
	flag.Parse()
	// Place your code here.
	if from == "" {
		_ = fmt.Errorf("-from arg is empty")
		os.Exit(1)
	}
	if to == "" {
		_ = fmt.Errorf("-to arg is empty")
		os.Exit(1)
	}
	if err := Copy(from, to, offset, limit); err != nil {
		_ = fmt.Errorf("file copying error: %w", err)
		os.Exit(1)
	}
	fmt.Printf("File %s successfully copied to %s", from, to)
}
