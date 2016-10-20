package deploy

import (
	"fmt"

	"io/ioutil"
	"strings"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/config"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/docker"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"github.com/d5/cc"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Command struct {
	kingpinApp   *kingpin.Application
	globalFlags  *flags.GlobalFlags
	commandFlags *Flags
	awsClient    *aws.Client
	dockerClient *docker.Client
}

func (c *Command) Init(ka *kingpin.Application, globalFlags *flags.GlobalFlags) *kingpin.CmdClause {
	c.kingpinApp = ka
	c.globalFlags = globalFlags

	cmd := ka.Command("deploy", "(deploy description goes here)")
	c.commandFlags = NewFlags(cmd)

	return cmd
}

func (c *Command) Run() error {
	// app configuration
	conf, err := c.getConfiguration()
	if err != nil {
		return c.exitWithError(err)
	}
	if conf == nil {
		return c.exitWithError("Configuration file was not found. You can use \"init\" command to create configuration file.")
	}

	// CLI flags
	if err := c.validateFlags(); err != nil {
		return c.exitWithError(err)
	}

	// AWS client
	c.awsClient = c.globalFlags.GetAWSClient()

	// test if target cluster is available to use
	clusterName := conf.ClusterName
	if err := c.isClusterAvailable(clusterName); err != nil {
		return c.exitWithErrorInfo(err, "https://github.com/coldbrewcloud/coldbrew-cli/wiki/Error:-Cluster-not-found")
	}

	// docker client
	console.Println("Docker Executable", cc.Green(conf.Docker.Bin))
	c.dockerClient = docker.NewClient(conf.Docker.Bin)
	if !c.dockerClient.DockerBinAvailable() {
		return c.exitWithErrorInfo(
			fmt.Errorf("Failed to find Docker binary [%s].", conf.Docker.Bin),
			"https://github.com/coldbrewcloud/coldbrew-cli/wiki/Error:-Docker-binary-not-found")
	}

	// prepare ECR repo (create one if needed)
	ecrRepoName := fmt.Sprintf("%s/%s", conf.AWS.ECRNamespace, conf.Name)
	ecrRepoURI, err := c.prepareECRRepo(ecrRepoName)
	if err != nil {
		return c.exitWithError(err)
	}
	console.Println("ECR Repository", cc.Green(ecrRepoURI))

	// prepare docker image (build one if needed)
	dockerImage := conv.S(c.commandFlags.DockerImage)
	imageTag := conv.S(c.commandFlags.AppVersion)
	if utils.IsBlank(dockerImage) {
		if err := c.buildDockerImage(ecrRepoURI, imageTag); err != nil {
			return err
		}
	}

	// push docker image to ECR
	if err := c.pushDockerImage(ecrRepoURI, imageTag); err != nil {
		return fmt.Errorf("failed to push docker image: %s", err.Error())
	}

	// prepare ECS cluster (create one if needed)
	if err := c.prepareECSCluster(conv.S(c.commandFlags.ECSClusterName)); err != nil {
		return err
	}

	// create/update ECS task definition
	ecsTaskDefinitionARN, err := c.updateECSTaskDefinition(fmt.Sprintf("%s:%s", ecrRepoURI, imageTag))
	if err != nil {
		return err
	}

	// create/update ECS service
	if err := c.createOrUpdateECSService(ecsTaskDefinitionARN); err != nil {
		return err
	}

	return nil
}

func (c *Command) exitWithError(err error) error {
	console.Errorln(cc.Red("Error:"), err.Error())
	return nil
}

func (c *Command) exitWithErrorInfo(err error, infoURL string) error {
	console.Errorln(cc.Red("Error:"), err.Error())
	console.Errorln(cc.BlackH("More Info:"), infoURL)
	return nil
}

func (c *Command) getConfiguration() (*config.Config, error) {
	// configuration file path
	configFilePath, err := c.globalFlags.GetConfigFile()
	if err != nil {
		return nil, err
	}
	if !utils.FileExists(configFilePath) {
		return nil, nil
	}

	// read configuration file
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read configuration file [%s]: %s\n", configFilePath, err.Error())
	}

	conf := &config.Config{}
	configFileFormat := strings.ToLower(conv.S(c.globalFlags.ConfigFileFormat))
	switch {
	case flags.GlobalFlagsConfigFileFormatYAML:
		if err := conf.FromYAML(data); err != nil {
			return nil, fmt.Errorf("Failed to read configuration file in YAML: %s\n", err.Error())
		}
	case flags.GlobalFlagsConfigFileFormatJSON:
		if err := conf.FromJSON(data); err != nil {
			return nil, fmt.Errorf("Failed to read configuration file in JSON: %s\n", err.Error())
		}
	default:
		return nil, fmt.Errorf("Unsupported configuration file format [%s]", configFileFormat)
	}

	// validation
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	return conf, nil
}
