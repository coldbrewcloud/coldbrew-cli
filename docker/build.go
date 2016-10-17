package docker

import (
	"fmt"

	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/exec"
	"github.com/d5/cc"
)

func (c *Client) BuildImage(buildPath, dockerfilePath, image, tag string) error {
	console.Println(c.dockerBin, "build", "-t", fmt.Sprintf("%s:%s", image, tag), "-f", dockerfilePath, buildPath)
	stdout, stderr, exit, err := exec.Exec(
		c.dockerBin,
		"build",
		"-t", fmt.Sprintf("%s:%s", image, tag),
		"-f", dockerfilePath,
		buildPath)
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
