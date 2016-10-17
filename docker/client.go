package docker

import "github.com/coldbrewcloud/coldbrew-cli/exec"

type Client struct {
	dockerBin string
}

func NewClient(dockerBin string) *Client {
	return &Client{dockerBin: dockerBin}
}

func (c *Client) DockerBinAvailable() bool {
	_, _, _, err := exec.Exec(c.dockerBin, "version")
	return err == nil
}
