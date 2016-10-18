package destroy

import (
	"errors"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"gopkg.in/alecthomas/kingpin.v2"
)

type DestroyCommand struct {
	appFlags    *flags.GlobalFlags
	deployFlags *flags.DeployFlags
	awsClient   *aws.Client
}

func (dc *DestroyCommand) Init(ka *kingpin.Application, appFlags *flags.GlobalFlags) *kingpin.CmdClause {
	dc.appFlags = appFlags

	cmd := ka.Command("destroy", "(destroy description goes here)")
	dc.deployFlags = flags.NewDeployFlags(cmd)

	dc.awsClient = aws.NewClient(*dc.appFlags.AWSRegion).WithCredentials(*dc.appFlags.AWSAccessKey, *dc.appFlags.AWSSecretKey)

	return cmd
}

func (dc *DestroyCommand) Run() error {
	return errors.New("destory command not implemented")
}
