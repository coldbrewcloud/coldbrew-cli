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

	// AWS client
	c.awsClient = c.globalFlags.GetAWSClient()

	conf := &config.Config{}
	conf.AWS = new(config.ConfigAWS)
	conf.Docker = new(config.ConfigDocker)

	appDir := conv.S(c.globalFlags.AppDirectory)
	if utils.IsBlank(appDir) {
		appDir = "."
	}
	appDir, err = filepath.Abs(appDir)
	if err != nil {
		return c.exitWithError(fmt.Errorf("Failed to resolve application directory [%s]: %s", appDir, err.Error()))
	}
	//console.Println("App Directory:", cc.Green(appDir))

	// app name
	conf.Name = c.askQuestion("Name of your application", "App Name", filepath.Base(appDir))

	// cluster name
	conf.ClusterName = c.askQuestion("Name of the cluster your application will be deployed", "Cluster Name", "cluster1")

	// app port
	input := c.askQuestion("Does your application expose TCP port? (Enter 0 if not)", "Port", "0")
	parsed, err := strconv.ParseUint(input, 10, 16)
	if err != nil {
		return c.exitWithError(fmt.Errorf("Invalid port number [%s]", input))
	}
	conf.Port = uint16(parsed)

	// cpu
	input = c.askQuestion("CPU allocation per unit (1core = 1.0)", "CPU", "0.5")
	parsedF, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return c.exitWithError(fmt.Errorf("Invalid CPU [%s]", input))
	}
	conf.CPU = parsedF

	// Memory
	conf.Memory = c.askQuestion("Memory allocation per unit", "Memory", "500m")

	// Units
	input = c.askQuestion("Number of application units", "Units", "1")
	parsed, err = strconv.ParseUint(input, 10, 16)
	if err != nil {
		return c.exitWithError(fmt.Errorf("Invalid units [%s]", input))
	}
	conf.Units = uint16(parsed)

	// Environment variables
	// ...

	// load balancer
	if conv.B(c.commandFlags.Default) || console.AskConfirm("Does your application need load balancing?") {
		conf.LoadBalancer = new(config.ConfigLoadBalancer)
		conf.LoadBalancer.HealthCheck = new(config.ConfigLoadBalancerHealthCheck)

		// https
		conf.LoadBalancer.IsHTTPS = false // TODO: not implemented

		// port
		input := c.askQuestion("Load balancer port number", "Load Balancer Port", "80")
		parsed, err := strconv.ParseUint(input, 10, 16)
		if err != nil || parsed == 0 {
			return c.exitWithError(fmt.Errorf("Invalid port number [%s]", input))
		}
		conf.LoadBalancer.Port = uint16(parsed)

		// health check
		conf.LoadBalancer.HealthCheck.Path = c.askQuestion("Health check destination path", "Health Check Path", "/")
		conf.LoadBalancer.HealthCheck.Status = c.askQuestion("HTTP codes to use when checking for a successful response", "Health Check Status", "200-299")
		conf.LoadBalancer.HealthCheck.Interval = c.askQuestion("Approximate amount of time between health checks of an individual instance", "Health Check Interval", "30s")
		conf.LoadBalancer.HealthCheck.Timeout = c.askQuestion("Amount of time during which no response from an instance means a failed health check", "Health Check Timeout", "5s")

		input = c.askQuestion("Number of consecutive health check successes required before considering an unhealthy instance to healthy.", "Healthy Limits", "5")
		parsed, err = strconv.ParseUint(input, 10, 16)
		if err != nil {
			return c.exitWithError(fmt.Errorf("Invalid number [%s]", input))
		}
		conf.LoadBalancer.HealthCheck.HealthyLimit = uint16(parsed)

		input = c.askQuestion("Number of consecutive health check failures required before considering an instance unhealthy.", "Unhealthy Limits", "5")
		parsed, err = strconv.ParseUint(input, 10, 16)
		if err != nil {
			return c.exitWithError(fmt.Errorf("Invalid number [%s]", input))
		}
		conf.LoadBalancer.HealthCheck.UnhealthyLimit = uint16(parsed)
	}

	// AWS
	{
		// elb lb name
		conf.AWS.ELBLoadBalancerName = c.askQuestion("ELB load balancer name", "ELB LB Name", conf.Name)

		// elb target name
		conf.AWS.ELBTargetName = c.askQuestion("ELB target name", "ELB Target Name", conf.Name)

		// elb security group
		conf.AWS.ELBSecurityGroup = c.askQuestion("Security group ID/name for ELB load balancer. Leave it blank to create default one.", "ELB Security Group", "")

		// ecr namespace
		conf.AWS.ECRNamespace = c.askQuestion("ECR namespace", "ECR Namespace", "coldbrew")
	}

	// Docker
	{
		conf.Docker.Bin = c.askQuestion("Docker executable path", "Docker Bin", "docker")
	}

	// config file path and format
	configFile := conv.S(c.globalFlags.ConfigFile)
	configFileFormat := strings.ToLower(conv.S(c.globalFlags.ConfigFileFormat))
	if utils.IsBlank(configFile) {
		configFile = "coldbrew.conf"

		if utils.IsBlank(configFileFormat) {
			configFileFormat = flags.GlobalFlagsConfigFileFormatYAML
		}
	} else if utils.IsBlank(configFileFormat) {
		switch strings.ToLower(filepath.Ext(configFile)) {
		case ".json":
			configFileFormat = flags.GlobalFlagsConfigFileFormatJSON
		default:
			configFileFormat = flags.GlobalFlagsConfigFileFormatYAML
		}
	}
	if !filepath.IsAbs(configFile) {
		configFile = filepath.Join(appDir, configFile)
	}

	// write to file
	var configData []byte
	switch configFileFormat {
	case flags.GlobalFlagsConfigFileFormatYAML:
		configData, err = conf.ToYAML()
		if err != nil {
			return c.exitWithError(fmt.Errorf("Failed to format configuration in YAML: %s", err.Error()))
		}
	case flags.GlobalFlagsConfigFileFormatJSON:
		configData, err = conf.ToJSONWithIndent()
		if err != nil {
			return c.exitWithError(fmt.Errorf("Failed to format configuration in JSON: %s", err.Error()))
		}
	default:
		return c.exitWithError(fmt.Errorf("Unsupported configuration file format [%s]", configFileFormat))
	}
	if err := ioutil.WriteFile(configFile, configData, 0644); err != nil {
		return c.exitWithError(fmt.Errorf("Failed to write configuration file [%s]: %s", configFile, err.Error()))
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

func (c *Command) exitWithError(err error) error {
	console.Errorln(cc.Red("Error:"), err.Error())
	return nil
}

func (c *Command) exitWithErrorInfo(err error, infoURL string) error {
	console.Errorln(cc.Red("Error:"), err.Error())
	console.Errorln(cc.BlackH("More Info:"), infoURL)
	return nil
}
