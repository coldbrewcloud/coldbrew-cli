package deploy

import (
	"fmt"
	"path/filepath"

	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

func (c *Command) buildDockerImage(image, tag string) error {
	buildPath, err := filepath.Abs(conv.S(c.commandFlags.AppPath))
	if err != nil {
		return fmt.Errorf("Failed to find app path [%s].", conv.S(c.commandFlags.AppPath))
	}

	dockerfilePath, err := filepath.Abs(conv.S(c.commandFlags.DockerfilePath))
	if err != nil {
		return fmt.Errorf("Failed to find Dockerilfe path [%s]", conv.S(c.commandFlags.DockerfilePath))
	}

	// docker build
	if err = c.dockerClient.BuildImage(buildPath, dockerfilePath, image, tag); err != nil {
		return err
	}

	return nil
}

func (c *Command) pushDockerImage(image, tag string) error {
	// docker login
	userName, password, proxyURL, err := c.awsClient.ECR().GetDockerLogin()
	if err != nil {
		return fmt.Errorf("Failed to retrieve docker login info: %s", err.Error())
	}
	if err := c.dockerClient.Login(userName, password, proxyURL); err != nil {
		return fmt.Errorf("Docker login failed: %s", err.Error())
	}

	// docker push
	if err = c.dockerClient.PushImage(image, tag); err != nil {
		return fmt.Errorf("Failed to push docker image: %s", err.Error())
	}

	return nil
}
