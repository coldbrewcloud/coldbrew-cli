package core

import "fmt"

const defaultPrefix = "coldbrew-"

const (
	EC2AssumeRolePolicy = `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"ec2.amazonaws.com"},"Action": "sts:AssumeRole"}]}`
	ECSAssumeRolePolicy = `{"Version":"2008-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"ecs.amazonaws.com"},"Action": "sts:AssumeRole"}]}`

	AdministratorAccessPolicyARN = "arn:aws:iam::aws:policy/AdministratorAccess"
	ECSServiceRolePolicyARN      = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceRole"
)

func DefaultECSClusterName(clusterName string) string {
	return fmt.Sprintf("%s%s", defaultPrefix, clusterName)
}

func DefaultLaunchConfigurationName(clusterName string) string {
	return fmt.Sprintf("%s%s-lc", defaultPrefix, clusterName)
}

func DefaultAutoScalingGroupName(clusterName string) string {
	return fmt.Sprintf("%s%s-asg", defaultPrefix, clusterName)
}

func DefaultInstanceProfileName(clusterName string) string {
	return fmt.Sprintf("%s%s-instance-profile", defaultPrefix, clusterName)
}

func DefaultInstanceSecurityGroupName(clusterName string) string {
	return fmt.Sprintf("%s%s-instance-sg", defaultPrefix, clusterName)
}

func DefaultECSServiceRoleName(clusterName string) string {
	return fmt.Sprintf("%s%s-ecs-service-role", defaultPrefix, clusterName)
}

func DefaultContainerInstanceType() string {
	return "t2.micro"
}
