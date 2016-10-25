package console

import (
	"bufio"
	"os"
	"strings"
)

func AskConfirm(message string, defaultYes bool) bool {
	return AskConfirmWithNote(message, defaultYes, "")
}

func AskConfirmWithNote(message string, defaultYes bool, note string) bool {
	reader := bufio.NewReader(os.Stdin)

	if note != "" {
		stdout("%s\n", ColorFnAskConfirmNote(note))
	}

	for {
		if defaultYes {
			stdout("%s %s [%s/%s]: ",
				ColorFnMarkQuestion(MarkQuestion),
				ColorFnAskConfirmMain(message),
				ColorFnAskConfirmDefaultAnswer("Y"),
				ColorFnAskConfirmAnswer("n"))
		} else {
			stdout("%s %s [%s/%s]: ",
				ColorFnMarkQuestion(MarkQuestion),
				ColorFnAskConfirmMain(message),
				ColorFnAskConfirmAnswer("y"),
				ColorFnAskConfirmDefaultAnswer("N"))
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
	return AskQuestionWithNote(message, defaultValue, "")
}

func AskQuestionWithNote(message, defaultValue, note string) string {
	reader := bufio.NewReader(os.Stdin)

	if note != "" {
		stdout("%s\n", ColorFnAskQuestionNote(note))
	}

	stdout("%s %s [%s]: ",
		ColorFnMarkQuestion(MarkQuestion),
		ColorFnAskQuestionMain(message),
		ColorFnAskQuestionDefaultValue(defaultValue))

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
