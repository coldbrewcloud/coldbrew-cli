package console

import (
	"bufio"
	"os"
	"strings"

	"github.com/d5/cc"
)

func AskConfirm(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		stdout("%s %s: ", cc.Blue(message), "[y/N]")

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

	stdout("%s %s: ", cc.Blue(message), "["+defaultValue+"]")

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
