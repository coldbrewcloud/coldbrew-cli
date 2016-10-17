package exec

import (
	"bufio"
	"errors"
	"os/exec"
)

type ExecCallback func(stdout, stderr *string, exitError *exec.ExitError, err error)

func Exec(name string, args ...string) (stdout chan string, stderr chan string, exit chan error, err error) {
	if name == "" {
		return nil, nil, nil, errors.New("name is empty")
	}

	stdout = make(chan string)
	stderr = make(chan string)
	exit = make(chan error)

	cmd := exec.Command(name, args...)

	// redirect std out
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			stdout <- line
		}
		if err := scanner.Err(); err != nil {
			// ignored
		}
	}()

	// redirect std err
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			stderr <- line
		}
		if err := scanner.Err(); err != nil {
			// ignored
		}
	}()

	// start command
	if err := cmd.Start(); err != nil {
		return nil, nil, nil, err
	}

	// wait until it exits
	go func() {
		exit <- cmd.Wait()
	}()

	return
}
