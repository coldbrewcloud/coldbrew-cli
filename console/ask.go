package console

import (
	"bufio"
	"os"
	"strings"
)

func AskConfirm(message string, defaultYes bool) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		if defaultYes {
			stdout("%s %s: ", message, "[YES/no]")
		} else {
			stdout("%s %s: ", message, "[yes/NO]")
		}

		response, err := reader.ReadString('\n')
		if err != nil {
			stderr("Error: %s\n", err.Error())
			return false
		}

		switch strings.ToLower(strings.TrimSpace(response)) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		case "":
			return defaultYes
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
