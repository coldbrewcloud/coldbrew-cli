package clusterstatus

import (
	"strings"

	"fmt"

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

	cmd := ka.Command("cluster-status", "(cluster-status description goes here)")
	c.commandFlags = NewFlags(cmd)

	c.clusterNameArg = cmd.Arg("cluster-name", "Cluster name").Required().String()

	return cmd
}

func (c *Command) Run() error {
	c.awsClient = c.globalFlags.GetAWSClient()

	// AWS networking
	regionName, vpcID, subnetIDs, err := c.getAWSInfo()
	if err != nil {
		return console.ExitWithError(err)
	}

	// cluster name
	clusterName := strings.TrimSpace(conv.S(c.clusterNameArg))
	console.Info("Cluster")
	console.DetailWithResource("Name", clusterName)

	// AWS env
	console.Info("AWS")
	console.DetailWithResource("Region", regionName)
	console.DetailWithResource("VPC", vpcID)
	console.DetailWithResource("Subnets", strings.Join(subnetIDs, " "))

	// launch config and auto scaling group
	console.Info("Auto Scaling")
	showContainerInstanceDetails := false

	// launch configuration
	launchConfigName := core.DefaultLaunchConfigurationName(clusterName)
	launchConfig, err := c.awsClient.AutoScaling().RetrieveLaunchConfiguration(launchConfigName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve Launch Configuration [%s]: %s", launchConfigName, err.Error())
	}
	if launchConfig == nil {
		console.DetailWithResourceNote("EC2 Launch Configuration", launchConfigName, "(not found)", true)
	} else {
		console.DetailWithResource("EC2 Launch Configuration", launchConfigName)
		showContainerInstanceDetails = true
	}

	// auto scaling group
	autoScalingGroupName := core.DefaultAutoScalingGroupName(clusterName)
	autoScalingGroup, err := c.awsClient.AutoScaling().RetrieveAutoScalingGroup(autoScalingGroupName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve Auto Scaling Group [%s]: %s", autoScalingGroupName, err.Error())
	}
	if autoScalingGroup == nil {
		console.DetailWithResourceNote("EC2 Auto Scaling Group", autoScalingGroupName, "(not found)", true)
	} else if utils.IsBlank(conv.S(autoScalingGroup.Status)) {
		console.DetailWithResource("EC2 Auto Scaling Group", autoScalingGroupName)
		console.DetailWithResource("EC2 Instances (current/desired/min/max)",
			fmt.Sprintf("%d/%d/%d/%d",
				len(autoScalingGroup.Instances),
				conv.I64(autoScalingGroup.DesiredCapacity),
				conv.I64(autoScalingGroup.MinSize),
				conv.I64(autoScalingGroup.MaxSize)))
	} else {
		console.DetailWithResourceNote("EC2 Auto Scaling Group", autoScalingGroupName, "(deleting)", true)
	}

	// ECS
	console.Info("ECS")
	showECSClusterDetails := false

	// ecs cluster
	ecsClusterName := core.DefaultECSClusterName(clusterName)
	ecsCluster, err := c.awsClient.ECS().RetrieveCluster(ecsClusterName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve ECS Cluster [%s]: %s", ecsClusterName, err.Error())
	}
	if ecsCluster == nil || conv.S(ecsCluster.Status) == "INACTIVE" {
		console.DetailWithResourceNote("ECS Cluster", ecsClusterName, "(not found)", true)
	} else {
		console.DetailWithResource("ECS Cluster", ecsClusterName)
		showECSClusterDetails = true
	}

	// ecs service role
	ecsServiceRoleName := core.DefaultECSServiceRoleName(clusterName)
	ecsServiceRole, err := c.awsClient.IAM().RetrieveRole(ecsServiceRoleName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve IAM Role [%s]: %s", ecsServiceRoleName, err.Error())
	}
	if ecsServiceRole == nil {
		console.DetailWithResourceNote("IAM Role for ECS Services", ecsServiceRoleName, "(not found)", true)
	} else {
		console.DetailWithResource("IAM Role for ECS Services", ecsServiceRoleName)
	}

	// ecs cluster details
	if showECSClusterDetails {
		console.DetailWithResource("ECS Services", conv.I64S(conv.I64(ecsCluster.ActiveServicesCount)))
		console.DetailWithResource("ECS Tasks (running/pending)",
			fmt.Sprintf("%d/%d",
				conv.I64(ecsCluster.RunningTasksCount),
				conv.I64(ecsCluster.PendingTasksCount)))
		console.DetailWithResource("ECS Container Instances", conv.I64S(conv.I64(ecsCluster.RegisteredContainerInstancesCount)))

	}

	// container instances
	if showContainerInstanceDetails {
		console.Info("ECS Container Instances")

		instanceProfileARN := conv.S(launchConfig.IamInstanceProfile)
		console.DetailWithResource("IAM Instance Profile", aws.GetIAMInstanceProfileNameFromARN(instanceProfileARN))

		console.DetailWithResource("Instance Type", conv.S(launchConfig.InstanceType))
		console.DetailWithResource("Image ID", conv.S(launchConfig.ImageId))
		console.DetailWithResource("Key Pair", conv.S(launchConfig.KeyName))

		securityGroupIDs := []string{}
		for _, sg := range launchConfig.SecurityGroups {
			securityGroupIDs = append(securityGroupIDs, conv.S(sg))
		}
		securityGroups, err := c.awsClient.EC2().RetrieveSecurityGroups(securityGroupIDs)
		if err != nil {
			return console.ExitWithErrorString("Failed to retrieve Security Groups [%s]: %s", strings.Join(securityGroupIDs, ","), err.Error())
		}
		securityGroupNames := []string{}
		for _, sg := range securityGroups {
			securityGroupNames = append(securityGroupNames, conv.S(sg.GroupName))
		}
		console.DetailWithResource("Security Groups", strings.Join(securityGroupNames, " "))
	}

	return nil
}
