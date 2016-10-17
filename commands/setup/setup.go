package setup

import (
	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"gopkg.in/alecthomas/kingpin.v2"
)

type SetupCommand struct {
	appFlags    *flags.AppFlags
	deployFlags *flags.DeployFlags
	awsClient   *aws.Client
}

func (sc *SetupCommand) Init(ka *kingpin.Application, appFlags *flags.AppFlags) *kingpin.CmdClause {
	sc.appFlags = appFlags

	cmd := ka.Command("setup", "(setup description goes here)")
	sc.deployFlags = flags.NewDeployFlags(cmd)

	sc.awsClient = aws.NewClient(*sc.appFlags.AWSRegion).WithCredentials(*sc.appFlags.AWSAccessKey, *sc.appFlags.AWSSecretKey)

	return cmd
}

func (sc *SetupCommand) Run() error {
	return nil
}
