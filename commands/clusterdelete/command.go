package clusterdelete

import (
	"fmt"
	"strings"
	"time"

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
}

func (c *Command) Init(ka *kingpin.Application, globalFlags *flags.GlobalFlags) *kingpin.CmdClause {
	c.globalFlags = globalFlags

	cmd := ka.Command("cluster-delete", "(cluster-delete description goes here)")
	c.commandFlags = NewFlags(cmd)

	c.clusterNameArg = cmd.Arg("cluster-name", "Cluster name").Required().String()

	return cmd
}

func (c *Command) Run() error {
	c.awsClient = c.globalFlags.GetAWSClient()

	clusterName := strings.TrimSpace(conv.S(c.clusterNameArg))

	console.Info("Determining AWS resources that need to be deleted...")
	deleteECSCluster := false
	deleteECSServiceRole := false
	deleteInstanceProfile := false
	deleteInstanceSecurityGroups := false
	deleteLaunchConfiguration := false
	deleteAutoScalingGroup := false

	// ECS cluster
	ecsClusterName := core.DefaultECSClusterName(clusterName)
	ecsCluster, err := c.awsClient.ECS().RetrieveCluster(ecsClusterName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve ECS Cluster [%s]: %s", ecsClusterName, err.Error())
	}
	if ecsCluster != nil && conv.S(ecsCluster.Status) != "INACTIVE" {
		deleteECSCluster = true
		console.DetailWithResource("ECS Cluster", ecsClusterName)
	}

	// ECS service role
	ecsServiceRoleName := core.DefaultECSServiceRoleName(clusterName)
	ecsServiceRole, err := c.awsClient.IAM().RetrieveRole(ecsServiceRoleName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve IAM Role [%s]: %s", ecsServiceRoleName, err.Error())
	}
	if ecsServiceRole != nil {
		deleteECSServiceRole = true
		console.DetailWithResource("IAM Role for ECS Services", ecsServiceRoleName)
	}

	// launch configuration
	lcName := core.DefaultLaunchConfigurationName(clusterName)
	launchConfiguration, err := c.awsClient.AutoScaling().RetrieveLaunchConfiguration(lcName)
	if err != nil {
		return console.ExitWithErrorString("Failed to delete Launch Configuration [%s]: %s", lcName, err.Error())
	}
	if launchConfiguration != nil {
		deleteLaunchConfiguration = true
		console.DetailWithResource("EC2 Launch Configuration for ECS Container Instances", lcName)
	}

	// auto scaling group
	asgName := core.DefaultAutoScalingGroupName(clusterName)
	autoScalingGroup, err := c.awsClient.AutoScaling().RetrieveAutoScalingGroup(asgName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve Auto Scaling Group [%s]: %s", asgName, err.Error())
	}
	if autoScalingGroup != nil && utils.IsBlank(conv.S(autoScalingGroup.Status)) {
		tags, err := c.awsClient.AutoScaling().RetrieveTagsForAutoScalingGroup(asgName)
		if err != nil {
			return console.ExitWithErrorString("Failed to retrieve tags for EC2 Auto Scaling Group [%s]: %s", asgName, err.Error())
		}
		if _, ok := tags[core.AWSTagNameCreatedTimestamp]; ok {
			deleteAutoScalingGroup = true
			console.DetailWithResource("EC2 Auto Scaling Group for ECS Container Instances", asgName)
		}
	}

	// instance profile
	instanceProfileName := core.DefaultInstanceProfileName(clusterName)
	instanceProfile, err := c.awsClient.IAM().RetrieveInstanceProfile(instanceProfileName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve Instance Profile [%s]: %s", instanceProfileName, err.Error())
	}
	if instanceProfile != nil {
		deleteInstanceProfile = true
		console.DetailWithResource("IAM Instance Profile for ECS Container Instances", instanceProfileName)
	}

	// instance security group
	sgName := core.DefaultInstanceSecurityGroupName(clusterName)
	securityGroup, err := c.awsClient.EC2().RetrieveSecurityGroupByName(sgName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve Security Group [%s]: %s", sgName, err.Error())
	}
	if securityGroup != nil {
		tags, err := c.awsClient.EC2().RetrieveTags(conv.S(securityGroup.GroupId))
		if err != nil {
			return console.ExitWithErrorString("Failed to retrieve tags for EC2 Security Group [%s]: %s", sgName, err.Error())
		}
		if _, ok := tags[core.AWSTagNameCreatedTimestamp]; ok {
			deleteInstanceSecurityGroups = true
			console.DetailWithResource("EC2 Security Group for ECS Container Instances", sgName)
		}
	}

	if !deleteECSServiceRole && !deleteECSCluster && !deleteLaunchConfiguration && !deleteAutoScalingGroup &&
		!deleteInstanceProfile && !deleteInstanceSecurityGroups {
		console.Info("Looks like everything's already cleaned up.")
		return nil
	}

	console.Blank()

	// confirmation
	if !conv.B(c.commandFlags.ForceDelete) && !console.AskConfirm("Do you want to delete these resources?", false) {
		return nil
	}

	console.Blank()

	// delete auto scaling group
	if deleteAutoScalingGroup {
		console.UpdatingResource("Terminating instances in EC2 Auto Scaling Group", asgName, true)

		if err := c.scaleDownAutoScalingGroup(autoScalingGroup); err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(console.ColorFnResourceNegative("Error:"), err.Error())
			} else {
				return console.ExitWithError(err)
			}
		} else {
			console.RemovingResource("Deleting EC2 Auto Scaling Group", asgName, true)
			if err := c.awsClient.AutoScaling().DeleteAutoScalingGroup(asgName, true); err != nil {
				if conv.B(c.commandFlags.ContinueOnError) {
					console.Errorln(console.ColorFnResourceNegative("Error:"), err.Error())
				} else {
					return console.ExitWithError(err)
				}
			}
		}
	}

	// delete launch configuration
	if deleteLaunchConfiguration {
		console.RemovingResource("Deleting EC2 Launch Configuration", lcName, false)

		if err := c.awsClient.AutoScaling().DeleteLaunchConfiguration(lcName); err != nil {
			err = fmt.Errorf("Failed to delete Launch Configuration [%s]: %s", lcName, err.Error())
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(console.ColorFnResourceNegative("Error:"), err.Error())
			} else {
				return console.ExitWithError(err)
			}
		}
	}

	// delete instance profile
	if deleteInstanceProfile {
		console.RemovingResource("Deleting IAM Instance Profile", instanceProfileName, false)

		if err := c.deleteDefaultInstanceProfile(instanceProfileName); err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(console.ColorFnResourceNegative("Error:"), err.Error())
			} else {
				return console.ExitWithError(err)
			}
		}
	}

	// delete instance security groups
	if deleteInstanceSecurityGroups {
		console.RemovingResource("Deleting EC2 Security Group", sgName, false)

		err = utils.RetryOnAWSErrorCode(func() error {
			return c.awsClient.EC2().DeleteSecurityGroup(conv.S(securityGroup.GroupId))
		}, []string{"DependencyViolation", "ResourceInUse"}, time.Second, 1*time.Minute)
		if err != nil {
			err = fmt.Errorf("Failed to delete Security Group [%s]: %s", sgName, err.Error())
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(console.ColorFnResourceNegative("Error:"), err.Error())
			} else {
				return console.ExitWithError(err)
			}
		}
	}

	// delete ECS cluster
	if deleteECSCluster {
		console.RemovingResource("Deleting ECS Cluster", ecsClusterName, false)

		if err := c.awsClient.ECS().DeleteCluster(ecsClusterName); err != nil {
			//if awsErr, ok := err.(awserr.Error); ok {
			//	if awsErr.Code() == "ClusterContainsContainerInstancesException" {
			//	}
			//}
			//
			err = fmt.Errorf("Failed to delete ECS Cluster [%s]: %s", ecsServiceRoleName, err.Error())
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(console.ColorFnResourceNegative("Error:"), err.Error())
			} else {
				return console.ExitWithError(err)
			}
		}
	}

	// delete ECS service role
	if deleteECSServiceRole {
		console.RemovingResource("Deleting IAM Role", ecsServiceRoleName, false)

		if err := c.deleteECSServiceRole(ecsServiceRoleName); err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(console.ColorFnResourceNegative("Error:"), err.Error())
			} else {
				return console.ExitWithError(err)
			}
		}
	}

	return nil
}
