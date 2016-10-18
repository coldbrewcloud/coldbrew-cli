package clustercreate

import (
	"errors"
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

	c.clusterNameArg = cmd.Arg("cluster-name", "Cluster name").Required().String()

	return cmd
}

func (c *Command) Run(cfg *config.Config) error {
	c.awsClient = aws.NewClient(conv.S(c.globalFlags.AWSRegion), conv.S(c.globalFlags.AWSAccessKey), conv.S(c.globalFlags.AWSSecretKey))

	// cluster name
	clusterName := strings.TrimSpace(conv.S(c.clusterNameArg))
	//if clusterName == "" {
	//	clusterName = console.AskQuestion("Enter cluster name", "cluster1")
	//}
	console.Println("Cluster")
	console.Println(" ", cc.BlackH("Name"), cc.Green(clusterName))

	// AWS networking
	regionName, vpcID, subnetIDs, err := c.getAWSNetwork()
	if err != nil {
		return c.exitWithError(err)
	}
	console.Println("AWS")
	console.Println(" ", cc.BlackH("Region"), cc.Green(regionName))
	console.Println(" ", cc.BlackH("VPC"), cc.Green(vpcID))
	console.Println(" ", cc.BlackH("Subnets"), cc.Green(strings.Join(subnetIDs, " ")))

	console.Println("Container Instances")

	// instance profile
	instanceProfileName := strings.TrimSpace(conv.S(c.commandFlags.InstanceProfile))
	instanceProfileCreated := false
	instanceProfileARN := ""
	if instanceProfileName == "" {
		instanceProfileName = fmt.Sprintf("coldbrew_%s_instance_profile", clusterName)

		existingInstanceProfile, err := c.awsClient.IAM().RetrieveInstanceProfile(instanceProfileName)
		if existingInstanceProfile != nil && err == nil {
			if conv.B(c.commandFlags.SkipExisting) {
				instanceProfileARN = conv.S(existingInstanceProfile.Arn)
			} else {
				return c.exitWithError(fmt.Errorf("Instance Profile [%s] already exists.", instanceProfileName))
			}
		} else {
			var err error
			instanceProfileARN, err = c.createFullAccessInstanceProfile(instanceProfileName)
			if err != nil {
				return c.exitWithError(fmt.Errorf("Failed to create IAM Instance Profile [%s]: %s", instanceProfileName, err.Error()))
			}
			instanceProfileCreated = true
		}
	} else {
		instanceProfile, err := c.awsClient.IAM().RetrieveInstanceProfile(instanceProfileName)
		if err != nil {
			return c.exitWithError(err)
		}
		instanceProfileARN = conv.S(instanceProfile.Arn)
	}

	console.Print(" ", cc.BlackH("Profile"), cc.Green(instanceProfileName), "")
	if instanceProfileCreated {
		console.Println(cc.Yellow("(created)"))
	} else {
		console.Println(cc.BlackH("(existing/skipped)"))
	}

	// container instance type
	instanceType := strings.TrimSpace(conv.S(c.commandFlags.InstanceType))
	if instanceType == "" {
		instanceType = console.AskQuestion("Enter instance type", defaultInstanceType)
	}

	console.Println(" ", cc.BlackH("Type"), cc.Green(instanceType))

	// container instance image ID
	imageID := c.getClusterImageID(conv.S(c.globalFlags.AWSRegion))
	if imageID == "" {
		return c.exitWithError(errors.New("No defatul instance image found"))
	}

	console.Println(" ", cc.BlackH("Image"), cc.Green(imageID))

	// key pair
	keyPairName := strings.TrimSpace(conv.S(c.commandFlags.KeyPairName))
	keyPairInfo, err := c.awsClient.EC2().RetrieveKeyPair(keyPairName)
	if err != nil {
		return c.exitWithError(fmt.Errorf("Failed to retrieve key pair info [%s]: %s", keyPairName, err.Error()))
	}
	if keyPairInfo == nil {
		return c.exitWithError(fmt.Errorf("Key pair [%s] was not found\n", keyPairName))
	}

	console.Println(" ", cc.BlackH("Keypair"), cc.Green(keyPairName))

	// container instances security group
	sgName := fmt.Sprintf("coldbrew_%s_sg", clusterName)
	sgCreated := false
	sgID := ""
	existingSG, err := c.awsClient.EC2().RetrieveSecurityGroupByName(sgName)
	if existingSG != nil && err == nil {
		// SG already exists
		if conv.B(c.commandFlags.SkipExisting) {
			sgID = conv.S(existingSG.GroupId)
		} else {
			return c.exitWithError(fmt.Errorf("Security Group [%s] already exists.", sgName))
		}
	} else {
		sgID, err = c.awsClient.EC2().CreateSecurityGroup(sgName, sgName, vpcID)
		if err != nil {
			return c.exitWithError(fmt.Errorf("Failed to create EC2 Security Group [%s] for container instances: %s", sgName, err.Error()))
		}
		if err := c.awsClient.EC2().AddInboundToSecurityGroup(sgID, ec2.SecurityGroupProtocolTCP, 22, 22, "0.0.0.0/0"); err != nil {
			return c.exitWithError(fmt.Errorf("Failed to add SSH inbound rule to Security Group [%s]: %s", sgName, err.Error()))
		}
		sgCreated = true
	}

	console.Print(" ", cc.BlackH("Security Group"), cc.Green(sgName), "")
	if sgCreated {
		console.Println(cc.Yellow("(created)"))
	} else {
		console.Println(cc.BlackH("(existing/skipped)"))
	}

	ecsClusterName := fmt.Sprintf("coldbrew_%s", clusterName)
	instanceUserData := c.getInstanceUserData(ecsClusterName)
	initialCapacity := conv.U16(c.commandFlags.InitialCapacity)

	console.Println("Auto Scaling")

	// create launch configuration
	lcName := fmt.Sprintf("coldbrew_%s_lc", clusterName)
	lcCreated := false
	existingLC, err := c.awsClient.AutoScaling().RetrieveLaunchConfiguration(lcName)
	if existingLC != nil && err == nil {
		// launch configuration already exists
		if !conv.B(c.commandFlags.SkipExisting) {
			return c.exitWithError(fmt.Errorf("Launch Configuration [%s] already exists.", lcName))
		}
	} else {
		err := c.awsClient.AutoScaling().CreateLaunchConfiguration(lcName, instanceType, imageID, []string{sgID}, keyPairName, instanceProfileARN, instanceUserData)
		if err != nil {
			return c.exitWithError(fmt.Errorf("Failed to create EC2 Launch Configuration [%s]: %s", lcName, err.Error()))
		}
		lcCreated = true
	}

	console.Print(" ", cc.BlackH("Launch Config"), cc.Green(lcName), "")
	if lcCreated {
		console.Println(cc.Yellow("(created)"))
	} else {
		console.Println(cc.BlackH("(existing/skipped)"))
	}

	// create auto scaling group
	asgName := fmt.Sprintf("coldbrew_%s_asg", clusterName)
	asgCreated := false
	existingASG, err := c.awsClient.AutoScaling().RetrieveAutoScalingGroup(asgName)
	if existingASG != nil && err == nil {
		// auto scaling group already exists
		if !conv.B(c.commandFlags.SkipExisting) {
			return c.exitWithError(fmt.Errorf("Auto Scaling Group [%s] already exists.", asgName))
		}
	} else {
		err = c.awsClient.AutoScaling().CreateAutoScalingGroup(asgName, lcName, subnetIDs, 0, initialCapacity, initialCapacity)
		if err != nil {
			return c.exitWithError(fmt.Errorf("Failed to create EC2 Auto Scaling Group [%s]: %s", asgName, err.Error()))
		}
		asgCreated = true
	}

	console.Print(" ", cc.BlackH("Auto Scaling Group"), cc.Green(asgName), "")
	if asgCreated {
		console.Println(cc.Yellow("(created)"))
	} else {
		console.Println(cc.BlackH("(existing/skipped)"))
	}

	console.Println("ECS")

	// create ECS cluster
	ecsCluster, err := c.awsClient.ECS().RetrieveCluster(ecsClusterName)
	ecsClusterCreated := false
	if ecsCluster == nil || err != nil {
		if _, err := c.awsClient.ECS().CreateCluster(ecsClusterName); err != nil {
			return c.exitWithError(fmt.Errorf("Failed to create ECS Cluster [%s]: %s", ecsClusterName, err.Error()))
		}
		ecsClusterCreated = true
	}

	console.Print(" ", cc.BlackH("Cluster"), cc.Green(ecsClusterName), "")
	if ecsClusterCreated {
		console.Println(cc.Yellow("(created)"))
	} else {
		console.Println(cc.BlackH("(existing/skipped)"))
	}

	// create ECS service role (IAM)
	ecsServiceRoleName := fmt.Sprintf("coldbrew_%s_ecs_service_role", clusterName)
	ecsServiceRoleCreated := false
	existingRole, err := c.awsClient.IAM().RetrieveRole(ecsServiceRoleName)
	if existingRole == nil || err != nil {
		_, err := c.createECSServiceRole(ecsServiceRoleName)
		if err != nil {
			return c.exitWithError(fmt.Errorf("Failed to create IAM role [%s]: %s", ecsServiceRoleName, err.Error()))
		}
		ecsServiceRoleCreated = true
	}

	console.Print(" ", cc.BlackH("Service Role"), cc.Green(ecsServiceRoleName), "")
	if ecsServiceRoleCreated {
		console.Println(cc.Yellow("(created)"))
	} else {
		console.Println(cc.BlackH("(existing/skipped)"))
	}

	return nil
}

