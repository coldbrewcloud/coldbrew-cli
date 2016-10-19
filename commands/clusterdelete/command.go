package clusterdelete

import (
	"fmt"
	"strings"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/config"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/core/clusters"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"github.com/d5/cc"
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

func (c *Command) Run(cfg *config.Config) error {
	c.awsClient = aws.NewClient(conv.S(c.globalFlags.AWSRegion), conv.S(c.globalFlags.AWSAccessKey), conv.S(c.globalFlags.AWSSecretKey))

	clusterName := strings.TrimSpace(conv.S(c.clusterNameArg))

	console.Println("Collecting resources to delete...")
	deleteECSCluster := false
	deleteECSServiceRole := false
	deleteInstanceProfile := false
	deleteInstanceSecurityGroups := false
	deleteLaunchConfiguration := false
	deleteAutoScalingGroup := false

	// ECS cluster
	ecsClusterName := clusters.DefaultECSClusterName(clusterName)
	ecsCluster, err := c.awsClient.ECS().RetrieveCluster(ecsClusterName)
	if err != nil {
		return c.exitWithError(fmt.Errorf("Failed to retrieve ECS Cluster [%s]: %s", ecsClusterName, err.Error()))
	}
	if ecsCluster != nil && conv.S(ecsCluster.Status) != "INACTIVE" {
		deleteECSCluster = true
		console.Println(" ", cc.BlackH("ECS Cluster"), cc.Red(ecsClusterName))
	}

	// ECS service role
	ecsServiceRoleName := clusters.DefaultECSServiceRoleName(clusterName)
	ecsServiceRole, err := c.awsClient.IAM().RetrieveRole(ecsServiceRoleName)
	if err != nil {
		return c.exitWithError(fmt.Errorf("Failed to retrieve IAM Role [%s]: %s", ecsServiceRoleName, err.Error()))
	}
	if ecsServiceRole != nil {
		deleteECSServiceRole = true
		console.Println(" ", cc.BlackH("ECS Service Rike"), cc.Red(ecsServiceRoleName))
	}

	// launch configuration
	lcName := clusters.DefaultLaunchConfigurationName(clusterName)
	launchConfiguration, err := c.awsClient.AutoScaling().RetrieveLaunchConfiguration(lcName)
	if err != nil {
		return c.exitWithError(fmt.Errorf("Failed to delete Launch Configuration [%s]: %s", lcName, err.Error()))
	}
	if launchConfiguration != nil {
		deleteLaunchConfiguration = true
		console.Println(" ", cc.BlackH("Launch Config"), cc.Red(lcName))
	}

	// auto scaling group
	asgName := clusters.DefaultAutoScalingGroupName(clusterName)
	autoScalingGroup, err := c.awsClient.AutoScaling().RetrieveAutoScalingGroup(asgName)
	if err != nil {
		return c.exitWithError(fmt.Errorf("Failed to retrieve Auto Scaling Group [%s]: %s", asgName, err.Error()))
	}
	if autoScalingGroup != nil && utils.IsBlank(conv.S(autoScalingGroup.Status)) {
		deleteAutoScalingGroup = true
		console.Println(" ", cc.BlackH("Auto Scaling Group"), cc.Red(asgName))
	}

	// instance profile
	instanceProfileName := clusters.DefaultInstanceProfileName(clusterName)
	instanceProfile, err := c.awsClient.IAM().RetrieveInstanceProfile(instanceProfileName)
	if err != nil {
		return c.exitWithError(fmt.Errorf("Failed to retrieve Instance Profile [%s]: %s", instanceProfileName, err.Error()))
	}
	if instanceProfile != nil {
		deleteInstanceProfile = true
		console.Println(" ", cc.BlackH("Instance Profile"), cc.Red(instanceProfileName))
	}

	// instance security group
	sgName := clusters.DefaultInstnaceSecurityGroupName(clusterName)
	securityGroup, err := c.awsClient.EC2().RetrieveSecurityGroupByName(sgName)
	if err != nil {
		return c.exitWithError(fmt.Errorf("Failed to retrieve Security Group [%s]: %s", sgName, err.Error()))
	}
	if securityGroup != nil {
		deleteInstanceSecurityGroups = true
		console.Println(" ", cc.BlackH("Instance Security Group"), cc.Red(sgName))
	}

	// confirmation
	if !conv.B(c.commandFlags.ForceDelete) && !console.AskConfirm("Do you want to delete these resources?") {
		return nil
	}

	// delete auto scaling group
	if deleteAutoScalingGroup {
		console.Printf("Scaling down Auto Scaling Group [%s]... %s\n", asgName, cc.BlackH("(this may take some time)"))

		if err := c.scaleDownAutoScalingGroup(autoScalingGroup); err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(cc.Red("Error: "), err.Error())
			} else {
				return c.exitWithError(err)
			}
		} else {
			console.Printf("Deleting Auto Scaling Group [%s]... %s\n", asgName, cc.BlackH("(this may take some time)"))
			if err := c.awsClient.AutoScaling().DeleteAutoScalingGroup(asgName, true); err != nil {
				if conv.B(c.commandFlags.ContinueOnError) {
					console.Errorln(cc.Red("Error: "), err.Error())
				} else {
					return c.exitWithError(err)
				}
			}
		}
	}

	// delete launch configuration
	if deleteLaunchConfiguration {
		console.Printf("Deleting Launch Configuration [%s]...\n", lcName)

		if err := c.awsClient.AutoScaling().DeleteLaunchConfiguration(lcName); err != nil {
			err = fmt.Errorf("Failed to delete Launch Configuration [%s]: %s", lcName, err.Error())
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(cc.Red("Error: "), err.Error())
			} else {
				return c.exitWithError(err)
			}
		}
	}

	// delete instance profile
	if deleteInstanceProfile {
		console.Printf("Deleting Instance Profile [%s]...\n", instanceProfileName)

		if err := c.deleteFullAccessInstanceProfile(instanceProfileName); err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(cc.Red("Error: "), err.Error())
			} else {
				return c.exitWithError(err)
			}
		}
	}

	// delete instance security groups
	if deleteInstanceSecurityGroups {
		console.Printf("Deleting Instance Security Group [%s]...\n", sgName)

		if err := c.awsClient.EC2().DeleteSecurityGroup(conv.S(securityGroup.GroupId)); err != nil {
			err = fmt.Errorf("Failed to delete Security Group [%s]: %s", sgName, err.Error())
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(cc.Red("Error: "), err.Error())
			} else {
				return c.exitWithError(err)
			}
		}
	}

	// delete ECS cluster
	if deleteECSCluster {
		console.Printf("Deleting ECS Cluster [%s]...\n", ecsClusterName)

		if err := c.awsClient.ECS().DeleteCluster(ecsClusterName); err != nil {
			err = fmt.Errorf("Failed to delete ECS Cluster [%s]: %s", ecsServiceRoleName, err.Error())
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(cc.Red("Error: "), err.Error())
			} else {
				return c.exitWithError(err)
			}
		}
	}

	// delete ECS service role
	if deleteECSServiceRole {
		console.Printf("Deleting ECS Service Role [%s]...\n", ecsServiceRoleName)

		if err := c.deleteECSServiceRole(ecsServiceRoleName); err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(cc.Red("Error: "), err.Error())
			} else {
				return c.exitWithError(err)
			}
		}
	}

	return nil
}

func (c *Command) exitWithError(err error) error {
	console.Errorln(cc.Red("Error:"), err.Error())
	return nil
}
