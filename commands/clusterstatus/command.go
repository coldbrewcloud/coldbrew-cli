package clusterstatus

import (
	"strings"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/core"
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
	console.Println("Cluster")
	console.Println(" ", cc.BlackH("Name"), cc.Green(clusterName))

	// AWS env
	console.Println("AWS")
	console.Println(" ", cc.BlackH("Region"), cc.Green(regionName))
	console.Println(" ", cc.BlackH("VPC"), cc.Green(vpcID))
	console.Println(" ", cc.BlackH("Subnets"), cc.Green(strings.Join(subnetIDs, " ")))

	// launch config and auto scaling group
	console.Println("Auto Scaling")
	showContainerInstanceDetails := false

	// launch configuration
	launchConfigName := core.DefaultLaunchConfigurationName(clusterName)
	launchConfig, err := c.awsClient.AutoScaling().RetrieveLaunchConfiguration(launchConfigName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve Launch Configuration [%s]: %s", launchConfigName, err.Error())
	}
	if launchConfig == nil {
		console.Println(" ", cc.BlackH("Launch Config"), cc.Green(launchConfigName), cc.Red("(not found)"))
	} else {
		console.Println(" ", cc.BlackH("Launch Config"), cc.Green(launchConfigName))
		showContainerInstanceDetails = true
	}

	// auto scaling group
	autoScalingGroupName := core.DefaultAutoScalingGroupName(clusterName)
	autoScalingGroup, err := c.awsClient.AutoScaling().RetrieveAutoScalingGroup(autoScalingGroupName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve Auto Scaling Group [%s]: %s", autoScalingGroupName, err.Error())
	}
	if autoScalingGroup == nil {
		console.Println(" ", cc.BlackH("Auto Scaling Group"), cc.Green(autoScalingGroupName), cc.Red("(not found)"))
	} else if utils.IsBlank(conv.S(autoScalingGroup.Status)) {
		console.Println(" ", cc.BlackH("Auto Scaling Group"), cc.Green(autoScalingGroupName))
		console.Println(" ", cc.BlackH("Instances"),
			cc.BlackH("Current"), cc.Green(conv.I64S(int64(len(autoScalingGroup.Instances)))),
			cc.BlackH("Desired"), cc.Green(conv.I64S(conv.I64(autoScalingGroup.DesiredCapacity))),
			cc.BlackH("Min"), cc.Green(conv.I64S(conv.I64(autoScalingGroup.MinSize))),
			cc.BlackH("Max"), cc.Green(conv.I64S(conv.I64(autoScalingGroup.MaxSize))))
	} else {
		console.Println(" ", cc.BlackH("Auto Scaling Group"), cc.Green(autoScalingGroupName), cc.Red("(deleting)"))
	}

	// ECS
	console.Println("ECS")
	showECSClusterDetails := false

	// ecs cluster
	ecsClusterName := core.DefaultECSClusterName(clusterName)
	ecsCluster, err := c.awsClient.ECS().RetrieveCluster(ecsClusterName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve ECS Cluster [%s]: %s", ecsClusterName, err.Error())
	}
	if ecsCluster == nil || conv.S(ecsCluster.Status) == "INACTIVE" {
		console.Println(" ", cc.BlackH("Cluster"), cc.Green(ecsClusterName), cc.Red("(not found)"))
	} else {
		console.Println(" ", cc.BlackH("Cluster"), cc.Green(ecsClusterName))
		showECSClusterDetails = true
	}

	// ecs service role
	ecsServiceRoleName := core.DefaultECSServiceRoleName(clusterName)
	ecsServiceRole, err := c.awsClient.IAM().RetrieveRole(ecsServiceRoleName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve IAM Role [%s]: %s", ecsServiceRoleName, err.Error())
	}
	if ecsServiceRole == nil {
		console.Println(" ", cc.BlackH("Service Role"), cc.Green(ecsServiceRoleName), cc.Red("(not found)"))
	} else {
		console.Println(" ", cc.BlackH("Service Role"), cc.Green(ecsServiceRoleName))
	}

	// ecs cluster details
	if showECSClusterDetails {
		console.Println(" ", cc.BlackH("Services"), cc.Green(conv.I64S(conv.I64(ecsCluster.ActiveServicesCount))))
		console.Println(" ", cc.BlackH("Tasks"),
			cc.BlackH("Running"), cc.Green(conv.I64S(conv.I64(ecsCluster.RunningTasksCount))),
			cc.BlackH("Pending"), cc.Green(conv.I64S(conv.I64(ecsCluster.PendingTasksCount))))
		console.Println(" ", cc.BlackH("Container Instances"), cc.Green(conv.I64S(conv.I64(ecsCluster.RegisteredContainerInstancesCount))))

	}

	// container instances
	if showContainerInstanceDetails {
		console.Println("Container Instances")

		instanceProfileARN := conv.S(launchConfig.IamInstanceProfile)
		console.Println(" ", cc.BlackH("Profile"), cc.Green(aws.GetIAMInstanceProfileNameFromARN(instanceProfileARN)))

		console.Println(" ", cc.BlackH("Type"), cc.Green(conv.S(launchConfig.InstanceType)))
		console.Println(" ", cc.BlackH("Image"), cc.Green(conv.S(launchConfig.ImageId)))
		console.Println(" ", cc.BlackH("Type"), cc.Green(conv.S(launchConfig.KeyName)))

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
		console.Println(" ", cc.BlackH("Security Groups"), cc.Green(strings.Join(securityGroupNames, " ")))
	}

	return nil
}
