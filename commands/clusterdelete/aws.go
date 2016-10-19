package clusterdelete

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/coldbrewcloud/coldbrew-cli/core/clusters"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

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

func (c *Command) scaleDownAutoScalingGroup(autoScalingGroup *autoscaling.Group) error {
	if autoScalingGroup == nil {
		return nil
	}

	asgName := conv.S(autoScalingGroup.AutoScalingGroupName)
	if err := c.awsClient.AutoScaling().SetAutoScalingGroupDesiredCapacity(asgName, 0); err != nil {
		return fmt.Errorf("Failed to change Auto Scaling Group [%s] desired capacity to 0: %s", asgName, err.Error())
	}

	maxTries := 300 // ~ 5-6 mins
	for i := 0; i < maxTries; i++ {
		asg, err := c.awsClient.AutoScaling().RetrieveAutoScalingGroup(asgName)
		if err != nil {
			return fmt.Errorf("Failed to retrieve Auto Scaling Group [%s]: %s", asgName, err.Error())
		}

		if asg == nil {
			return fmt.Errorf("Auto Scaling Group [%s] not found", asgName)
		}

		if len(asg.Instances) == 0 {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func (c *Command) deleteECSServiceRole(roleName string) error {
	if err := c.awsClient.IAM().DetachRolePolicy(clusters.ECSServiceRolePolicyARN, roleName); err != nil {
		return fmt.Errorf("Failed to detach ECS Service Role Policy from IAM Role [%s]: %s", roleName, err.Error())
	}

	if err := c.awsClient.IAM().DeleteRole(roleName); err != nil {
		return fmt.Errorf("Failed to delete IAM Role [%s]: %s", roleName, err.Error())
	}

	return nil
}

func (c *Command) deleteDefaultInstanceProfile(profileName string) error {
	if err := c.awsClient.IAM().RemoveRoleFromInstanceProfile(profileName, profileName); err != nil {
		return fmt.Errorf("Failed to remove IAM Role [%s] from Instance Profile [%s]: %s", profileName, profileName, err.Error())
	}

	if err := c.awsClient.IAM().DetachRolePolicy(clusters.AdministratorAccessPolicyARN, profileName); err != nil {
		return fmt.Errorf("Failed to detach Administrator Access Policy from IAM Role [%s]: %s", profileName, err.Error())
	}

	if err := c.awsClient.IAM().DeleteRole(profileName); err != nil {
		return fmt.Errorf("Failed to delete IAM Role [%s]: %s", profileName, err.Error())
	}

	if err := c.awsClient.IAM().DeleteInstanceProfile(profileName); err != nil {
		return fmt.Errorf("Failed to delete Instance Profile [%s]: %s", profileName, err.Error())
	}

	return nil
}
