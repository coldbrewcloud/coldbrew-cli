package deploy

import (
	"fmt"
	"path/filepath"

	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"github.com/d5/cc"
)

func (c *Command) buildDockerImage(image string) error {
	buildPath, err := c.globalFlags.GetApplicationDirectory()
	if err != nil {
		return err
	}

	dockerfilePath := conv.S(c._commandFlags.DockerfilePath)
	if utils.IsBlank(dockerfilePath) {
		dockerfilePath = "Dockerfile"
	}

	if !filepath.IsAbs(dockerfilePath) {
		var err error
		dockerfilePath, err = filepath.Abs(dockerfilePath)
		if err != nil {
			return fmt.Errorf("Error retrieving absolute path [%s]: %s", dockerfilePath, err.Error())
		}
	}

	// docker build
	if err = c.dockerClient.BuildImage(buildPath, dockerfilePath, image); err != nil {
		return err
	}

	return nil
}

func (c *Command) pushDockerImage(image string) error {
	console.Printf("Authenticating to push to ECR Repository...\n")
	// docker login
	userName, password, proxyURL, err := c.awsClient.ECR().GetDockerLogin()
	if err != nil {
		return fmt.Errorf("Failed to retrieve docker login info: %s", err.Error())
	}
	if err := c.dockerClient.Login(userName, password, proxyURL); err != nil {
		return fmt.Errorf("Docker login [%s] failed: %s", userName, err.Error())
	}

	// docker push
	console.Printf("Pushing Docker image [%s]...\n", cc.Green(image))
	if err = c.dockerClient.PushImage(image); err != nil {
		return fmt.Errorf("Failed to push Docker image [%s]: %s", image, err.Error())
	}

	return nil
}
