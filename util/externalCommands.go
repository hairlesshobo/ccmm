// =================================================================================
//
//		ccmm - https://www.foxhollow.cc/projects/ccmm/
//
//	 go-import-media, aka gim, is a tool for automatically importing media
//	 from removable disks into a predefined folder structure automatically.
//
//		Copyright (c) 2024 Steve Cross <flip@foxhollow.cc>
//
//		Licensed under the Apache License, Version 2.0 (the "License");
//		you may not use this file except in compliance with the License.
//		You may obtain a copy of the License at
//
//		     http://www.apache.org/licenses/LICENSE-2.0
//
//		Unless required by applicable law or agreed to in writing, software
//		distributed under the License is distributed on an "AS IS" BASIS,
//		WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//		See the License for the specific language governing permissions and
//		limitations under the License.
//
// =================================================================================
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
