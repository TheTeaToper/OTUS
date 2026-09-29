package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	// Place your code here
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	environments := make(Environment)
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}
		fileName := dirEntry.Name()
		if strings.Contains(fileName, "=") {
			continue
		}
		fileContent, err := os.ReadFile(filepath.Join(dir, fileName)) // #nosec
		if err != nil {
			return nil, err
		}
		if len(fileContent) == 0 {
			environments[fileName] = EnvValue{NeedRemove: true}
			continue
		}

		firstString := fileContent
		if firstNewLineIndex := bytes.IndexByte(fileContent, '\n'); firstNewLineIndex > -1 {
			firstString = fileContent[:firstNewLineIndex]
		}

		firstString = bytes.ReplaceAll(firstString, []byte{0x00}, []byte{'\n'})

		resultString := strings.TrimRight(string(firstString), " \t")

		environments[fileName] = EnvValue{
			Value:      resultString,
			NeedRemove: false,
		}
	}
	return environments, nil
}
