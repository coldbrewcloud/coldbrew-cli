package delete

import (
	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"gopkg.in/alecthomas/kingpin.v2"
)

type DeleteCommand struct {
	appFlags    *flags.AppFlags
	deployFlags *flags.DeployFlags
	awsClient   *aws.Client
}

func (dc *DeleteCommand) Init(ka *kingpin.Application, appFlags *flags.AppFlags) *kingpin.CmdClause {
	dc.appFlags = appFlags

	cmd := ka.Command("delete", "(delete description goes here)")
	dc.deployFlags = flags.NewDeployFlags(cmd)

	dc.awsClient = aws.NewClient(*dc.appFlags.AWSRegion).WithCredentials(*dc.appFlags.AWSAccessKey, *dc.appFlags.AWSSecretKey)

	return cmd
}

func (dc *DeleteCommand) Run() error {
	return nil
}
