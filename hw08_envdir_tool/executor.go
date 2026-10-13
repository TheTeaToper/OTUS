package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	// Place your code here.
	if len(cmd) == 0 {
		return 1
	}

	toolEnvironments := make(map[string]string)
	for _, envString := range os.Environ() {
		envKV := strings.SplitN(envString, "=", 2)
		toolEnvironments[envKV[0]] = envKV[1]
	}

	for envKey, envValue := range env {
		if envValue.NeedRemove {
			delete(toolEnvironments, envKey)
		} else {
			toolEnvironments[envKey] = envValue.Value
		}
	}

	resultEnvironments := make([]string, 0, len(toolEnvironments))
	for key, value := range toolEnvironments {
		resultEnvironments = append(resultEnvironments, key+"="+value)
	}

	command := exec.CommandContext(context.Background(), cmd[0], cmd[1:]...) // #nosec
	command.Env = resultEnvironments
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}
