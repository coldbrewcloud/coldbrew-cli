package console

import (
	"bufio"
	"os"
	"strings"
)

func AskConfirm(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		stdout("%s %s: ", message, "[y/N]")

		response, err := reader.ReadString('\n')
		if err != nil {
			stderr("Error: %s\n", err.Error())
			return false
		}

		switch strings.ToLower(strings.TrimSpace(response)) {
		case "y", "yes":
			return true
		case "", "n", "no":
			return false
		}
	}
}

func AskQuestion(message, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)

	stdout("%s %s: ", message, "["+defaultValue+"]")

	response, err := reader.ReadString('\n')
	if err != nil {
		stderr("Error: %s\n", err.Error())
		return ""
	}

	response = strings.TrimSpace(response)
	if response == "" {
		return defaultValue
	} else {
		return response
	}
}
