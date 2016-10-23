package console

import (
	"fmt"
	"os"

	"github.com/coldbrewcloud/coldbrew-cli/core"
)

func ExitWithErrorString(format string, a ...interface{}) error {
	return ExitWithError(fmt.Errorf(format, a))
}

func ExitWithError(err error) error {
	if ei, ok := err.(*core.Error); ok {
		errorfFn("%s %s %s\n",
			ColorFnErrorHeader("Error:"),
			ColorFnErrorMessage(ei.Error()),
			ColorFnSideNote("(more info: "+ei.ExtraInfo()+")"))
	} else {
		errorfFn("%s %s\n",
			ColorFnErrorHeader("Error:"),
			ColorFnErrorMessage(err.Error()))
	}
	os.Exit(100)
	return nil
}

func Error(message string) {
	errorfFn("%s %s\n",
		ColorFnErrorHeader("Error:"),
		ColorFnErrorMessage(message))
}
