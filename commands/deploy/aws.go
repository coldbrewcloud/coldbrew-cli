package deploy

import (
	"fmt"
	"math"

	"github.com/coldbrewcloud/coldbrew-cli/aws/ec2"
	"github.com/coldbrewcloud/coldbrew-cli/aws/ecs"
	"github.com/coldbrewcloud/coldbrew-cli/aws/elb"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"github.com/d5/cc"
)

func (c *Command) isClusterAvailable(clusterName string) error {
	// check ECS cluster
	ecsClusterName := core.DefaultECSClusterName(clusterName)
	ecsCluster, err := c.awsClient.ECS().RetrieveCluster(ecsClusterName)
	if err != nil {
		return fmt.Errorf("Failed to retrieve ECS Cluster [%s]: %s", ecsClusterName, err.Error())
	}
	if ecsCluster == nil || conv.S(ecsCluster.Status) == "INACTIVE" {
		return fmt.Errorf("ECS Cluster [%s] not found", ecsClusterName)
	}

	// check ECS service role
	ecsServiceRoleName := core.DefaultECSServiceRoleName(clusterName)
	ecsServiceRole, err := c.awsClient.IAM().RetrieveRole(ecsServiceRoleName)
	if err != nil {
		return fmt.Errorf("Failed to retrieve IAM Role [%s]: %s", ecsServiceRoleName, err.Error())
	}
	if ecsServiceRole == nil {
		return fmt.Errorf("IAM Role [%s] not found", ecsServiceRoleName)
	}

	return nil
}

func (c *Command) prepareECRRepo(repoName string) (string, error) {
	ecrRepo, err := c.awsClient.ECR().RetrieveRepository(repoName)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve ECR repository [%s]: %s", repoName, err.Error())
	}

	if ecrRepo == nil {
		ecrRepo, err = c.awsClient.ECR().CreateRepository(repoName)
		if err != nil {
			return "", fmt.Errorf("Failed to create ECR repository [%s]: %s", repoName, err.Error())
		}
	}

	return *ecrRepo.RepositoryUri, nil
}

func (c *Command) prepareECSServiceRole(roleName string) error {
	const (
		ecsAssumeRolePolicy     = `{"Version": "2008-10-17", "Statement": [{"Sid": "", "Effect": "Allow", "Principal": {"Service": "ecs.amazonaws.com"},"Action": "sts:AssumeRole"}]}`
		ecsServiceRolePolicyARN = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceRole"
	)

	iamRole, err := c.awsClient.IAM().RetrieveRole(roleName)
	if err != nil {
		return fmt.Errorf("Failed to IAM role [%s]: %s", roleName, err.Error())
	}

	if iamRole == nil {
		iamRole, err = c.awsClient.IAM().CreateRole(ecsAssumeRolePolicy, roleName)
		if err != nil {
			return fmt.Errorf("Failed to create IAM role [%s]: %s", roleName, err.Error())
		}
		if err := c.awsClient.IAM().AttachRolePolicy(ecsServiceRolePolicyARN, roleName); err != nil {
			return fmt.Errorf("Failed to attach policy to role [%s]: %s", roleName, err.Error())
		}
	}

	return nil
}

func (c *Command) updateECSTaskDefinition(dockerImageFullURI string) (string, error) {
	// port mappings
	var portMappings []ecs.PortMapping
	if conv.U16(c.conf.Port) > 0 {
		portMappings = []ecs.PortMapping{
			{
				ContainerPort: conv.U16(c.conf.Port),
				Protocol:      "tcp",
			},
		}
	}

	ecsTaskDefinitionName := core.DefaultECSTaskDefinitionName(conv.S(c.conf.Name))
	ecsTaskContainerName := core.DefaultECSTaskMainContainerName(conv.S(c.conf.Name))
	cpu := uint64(math.Ceil(conv.F64(c.conf.CPU) * 1024.0))
	memory, err := core.ParseSizeExpression(conv.S(c.conf.Memory))
	if err != nil {
		return "", err
	}
	memory /= 1000 * 1000
	useCloudWatchLogs := false

	console.Printf("Updating ECS Task Definition [%s]...\n", cc.Green(ecsTaskContainerName))
	ecsTaskDef, err := c.awsClient.ECS().UpdateTaskDefinition(
		ecsTaskDefinitionName,
		dockerImageFullURI,
		ecsTaskContainerName,
		cpu,
		memory,
		c.conf.Env,
		portMappings,
		useCloudWatchLogs)
	if err != nil {
		return "", fmt.Errorf("Failed to update ECS task definition: %s", err.Error())
	}

	return conv.S(ecsTaskDef.TaskDefinitionArn), nil
}

