package clustercreate

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/core/clusters"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

func (c *Command) getAWSInfo() (string, string, []string, error) {
	regionName, vpcID, err := c.globalFlags.GetAWSRegionAndVPCID()
	if err != nil {
		return "", "", nil, err
	}

	// Subnet IDs
	subnetIDs, err := c.awsClient.EC2().ListVPCSubnets(vpcID)
	if err != nil {
		return "", "", nil, fmt.Errorf("Failed to list subnets of VPC [%s]: %s", vpcID, err.Error())
	}
	if len(subnetIDs) == 0 {
		return "", "", nil, fmt.Errorf("VPC [%s] does not have any subnets.", vpcID)
	}

	return regionName, vpcID, subnetIDs, nil
}

func (c *Command) getClusterImageID(region string) string {
	switch region {
	case aws.AWSRegionUSEast1:
		return "ami-40286957"
	case aws.AWSRegionUSWest1:
		return "ami-20fab440"
	case aws.AWSRegionUSWest2:
		return "ami-562cf236"
	case aws.AWSRegionEUWest1:
		return "ami-175f1964"
	case aws.AWSRegionEUCentral1:
		return "ami-c55ea2aa"
	case aws.AWSRegionAPNorthEast1:
		return "ami-010ed160"
	case aws.AWSRegionAPSouthEast1:
		return "ami-438b2f20"
	case aws.AWSRegionAPSouthEast2:
		return "ami-862211e5"
	default:
		return ""
	}
}

func (c *Command) getInstanceUserData(ecsClusterName string) string {
	userData := fmt.Sprintf(`#!/bin/bash
echo ECS_CLUSTER=%s >> /etc/ecs/ecs.config`, ecsClusterName)
	return base64.StdEncoding.EncodeToString([]byte(userData))
}

func (c *Command) createDefaultInstanceProfile(profileName string) (string, error) {
	_, err := c.awsClient.IAM().CreateRole(clusters.EC2AssumeRolePolicy, profileName)
	if err != nil {
		return "", fmt.Errorf("Failed to create IAM Role [%s]: %s", profileName, err.Error())
	}
	if err := c.awsClient.IAM().AttachRolePolicy(clusters.AdministratorAccessPolicyARN, profileName); err != nil {
		return "", fmt.Errorf("Failed to attach policy to IAM Role [%s]: %s", profileName, err.Error())
	}

	iamInstanceProfile, err := c.awsClient.IAM().CreateInstanceProfile(profileName)
	if err != nil {
		return "", fmt.Errorf("Failed to create IAM Instance Profile [%s]: %s", profileName, err.Error())
	}
	if iamInstanceProfile == nil {
		return "", fmt.Errorf("Failed to create IAM Instance Profile [%s]: empty result", profileName)
	}
	if err := c.awsClient.IAM().AddRoleToInstanceProfile(profileName, profileName); err != nil {
		return "", fmt.Errorf("Failed to add IAM Role [%s] to IAM Instance Profile [%s]: %s", profileName, profileName, err.Error())
	}

	return conv.S(iamInstanceProfile.Arn), nil
}

func (c *Command) createECSServiceRole(roleName string) (string, error) {
	iamRole, err := c.awsClient.IAM().CreateRole(clusters.ECSAssumeRolePolicy, roleName)
	if err != nil {
		return "", fmt.Errorf("Failed to create IAM Role [%s]: %s", roleName, err.Error())
	}
	if err := c.awsClient.IAM().AttachRolePolicy(clusters.ECSServiceRolePolicyARN, roleName); err != nil {
		return "", fmt.Errorf("Failed to attach policy to IAM Role [%s]: %s", roleName, err.Error())
	}

	return conv.S(iamRole.Arn), nil
}

func (c *Command) waitAutoScalingGroupDeletion(autoScalingGroupName string) error {
	maxRetries := 60
	for i := 0; i < maxRetries; i++ {
		autoScalingGroup, err := c.awsClient.AutoScaling().RetrieveAutoScalingGroup(autoScalingGroupName)
		if err != nil {
			return fmt.Errorf("Failed to retrieve Auto Scaling Group [%s]: %s", autoScalingGroupName, err.Error())
		}
		if autoScalingGroup == nil {
			break
		}

		time.Sleep(1 * time.Second)
	}
	return nil
}
