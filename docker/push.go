package docker

import (
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/exec"
	"github.com/d5/cc"
)

func (c *Client) PushImage(image string) error {
	console.Println(c.dockerBin, "push", image)
	stdout, stderr, exit, err := exec.Exec(c.dockerBin, "push", image)
	if err != nil {
		return err
	}

	for {
		select {
		case line := <-stdout:
			console.Println(cc.BlackH(line))
		case line := <-stderr:
			console.Errorln(cc.RedL(line))
		case exitErr := <-exit:
			return exitErr
		}
	}
}
