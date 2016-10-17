package deploy

import (
	"fmt"
	"path/filepath"

	"github.com/coldbrewcloud/coldbrew-cli/docker"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

func (dc *DeployCommand) buildDockerImage(dockerClient *docker.Client, image, tag string) error {
	buildPath, err := filepath.Abs(conv.S(dc.deployFlags.AppPath))
	if err != nil {
		return fmt.Errorf("Failed to find app path [%s].", conv.S(dc.deployFlags.AppPath))
	}

	dockerfilePath, err := filepath.Abs(conv.S(dc.deployFlags.DockerfilePath))
	if err != nil {
		return fmt.Errorf("Failed to find Dockerilfe path [%s]", conv.S(dc.deployFlags.DockerfilePath))
	}

	// docker build
	if err = dockerClient.BuildImage(buildPath, dockerfilePath, image, tag); err != nil {
		return err
	}

	return nil
}

func (dc *DeployCommand) pushDockerImage(dockerClient *docker.Client, image, tag string) error {
	// docker login
	userName, password, proxyURL, err := dc.awsClient.ECR().GetDockerLogin()
	if err != nil {
		return fmt.Errorf("Failed to retrieve docker login info: %s", err.Error())
	}
	if err := dockerClient.Login(userName, password, proxyURL); err != nil {
		return fmt.Errorf("Docker login failed: %s", err.Error())
	}

	// docker push
	if err = dockerClient.PushImage(image, tag); err != nil {
		return fmt.Errorf("Failed to push docker image: %s", err.Error())
	}

	return nil
}
