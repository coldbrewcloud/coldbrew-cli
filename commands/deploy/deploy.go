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
	"github.com/d5/cc"
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
		return core.ExitWithError(err)
	}
	configData, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return core.ExitWithErrorString("Failed to read configuration file [%s]: %s", configFilePath, err.Error())
	}
	c.conf, err = config.Load(configData, conv.S(c.globalFlags.ConfigFileFormat), core.DefaultAppName(configFilePath))
	if err != nil {
		return core.ExitWithError(err)
	}

	// CLI flags validation
	if err := c.validateFlags(c._commandFlags); err != nil {
		return core.ExitWithError(core.NewErrorExtraInfo(err, "https://github.com/coldbrewcloud/coldbrew-cli/wiki/Command:-deploy"))
	}

	// merge flags into main configuration
	c.conf = c.mergeFlagsIntoConfiguration(c.conf, c._commandFlags)

	// AWS client
	c.awsClient = c.globalFlags.GetAWSClient()

	// test if target cluster is available to use
	console.Printf("Checking if cluster [%s] is available...\n", cc.Green(conv.S(c.conf.ClusterName)))
	if err := c.isClusterAvailable(conv.S(c.conf.ClusterName)); err != nil {
		return core.ExitWithError(core.NewErrorExtraInfo(err, "https://github.com/coldbrewcloud/coldbrew-cli/wiki/Error:-Cluster-not-found"))
	}

	// docker client
	c.dockerClient = docker.NewClient(conv.S(c.conf.Docker.Bin))
	if !c.dockerClient.DockerBinAvailable() {
		return core.ExitWithError(core.NewErrorExtraInfo(
			fmt.Errorf("Failed to find Docker binary [%s].", c.conf.Docker.Bin),
			"https://github.com/coldbrewcloud/coldbrew-cli/wiki/Error:-Docker-binary-not-found"))
	}

	// prepare ECR repo (create one if needed)
	console.Println("Setting up ECR Repository...")
	ecrRepoURI, err := c.prepareECRRepo(conv.S(c.conf.AWS.ECRRepositoryName))
	if err != nil {
		return core.ExitWithError(err)
	}
	console.Println("ECR Repository", cc.Green(ecrRepoURI))

	// prepare docker image (build one if needed)
	dockerImage := conv.S(c._commandFlags.DockerImage)
	if utils.IsBlank(dockerImage) { // build local docker image
		dockerImage = fmt.Sprintf("%s:localbuild", ecrRepoURI)
		console.Printf("Start building Docker image [%s]...\n", dockerImage)
		if err := c.buildDockerImage(dockerImage); err != nil {
			return core.ExitWithError(err)
		}
	} else { // use local docker image
		// if needed, re-tag local image so it can be pushed to target ECR repo
		m := core.DockerImageURIRE.FindAllStringSubmatch(dockerImage, -1)
		if len(m) != 0 {
			return core.ExitWithErrorString("Invalid Docker image [%s]", dockerImage)
		}
		if m[0][1] != ecrRepoURI {
			newImage := fmt.Sprintf("%s:%s", ecrRepoURI, m[0][1])

			console.Printf("Tagging Docker image [%s] to [%s]...\n", dockerImage, newImage)
			if err := c.dockerClient.TagImage(dockerImage, newImage); err != nil {
				return core.ExitWithError(err)
			}

			dockerImage = newImage
		}
	}

	// push docker image to ECR
	if err := c.pushDockerImage(dockerImage); err != nil {
		return core.ExitWithError(err)
	}

	// create/update ECS task definition
	ecsTaskDefinitionARN, err := c.updateECSTaskDefinition(dockerImage)
	if err != nil {
		return core.ExitWithError(err)
	}

	// create/update ECS service
	if err := c.createOrUpdateECSService(ecsTaskDefinitionARN); err != nil {
		return core.ExitWithError(err)
	}

	return nil
}

func (c *Command) mergeFlagsIntoConfiguration(conf *config.Config, flags *Flags) *config.Config {
	if conv.I64(flags.Units) >= 0 {
		conf.Units = conv.U16P(uint16(conv.I64(flags.Units)))
	}

	if conv.F64(flags.CPU) >= 0 {
		conf.CPU = conv.F64P(conv.F64(flags.CPU))
	}

	if !utils.IsBlank(conv.S(flags.Memory)) {
		conf.Memory = conv.SP(conv.S(flags.Memory))
	}

	// envs
	for ek, ev := range *flags.Envs {
		conf.Env[ek] = ev
	}

	return conf
}
