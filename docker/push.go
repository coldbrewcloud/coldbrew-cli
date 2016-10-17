package docker

import (
	"fmt"

	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/exec"
	"github.com/d5/cc"
)

func (c *Client) PushImage(image, tag string) error {
	console.Println(c.dockerBin, "push", fmt.Sprintf("%s:%s", image, tag))
	stdout, stderr, exit, err := exec.Exec(c.dockerBin, "push", fmt.Sprintf("%s:%s", image, tag))
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
