package main

import (
	"fmt"
	"os"

	"io/ioutil"
	"strings"

	"github.com/coldbrewcloud/coldbrew-cli/commands"
	"github.com/coldbrewcloud/coldbrew-cli/commands/clustercreate"
	"github.com/coldbrewcloud/coldbrew-cli/commands/deploy"
	"github.com/coldbrewcloud/coldbrew-cli/config"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	appName = "coldbrew"
	appHelp = "(some application description goes here)"
)

type CLIApp struct {
	kingpinApp *kingpin.Application
	appFlags   *flags.GlobalFlags
	commands   map[string]commands.Command
}

func main() {
	kingpinApp := kingpin.New(appName, appHelp)
	kingpinApp.Version(Version)
	globalFlags := flags.NewGlobalFlags(kingpinApp)

	cmds := make(map[string]commands.Command)

	// register commands
	{
		// deploy
		deployCommand := &deploy.DeployCommand{}
		kpDeployCommand := deployCommand.Init(kingpinApp, globalFlags)
		cmds[kpDeployCommand.FullCommand()] = deployCommand

		// cluster-create
		clusterCreateCommand := &clustercreate.Command{}
		kpClusterCreateCommand := clusterCreateCommand.Init(kingpinApp, globalFlags)
		cmds[kpClusterCreateCommand.FullCommand()] = clusterCreateCommand
	}

	// parse CLI inputs
	cmd, err := kingpinApp.Parse(os.Args[1:])
	if err != nil {
		console.Errorf(err.Error())
		os.Exit(5)
	}

	// setup debug logging
	console.EnableDebugf(*globalFlags.Verbose, "")

	// load configuration file
	var cfg *config.Config
	{
		configFile := conv.S(globalFlags.ConfigFile)
		if !utils.IsBlank(configFile) {
			data, err := ioutil.ReadFile(configFile)
			if err != nil {
				console.Errorf("Failed to read configuration file [%s]: %s\n", configFile, err.Error())
				os.Exit(10)
			}

			configFileFormat := strings.ToLower(conv.S(globalFlags.ConfigFileFormat))
			if configFileFormat == flags.GlobalFlagsConfigFileFormatYAML {
				if err := cfg.FromYAML(data); err != nil {
					console.Errorf("Failed to read configuration file in YAML: %s\n", err.Error())
					os.Exit(11)
				}
			} else if configFileFormat == flags.GlobalFlagsConfigFileFormatJSON {
				if err := cfg.FromJSON(data); err != nil {
					console.Errorf("Failed to read configuration file in JSON: %s\n", err.Error())
					os.Exit(11)
				}
			} else {
				console.Errorf("Unknown configuration file format: %s\n", configFileFormat)
				os.Exit(12)
			}
		}
	}

	// execute command
	for n, c := range cmds {
		if n == cmd {
			if err := c.Run(cfg); err != nil {
				console.Errorf("Error: %s\n", err.Error())
				os.Exit(10)
			}
			os.Exit(0)
		}
	}

	panic(fmt.Errorf("Unknown command: %s", cmd))
}