func (c *Command) createOrUpdateECSService(ecsTaskDefinitionARN string) error {
	ecsClusterName := core.DefaultECSClusterName(conv.S(c.conf.ClusterName))
	ecsServiceName := core.DefaultECSServiceName(conv.S(c.conf.Name))

	ecsService, err := c.awsClient.ECS().RetrieveService(ecsClusterName, ecsServiceName)
	if err != nil {
		return fmt.Errorf("Failed to retrieve ECS Service [%s/%s]: %s", ecsClusterName, ecsServiceName, err.Error())
	}

	if ecsService != nil && conv.S(ecsService.Status) == "ACTIVE" {
		// TODO: handle the case where configuration requires changes in ECS Service
		// E.g. ask user to re-create the Service
		if err := c.updateECSService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN); err != nil {
			return err
		}
	} else {
		if err := c.createECSService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN); err != nil {
			return err
		}
	}

	return nil
}

func (c *Command) createECSService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN string) error {
	ecsServiceRoleName := core.DefaultECSServiceRoleName(conv.S(c.conf.ClusterName))
	ecsTaskContainerName := conv.S(c.conf.Name)
	ecsTaskContainerPort := conv.U16(c.conf.Port)

	var loadBalancers []*ecs.LoadBalancer
	if conv.U16(c.conf.Port) > 0 && c.conf.LoadBalancer.Enabled {
		// prepare load balancer (create one if needed)
		loadBalancer, err := c.prepareELBLoadBalancer(
			conv.S(c.conf.AWS.ELBLoadBalancerName),
			conv.S(c.conf.AWS.ELBTargetGroupName),
			conv.U16(c.conf.LoadBalancer.Port),
			ecsServiceRoleName,
			ecsTaskContainerName,
			ecsTaskContainerPort)
		if err != nil {
			return err
		}

		loadBalancers = []*ecs.LoadBalancer{loadBalancer}
	}

	console.Printf("Creating ECS Service [%s]...\n", cc.Green(ecsServiceName))
	_, err := c.awsClient.ECS().CreateService(
		ecsClusterName, ecsServiceName, ecsTaskDefinitionARN, conv.U16(c.conf.Units),
		loadBalancers, ecsServiceRoleName)
	if err != nil {
		return fmt.Errorf("Failed to create ECS service [%s/%s]: %s", ecsClusterName, ecsServiceName, err.Error())
	}

	return nil
}

func (c *Command) prepareELBLoadBalancer(elbLoadBalancerName, elbTargetGroupName string, elbPort uint16, ecsServiceRoleName, ecsTaskContainerName string, ecsTaskContainerPort uint16) (*ecs.LoadBalancer, error) {
	_, vpcID, err := c.globalFlags.GetAWSRegionAndVPCID()
	if err != nil {
		return nil, err
	}

	elbLoadBalancer, err := c.awsClient.ELB().RetrieveLoadBalancer(elbLoadBalancerName)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve ELB Load Balancer [%s]: %s", elbLoadBalancerName, err.Error())
	}
	if elbLoadBalancer != nil {
		elbTargetGroup, err := c.awsClient.ELB().RetrieveTargetGroupByName(elbTargetGroupName)
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve ELB Target Group [%s]: %s", elbTargetGroupName, err.Error())
		}
		if elbTargetGroup == nil {
			return nil, fmt.Errorf("ELB Target Group [%s] was not found.", elbTargetGroupName)
		}

		return &ecs.LoadBalancer{
			ELBTargetGroupARN: conv.S(elbTargetGroup.TargetGroupArn),
			TaskContainerName: ecsTaskContainerName,
			TaskContainerPort: ecsTaskContainerPort,
		}, nil
	} else {
		// create ELB target group
		elbTargetGroupARN, err := c.createELBTargetGroup(elbTargetGroupName)
		if err != nil {
			return nil, err
		}

		// create security group for ELB (if needed)
		elbSecurityGroupID, err := c.prepareLoadBalancerSecurityGroup(vpcID, elbPort)
		if err != nil {
			return nil, err
		}

		// create ELB load balancer
		elbLoadBalancerARN, err := c.createELBLoadBalancer(elbLoadBalancerName, vpcID, elbSecurityGroupID)
		if err != nil {
			return nil, err
		}

		// create listen between load balancer and target group
		err = c.awsClient.ELB().CreateListener(elbLoadBalancerARN, elbTargetGroupARN, elbPort, "HTTP")
		if err != nil {
			return nil, fmt.Errorf("Failed to create ELB Listener: %s", err.Error())
		}

		return &ecs.LoadBalancer{
			ELBTargetGroupARN: elbTargetGroupARN,
			TaskContainerName: ecsTaskContainerName,
			TaskContainerPort: ecsTaskContainerPort,
		}, nil
	}
}

func (c *Command) createELBTargetGroup(targetGroupName string) (string, error) {
	_, vpcID, err := c.globalFlags.GetAWSRegionAndVPCID()
	if err != nil {
		return "", err
	}

	checkInterval, err := core.ParseTimeExpression(conv.S(c.conf.LoadBalancer.HealthCheck.Interval))
	if err != nil {
		return "", nil
	}

	timeout, err := core.ParseTimeExpression(conv.S(c.conf.LoadBalancer.HealthCheck.Timeout))
	if err != nil {
		return "", nil
	}

	healthCheck := &elb.HealthCheckParams{
		CheckIntervalSeconds:    uint16(checkInterval),
		CheckPath:               conv.S(c.conf.LoadBalancer.HealthCheck.Path),
		Protocol:                "HTTP",
		ExpectedHTTPStatusCodes: conv.S(c.conf.LoadBalancer.HealthCheck.Status),
		CheckTimeoutSeconds:     uint16(timeout),
		HealthyThresholdCount:   conv.U16(c.conf.LoadBalancer.HealthCheck.HealthyLimit),
		UnhealthyThresholdCount: conv.U16(c.conf.LoadBalancer.HealthCheck.UnhealthyLimit),
	}

	targetGroup, err := c.awsClient.ELB().CreateTargetGroup(targetGroupName, 80, "HTTP", vpcID, healthCheck)
	if err != nil {
		return "", fmt.Errorf("Failed to create ELB target group [%s]: %s", targetGroupName, err.Error())
	}

	return conv.S(targetGroup.TargetGroupArn), nil
}

