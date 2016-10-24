package clusterscale

import (
	"fmt"

	"time"

	"strings"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Command struct {
	globalFlags    *flags.GlobalFlags
	commandFlags   *Flags
	awsClient      *aws.Client
	clusterNameArg *string
	instanceCount  *uint16
}

func (c *Command) Init(ka *kingpin.Application, globalFlags *flags.GlobalFlags) *kingpin.CmdClause {
	c.globalFlags = globalFlags

	cmd := ka.Command("cluster-scale",
		"See: "+console.ColorFnHelpLink("https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-scale"))
	c.commandFlags = NewFlags(cmd)

	c.clusterNameArg = cmd.Arg("cluster-name", "Cluster name").Required().String()

	c.instanceCount = cmd.Arg("instance-count", "Number of instances").Uint16()

	return cmd
}

func (c *Command) Run() error {
	c.awsClient = c.globalFlags.GetAWSClient()

	clusterName := strings.TrimSpace(conv.S(c.clusterNameArg))
	if !core.ClusterNameRE.MatchString(clusterName) {
		return console.ExitWithError(core.NewErrorExtraInfo(
			fmt.Errorf("Invalid cluster name [%s]", clusterName), "https://github.com/coldbrewcloud/coldbrew-cli/wiki/Configuration-File#cluster"))
	}

	autoScalingGroupName := core.DefaultAutoScalingGroupName(clusterName)

	console.Info("Auto Scaling Group")
	console.DetailWithResource("Name", autoScalingGroupName)

	autoScalingGroup, err := c.awsClient.AutoScaling().RetrieveAutoScalingGroup(autoScalingGroupName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve EC2 Auto Scaling Group [%s]: %s", autoScalingGroupName, err.Error())
	}
	if autoScalingGroup == nil {
		return console.ExitWithErrorString("EC2 Auto Scaling Group [%s] was not found.", autoScalingGroupName)
	}
	if !utils.IsBlank(conv.S(autoScalingGroup.Status)) {
		return console.ExitWithErrorString("EC2 Auto Scaling Group [%s] is being deleted: %s", autoScalingGroupName, conv.S(autoScalingGroup.Status))
	}

	currentTarget := uint16(conv.I64(autoScalingGroup.DesiredCapacity))
	newTarget := conv.U16(c.instanceCount)

	console.DetailWithResource("Current Target", fmt.Sprintf("%d", currentTarget))
	console.DetailWithResource("New Target", fmt.Sprintf("%d", newTarget))

	console.Blank()

	if currentTarget < newTarget {
		if err := c.scaleOut(autoScalingGroupName, currentTarget, newTarget); err != nil {
			return console.ExitWithError(err)
		}
	} else if currentTarget > newTarget {
		if err := c.scaleIn(autoScalingGroupName, currentTarget, newTarget); err != nil {
			return console.ExitWithError(err)
		}
	} else {

	}

	return nil
}

func (c *Command) scaleOut(autoScalingGroupName string, currentTarget, newTarget uint16) error {
	console.UpdatingResource(fmt.Sprintf("Updating desired capacity to %d", newTarget), autoScalingGroupName, true)

	err := c.awsClient.AutoScaling().UpdateAutoScalingGroupCapacity(autoScalingGroupName, 0, newTarget, newTarget)
	if err != nil {
		return fmt.Errorf("Failed to update capacity of EC2 Auto Scaling Group [%s]: %s", autoScalingGroupName, err.Error())
	}

	err = utils.Retry(func() (bool, error) {
		autoScalingGroup, err := c.awsClient.AutoScaling().RetrieveAutoScalingGroup(autoScalingGroupName)
		if err != nil {
			return false, fmt.Errorf("Failed to retrieve EC2 Auto Scaling Group [%s]: %s", autoScalingGroupName, err.Error())
		}
		if autoScalingGroup == nil {
			return false, fmt.Errorf("EC2 Auto Scaling Group [%s] was not found.", autoScalingGroupName)
		}
		if uint16(len(autoScalingGroup.Instances)) == newTarget {
			return false, nil
		}
		return true, nil
	}, time.Second, 5*time.Minute)
	if err != nil {
		return err
	}

	console.Blank()
	//console.Info(fmt.Sprintf("EC2 Auto Scaling Group [%s] now has %d instances.", autoScalingGroupName, newTarget))

	return nil
}

func (c *Command) scaleIn(autoScalingGroupName string, currentTarget, newTarget uint16) error {
	console.UpdatingResource(fmt.Sprintf("Updating desired capacity to %d", newTarget), autoScalingGroupName, true)

	err := c.awsClient.AutoScaling().UpdateAutoScalingGroupCapacity(autoScalingGroupName, 0, newTarget, newTarget)
	if err != nil {
		return fmt.Errorf("Failed to update capacity of EC2 Auto Scaling Group [%s]: %s", autoScalingGroupName, err.Error())
	}

	err = utils.Retry(func() (bool, error) {
		autoScalingGroup, err := c.awsClient.AutoScaling().RetrieveAutoScalingGroup(autoScalingGroupName)
		if err != nil {
			return false, fmt.Errorf("Failed to retrieve EC2 Auto Scaling Group [%s]: %s", autoScalingGroupName, err.Error())
		}
		if autoScalingGroup == nil {
			return false, fmt.Errorf("EC2 Auto Scaling Group [%s] was not found.", autoScalingGroupName)
		}
		if uint16(len(autoScalingGroup.Instances)) == newTarget {
			return false, nil
		}
		return true, nil
	}, time.Second, 5*time.Minute)
	if err != nil {
		return err
	}

	console.Blank()
	//console.Info(fmt.Sprintf("EC2 Auto Scaling Group [%s] now has %d instances.", autoScalingGroupName, newTarget))

	return nil
}
