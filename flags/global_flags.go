package flags

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	GlobalFlagsConfigFileFormatJSON = "json"
	GlobalFlagsConfigFileFormatYAML = "yaml"
)

type GlobalFlags struct {
	AppDirectory     *string `json:"app-dir,omitempty"`
	ConfigFile       *string `json:"config,omitempty"`
	ConfigFileFormat *string `json:"config-format,omitempty"`
	DisableColoring  *bool   `json:"disable-color,omitempty"`
	Verbose          *bool   `json:"verbose,omitempty"`
	AWSAccessKey     *string `json:"aws-access-key,omitempty"`
	AWSSecretKey     *string `json:"aws-secret-key,omitempty"`
	AWSRegion        *string `json:"aws-region,omitempty"`
	AWSVPC           *string `json:"aws-vpc,omitempty"`
}

func NewGlobalFlags(ka *kingpin.Application) *GlobalFlags {
	return &GlobalFlags{
		ConfigFile:       ka.Flag("config", "Configuration file path").Short('C').String(),
		ConfigFileFormat: ka.Flag("config-format", "Configuraiton file format (JSON/YAML)").Default(GlobalFlagsConfigFileFormatYAML).String(),
		AppDirectory:     ka.Flag("app-dir", "Application directory").Short('D').Default(".").String(),
		DisableColoring:  ka.Flag("disable-color", "Disable color outputs").Bool(),
		Verbose:          ka.Flag("verbose", "Enable verbose logging").Short('V').Default("false").Bool(),
		AWSAccessKey:     ka.Flag("aws-access-key", "AWS access key (default: $AWS_ACCESS_KEY_ID)").Envar("AWS_ACCESS_KEY_ID").Default("").String(),
		AWSSecretKey:     ka.Flag("aws-secret-key", "AWS secret key (default: $AWS_SECRET_ACCESS_KEY)").Envar("AWS_SECRET_ACCESS_KEY").Default("").String(),
		AWSRegion:        ka.Flag("aws-region", "AWS region name (default: $AWS_REGION)").Envar("AWS_REGION").Default("us-west-2").String(),
		AWSVPC:           ka.Flag("aws-vpc", "AWS VPC ID").Envar("AWS_VPC").String(),
	}
}

// GetApplicationDirectory returns an absolute path of the application directory.
func (gf *GlobalFlags) GetApplicationDirectory() (string, error) {
	appDir := conv.S(gf.AppDirectory)
	if utils.IsBlank(appDir) {
		appDir = "." // default: current working directory
	}

	// resolve to absolute path
	absPath, err := filepath.Abs(appDir)
	if err != nil {
		return "", fmt.Errorf("Error retrieving absolute path [%s]: %s", appDir, err.Error())
	}

	return absPath, nil
}

// GetConfigFile returns an absolute path of the configuration file.
func (gf *GlobalFlags) GetConfigFile() (string, error) {
	configFile := conv.S(gf.ConfigFile)

	// if specified config file is absolute path, just use it
	if !utils.IsBlank(configFile) && filepath.IsAbs(configFile) {
		return configFile, nil
	}

	if utils.IsBlank(configFile) {
		configFile = "./coldbrew.conf" // default: coldbrew.conf
	}

	// join with application directory
	appDir, err := gf.GetApplicationDirectory()
	if err != nil {
		return "", err
	}

	return filepath.Join(appDir, configFile), nil
}

func (gf *GlobalFlags) GetAWSClient() *aws.Client {
	return aws.NewClient(conv.S(gf.AWSRegion), conv.S(gf.AWSAccessKey), conv.S(gf.AWSSecretKey))
}

func (gf *GlobalFlags) GetAWSRegionAndVPCID() (string, string, error) {
	if utils.IsBlank(conv.S(gf.AWSRegion)) {
		return "", "", errors.New("AWS region cannot be blank.")
	}

	awsClient := gf.GetAWSClient()

	// VPC ID explicitly specified: make sure it's really there
	if !utils.IsBlank(conv.S(gf.AWSVPC)) {
		vpc, err := awsClient.EC2().RetrieveVPC(conv.S(gf.AWSVPC))
		if err != nil {
			return "", "", fmt.Errorf("Failed to retrieve VPC [%s]: %s", conv.S(gf.AWSVPC), err.Error())
		}
		if vpc == nil {
			return "", "", fmt.Errorf("VPC [%s] was not found.", conv.S(gf.AWSVPC))
		}
		return conv.S(gf.AWSRegion), conv.S(gf.AWSVPC), nil
	}

	// if VPC is not specified, try to find account default VPC
	defaultVPC, err := awsClient.EC2().RetrieveDefaultVPC()
	if err != nil {
		return "", "", fmt.Errorf("Failed to retrieve default VPC: %s", err.Error())
	}
	if defaultVPC == nil {
		return "", "", errors.New("Your AWS account does not have default VPC. You must explicitly specify VPC ID using --aws-vpc flag.")
	}
	return conv.S(gf.AWSRegion), conv.S(defaultVPC.VpcId), nil
}
