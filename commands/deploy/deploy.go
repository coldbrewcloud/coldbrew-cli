package deploy

import (
	"fmt"
	"io/ioutil"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/config"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/docker"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Command struct {
	kingpinApp    *kingpin.Application
	globalFlags   *flags.GlobalFlags
	_commandFlags *Flags // NOTE: this name intentionally starts with underscore because main configuration (conf) should be used throughout Run() after merging them
	awsClient     *aws.Client
	dockerClient  *docker.Client
	conf          *config.Config
}

func (c *Command) Init(ka *kingpin.Application, globalFlags *flags.GlobalFlags) *kingpin.CmdClause {
	c.kingpinApp = ka
	c.globalFlags = globalFlags

	cmd := ka.Command("deploy", "(deploy description goes here)")
	c._commandFlags = NewFlags(cmd)

	return cmd
}

func (c *Command) Run() error {
	var err error

	// app configuration
	configFilePath, err := c.globalFlags.GetConfigFile()
	if err != nil {
		return console.ExitWithError(err)
	}
	configData, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return console.ExitWithErrorString("Failed to read configuration file [%s]: %s", configFilePath, err.Error())
	}
	c.conf, err = config.Load(configData, conv.S(c.globalFlags.ConfigFileFormat), core.DefaultAppName(configFilePath))
	if err != nil {
		return console.ExitWithError(err)
	}

	// CLI flags validation
	if err := c.validateFlags(c._commandFlags); err != nil {
		return console.ExitWithError(core.NewErrorExtraInfo(err, "https://github.com/coldbrewcloud/coldbrew-cli/wiki/Command:-deploy"))
	}

	// merge flags into main configuration
	c.conf = c.mergeFlagsIntoConfiguration(c.conf, c._commandFlags)

	// AWS client
	c.awsClient = c.globalFlags.GetAWSClient()

	// test if target cluster is available to use
	console.ProcessingOnResource("Checking cluster availability", conv.S(c.conf.ClusterName), false)
	if err := c.isClusterAvailable(conv.S(c.conf.ClusterName)); err != nil {
		return console.ExitWithError(core.NewErrorExtraInfo(err, "https://github.com/coldbrewcloud/coldbrew-cli/wiki/Error:-Cluster-not-found"))
	}

	// docker client
	c.dockerClient = docker.NewClient(conv.S(c.conf.Docker.Bin))
	if !c.dockerClient.DockerBinAvailable() {
		return console.ExitWithError(core.NewErrorExtraInfo(
			fmt.Errorf("Failed to find Docker binary [%s].", c.conf.Docker.Bin),
			"https://github.com/coldbrewcloud/coldbrew-cli/wiki/Error:-Docker-binary-not-found"))
	}

	// prepare ECR repo (create one if needed)
	ecrRepoURI, err := c.prepareECRRepo(conv.S(c.conf.AWS.ECRRepositoryName))
	if err != nil {
		return console.ExitWithError(err)
	}

	// prepare docker image (build one if needed)
	dockerImage := conv.S(c._commandFlags.DockerImage)
	if utils.IsBlank(dockerImage) { // build local docker image
		dockerImage = fmt.Sprintf("%s:latest", ecrRepoURI)
		console.ProcessingOnResource("Building Docker image", dockerImage, true)
		if err := c.buildDockerImage(dockerImage); err != nil {
			return console.ExitWithError(err)
		}
	} else { // use local docker image
		// if needed, re-tag local image so it can be pushed to target ECR repo
		m := core.DockerImageURIRE.FindAllStringSubmatch(dockerImage, -1)
		if len(m) != 1 {
			return console.ExitWithErrorString("Invalid Docker image [%s]", dockerImage)
		}
		if m[0][1] != ecrRepoURI {
			tag := m[0][2]
			if tag == "" {
				tag = "latest"
			}
			newImage := fmt.Sprintf("%s:%s", ecrRepoURI, tag)

			console.AddingResource("Tagging Docker image", fmt.Sprintf("%s -> %s", dockerImage, newImage), false)
			if err := c.dockerClient.TagImage(dockerImage, newImage); err != nil {
				return console.ExitWithError(err)
			}

			dockerImage = newImage
		}
	}

	// push docker image to ECR
	if err := c.pushDockerImage(dockerImage); err != nil {
		return console.ExitWithError(err)
	}

	// create/update ECS task definition
	ecsTaskDefinitionARN, err := c.updateECSTaskDefinition(dockerImage)
	if err != nil {
		return console.ExitWithError(err)
	}

	// create/update ECS service
	if err := c.createOrUpdateECSService(ecsTaskDefinitionARN); err != nil {
		return console.ExitWithError(err)
	}

	console.Blank()
	console.Info("Application deployment completed.")

	return nil
}

func (c *Command) isClusterAvailable(clusterName string) error {
	// check ECS cluster
	ecsClusterName := core.DefaultECSClusterName(clusterName)
	ecsCluster, err := c.awsClient.ECS().RetrieveCluster(ecsClusterName)
	if err != nil {
		return fmt.Errorf("Failed to retrieve ECS Cluster [%s]: %s", ecsClusterName, err.Error())
	}
	if ecsCluster == nil || conv.S(ecsCluster.Status) == "INACTIVE" {
		return fmt.Errorf("ECS Cluster [%s] not found", ecsClusterName)
	}

	// check ECS service role
	ecsServiceRoleName := core.DefaultECSServiceRoleName(clusterName)
	ecsServiceRole, err := c.awsClient.IAM().RetrieveRole(ecsServiceRoleName)
	if err != nil {
		return fmt.Errorf("Failed to retrieve IAM Role [%s]: %s", ecsServiceRoleName, err.Error())
	}
	if ecsServiceRole == nil {
		return fmt.Errorf("IAM Role [%s] not found", ecsServiceRoleName)
	}

	return nil
}

func (c *Command) prepareECRRepo(repoName string) (string, error) {
	ecrRepo, err := c.awsClient.ECR().RetrieveRepository(repoName)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve ECR repository [%s]: %s", repoName, err.Error())
	}

	if ecrRepo == nil {
		console.AddingResource("Creating ECR Repository", repoName, false)
		ecrRepo, err = c.awsClient.ECR().CreateRepository(repoName)
		if err != nil {
			return "", fmt.Errorf("Failed to create ECR repository [%s]: %s", repoName, err.Error())
		}
	}

	return *ecrRepo.RepositoryUri, nil
}
