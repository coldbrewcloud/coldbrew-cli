package clustercreate

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/aws/ec2"
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

	cmd := ka.Command("cluster-create", "(cluster-create description goes here)")
	c.commandFlags = NewFlags(cmd)

	c.clusterNameArg = cmd.Arg("cluster-name", "Cluster name").Required().String()

	return cmd
}

func (c *Command) Run() error {
	c.awsClient = c.globalFlags.GetAWSClient()

	// AWS networking
	_, vpcID, subnetIDs, err := c.getAWSInfo()
	if err != nil {
		return c.exitWithError(err)
	}

	clusterName := strings.TrimSpace(conv.S(c.clusterNameArg))

	console.Println("Identifying resources to create...")
	createECSCluster := false
	createECSServiceRole := false
	createInstanceProfile := false
	createInstanceSecurityGroup := false
	createLaunchConfiguration := false
	createAutoScalingGroup := false

	// ECS cluster
	ecsClusterName := clusters.DefaultECSClusterName(clusterName)
	ecsCluster, err := c.awsClient.ECS().RetrieveCluster(ecsClusterName)
	if err != nil {
		return c.exitWithError(fmt.Errorf("Failed to retrieve ECS Cluster [%s]: %s", ecsClusterName, err.Error()))
	}
	if ecsCluster == nil || conv.S(ecsCluster.Status) == "INACTIVE" {
		createECSCluster = true
		console.Println(" ", cc.BlackH("ECS Cluster"), cc.Green(ecsClusterName))
	}

	// ECS service role
	ecsServiceRoleName := clusters.DefaultECSServiceRoleName(clusterName)
	ecsServiceRole, err := c.awsClient.IAM().RetrieveRole(ecsServiceRoleName)
	if err != nil {
		return c.exitWithError(fmt.Errorf("Failed to retrieve IAM Role [%s]: %s", ecsServiceRoleName, err.Error()))
	}
	if ecsServiceRole == nil {
		createECSServiceRole = true
		console.Println(" ", cc.BlackH("ECS Service Rike"), cc.Green(ecsServiceRoleName))
	}

	// launch configuration
	launchConfigName := clusters.DefaultLaunchConfigurationName(clusterName)
	launchConfig, err := c.awsClient.AutoScaling().RetrieveLaunchConfiguration(launchConfigName)
	if err != nil {
		return c.exitWithError(fmt.Errorf("Failed to delete Launch Configuration [%s]: %s", launchConfigName, err.Error()))
	}
	if launchConfig == nil {
		createLaunchConfiguration = true
		console.Println(" ", cc.BlackH("Launch Config"), cc.Green(launchConfigName))
	}

	// auto scaling group
	autoScalingGroupName := clusters.DefaultAutoScalingGroupName(clusterName)
	autoScalingGroup, err := c.awsClient.AutoScaling().RetrieveAutoScalingGroup(autoScalingGroupName)
	if err != nil {
		return c.exitWithError(fmt.Errorf("Failed to retrieve Auto Scaling Group [%s]: %s", autoScalingGroupName, err.Error()))
	}
	if autoScalingGroup == nil || !utils.IsBlank(conv.S(autoScalingGroup.Status)) {
		createAutoScalingGroup = true
		console.Println(" ", cc.BlackH("Auto Scaling Group"), cc.Green(autoScalingGroupName))
	}

	// instance profile
	instanceProfileName := clusters.DefaultInstanceProfileName(clusterName)
	if !utils.IsBlank(conv.S(c.commandFlags.InstanceProfile)) {
		instanceProfileName = conv.S(c.commandFlags.InstanceProfile)
	} else {
		instanceProfile, err := c.awsClient.IAM().RetrieveInstanceProfile(instanceProfileName)
		if err != nil {
			return c.exitWithError(fmt.Errorf("Failed to retrieve Instance Profile [%s]: %s", instanceProfileName, err.Error()))
		}
		if instanceProfile == nil {
			createInstanceProfile = true
			console.Println(" ", cc.BlackH("Instance Profile"), cc.Green(instanceProfileName))
		}
	}

	// instance security group
	instanceSecurityGroupName := clusters.DefaultInstanceSecurityGroupName(clusterName)
	instanceSecurityGroup, err := c.awsClient.EC2().RetrieveSecurityGroupByName(instanceSecurityGroupName)
	instanceSecurityGroupID := ""
	if err != nil {
		return c.exitWithError(fmt.Errorf("Failed to retrieve Security Group [%s]: %s", instanceSecurityGroupName, err.Error()))
	}
	if instanceSecurityGroup == nil {
		createInstanceSecurityGroup = true
		console.Println(" ", cc.BlackH("Instance Security Group"), cc.Green(instanceSecurityGroupName))
	} else {
		instanceSecurityGroupID = conv.S(instanceSecurityGroup.GroupId)
	}

	if !createECSServiceRole && !createECSCluster && !createLaunchConfiguration && !createAutoScalingGroup &&
		!createInstanceProfile && !createInstanceSecurityGroup {
		console.Println("Looks like everything is already up and running!")
		return nil
	}

	// confirmation
	if !conv.B(c.commandFlags.ForceCreate) && !console.AskConfirm("Do you want to create these resources?") {
		return nil
	}

	// create instance profile
	if createInstanceProfile {
		console.Printf("Creating Instance Profile [%s]...\n", cc.Green(instanceProfileName))

		if _, err = c.createDefaultInstanceProfile(instanceProfileName); err != nil {
			return c.exitWithError(fmt.Errorf("Failed to create Instance Profile [%s]: %s", instanceProfileName, err.Error()))
		}
	}

	// create instance security group
	if createInstanceSecurityGroup {
		console.Printf("Creating Security Group [%s]...\n", cc.Green(instanceSecurityGroupName))

		var err error
		instanceSecurityGroupID, err = c.awsClient.EC2().CreateSecurityGroup(instanceSecurityGroupName, instanceSecurityGroupName, vpcID)
		if err != nil {
			return c.exitWithError(fmt.Errorf("Failed to create EC2 Security Group [%s] for container instances: %s", instanceSecurityGroupName, err.Error()))
		}
		if err := c.awsClient.EC2().AddInboundToSecurityGroup(instanceSecurityGroupID, ec2.SecurityGroupProtocolTCP, 22, 22, "0.0.0.0/0"); err != nil {
			return c.exitWithError(fmt.Errorf("Failed to add SSH inbound rule to Security Group [%s]: %s", instanceSecurityGroupName, err.Error()))
		}
	}

	// create launch configuration
	if createLaunchConfiguration {
		console.Printf("Creating Launch Configuration [%s]...\n", cc.Green(launchConfigName))

		// key pair
		keyPairName := strings.TrimSpace(conv.S(c.commandFlags.KeyPairName))
		keyPairInfo, err := c.awsClient.EC2().RetrieveKeyPair(keyPairName)
		if err != nil {
			return c.exitWithError(fmt.Errorf("Failed to retrieve key pair info [%s]: %s", keyPairName, err.Error()))
		}
		if keyPairInfo == nil {
			return c.exitWithError(fmt.Errorf("Key pair [%s] was not found\n", keyPairName))
		}

		// container instance type
		instanceType := strings.TrimSpace(conv.S(c.commandFlags.InstanceType))
		if instanceType == "" {
			instanceType = console.AskQuestion("Enter instance type", clusters.DefaultContainerInstanceType())
		}

		// container instance image ID
		imageID := c.getClusterImageID(conv.S(c.globalFlags.AWSRegion))
		if imageID == "" {
			return c.exitWithError(errors.New("No defatul instance image found"))
		}

		instanceUserData := c.getInstanceUserData(ecsClusterName)

		// NOTE: sometimes resources created (e.g. InstanceProfile) do not become available immediately.
		// So here we retry up to 10 times just to be safe.
		var lastErr error
		for i := 0; i < 10; i++ {
			err := c.awsClient.AutoScaling().CreateLaunchConfiguration(launchConfigName, instanceType, imageID, []string{instanceSecurityGroupID}, keyPairName, instanceProfileName, instanceUserData)
			if err != nil {
				lastErr = err
			} else {
				lastErr = nil
				break
			}

			time.Sleep(1 * time.Second)
		}
		if lastErr != nil {
			return c.exitWithError(fmt.Errorf("Failed to create EC2 Launch Configuration [%s]: %s", launchConfigName, lastErr.Error()))
		}
	}

	// create auto scaling group
	if createAutoScalingGroup {
		console.Printf("Creating Auto Scaling Group [%s]...\n", cc.Green(autoScalingGroupName))

		// if existing auto scaling group is currently pending delete, wait a bit so it gets fully deleted
		if autoScalingGroup != nil && !utils.IsBlank(conv.S(autoScalingGroup.Status)) {
			if err := c.waitAutoScalingGroupDeletion(autoScalingGroupName); err != nil {
				return c.exitWithError(err)
			}
		}

		initialCapacity := conv.U16(c.commandFlags.InitialCapacity)

		err = c.awsClient.AutoScaling().CreateAutoScalingGroup(autoScalingGroupName, launchConfigName, subnetIDs, 0, initialCapacity, initialCapacity)
		if err != nil {
			return c.exitWithError(fmt.Errorf("Failed to create EC2 Auto Scaling Group [%s]: %s", autoScalingGroupName, err.Error()))
		}
	}

	// create ECS cluster
	if createECSCluster {
		console.Printf("Creating ECS Cluster [%s]...\n", cc.Green(ecsClusterName))

		if _, err := c.awsClient.ECS().CreateCluster(ecsClusterName); err != nil {
			return c.exitWithError(fmt.Errorf("Failed to create ECS Cluster [%s]: %s", ecsClusterName, err.Error()))
		}
	}

	// create ECS service role
	if createECSServiceRole {
		console.Printf("Creating ECS Service Role [%s]...\n", cc.Green(ecsServiceRoleName))

		if _, err := c.createECSServiceRole(ecsServiceRoleName); err != nil {
			return c.exitWithError(fmt.Errorf("Failed to create IAM role [%s]: %s", ecsServiceRoleName, err.Error()))
		}
	}

	return nil
}

func (c *Command) exitWithError(err error) error {
	console.Errorln(cc.Red("Error:"), err.Error())
	return nil
}
