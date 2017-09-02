package clustercreate

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

const (
	defaultECSContainerInstanceImageIDBaseURL = "https://s3-us-west-2.amazonaws.com/files.coldbrewcloud.com/coldbrew-cli/ecs-ci-ami/default/"
	defaultECSContainerInstanceImageOwnerID   = "865092420289"
)

var defaultImageID = map[string]string{
	aws.AWSRegionAPNorthEast1: "ami-3217ed54",
	aws.AWSRegionAPSouthEast1: "ami-b30b67d0",
	aws.AWSRegionAPSouthEast2: "ami-5f38dd3d",
	aws.AWSRegionEUCentral1:   "ami-3645f059",
	aws.AWSRegionEUWest1:      "ami-d104c1a8",
	aws.AWSRegionUSEast1:      "ami-c25a4eb9",
	aws.AWSRegionUSEast2:      "ami-498dae2c",
	aws.AWSRegionUSWest1:      "ami-fdcefa9d",
	aws.AWSRegionUSWest2:      "ami-1d28dd65",
}

var defaultECSContainerInstanceAmazonImageID = map[string]string{
	aws.AWSRegionUSEast1:      "ami-1924770e",
	aws.AWSRegionUSEast2:      "ami-bd3e64d8",
	aws.AWSRegionUSWest1:      "ami-7f004b1f",
	aws.AWSRegionUSWest2:      "ami-56ed4936",
	aws.AWSRegionEUWest1:      "ami-c8337dbb",
	aws.AWSRegionEUCentral1:   "ami-dd12ebb2",
	aws.AWSRegionAPNorthEast1: "ami-c8b016a9",
	aws.AWSRegionAPSouthEast1: "ami-6d22840e",
	aws.AWSRegionAPSouthEast2: "ami-73407d10",
}

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

func (c *Command) retrieveDefaultECSContainerInstancesImageID(region string) string {
	/*defaultImages, err := c.awsClient.EC2().FindImage(defaultECSContainerInstanceImageOwnerID, core.AWSTagNameCreatedTimestamp)
	if err == nil {
		var latestImage *ec2.Image
		var latestImageCreationTime string

		for _, image := range defaultImages {
			if conv.S(image.OwnerId) == defaultECSContainerInstanceImageOwnerID {
				if latestImage == nil {
					latestImageCreationTime = getCreationTimeFromTags(image.Tags)
					if latestImageCreationTime != "" {
						latestImage = image
					}
				} else {
					creationTime := getCreationTimeFromTags(image.Tags)
					if creationTime != "" {
						if strings.Compare(latestImageCreationTime, creationTime) < 0 {
							latestImage = image
							latestImageCreationTime = creationTime
						}
					}
				}
			}
		}

		if latestImage != nil {
			return conv.S(latestImage.ImageId)
		}
	}*/
	if imageID, ok := defaultImageID[region]; ok {
		return imageID
	}

	// if failed to find coldbrew-cli default image, use Amazon ECS optimized image as fallback
	console.Error("Failed to retrieve default image ID for ECS Container Instances. Amazon ECS Optimized AMI will be used instead.")
	if imageID, ok := defaultECSContainerInstanceAmazonImageID[region]; ok {
		return imageID
	}
	return ""
}

func (c *Command) getDefaultInstanceUserData(ecsClusterName string) string {
	userData := fmt.Sprintf(`#!/bin/bash
echo ECS_CLUSTER=%s >> /etc/ecs/ecs.config`, ecsClusterName)
	return base64.StdEncoding.EncodeToString([]byte(userData))
}

func (c *Command) createDefaultInstanceProfile(profileName string) (string, error) {
	_, err := c.awsClient.IAM().CreateRole(core.EC2AssumeRolePolicy, profileName)
	if err != nil {
		return "", fmt.Errorf("Failed to create IAM Role [%s]: %s", profileName, err.Error())
	}
	if err := c.awsClient.IAM().AttachRolePolicy(core.AdministratorAccessPolicyARN, profileName); err != nil {
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
	iamRole, err := c.awsClient.IAM().CreateRole(core.ECSAssumeRolePolicy, roleName)
	if err != nil {
		return "", fmt.Errorf("Failed to create IAM Role [%s]: %s", roleName, err.Error())
	}
	if err := c.awsClient.IAM().AttachRolePolicy(core.ECSServiceRolePolicyARN, roleName); err != nil {
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

func getCreationTimeFromTags(tags []*ec2.Tag) string {
	for _, tag := range tags {
		if conv.S(tag.Key) == core.AWSTagNameCreatedTimestamp {
			return conv.S(tag.Value)
			break
		}
	}
	return ""
}
