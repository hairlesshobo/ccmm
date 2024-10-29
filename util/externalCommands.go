package util

import (
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
