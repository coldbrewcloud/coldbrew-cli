package deploy

import (
	"fmt"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/docker"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"gopkg.in/alecthomas/kingpin.v2"
)

type DeployCommand struct {
	appFlags    *flags.AppFlags
	deployFlags *flags.DeployFlags
	awsClient   *aws.Client
}

func (dc *DeployCommand) Init(ka *kingpin.Application, appFlags *flags.AppFlags) *kingpin.CmdClause {
	dc.appFlags = appFlags

	cmd := ka.Command("deploy", "Deploy app")
	dc.deployFlags = flags.NewDeployFlags(cmd)

	dc.awsClient = aws.NewClient(*dc.appFlags.AWSRegion).WithCredentials(*dc.appFlags.AWSAccessKey, *dc.appFlags.AWSSecretKey)

	return cmd
}

func (dc *DeployCommand) Run() error {
	if err := dc.validateFlags(); err != nil {
		return err
	}

	// docker client
	console.Debugln("docker path:", conv.S(dc.deployFlags.DockerBinPath))
	dockerClient := docker.NewClient(conv.S(dc.deployFlags.DockerBinPath))
	if !dockerClient.DockerBinAvailable() {
		return fmt.Errorf("docker path [%s] not found.", conv.S(dc.deployFlags.DockerBinPath))
	}

	// prepare ECR repo (create one if needed)
	ecrRepoName := fmt.Sprintf("%s/%s", conv.S(dc.deployFlags.ECRNamespace), conv.S(dc.deployFlags.AppName))
	ecrRepoURI, err := dc.prepareECRRepo(ecrRepoName)
	if err != nil {
		return err
	}

	// prepare docker image (build one if needed)
	dockerImage := conv.S(dc.deployFlags.DockerImage)
	imageTag := conv.S(dc.deployFlags.AppVersion)
	if utils.IsBlank(dockerImage) {
		if err := dc.buildDockerImage(dockerClient, ecrRepoURI, imageTag); err != nil {
			return err
		}
	}

	// push docker image to ECR
	if err := dc.pushDockerImage(dockerClient, ecrRepoURI, imageTag); err != nil {
		return fmt.Errorf("failed to push docker image: %s", err.Error())
	}

	// prepare ECS cluster (create one if needed)
	if err := dc.prepareECSCluster(conv.S(dc.deployFlags.ECSClusterName)); err != nil {
		return err
	}

	// create/update ECS task definition
	ecsTaskDefinitionARN, err := dc.updateECSTaskDefinition(fmt.Sprintf("%s:%s", ecrRepoURI, imageTag))
	if err != nil {
		return err
	}

	// create/update ECS service
	if err := dc.createOrUpdateECSService(ecsTaskDefinitionARN); err != nil {
		return err
	}

	return nil
}