func (c *Command) createELBLoadBalancer(name, vpcID, securityGroupID string) (string, error) {
	subnetIDs, err := c.awsClient.EC2().ListVPCSubnets(vpcID)
	if err != nil {
		return "", fmt.Errorf("Failed to list subnets: %s", err.Error())
	}

	lb, err := c.awsClient.ELB().CreateLoadBalancer(name, true, []string{securityGroupID}, subnetIDs)
	if err != nil {
		return "", fmt.Errorf("Failed to create ELB Load Balancer [%s]: %s", name, err.Error())
	}

	return conv.S(lb.LoadBalancerArn), nil
}

func (c *Command) updateECSService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN string) error {
	console.Printf("Updating ECS Service [%s]...\n", cc.Green(ecsServiceName))
	_, err := c.awsClient.ECS().UpdateService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN, conv.U16(c.conf.Units))
	if err != nil {
		return fmt.Errorf("Failed to update ECS service [%s/%s]: %s", ecsClusterName, ecsServiceName, err.Error())
	}

	return nil
}

func (c *Command) prepareLoadBalancerSecurityGroup(vpcID string, elbPort uint16) (string, error) {
	if utils.IsBlank(conv.S(c.conf.AWS.ELBSecurityGroup)) {
		// use default security group
		securityGroupName := fmt.Sprintf("%s-sg", conv.S(c.conf.AWS.ELBSecurityGroup))

		// test if default security group exists or not
		securityGroup, err := c.awsClient.EC2().RetrieveSecurityGroupByNameOrID(securityGroupName)
		if err != nil {
			return "", fmt.Errorf("Failed to retrieve EC2 Security Group [%s]: %s", securityGroupName, err.Error())
		}

		// default security group exists, return its group ID
		if securityGroup != nil {
			return conv.S(securityGroup.GroupId), nil
		}

		// create default security group
		securityGroupID, err := c.awsClient.EC2().CreateSecurityGroup(securityGroupName, securityGroupName, vpcID)
		if err != nil {
			return "", fmt.Errorf("Failed to create EC2 Security Group [%s]: %s", securityGroupName, err.Error())
		}

		// add load balancer inbound rule
		err = c.awsClient.EC2().AddInboundToSecurityGroup(
			securityGroupID,
			ec2.SecurityGroupProtocolTCP,
			elbPort, elbPort, "0.0.0.0/0")
		if err != nil {
			return "", fmt.Errorf("Failed to add inbound rule to EC2 Security Group [%s]: %s", securityGroupName, err.Error())
		}

		// add inbound rule to ECS instance security group
		ecsInstancesSecurityGroupName := core.DefaultInstanceSecurityGroupName(conv.S(c.conf.ClusterName))
		ecsInstancesSecurityGroup, err := c.awsClient.EC2().RetrieveSecurityGroupByName(ecsInstancesSecurityGroupName)
		if err != nil {
			return "", fmt.Errorf("Failed to retrieve EC2 Security Group [%s]: %s", ecsInstancesSecurityGroupName, err.Error())
		}
		if ecsInstancesSecurityGroup == nil {
			return "", fmt.Errorf("EC2 Security Group [%s] for ECS Container Instances was not found.", ecsInstancesSecurityGroupName)
		}
		err = c.awsClient.EC2().AddInboundToSecurityGroup(
			conv.S(ecsInstancesSecurityGroup.GroupId),
			ec2.SecurityGroupProtocolTCP,
			0, 0, securityGroupID)
		if err != nil {
			return "", fmt.Errorf("Failed to add inbound rule to EC2 Security Group [%s]: %s", ecsInstancesSecurityGroupName, err.Error())
		}

		return securityGroupID, nil

	} else {
		// make sure user-specified Security Group does exist
		securityGroup, err := c.awsClient.EC2().RetrieveSecurityGroupByNameOrID(conv.S(c.conf.AWS.ELBSecurityGroup))
		if err != nil {
			return "", fmt.Errorf("Failed to retrieve EC2 Security Group [%s]: %s", conv.S(c.conf.AWS.ELBSecurityGroup), err.Error())
		}
		if securityGroup == nil {
			return "", fmt.Errorf("EC2 Security Group [%s] was not found.", conv.S(c.conf.AWS.ELBSecurityGroup))
		}

		return conv.S(securityGroup.GroupId), nil
	}
}
