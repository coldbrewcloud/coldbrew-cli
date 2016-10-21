package main

import (
	"fmt"
	"os"

	"github.com/coldbrewcloud/coldbrew-cli/commands"
	"github.com/coldbrewcloud/coldbrew-cli/commands/clustercreate"
	"github.com/coldbrewcloud/coldbrew-cli/commands/clusterdelete"
	"github.com/coldbrewcloud/coldbrew-cli/commands/clusterstatus"
	"github.com/coldbrewcloud/coldbrew-cli/commands/create"
	"github.com/coldbrewcloud/coldbrew-cli/commands/deploy"
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
		deployCommand := &deploy.Command{}
		kpDeployCommand := deployCommand.Init(kingpinApp, globalFlags)
		cmds[kpDeployCommand.FullCommand()] = deployCommand

		// create / init
		createCommand := &create.Command{}
		kpCreateCommand := createCommand.Init(kingpinApp, globalFlags)
		cmds[kpCreateCommand.FullCommand()] = createCommand

		// cluster-create
		clusterCreateCommand := &clustercreate.Command{}
		kpClusterCreateCommand := clusterCreateCommand.Init(kingpinApp, globalFlags)
		cmds[kpClusterCreateCommand.FullCommand()] = clusterCreateCommand

		// cluster-status
		clusterStatusCommand := &clusterstatus.Command{}
		kpClusterStatusCommand := clusterStatusCommand.Init(kingpinApp, globalFlags)
		cmds[kpClusterStatusCommand.FullCommand()] = clusterStatusCommand

		// cluster-delete
		clusterDeleteCommand := &clusterdelete.Command{}
		kpClusterDeleteCommand := clusterDeleteCommand.Init(kingpinApp, globalFlags)
		cmds[kpClusterDeleteCommand.FullCommand()] = clusterDeleteCommand
	}

	// parse CLI inputs
	cmd, err := kingpinApp.Parse(os.Args[1:])
	if err != nil {
		console.Errorf(err.Error())
		os.Exit(5)
	}

	// setup debug logging
	console.EnableDebugf(*globalFlags.Verbose, "")

	// execute command
	for n, c := range cmds {
		if n == cmd {
			if err := c.Run(); err != nil {
				console.Errorf("Error: %s\n", err.Error())
				os.Exit(40)
			}
			os.Exit(0)
		}
	}

	panic(fmt.Errorf("Unknown command: %s", cmd))
}
