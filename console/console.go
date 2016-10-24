package console

import (
	"fmt"
	"os"
	"strings"

	"github.com/coldbrewcloud/coldbrew-cli/core"
)

func stdout(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(os.Stdout, format, a...)
}

func stderr(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(os.Stderr, format, a...)
}

func noop(string, ...interface{}) (int, error) {
	return 0, nil
}

var (
	debugfFn       = noop
	debugLogPrefix = ""

	printfFn = stdout
	errorfFn = stderr
)

func EnablePrintf(enable bool) {
	if enable {
		printfFn = stdout
	} else {
		printfFn = noop
	}
}

func EnableErrorf(enable bool) {
	if enable {
		errorfFn = stderr
	} else {
		errorfFn = noop
	}
}

func EnableDebugf(enable bool, prefix string) {
	if enable {
		debugfFn = stdout
		debugLogPrefix = prefix
	} else {
		debugfFn = noop
		debugLogPrefix = ""
	}
}

func Debug(tokens ...string) (int, error) {
	return debugfFn(debugLogPrefix + strings.Join(tokens, " "))
}

func Debugln(tokens ...string) (int, error) {
	return debugfFn(debugLogPrefix + strings.Join(tokens, " ") + "\n")
}

func Debugf(format string, a ...interface{}) (int, error) {
	if debugLogPrefix != "" {
		return debugfFn(debugLogPrefix+format, a...)
	} else {
		return debugfFn(format, a...)
	}

}

func ExitWithErrorString(format string, a ...interface{}) error {
	return ExitWithError(fmt.Errorf(format, a...))
}

func ExitWithError(err error) error {
	errorfFn("\n")
	if ei, ok := err.(*core.Error); ok {
		errorfFn("%s %s\n       %s\n",
			ColorFnErrorHeader("Error:"),
			ColorFnErrorMessage(ei.Error()),
			ColorFnSideNote("(See: "+ei.ExtraInfo()+")"))
	} else {
		errorfFn("%s %s\n",
			ColorFnErrorHeader("Error:"),
			ColorFnErrorMessage(err.Error()))
	}
	errorfFn("\n")

	os.Exit(100)
	return nil
}

func Error(message string) {
	errorfFn("%s %s\n",
		ColorFnErrorHeader("Error:"),
		ColorFnErrorMessage(message))
}
