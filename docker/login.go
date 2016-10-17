package docker

import (
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/exec"
	"github.com/d5/cc"
)

func (c *Client) Login(userName, password, proxyURL string) error {
	console.Println(c.dockerBin, "login", "-u", userName, "-p", password, proxyURL)
	stdout, stderr, exit, err := exec.Exec(c.dockerBin, "login", "-u", userName, "-p", password, proxyURL)
	if err != nil {
		return err
	}

	for {
		select {
		case line := <-stdout:
			console.Println(cc.BlackH(line))
		case line := <-stderr:
			console.Println(cc.RedL(line))
		case exitErr := <-exit:
			return exitErr
		}
	}
}
