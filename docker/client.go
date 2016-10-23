package docker

import (
	"strings"

	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/exec"
	"github.com/d5/cc"
)

type Client struct {
	dockerBin     string
	outputIndent  string
	execColorFn   func(string, ...interface{}) string
	outputColorFn func(string, ...interface{}) string
	errorColorFn  func(string, ...interface{}) string
}

func NewClient(dockerBin string) *Client {
	return &Client{
		dockerBin:     dockerBin,
		outputIndent:  "  ",
		outputColorFn: cc.BlackH,
		errorColorFn:  cc.Red,
		execColorFn:   cc.YellowL,
	}
}

func (c *Client) DockerBinAvailable() bool {
	_, _, _, err := exec.Exec(c.dockerBin, "version")
	return err == nil
}

func (c *Client) PrintVersion() error {
	return c.exec(c.dockerBin, "--version")
}

func (c *Client) BuildImage(buildPath, dockerfilePath, image string) error {
	return c.exec(c.dockerBin, "build", "-t", image, "-f", dockerfilePath, buildPath)
}

func (c *Client) Login(userName, password, proxyURL string) error {
	// NOTE: use slightly different implementation to hide password in output
	//return c.exec(c.dockerBin, "login", "-u", userName, "-p", password, proxyURL)

	console.Blank()
	console.ShellCommand(c.dockerBin + " login -u " + userName + " -p ****** " + proxyURL)

	stdout, stderr, exit, err := exec.Exec(c.dockerBin, "login", "-u", userName, "-p", password, proxyURL)
	if err != nil {
		return err
	}

	for {
		select {
		case line := <-stdout:
			console.ShellOutput(line)
		case line := <-stderr:
			console.ShellError(line)
		case exitErr := <-exit:
			console.Blank()
			return exitErr
		}
	}

	return nil
}

func (c *Client) PushImage(image string) error {
	return c.exec(c.dockerBin, "push", image)
}

func (c *Client) TagImage(src, dest string) error {
	return c.exec(c.dockerBin, "tag", src, dest)
}

func (c *Client) exec(name string, args ...string) error {
	console.Blank()
	console.ShellCommand(name + " " + strings.Join(args, " "))

	stdout, stderr, exit, err := exec.Exec(name, args...)
	if err != nil {
		return err
	}

	for {
		select {
		case line := <-stdout:
			console.ShellOutput(line)
		case line := <-stderr:
			console.ShellError(line)
		case exitErr := <-exit:
			console.Blank()
			return exitErr
		}
	}

	return nil
}
