package clustercreate

import (
	"fmt"

	"encoding/base64"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

const (
	ec2AssumeRolePolicy = `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"ec2.amazonaws.com"},"Action": "sts:AssumeRole"}]}`
	ecsAssumeRolePolicy = `{"Version":"2008-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"ecs.amazonaws.com"},"Action": "sts:AssumeRole"}]}`

	administratorAccessPolicyARN = "arn:aws:iam::aws:policy/AdministratorAccess"
	ecsServiceRolePolicyARN      = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceRole"
)

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

func (c *Command) createFullAccessInstanceProfile(profileName string) (string, error) {
	_, err := c.awsClient.IAM().CreateRole(ec2AssumeRolePolicy, profileName)
	if err != nil {
		return "", fmt.Errorf("Failed to create IAM Role [%s]: %s", profileName, err.Error())
	}
	if err := c.awsClient.IAM().AttachRolePolicy(administratorAccessPolicyARN, profileName); err != nil {
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
	iamRole, err := c.awsClient.IAM().CreateRole(ecsAssumeRolePolicy, roleName)
	if err != nil {
		return "", fmt.Errorf("Failed to create IAM Role [%s]: %s", roleName, err.Error())
	}
	if err := c.awsClient.IAM().AttachRolePolicy(ecsServiceRolePolicyARN, roleName); err != nil {
		return "", fmt.Errorf("Failed to attach policy to IAM Role [%s]: %s", roleName, err.Error())
	}

	return conv.S(iamRole.Arn), nil
}
