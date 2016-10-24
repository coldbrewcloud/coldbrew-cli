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

	cmd := ka.Command("cluster-status",
		"See: "+console.ColorFnHelpLink("https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-status"))
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
	if !core.ClusterNameRE.MatchString(clusterName) {
		return console.ExitWithError(core.NewErrorExtraInfo(
			fmt.Errorf("Invalid cluster name [%s]", clusterName), "https://github.com/coldbrewcloud/coldbrew-cli/wiki/Configuration-File#cluster"))
	}

	console.Info("Cluster")
	console.DetailWithResource("Name", clusterName)

	// AWS env
	console.Info("AWS")
	console.DetailWithResource("Region", regionName)
	console.DetailWithResource("VPC", vpcID)
	console.DetailWithResource("Subnets", strings.Join(subnetIDs, " "))

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

	// launch config and auto scaling group
	console.Info("Auto Scaling")

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

		instanceProfileARN := conv.S(launchConfig.IamInstanceProfile)
		console.DetailWithResource("  IAM Instance Profile", aws.GetIAMInstanceProfileNameFromARN(instanceProfileARN))

		console.DetailWithResource("  Instance Type", conv.S(launchConfig.InstanceType))
		console.DetailWithResource("  Image ID", conv.S(launchConfig.ImageId))
		console.DetailWithResource("  Key Pair", conv.S(launchConfig.KeyName))

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
		console.DetailWithResource("  Security Groups", strings.Join(securityGroupNames, " "))
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
		console.DetailWithResource("  Instances (current/desired/min/max)",
			fmt.Sprintf("%d/%d/%d/%d",
				len(autoScalingGroup.Instances),
				conv.I64(autoScalingGroup.DesiredCapacity),
				conv.I64(autoScalingGroup.MinSize),
				conv.I64(autoScalingGroup.MaxSize)))
	} else {
		console.DetailWithResourceNote("EC2 Auto Scaling Group", autoScalingGroupName, "(deleting)", true)
	}

	// ECS Container Instances
	if ecsCluster != nil && !conv.B(c.commandFlags.ExcludeContainerInstanceInfos) {
		containerInstanceARNs, err := c.awsClient.ECS().ListContainerInstanceARNs(ecsClusterName)
		if err != nil {
			return console.ExitWithErrorString("Failed to list ECS Container Instances: %s", err.Error())
		}
		containerInstances, err := c.awsClient.ECS().RetrieveContainerInstances(ecsClusterName, containerInstanceARNs)
		if err != nil {
			return console.ExitWithErrorString("Failed to retrieve ECS Container Instances: %s", err.Error())
		}

		// retrieve EC2 Instance info
		ec2InstanceIDs := []string{}
		for _, ci := range containerInstances {
			ec2InstanceIDs = append(ec2InstanceIDs, conv.S(ci.Ec2InstanceId))
		}
		ec2Instances, err := c.awsClient.EC2().RetrieveInstances(ec2InstanceIDs)
		if err != nil {
			return console.ExitWithErrorString("Failed to retrieve EC2 Instances: %s", err.Error())
		}

		for _, ci := range containerInstances {
			console.Info("ECS Container Instance")

			console.DetailWithResource("ID", aws.GetECSContainerInstanceIDFromARN(conv.S(ci.ContainerInstanceArn)))

			if conv.B(ci.AgentConnected) {
				console.DetailWithResource("Status", conv.S(ci.Status))
			} else {
				console.DetailWithResourceNote("Status", conv.S(ci.Status), "(agent not connected)", true)
			}

			console.DetailWithResource("Tasks (running/pending)", fmt.Sprintf("%d/%d",
				conv.I64(ci.RunningTasksCount),
				conv.I64(ci.PendingTasksCount)))

			var registeredCPU, registeredMemory, remainingCPU, remainingMemory int64
			for _, rr := range ci.RegisteredResources {
				switch strings.ToLower(conv.S(rr.Name)) {
				case "cpu":
					registeredCPU = conv.I64(rr.IntegerValue)
				case "memory":
					registeredMemory = conv.I64(rr.IntegerValue)
				}
			}
			for _, rr := range ci.RemainingResources {
				switch strings.ToLower(conv.S(rr.Name)) {
				case "cpu":
					remainingCPU = conv.I64(rr.IntegerValue)
				case "memory":
					remainingMemory = conv.I64(rr.IntegerValue)
				}
			}

			console.DetailWithResource("CPU (remaining/registered)", fmt.Sprintf("%.2f/%.2f",
				float64(remainingCPU)/1024.0, float64(registeredCPU)/1024.0))
			console.DetailWithResource("Memory (remaining/registered)", fmt.Sprintf("%dM/%dM,",
				remainingMemory, registeredMemory))

			console.DetailWithResource("EC2 Instance ID", conv.S(ci.Ec2InstanceId))
			for _, ei := range ec2Instances {
				if conv.S(ei.InstanceId) == conv.S(ci.Ec2InstanceId) {
					if !utils.IsBlank(conv.S(ei.PrivateIpAddress)) {
						console.DetailWithResource("  Private IP", conv.S(ei.PrivateIpAddress))
					}
					if !utils.IsBlank(conv.S(ei.PublicIpAddress)) {
						console.DetailWithResource("  Public IP", conv.S(ei.PublicIpAddress))
					}
					break
				}
			}
		}
	}

	return nil
}
