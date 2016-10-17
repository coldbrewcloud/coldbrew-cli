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

func Print(tokens ...string) (int, error) {
	return printfFn(strings.Join(tokens, " "))
}

func Println(tokens ...string) (int, error) {
	return printfFn(strings.Join(tokens, " ") + "\n")
}

func Printf(format string, a ...interface{}) (int, error) {
	return printfFn(format, a...)
}

func Error(tokens ...string) (int, error) {
	return errorfFn(strings.Join(tokens, " "))
}

func Errorln(tokens ...string) (int, error) {
	return errorfFn(strings.Join(tokens, " ") + "\n")
}

func Errorf(format string, a ...interface{}) (int, error) {
	return errorfFn(format, a...)
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
