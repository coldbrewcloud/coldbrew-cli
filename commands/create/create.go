package create

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/config"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"github.com/d5/cc"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Command struct {
	globalFlags  *flags.GlobalFlags
	commandFlags *Flags
	awsClient    *aws.Client
}

func (c *Command) Init(ka *kingpin.Application, globalFlags *flags.GlobalFlags) *kingpin.CmdClause {
	c.globalFlags = globalFlags

	cmd := ka.Command("init", "(init description goes here)").Alias("create")
	c.commandFlags = NewFlags(cmd)

	return cmd
}

func (c *Command) Run() error {
	var err error

	appDirectory, err := c.globalFlags.GetApplicationDirectory()
	if err != nil {
		return err
	}

	// AWS client
	c.awsClient = c.globalFlags.GetAWSClient()

	// default config
	defConf := config.DefaultConfig(appDirectory)

	conf := &config.Config{}
	conf.AWS = new(config.ConfigAWS)
	conf.Docker = new(config.ConfigDocker)

	// app name
	conf.Name = conv.SP(c.askQuestion("Name of your application", "App Name", conv.S(defConf.Name)))

	// cluster name
	conf.ClusterName = conv.SP(c.askQuestion("Name of the cluster your application will be deployed", "Cluster Name", conv.S(defConf.ClusterName)))

	// app port
	input := c.askQuestion("Does your application expose TCP port? (Enter 0 if not)", "Port", fmt.Sprintf("%d", conv.U16(defConf.Port)))
	parsed, err := strconv.ParseUint(input, 10, 16)
	if err != nil {
		return core.ExitWithError(fmt.Errorf("Invalid port number [%s]", input))
	}
	conf.Port = conv.U16P(uint16(parsed))

	// cpu
	input = c.askQuestion("CPU allocation per unit (1core = 1.0)", "CPU", fmt.Sprintf("%.2f", conv.F64(defConf.CPU)))
	parsedF, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return core.ExitWithError(fmt.Errorf("Invalid CPU [%s]", input))
	}
	conf.CPU = conv.F64P(parsedF)

	// Memory
	conf.Memory = conv.SP(c.askQuestion("Memory allocation per unit", "Memory", conv.S(defConf.Memory)))

	// Units
	input = c.askQuestion("Number of application units", "Units", fmt.Sprintf("%d", conv.U16(defConf.Units)))
	parsed, err = strconv.ParseUint(input, 10, 16)
	if err != nil {
		return core.ExitWithError(fmt.Errorf("Invalid units [%s]", input))
	}
	conf.Units = conv.U16P(uint16(parsed))

	// load balancer
	if conv.B(c.commandFlags.Default) || console.AskConfirm("Does your application need load balancing?", true) {
		conf.LoadBalancer = new(config.ConfigLoadBalancer)
		conf.LoadBalancer.HealthCheck = new(config.ConfigLoadBalancerHealthCheck)

		// https
		conf.LoadBalancer.IsHTTPS = defConf.LoadBalancer.IsHTTPS

		// port
		input := c.askQuestion("Load balancer port number", "Load Balancer Port", fmt.Sprintf("%d", conv.U16(defConf.Port)))
		parsed, err := strconv.ParseUint(input, 10, 16)
		if err != nil || parsed == 0 {
			return core.ExitWithError(fmt.Errorf("Invalid port number [%s]", input))
		}
		conf.LoadBalancer.Port = conv.U16P(uint16(parsed))

		// health check
		conf.LoadBalancer.HealthCheck.Path = conv.SP(c.askQuestion("Health check destination path", "Health Check Path", conv.S(defConf.LoadBalancer.HealthCheck.Path)))
		conf.LoadBalancer.HealthCheck.Status = conv.SP(c.askQuestion("HTTP codes to use when checking for a successful response", "Health Check Status", conv.S(defConf.LoadBalancer.HealthCheck.Status)))
		conf.LoadBalancer.HealthCheck.Interval = conv.SP(c.askQuestion("Approximate amount of time between health checks of an individual instance", "Health Check Interval", conv.S(defConf.LoadBalancer.HealthCheck.Interval)))
		conf.LoadBalancer.HealthCheck.Timeout = conv.SP(c.askQuestion("Amount of time during which no response from an instance means a failed health check", "Health Check Timeout", conv.S(defConf.LoadBalancer.HealthCheck.Timeout)))

		input = c.askQuestion("Number of consecutive health check successes required before considering an unhealthy instance to healthy.", "Healthy Limits", fmt.Sprintf("%d", conv.U16(defConf.LoadBalancer.HealthCheck.HealthyLimit)))
		parsed, err = strconv.ParseUint(input, 10, 16)
		if err != nil {
			return core.ExitWithError(fmt.Errorf("Invalid number [%s]", input))
		}
		conf.LoadBalancer.HealthCheck.HealthyLimit = conv.U16P(uint16(parsed))

		input = c.askQuestion("Number of consecutive health check failures required before considering an instance unhealthy.", "Unhealthy Limits", fmt.Sprintf("%d", conv.U16(defConf.LoadBalancer.HealthCheck.UnhealthyLimit)))
		parsed, err = strconv.ParseUint(input, 10, 16)
		if err != nil {
			return core.ExitWithError(fmt.Errorf("Invalid number [%s]", input))
		}
		conf.LoadBalancer.HealthCheck.UnhealthyLimit = conv.U16P(uint16(parsed))
	}

	// AWS
	{
		// elb lb name
		conf.AWS.ELBLoadBalancerName = conv.SP(c.askQuestion("ELB load balancer name", "ELB Load Balancer Name", conv.S(defConf.AWS.ELBLoadBalancerName)))

		// elb target name
		conf.AWS.ELBTargetGroupName = conv.SP(c.askQuestion("ELB target name", "ELB Target Group Name", conv.S(defConf.AWS.ELBTargetGroupName)))

		// elb security group
		conf.AWS.ELBSecurityGroup = conv.SP(c.askQuestion("Security group ID/name for ELB load balancer. Leave it blank to create default one.", "ELB Security Group", conv.S(defConf.AWS.ELBSecurityGroup)))

		// ecr repo name
		conf.AWS.ECRRepositoryName = conv.SP(c.askQuestion("ECR repository name", "ECR Namespace", conv.S(defConf.AWS.ECRRepositoryName)))
	}

	// Docker
	{
		conf.Docker.Bin = conv.SP(c.askQuestion("Docker executable path", "Docker Bin", conv.S(defConf.Docker.Bin)))
	}

	// config file path and format
	configFile, err := c.globalFlags.GetConfigFile()
	if err != nil {
		return err
	}
	configFileFormat := strings.ToLower(conv.S(c.globalFlags.ConfigFileFormat))
	if utils.IsBlank(configFileFormat) {
		switch strings.ToLower(filepath.Ext(configFile)) {
		case ".json":
			configFileFormat = flags.GlobalFlagsConfigFileFormatJSON
		default:
			configFileFormat = flags.GlobalFlagsConfigFileFormatYAML
		}
	}

	// write to file
	var configData []byte
	switch configFileFormat {
	case flags.GlobalFlagsConfigFileFormatYAML:
		configData, err = conf.ToYAML()
		if err != nil {
			return core.ExitWithError(fmt.Errorf("Failed to format configuration in YAML: %s", err.Error()))
		}
	case flags.GlobalFlagsConfigFileFormatJSON:
		configData, err = conf.ToJSONWithIndent()
		if err != nil {
			return core.ExitWithError(fmt.Errorf("Failed to format configuration in JSON: %s", err.Error()))
		}
	default:
		return core.ExitWithError(fmt.Errorf("Unsupported configuration file format [%s]", configFileFormat))
	}
	if err := ioutil.WriteFile(configFile, configData, 0644); err != nil {
		return core.ExitWithError(fmt.Errorf("Failed to write configuration file [%s]: %s", configFile, err.Error()))
	}
	console.Println("Configuration file was successfully created:", cc.Green(configFile))

	return nil
}

func (c *Command) askQuestion(description, question, defaultValue string) string {
	if conv.B(c.commandFlags.Default) {
		console.Println(cc.Blue(question)+":", cc.Green(defaultValue))
		return defaultValue
	}

	return console.AskQuestion(fmt.Sprintf("%s\n%s", cc.BlackH(description), cc.Blue(question)), defaultValue)
}
