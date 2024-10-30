package util

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
)

func prepareCommand(command string, arg ...string) (string, []string) {
	// I'm sure this function isn't perfect, but it should simplify things for me a bit
	// so its what i am going to go with
	commandParts := strings.Split(command, " ")

	if len(commandParts) <= 0 {
		panic("No command was passed")
	}

	commandName := commandParts[0]
	var commandArgs []string

	for i := 1; i < len(commandParts); i++ {
		part := commandParts[i]

		if strings.HasPrefix(part, "%s") {
			// we found a placeholder
			wants_index, err := strconv.Atoi(strings.TrimPrefix(part, "%s"))
			if err != nil {
				panic("Error parsing index placeholder")
			}

			if len(arg) > wants_index {
				commandArgs = append(commandArgs, arg[wants_index])
			} else {
				panic("Invalid placeholder index provided")
			}
		} else {
			// no placeholder
			commandArgs = append(commandArgs, part)
		}
	}

	return commandName, commandArgs
}

func callExternalCommand(command string, arg ...string) (string, int, error) {
	cmdName, cmdArgs := prepareCommand(command, arg...)
	slog.Debug("Calling external command", slog.String("command", cmdName), slog.Any("args", cmdArgs))

	cmd := exec.Command(cmdName, cmdArgs...)
	output, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			slog.Debug(fmt.Sprintf("Exit Status: %d", exiterr.ExitCode()))
			slog.Debug(fmt.Sprintf("stderr output: %s", string(exiterr.Stderr)))
			return string(exiterr.Stderr), exiterr.ExitCode(), err
		} else {
			slog.Warn(fmt.Sprintf("Error occurred while calling '%s' command: %s", cmdName, err.Error()))
			return "", -666, err
		}
	}

	return string(output), 0, nil
}
