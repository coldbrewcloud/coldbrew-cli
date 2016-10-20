package deploy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/coldbrewcloud/coldbrew-cli/aws/ec2"
	"github.com/coldbrewcloud/coldbrew-cli/aws/ecs"
	"github.com/coldbrewcloud/coldbrew-cli/aws/elb"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/core/clusters"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

func (c *Command) isClusterAvailable(clusterName string) error {
	// check ECS cluster
	ecsClusterName := clusters.DefaultECSClusterName(clusterName)
	ecsCluster, err := c.awsClient.ECS().RetrieveCluster(ecsClusterName)
	if err != nil {
		return fmt.Errorf("Failed to retrieve ECS Cluster [%s]: %s", ecsClusterName, err.Error())
	}
	if ecsCluster == nil || conv.S(ecsCluster.Status) == "INACTIVE" {
		return fmt.Errorf("ECS Cluster [%s] not found", ecsClusterName)
	}

	// check ECS service role
	ecsServiceRoleName := clusters.DefaultECSServiceRoleName(clusterName)
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

func (c *Command) prepareECSCluster(clusterName string) error {
	cluster, err := c.awsClient.ECS().RetrieveCluster(clusterName)
	if err != nil {
		return fmt.Errorf("Failed to retrieve ECS cluster [%s]: %s", clusterName, err.Error())
	}

	if cluster == nil {
		if _, err := c.awsClient.ECS().CreateCluster(clusterName); err != nil {
			return fmt.Errorf("Failed to create ECS cluster [%s]: %s", clusterName, err.Error())
		}
	}

	return nil
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
	envs, err := c.populateEnvVars()
	if err != nil {
		return "", err
	}

	// port mappings
	// TODO: support multiple ports exposing
	// TODO: support UDP protocol
	var portMappings []ecs.PortMapping
	if conv.U16(c.commandFlags.AppPort) > 0 {
		portMappings = []ecs.PortMapping{
			{
				ContainerPort: conv.U16(c.commandFlags.AppPort),
				Protocol:      "tcp",
			},
		}
	}

	ecsTaskDefinitionName := conv.S(c.commandFlags.AppName)
	ecsTaskContainerName := conv.S(c.commandFlags.AppName)
	cpu := conv.U64(c.commandFlags.CPU)
	memory := conv.U64(c.commandFlags.Memory)
	useCloudWatchLogs := *c.commandFlags.CloudWatchLogs

	ecsTaskDef, err := c.awsClient.ECS().UpdateTaskDefinition(
		ecsTaskDefinitionName,
		dockerImageFullURI,
		ecsTaskContainerName,
		cpu,
		memory,
		envs,
		portMappings,
		useCloudWatchLogs)
	if err != nil {
		return "", fmt.Errorf("Failed to update ECS task definition: %s", err.Error())
	}

	return conv.S(ecsTaskDef.TaskDefinitionArn), nil
}

func (c *Command) createOrUpdateECSService(ecsTaskDefinitionARN string) error {
	ecsClusterName := conv.S(c.commandFlags.ECSClusterName)
	ecsServiceName := conv.S(c.commandFlags.AppName)

	ecsService, err := c.awsClient.ECS().RetrieveService(ecsClusterName, ecsServiceName)
	if err != nil {
		return fmt.Errorf("Failed to retrieve ECS service [%s/%s]: %s", ecsClusterName, ecsServiceName, err.Error())
	}

	if ecsService != nil && *ecsService.Status != "INACTIVE" {
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
	// TODO: support multiple load balancers

	loadBalancerName := conv.S(c.commandFlags.LoadBalancerName)
	ecsServiceRoleName := conv.S(c.commandFlags.ECSServiceRoleName)
	ecsTaskContainerName := conv.S(c.commandFlags.AppName) // TODO: implicitly known; should have been retrieved from task definition
	ecsTaskContainerPort := conv.U16(c.commandFlags.AppPort)

	var loadBalancers []*ecs.LoadBalancer
	if !utils.IsBlank(loadBalancerName) {
		// prepare load balancer (create one if needed)
		loadBalancer, err := c.prepareELBLoadBalancer(loadBalancerName, ecsServiceRoleName, ecsTaskContainerName, ecsTaskContainerPort)
		if err != nil {
			return err
		}
		loadBalancers = []*ecs.LoadBalancer{loadBalancer}

		// prepare ECS service role (create one if needed)
		if err := c.prepareECSServiceRole(ecsServiceRoleName); err != nil {
			return err
		}
	}

	desiredUnits := conv.U16(c.commandFlags.Units)

	ecsService, err := c.awsClient.ECS().CreateService(
		ecsClusterName, ecsServiceName, ecsTaskDefinitionARN, desiredUnits,
		loadBalancers, ecsServiceRoleName)
	if err != nil {
		return fmt.Errorf("Failed to create ECS service [%s/%s]: %s", ecsClusterName, ecsServiceName, err.Error())
	}

	console.Println("ECS service created", conv.S(ecsService.ServiceName), conv.S(ecsService.ServiceArn))

	return nil
}

func (c *Command) prepareELBLoadBalancer(lbName string, ecsServiceRoleName, ecsTaskContainerName string, ecsTaskContainerPort uint16) (*ecs.LoadBalancer, error) {
	elbTargetGroupName := lbName

	elbDesc, err := c.awsClient.ELB().RetrieveLoadBalancer(lbName)
	if err != nil {
		fmt.Errorf("Failed to retrieve ELB load balancer [%s]: %s", lbName, err.Error())
	}
	if elbDesc != nil {
		elbTargetGroup, err := c.awsClient.ELB().RetrieveTargetGroup(
			conv.S(elbDesc.LoadBalancerArn),
			elbTargetGroupName)
		if err != nil {
			fmt.Errorf("Failed to retrieve ELB target group [%s]: %s", elbTargetGroupName, err.Error())
		}

		return &ecs.LoadBalancer{
			ELBTargetGroupARN: conv.S(elbTargetGroup.TargetGroupArn),
			TaskContainerName: ecsTaskContainerName,
			TaskContainerPort: ecsTaskContainerPort,
		}, nil
	} else {
		// create ELB target group
		elbTargetGroupARN, err := c.createELBTargetGroup(lbName)
		if err != nil {
			return nil, err
		}

		// load balancer port
		lbPort := uint16(80)

		securityGroupName := fmt.Sprintf("elb-%s", lbName)
		securityGroupID, err := c.awsClient.EC2().CreateSecurityGroup(securityGroupName, securityGroupName, conv.S(c.commandFlags.VPCID))
		if err != nil {
			return nil, fmt.Errorf("Failed to create security group [%s]: %s", securityGroupName, err.Error())
		}
		err = c.awsClient.EC2().AddInboundToSecurityGroup(securityGroupID, ec2.SecurityGroupProtocolTCP, lbPort, lbPort, "0.0.0.0/0")
		if err != nil {
			return nil, fmt.Errorf("Failed to add incoming rule to security group [%s]: %s", securityGroupID, err.Error())
		}

		// create ELB load balancer
		elbLoadBalancerARN, err := c.createELBLoadBalancer(lbName, securityGroupID)
		if err != nil {
			return nil, err
		}

		err = c.awsClient.ELB().CreateListener(elbLoadBalancerARN, elbTargetGroupARN, lbPort, "HTTP")
		if err != nil {
			return nil, fmt.Errorf("Failed to create ELB listener: %s", err.Error())
		}

		return &ecs.LoadBalancer{
			ELBTargetGroupARN: elbTargetGroupARN,
			TaskContainerName: ecsTaskContainerName,
			TaskContainerPort: ecsTaskContainerPort,
		}, nil
	}
}

func (c *Command) createELBTargetGroup(targetGroupName string) (string, error) {
	healthCheck := &elb.HealthCheckParams{
		CheckIntervalSeconds:    30,
		CheckPath:               conv.S(c.commandFlags.HealthCheckPath),
		Protocol:                "HTTP",
		ExpectedHTTPStatusCodes: "200-299",
		CheckTimeoutSeconds:     10,
		HealthyThresholdCount:   5,
		UnhealthyThresholdCount: 2,
	}

	vpcID := conv.S(c.commandFlags.VPCID)

	targetGroup, err := c.awsClient.ELB().CreateTargetGroup(targetGroupName, 80, "HTTP", vpcID, healthCheck)
	if err != nil {
		return "", fmt.Errorf("Failed to create ELB target group [%s]: %s", targetGroupName, err.Error())
	}

	return conv.S(targetGroup.TargetGroupArn), nil
}

func (c *Command) createELBLoadBalancer(loadBalancerName, securityGroupID string) (string, error) {
	vpcID := conv.S(c.commandFlags.VPCID)
	subnetIDs, err := c.awsClient.EC2().ListVPCSubnets(vpcID)
	if err != nil {
		return "", fmt.Errorf("Failed to list subnets: %s", err.Error())
	}

	lb, err := c.awsClient.ELB().CreateLoadBalancer(loadBalancerName, true, []string{securityGroupID}, subnetIDs)
	if err != nil {
		return "", fmt.Errorf("Failed to create ELB load balancer [%s]: %s", loadBalancerName, err.Error())
	}

	return conv.S(lb.LoadBalancerArn), nil
}

func (c *Command) updateECSService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN string) error {
	desiredUnits := conv.U16(c.commandFlags.Units)

	ecsService, err := c.awsClient.ECS().UpdateService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN, desiredUnits)
	if err != nil {
		return fmt.Errorf("Failed to update ECS service [%s/%s]: %s", ecsClusterName, ecsServiceName, err.Error())
	}

	console.Println("ECS service updated", conv.S(ecsService.ServiceName), conv.S(ecsService.ServiceArn))

	return nil
}

func (c *Command) populateEnvVars() (map[string]string, error) {
	// load from envs file (JSON)
	envs := make(map[string]string)
	if !utils.IsBlank(conv.S(c.commandFlags.EnvsFile)) {
		data, err := ioutil.ReadFile(conv.S(c.commandFlags.EnvsFile))
		if err != nil {
			return nil, fmt.Errorf("Failed to read envs file [%s]: %s", conv.S(c.commandFlags.EnvsFile), err.Error())
		}

		if err := json.Unmarshal(data, &envs); err != nil {
			return nil, fmt.Errorf("Failed to parse envs file: %s", err.Error())
		}
	}

	// override from CLI params
	if c.commandFlags.Envs != nil {
		for ek, ev := range *c.commandFlags.Envs {
			envs[ek] = ev
		}
	}

	return envs, nil
}
