package flags

import (
	"strings"

	"fmt"
	"path/filepath"

	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	GlobalFlagsConfigFileFormatJSON = "json"
	GlobalFlagsConfigFileFormatYAML = "yaml"
)

type GlobalFlags struct {
	ConfigFile       *string `json:"config,omitempty"`
	ConfigFileFormat *string `json:"config-format,omitempty"`
	AppDirectory     *string `json:"app-dir,omitempty"`
	Verbose          *bool   `json:"verbose,omitempty"`
	AWSRegion        *string `json:"aws-region,omitempty"`
	AWSAccessKey     *string `json:"aws-access-key,omitempty"`
	AWSSecretKey     *string `json:"aws-secret-key,omitempty"`
}

func NewGlobalFlags(ka *kingpin.Application) *GlobalFlags {
	return &GlobalFlags{
		ConfigFile:       ka.Flag("config", "Configuration file path").Short('C').String(),
		ConfigFileFormat: ka.Flag("config-format", "Configuraiton file format (JSON/YAML)").Default(GlobalFlagsConfigFileFormatYAML).String(),
		AppDirectory:     ka.Flag("app-dir", "Application directory").Short('D').Default(".").String(),
		Verbose:          ka.Flag("verbose", "Enable verbose logging").Short('V').Default("false").Bool(),
		AWSRegion:        ka.Flag("aws-region", "AWS region name (default: $AWS_REGION)").Envar("AWS_REGION").Default("us-west-2").String(),
		AWSAccessKey:     ka.Flag("aws-access-key", "AWS access key (default: $AWS_ACCESS_KEY_ID)").Envar("AWS_ACCESS_KEY_ID").Default("").String(),
		AWSSecretKey:     ka.Flag("aws-secret-key", "AWS secret key (default: $AWS_SECRET_ACCESS_KEY)").Envar("AWS_SECRET_ACCESS_KEY").Default("").String(),
	}
}

func (gf *GlobalFlags) ResolveAppDirectory() (string, error) {
	appDir := "."
	if gf.AppDirectory != nil {
		appDir = strings.TrimSpace(conv.S(gf.AppDirectory))
	}

	absPath, err := filepath.Abs(appDir)
	if err != nil {
		return "", fmt.Errorf("Error resolving application directory [%s]: %s", appDir, err.Error())
	}

	return absPath, nil
}

func (gf *GlobalFlags) ResolveConfigFile() (string, error) {
	configFile := strings.TrimSpace(conv.S(gf.ConfigFile))

	// if specified config file is absolute path, just use it
	if configFile != "" && filepath.IsAbs(configFile) {
		return configFile, nil
	}

	// default: coldbrew.conf file in app directory
	if configFile == "" {
		configFile = "./coldbrew.conf"
	}

	appDir, err := gf.ResolveAppDirectory()
	if err != nil {
		return "", err
	}

	return filepath.Join(appDir, configFile), nil
}
