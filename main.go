package main

import (
	"fmt"
	"os"

	"github.com/coldbrewcloud/coldbrew-cli/commands"
	"github.com/coldbrewcloud/coldbrew-cli/commands/delete"
	"github.com/coldbrewcloud/coldbrew-cli/commands/deploy"
	"github.com/coldbrewcloud/coldbrew-cli/commands/setup"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	appName = "coldbrew"
	appHelp = "(some application description goes here)"
)

type CLIApp struct {
	kingpinApp *kingpin.Application
	appFlags   *flags.AppFlags
	commands   map[string]commands.Command
}

func main() {
	kingpinApp := kingpin.New(appName, appHelp)
	kingpinApp.Version(Version)
	appFlags := flags.NewAppFlags(kingpinApp)

	cmds := make(map[string]commands.Command)

	// register commands
	{
		// deploy
		deployCommand := &deploy.DeployCommand{}
		kpDeployCommand := deployCommand.Init(kingpinApp, appFlags)
		cmds[kpDeployCommand.FullCommand()] = deployCommand

		// setup
		setupCommand := &setup.SetupCommand{}
		kpSetupCommand := setupCommand.Init(kingpinApp, appFlags)
		cmds[kpSetupCommand.FullCommand()] = setupCommand

		// delete
		deleteCommand := &delete.DeleteCommand{}
		kpDeleteCommand := deleteCommand.Init(kingpinApp, appFlags)
		cmds[kpDeleteCommand.FullCommand()] = deleteCommand
	}

	// parse CLI inputs
	cmd, err := kingpinApp.Parse(os.Args[1:])
	if err != nil {
		console.Errorf(err.Error())
		os.Exit(5)
	}

	// setup debug logging
	console.EnableDebugf(*appFlags.Debug, *appFlags.DebugLogPrefix)

	// app flags
	/*
		if *appFlags.Debug {
			asMap, err := utils.AsMap(appFlags)
			if err != nil {
				console.Errorf("Error enumerating app flags: %s", err.Error())
				os.Exit(99)
			}

			for k, v := range asMap {
				console.Debugf("FLAG %s=%v\n", k, v)
			}
		}
	*/

	// execute command
	for n, c := range cmds {
		if n == cmd {
			if err := c.Run(); err != nil {
				console.Errorf("Error: %s\n", err.Error())
				os.Exit(10)
			}
			os.Exit(0)
		}
	}

	panic(fmt.Errorf("Unknown command: %s", cmd))
}
