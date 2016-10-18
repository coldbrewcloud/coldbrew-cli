package clustercreate

import (
	"fmt"
	"strings"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/aws/ec2"
	"github.com/coldbrewcloud/coldbrew-cli/config"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"github.com/d5/cc"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	defaultInstanceType = "t2.micro"
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

	c.clusterNameArg = cmd.Arg("cluster-name", "Cluster name").String()

	return cmd
}

func (c *Command) Run(cfg *config.Config) error {
	c.awsClient = aws.NewClient(conv.S(c.globalFlags.AWSRegion), conv.S(c.globalFlags.AWSAccessKey), conv.S(c.globalFlags.AWSSecretKey))

	// cluster name
	clusterName := strings.TrimSpace(conv.S(c.clusterNameArg))
	if clusterName == "" {
		clusterName = console.AskQuestion("Enter cluster name", "cluster1")
	}
	console.Println("Cluster name:", cc.Green(clusterName))

	// AWS region
	console.Println("AWS region:", cc.Green(conv.S(c.globalFlags.AWSRegion)))

	// VPC ID
	vpcID := strings.TrimSpace(conv.S(c.commandFlags.VPC))
	if vpcID == "" {
		// find/use default VPC for the account
		defaultVPC, err := c.awsClient.EC2().RetrieveDefaultVPC()
		if err != nil {
			console.Errorf("Failed to find the default VPC: %s\n", err.Error())
			return nil
		} else if defaultVPC == nil {
			console.Errorf("This AWS account does not have a default VPC configured. Specify VPC ID using --vpc flag.\n")
			return nil
		}

		vpcID = conv.S(defaultVPC.VpcId)
		console.Println("VPC:", cc.Green(vpcID), cc.BlackH("(default VPC)"))
	} else {
		vpcID = strings.TrimSpace(conv.S(c.commandFlags.VPC))

		vpc, err := c.awsClient.EC2().RetrieveVPC(vpcID)
		if err != nil {
			console.Errorf("Failed to retrieve VPC [%s]: %s\n", cc.Green(vpcID), err.Error())
			return nil
		}
		if vpc == nil {
			console.Errorf("VPC [%s] does not exist.\n", cc.Red(vpcID))
			return nil
		}

		console.Println("VPC:", cc.Green(vpcID))
	}

	// Subnet IDs
	subnetIDs, err := c.awsClient.EC2().ListVPCSubnets(vpcID)
	if err != nil {
		console.Errorf("Failed to list subnets in VPC [%s]: %s\n", cc.Red(vpcID), err.Error())
	}
	console.Println("Subnets:", cc.Green(strings.Join(subnetIDs, ", ")))

	// instance profile
	instanceProfileARN := strings.TrimSpace(conv.S(c.commandFlags.InstanceProfile))
	if instanceProfileARN == "" {
		instanceProfileName := fmt.Sprintf("coldbrew_%s_instance_profile", clusterName)
		console.Printf("Creating IAM Instance Profile [%s] ...\n", cc.Green(instanceProfileName))

		existingInstanceProfile, err := c.awsClient.IAM().RetrieveInstanceProfile(instanceProfileName)
		if existingInstanceProfile != nil && err == nil {
			if conv.B(c.commandFlags.ReuseResources) {
				console.Printf("  > Existing IAM Instance Profile [%s] will be used.\n", cc.Green(instanceProfileName))
				instanceProfileARN = conv.S(existingInstanceProfile.Arn)
			} else {
				// role already exists
				if console.AskConfirm(fmt.Sprintf("  > Profile [%s] already exists.\n  > Do you want to use it?", cc.Green(instanceProfileName))) {
					instanceProfileARN = conv.S(existingInstanceProfile.Arn)
				} else {
					console.Errorln("  > You can delete the existing profile, or, specify different IAM Instance Profile using --instance-profile flag.")
					return nil
				}
			}
		} else {
			var err error
			instanceProfileARN, err = c.createFullAccessInstanceProfile(instanceProfileName)
			if err != nil {
				console.Errorf("Failed to create IAM Instance Profile [%s]: %s", cc.Red(instanceProfileName), err.Error())
				return nil
			}
		}

	} else {
		instanceProfileARN = strings.TrimSpace(conv.S(c.commandFlags.InstanceProfile))
	}
	console.Println("IAM Instance Profile:", cc.Green(instanceProfileARN))

	// container instance image ID
	imageID := c.getClusterImageID(conv.S(c.globalFlags.AWSRegion))
	if imageID == "" {
		console.Errorf("Failed to find image ID in AWS region [%s]\n", cc.Green(conv.S(c.globalFlags.AWSRegion)))
		return nil
	}
	console.Printf("ECS Container Instance Image ID: %s\n", cc.Green(imageID))

	// container instance type
	instanceType := strings.TrimSpace(conv.S(c.commandFlags.InstanceType))
	if instanceType == "" {
		instanceType = console.AskQuestion("Enter EC2 Container Instance type", defaultInstanceType)
	}
	console.Println("ECS Container Instance type:", cc.Green(instanceType))

	// key pair
	keyPairName := strings.TrimSpace(conv.S(c.commandFlags.KeyPairName))
	if keyPairName == "" {
		console.Println("You did not specify the key pair, which is needed to access container instances.")
		if !console.AskConfirm("  > Are you sure you don't want to create cluster without key pair?") {
			console.Errorln("  > Use --keypair flag to specify the key pair name.")
			return nil
		}
	} else {
		keyPairInfo, err := c.awsClient.EC2().RetrieveKeyPair(keyPairName)
		if err != nil {
			console.Errorf("Failed to retrieve key pair info [%s]: %s\n", cc.Red(keyPairName), err.Error())
			return nil
		}
		if keyPairInfo == nil {
			console.Errorf("Key pair [%s] was not found\n", cc.Red(keyPairName))
			return nil
		}
	}
	console.Println("EC2 Key Pair name:", cc.Green(keyPairName))

	// container instances security group
	sgName := fmt.Sprintf("coldbrew_%s_sg", clusterName)
	sgID := ""
	console.Printf("Creating EC2 Security Group [%s] ...\n", cc.Green(sgName))
	existingSG, err := c.awsClient.EC2().RetrieveSecurityGroupByName(sgName)
	if existingSG != nil && err == nil {
		// SG already exists
		if conv.B(c.commandFlags.ReuseResources) {
			console.Printf("  > Existing EC2 Security Group [%s] will be used.\n", cc.Green(sgName))
			sgID = conv.S(existingSG.GroupId)
		} else {
			if console.AskConfirm(fmt.Sprintf("  > Security Group [%s] already exists.\n  > Do you want to use it?", cc.Green(sgName))) {
				sgID = conv.S(existingSG.GroupId)
			} else {
				console.Errorln("  > You can delete the existing security group and try again.")
				return nil
			}
		}
	} else {
		sgID, err = c.awsClient.EC2().CreateSecurityGroup(sgName, sgName, vpcID)
		if err != nil {
			console.Errorf("Failed to create EC2 Security Group [%s] for container instances: %", cc.Red(sgName), err.Error())
			return nil
		}
		if err := c.awsClient.EC2().AddInboundToSecurityGroup(sgID, ec2.SecurityGroupProtocolTCP, 22, 22, "0.0.0.0/0"); err != nil {
			console.Errorf("Failed to add inbound rule for SSH to EC2 Security Group [%s]: %s", cc.Red(sgName), err.Error())
			return nil
		}
	}

	ecsClusterName := fmt.Sprintf("coldbrew_%s", clusterName)

	// user data
	instanceUserData := c.getInstanceUserData(ecsClusterName)

	// create launch configuration
	lcName := fmt.Sprintf("coldbrew_%s_lc", clusterName)
	console.Printf("Creating EC2 Launch Configuration [%s] ...\n", cc.Green(lcName))
	existingLC, err := c.awsClient.AutoScaling().RetrieveLaunchConfiguration(lcName)
	if existingLC != nil && err == nil {
		// launch configuration already exists
		if conv.B(c.commandFlags.ReuseResources) {
			console.Printf("  > Existing EC2 Launch Configuration [%s] will be used.\n", cc.Green(lcName))
		} else {
			if !console.AskConfirm(fmt.Sprintf("  > EC2 Launch Configuration [%s] already exists.\n  > Do you want to use it?", cc.Green(lcName))) {
				console.Errorln("  > You can delete the existing EC2 Launch Configuration and try again.")
				return nil
			}
		}
	} else {
		err := c.awsClient.AutoScaling().CreateLaunchConfiguration(lcName, instanceType, imageID, []string{sgID}, keyPairName, instanceProfileARN, instanceUserData)
		if err != nil {
			console.Errorf("Failed to create EC2 Launch Configuration [%s]: %s\n", cc.Red(lcName), err.Error())
			return nil
		}
	}

	// initial instance count
	initialCapacity := conv.U16(c.commandFlags.InitialCapacity)

	// create auto scaling group
	asgName := fmt.Sprintf("coldbrew_%s_asg", clusterName)
	console.Printf("Creating EC2 Auto Scaling Group [%s] ...\n", cc.Green(asgName))
	existingASG, err := c.awsClient.AutoScaling().RetrieveAutoScalingGroup(asgName)
	if existingASG != nil && err == nil {
		// auto scaling group already exists
		if conv.B(c.commandFlags.ReuseResources) {
			console.Printf("  > Existing EC2 Auto Scaling Group [%s] will be used.\n", cc.Green(asgName))
		} else {
			if !console.AskConfirm(fmt.Sprintf("  > EC2 Auto Scaling Group [%s] already exists.\n  > Do you want to use it?", cc.Green(asgName))) {
				console.Errorln("  > You can delete the existing EC2 Auto Scaling Group and try again.")
				return nil
			}
		}
	} else {
		err = c.awsClient.AutoScaling().CreateAutoScalingGroup(asgName, lcName, subnetIDs, 0, initialCapacity, initialCapacity)
		if err != nil {
			console.Errorf("Failed to create EC2 Auto Scaling Group [%s]: %s\n", cc.Red(asgName), err.Error())
			return nil
		}
	}

	// create ECS cluster
	console.Printf("Creating ECS cluster [%s] ...\n", cc.Green(ecsClusterName))
	ecsCluster, err := c.awsClient.ECS().RetrieveCluster(ecsClusterName)
	if ecsCluster != nil && err == nil {
		console.Printf("  > Existing ECS Cluster [%s] will be used.\n", cc.Green(ecsClusterName))
	} else {
		if _, err := c.awsClient.ECS().CreateCluster(ecsClusterName); err != nil {
			console.Errorf("Failed to create ECS Cluster [$s]: %s\n", cc.Red(ecsClusterName), err.Error())
			return nil
		}
	}

	// create ECS service role (IAM)
	ecsServiceRoleName := fmt.Sprintf("coldbrew_%s_ecs_service_role", clusterName)
	console.Printf("Creating IAM Role [%s] for ECS service role ...\n", cc.Green(ecsServiceRoleName))
	existingRole, err := c.awsClient.IAM().RetrieveRole(ecsServiceRoleName)
	if existingRole != nil && err == nil {
		console.Printf("  > Existing IAM Role [%s] will be used.\n", cc.Green(ecsServiceRoleName))
	} else {
		_, err := c.createECSServiceRole(ecsServiceRoleName)
		if err != nil {
			console.Errorf("Failed to create IAM role [%s]: %s\n", cc.Red(ecsServiceRoleName), err.Error())
			return nil
		}
	}

	console.Printf("Cluster [%s] was created successfully. It can take some time until all new container instances full become available in the cluster.\n", cc.Green(clusterName))
	return nil
}
