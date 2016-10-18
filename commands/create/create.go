package create

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"path/filepath"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/config"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"gopkg.in/alecthomas/kingpin.v2"
)

type CreateCommand struct {
	globalFlags *flags.GlobalFlags
	createFlags *flags.CreateFlags
	awsClient   *aws.Client
}

func (cc *CreateCommand) Init(ka *kingpin.Application, globalFlags *flags.GlobalFlags) *kingpin.CmdClause {
	cc.globalFlags = globalFlags

	cmd := ka.Command("create", "(create description goes here)").Alias("init")
	cc.createFlags = flags.NewCreateFlags(cmd)

	cc.awsClient = aws.NewClient(*cc.globalFlags.AWSRegion).WithCredentials(*cc.globalFlags.AWSAccessKey, *cc.globalFlags.AWSSecretKey)

	return cmd
}

func (cc *CreateCommand) Run() error {
	configFilePath, err := cc.globalFlags.ResolveConfigFile()
	if err != nil {
		return err
	}

	if utils.FileExists(configFilePath) && !conv.B(cc.createFlags.OverwriteExisting) {
		return errors.New("Config file [%s] already exists. Use \"--overwrite\" flag if you want to overwrite to it.")
	}

	autoConfig, err := cc.createAutoConfig()
	if err != nil {
		return err
	}

	var configData []byte
	configFileFormat := strings.ToLower(conv.S(cc.globalFlags.ConfigFileFormat))
	if configFileFormat == flags.GlobalFlagsConfigFileFormatYAML {
		configData, err = autoConfig.ToYAML()
		if err != nil {
			return err
		}
	} else if configFileFormat == flags.GlobalFlagsConfigFileFormatJSON {
		configData, err = autoConfig.ToJSON()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Unknown config file format: %s", configFileFormat)
	}

	if err := ioutil.WriteFile(configFilePath, configData, 0644); err != nil {
		return fmt.Errorf("Failed to write config to file [%s]: %s", configFilePath, err.Error())
	}

	return nil
}

func (cc *CreateCommand) createAutoConfig() (*config.Config, error) {
	autoConfig := &config.Config{}

	// app name: derive from app directory name
	{
		appDir, err := cc.globalFlags.ResolveAppDirectory()
		if err != nil {
			return nil, err
		}
		tokens := filepath.SplitList(appDir)

		autoConfig.Name = tokens[len(tokens)-1]

		console.Debugln("Config name:", autoConfig.Name)
	}

	// port: use Dockerfile EXPOSE or 80
	{
	}

	return autoConfig, nil
}
