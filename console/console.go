package console

import (
	"fmt"
	"os"
	"strings"
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
