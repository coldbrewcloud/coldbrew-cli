package clustercreate

import (
	"fmt"
	"strings"
	"time"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/aws/ec2"
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

	cmd := ka.Command("cluster-create",
		"See: "+console.ColorFnHelpLink("https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-cluster-create"))
	c.commandFlags = NewFlags(cmd)

	c.clusterNameArg = cmd.Arg("cluster-name", "Cluster name").Required().String()

	return cmd
}

func (c *Command) Run() error {
	c.awsClient = c.globalFlags.GetAWSClient()

	clusterName := strings.TrimSpace(conv.S(c.clusterNameArg))
	if !core.ClusterNameRE.MatchString(clusterName) {
		return console.ExitWithError(core.NewErrorExtraInfo(
			fmt.Errorf("Invalid cluster name [%s]", clusterName), "https://github.com/coldbrewcloud/coldbrew-cli/wiki/Configuration-File#cluster"))
	}

	// AWS networking
	_, vpcID, subnetIDs, err := c.getAWSInfo()
	if err != nil {
		return console.ExitWithError(err)
	}

	// keypair
	keyPairName := ""
	if !conv.B(c.commandFlags.NoKeyPair) {
		strings.TrimSpace(conv.S(c.commandFlags.KeyPairName))
		if utils.IsBlank(keyPairName) {
			keyPairs, err := c.awsClient.EC2().ListKeyPairs()
			if err != nil {
				return console.ExitWithErrorString("Failed to list EC2 Key Pairs: %s", err.Error())
			}

			defaultKeyPairName := ""
			if len(keyPairs) > 0 {
				defaultKeyPairName = conv.S(keyPairs[0].KeyName)

				if conv.B(c.commandFlags.ForceCreate) {
					keyPairName = defaultKeyPairName
				} else {
					keyPairName = console.AskQuestionWithNote(
						"Enter EC2 Key Pair name",
						defaultKeyPairName,
						"EC2 Key Pair name is required to create a new cluster.")
				}
			} else {
				if !conv.B(c.commandFlags.ForceCreate) && !console.AskConfirmWithNote(
					"Do you still want to create cluster without EC2 Key Pair?",
					false,
					"Could not find any EC2 Key Pairs available to use.") {
					console.Info("Please create an EC2 Key Pair and try again.")
					return nil
				}
			}

		}

		// check if key pair exists
		if !utils.IsBlank(keyPairName) {
			keyPairInfo, err := c.awsClient.EC2().RetrieveKeyPair(keyPairName)
			if err != nil {
				return console.ExitWithErrorString("Failed to retrieve EC2 Key Pair [%s]: %s", keyPairName, err.Error())
			}
			if keyPairInfo == nil {
				return console.ExitWithErrorString("EC2 Key Pair [%s] was not found.", keyPairName)
			}
		}
	}

	console.Info("Determining AWS resources to create...")
	createECSCluster := false
	createECSServiceRole := false
	createInstanceProfile := false
	createInstanceSecurityGroup := false
	createLaunchConfiguration := false
	createAutoScalingGroup := false

	// ECS cluster
	ecsClusterName := core.DefaultECSClusterName(clusterName)
	ecsCluster, err := c.awsClient.ECS().RetrieveCluster(ecsClusterName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve ECS Cluster [%s]: %s", ecsClusterName, err.Error())
	}
	if ecsCluster == nil || conv.S(ecsCluster.Status) == "INACTIVE" {
		createECSCluster = true
		console.DetailWithResource("ECS Cluster", ecsClusterName)
	}

	// ECS service role
	ecsServiceRoleName := core.DefaultECSServiceRoleName(clusterName)
	ecsServiceRole, err := c.awsClient.IAM().RetrieveRole(ecsServiceRoleName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve IAM Role [%s]: %s", ecsServiceRoleName, err.Error())
	}
	if ecsServiceRole == nil {
		createECSServiceRole = true
		console.DetailWithResource("IAM Role for ECS Services", ecsServiceRoleName)
	}

	// launch configuration
	launchConfigName := core.DefaultLaunchConfigurationName(clusterName)
	launchConfig, err := c.awsClient.AutoScaling().RetrieveLaunchConfiguration(launchConfigName)
	if err != nil {
		return console.ExitWithErrorString("Failed to delete Launch Configuration [%s]: %s", launchConfigName, err.Error())
	}
	if launchConfig == nil {
		createLaunchConfiguration = true
		console.DetailWithResource("EC2 Launch Configuration for ECS Container Instances", launchConfigName)
	}

	// auto scaling group
	autoScalingGroupName := core.DefaultAutoScalingGroupName(clusterName)
	autoScalingGroup, err := c.awsClient.AutoScaling().RetrieveAutoScalingGroup(autoScalingGroupName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve Auto Scaling Group [%s]: %s", autoScalingGroupName, err.Error())
	}
	if autoScalingGroup == nil || !utils.IsBlank(conv.S(autoScalingGroup.Status)) {
		createAutoScalingGroup = true
		console.DetailWithResource("EC2 Auto Scaling Group for ECS Container Instances", autoScalingGroupName)
	}

	// instance profile
	instanceProfileName := core.DefaultInstanceProfileName(clusterName)
	if !utils.IsBlank(conv.S(c.commandFlags.InstanceProfile)) {
		instanceProfileName = conv.S(c.commandFlags.InstanceProfile)
	} else {
		instanceProfile, err := c.awsClient.IAM().RetrieveInstanceProfile(instanceProfileName)
		if err != nil {
			return console.ExitWithErrorString("Failed to retrieve Instance Profile [%s]: %s", instanceProfileName, err.Error())
		}
		if instanceProfile == nil {
			createInstanceProfile = true
			console.DetailWithResource("IAM Instance Profile for ECS Container Instances", instanceProfileName)
		}
	}

	// instance security group
	instanceSecurityGroupName := core.DefaultInstanceSecurityGroupName(clusterName)
	instanceSecurityGroup, err := c.awsClient.EC2().RetrieveSecurityGroupByName(instanceSecurityGroupName)
	instanceSecurityGroupID := ""
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve Security Group [%s]: %s", instanceSecurityGroupName, err.Error())
	}
	if instanceSecurityGroup == nil {
		createInstanceSecurityGroup = true
		console.DetailWithResource("EC2 Security Group for ECS Container Instances", instanceSecurityGroupName)
	} else {
		instanceSecurityGroupID = conv.S(instanceSecurityGroup.GroupId)
	}

	if !createECSServiceRole && !createECSCluster && !createLaunchConfiguration && !createAutoScalingGroup &&
		!createInstanceProfile && !createInstanceSecurityGroup {
		console.Info("Looks like everything is already up and running!")
		return nil
	}

	console.Blank()

	// confirmation
	if !conv.B(c.commandFlags.ForceCreate) && !console.AskConfirm("Do you want to create these resources?", false) {
		return nil
	}

	console.Blank()

	// create instance profile
	if createInstanceProfile {
		console.AddingResource("Creating IAM Instance Profile", instanceProfileName, false)

		if _, err = c.createDefaultInstanceProfile(instanceProfileName); err != nil {
			return console.ExitWithErrorString("Failed to create Instance Profile [%s]: %s", instanceProfileName, err.Error())
		}
	}

	// create instance security group
	if createInstanceSecurityGroup {
		console.AddingResource("Creating EC2 Security Group", instanceSecurityGroupName, false)

		var err error
		instanceSecurityGroupID, err = c.awsClient.EC2().CreateSecurityGroup(instanceSecurityGroupName, instanceSecurityGroupName, vpcID)
		if err != nil {
			return console.ExitWithErrorString("Failed to create EC2 Security Group [%s] for container instances: %s", instanceSecurityGroupName, err.Error())
		}

		err = utils.RetryOnAWSErrorCode(func() error {
			return c.awsClient.EC2().CreateTags(instanceSecurityGroupID, core.DefaultTagsForAWSResources(instanceSecurityGroupName))
		}, []string{"InvalidGroup.NotFound"}, time.Second, 5*time.Minute)
		if err != nil {
			return console.ExitWithErrorString("Failed to tag EC2 Security Group [%s]: %s", instanceSecurityGroupName, err.Error())
		}

		console.UpdatingResource(fmt.Sprintf("Adding inbound rule [%s:%d:%s] to EC2 Security Group",
			ec2.SecurityGroupProtocolTCP, 22, "0.0.0.0/0"),
			instanceSecurityGroupName, false)
		if err := c.awsClient.EC2().AddInboundToSecurityGroup(instanceSecurityGroupID, ec2.SecurityGroupProtocolTCP, 22, 22, "0.0.0.0/0"); err != nil {
			return console.ExitWithErrorString("Failed to add SSH inbound rule to Security Group [%s]: %s", instanceSecurityGroupName, err.Error())
		}
	}

	// create launch configuration
	if createLaunchConfiguration {
		console.AddingResource("Creating EC2 Launch Configuration", launchConfigName, true)

		// container instance type
		instanceType := strings.TrimSpace(conv.S(c.commandFlags.InstanceType))
		if instanceType == "" {
			defaultInstanceType := core.DefaultContainerInstanceType()

			if conv.B(c.commandFlags.ForceCreate) {
				instanceType = defaultInstanceType
			} else {
				instanceType = console.AskQuestionWithNote(
					"Enter instance type",
					defaultInstanceType,
					"EC2 Instance Type for ECS Container Instances")
			}
		}

		// container instance image ID
		imageID := c.retrieveDefaultECSContainerInstancesImageID(conv.S(c.globalFlags.AWSRegion))
		if imageID == "" {
			return console.ExitWithErrorString("No defatul instance image found")
		}

		instanceUserData := c.getInstanceUserData(ecsClusterName)

		// NOTE: sometimes resources created (e.g. InstanceProfile) do not become available immediately.
		err = utils.Retry(func() (bool, error) {
			err := c.awsClient.AutoScaling().CreateLaunchConfiguration(launchConfigName, instanceType, imageID, []string{instanceSecurityGroupID}, keyPairName, instanceProfileName, instanceUserData)
			if err == nil {
				return false, nil
			}
			return true, err
		}, time.Second, 5*time.Minute)
		if err != nil {
			return console.ExitWithErrorString("Failed to create EC2 Launch Configuration [%s]: %s", launchConfigName, err.Error())
		}
	}

	// create auto scaling group
	if createAutoScalingGroup {
		console.AddingResource("Creating EC2 Auto Scaling Group", autoScalingGroupName, true)

		// if existing auto scaling group is currently pending delete, wait a bit so it gets fully deleted
		if autoScalingGroup != nil && !utils.IsBlank(conv.S(autoScalingGroup.Status)) {
			if err := c.waitAutoScalingGroupDeletion(autoScalingGroupName); err != nil {
				return console.ExitWithError(err)
			}
		}

		initialCapacity := conv.U16(c.commandFlags.InitialCapacity)

		err = c.awsClient.AutoScaling().CreateAutoScalingGroup(autoScalingGroupName, launchConfigName, subnetIDs, 0, initialCapacity, initialCapacity)
		if err != nil {
			return console.ExitWithErrorString("Failed to create EC2 Auto Scaling Group [%s]: %s", autoScalingGroupName, err.Error())
		}

		if err := c.awsClient.AutoScaling().AddTagsToAutoScalingGroup(autoScalingGroupName, core.DefaultTagsForAWSResources(autoScalingGroupName), true); err != nil {
			return console.ExitWithErrorString("Failed to tag EC2 Auto Scaling Group [%s]: %s", autoScalingGroupName, err.Error())
		}
	}

	// create ECS cluster
	if createECSCluster {
		console.AddingResource("Creating ECS Cluster", ecsClusterName, false)

		if _, err := c.awsClient.ECS().CreateCluster(ecsClusterName); err != nil {
			return console.ExitWithErrorString("Failed to create ECS Cluster [%s]: %s", ecsClusterName, err.Error())
		}
	}

	// create ECS service role
	if createECSServiceRole {
		console.AddingResource("Creating IAM Role", ecsServiceRoleName, false)

		if _, err := c.createECSServiceRole(ecsServiceRoleName); err != nil {
			return console.ExitWithErrorString("Failed to create IAM role [%s]: %s", ecsServiceRoleName, err.Error())
		}
	}

	return nil
}