func (c *Command) getAWSNetwork() (string, string, []string, error) {
	regionName := strings.TrimSpace(conv.S(c.globalFlags.AWSRegion))

	// VPC ID
	vpcID := strings.TrimSpace(conv.S(c.commandFlags.VPC))
	if vpcID == "" {
		// find/use default VPC for the account
		defaultVPC, err := c.awsClient.EC2().RetrieveDefaultVPC()
		if err != nil {
			return "", "", nil, fmt.Errorf("Failed to retrieve default VPC: %s", err.Error())
		} else if defaultVPC == nil {
			return "", "", nil, errors.New("No default VPC configured")
		}

		vpcID = conv.S(defaultVPC.VpcId)
	} else {
		vpc, err := c.awsClient.EC2().RetrieveVPC(vpcID)
		if err != nil {
			return "", "", nil, fmt.Errorf("Failed to retrieve VPC [%s] info: %s", vpcID, err.Error())
		}
		if vpc == nil {
			return "", "", nil, fmt.Errorf("VPC [%s] not found", vpcID)
		}
	}

	// Subnet IDs
	subnetIDs, err := c.awsClient.EC2().ListVPCSubnets(vpcID)
	if err != nil {
		return "", "", nil, fmt.Errorf("Failed to list subnets of VPC [%s]: %s", vpcID, err.Error())
	}

	return regionName, vpcID, subnetIDs, nil
}

func (c *Command) exitWithError(err error) error {
	console.Errorln(cc.Red("Error:"), err.Error())
	return nil
}
