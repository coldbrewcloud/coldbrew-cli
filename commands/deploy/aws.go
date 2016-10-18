package deploy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/coldbrewcloud/coldbrew-cli/aws/ec2"
	"github.com/coldbrewcloud/coldbrew-cli/aws/ecs"
	"github.com/coldbrewcloud/coldbrew-cli/aws/elb"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

func (dc *DeployCommand) prepareECRRepo(repoName string) (string, error) {
	ecrRepo, err := dc.awsClient.ECR().RetrieveRepository(repoName)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve ECR repository [%s]: %s", repoName, err.Error())
	}

	if ecrRepo == nil {
		ecrRepo, err = dc.awsClient.ECR().CreateRepository(repoName)
		if err != nil {
			return "", fmt.Errorf("Failed to create ECR repository [%s]: %s", repoName, err.Error())
		}
	}

	return *ecrRepo.RepositoryUri, nil
}

func (dc *DeployCommand) prepareECSCluster(clusterName string) error {
	cluster, err := dc.awsClient.ECS().RetrieveCluster(clusterName)
	if err != nil {
		return fmt.Errorf("Failed to retrieve ECS cluster [%s]: %s", clusterName, err.Error())
	}

	if cluster == nil {
		if _, err := dc.awsClient.ECS().CreateCluster(clusterName); err != nil {
			return fmt.Errorf("Failed to create ECS cluster [%s]: %s", clusterName, err.Error())
		}
	}

	return nil
}

func (dc *DeployCommand) prepareECSServiceRole(roleName string) error {
	const (
		ecsAssumeRolePolicy     = `{"Version": "2008-10-17", "Statement": [{"Sid": "", "Effect": "Allow", "Principal": {"Service": "ecs.amazonaws.com"},"Action": "sts:AssumeRole"}]}`
		ecsServiceRolePolicyARN = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceRole"
	)

	iamRole, err := dc.awsClient.IAM().RetrieveRole(roleName)
	if err != nil {
		return fmt.Errorf("Failed to IAM role [%s]: %s", roleName, err.Error())
	}

	if iamRole == nil {
		iamRole, err = dc.awsClient.IAM().CreateRole(ecsAssumeRolePolicy, roleName)
		if err != nil {
			return fmt.Errorf("Failed to create IAM role [%s]: %s", roleName, err.Error())
		}
		if err := dc.awsClient.IAM().AttachRolePolicy(ecsServiceRolePolicyARN, roleName); err != nil {
			return fmt.Errorf("Failed to attach policy to role [%s]: %s", roleName, err.Error())
		}
	}

	return nil
}

func (dc *DeployCommand) updateECSTaskDefinition(dockerImageFullURI string) (string, error) {
	envs, err := dc.populateEnvVars()
	if err != nil {
		return "", err
	}

	// port mappings
	// TODO: support multiple ports exposing
	// TODO: support UDP protocol
	var portMappings []ecs.PortMapping
	if conv.U16(dc.deployFlags.ContainerPort) > 0 {
		portMappings = []ecs.PortMapping{
			{
				ContainerPort: conv.U16(dc.deployFlags.ContainerPort),
				Protocol:      "tcp",
			},
		}
	}

	ecsTaskDefinitionName := conv.S(dc.deployFlags.AppName)
	ecsTaskContainerName := conv.S(dc.deployFlags.AppName)
	cpu := conv.U64(dc.deployFlags.CPU)
	memory := conv.U64(dc.deployFlags.Memory)
	useCloudWatchLogs := *dc.deployFlags.CloudWatchLogs

	ecsTaskDef, err := dc.awsClient.ECS().UpdateTaskDefinition(
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

func (dc *DeployCommand) createOrUpdateECSService(ecsTaskDefinitionARN string) error {
	ecsClusterName := conv.S(dc.deployFlags.ECSClusterName)
	ecsServiceName := conv.S(dc.deployFlags.AppName)

	ecsService, err := dc.awsClient.ECS().RetrieveService(ecsClusterName, ecsServiceName)
	if err != nil {
		return fmt.Errorf("Failed to retrieve ECS service [%s/%s]: %s", ecsClusterName, ecsServiceName, err.Error())
	}

	if ecsService != nil && *ecsService.Status != "INACTIVE" {
		if err := dc.updateECSService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN); err != nil {
			return err
		}
	} else {
		if err := dc.createECSService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN); err != nil {
			return err
		}
	}

	return nil
}

func (dc *DeployCommand) createECSService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN string) error {
	// TODO: support multiple load balancers

	loadBalancerName := conv.S(dc.deployFlags.LoadBalancerName)
	ecsServiceRoleName := conv.S(dc.deployFlags.ECSServiceRoleName)
	ecsTaskContainerName := conv.S(dc.deployFlags.AppName) // TODO: implicitly known; should have been retrieved from task definition
	ecsTaskContainerPort := conv.U16(dc.deployFlags.ContainerPort)

	var loadBalancers []*ecs.LoadBalancer
	if !utils.IsBlank(loadBalancerName) {
		// prepare load balancer (create one if needed)
		loadBalancer, err := dc.prepareELBLoadBalancer(loadBalancerName, ecsServiceRoleName, ecsTaskContainerName, ecsTaskContainerPort)
		if err != nil {
			return err
		}
		loadBalancers = []*ecs.LoadBalancer{loadBalancer}

		// prepare ECS service role (create one if needed)
		if err := dc.prepareECSServiceRole(ecsServiceRoleName); err != nil {
			return err
		}
	}

	desiredUnits := conv.U16(dc.deployFlags.Units)

	ecsService, err := dc.awsClient.ECS().CreateService(
		ecsClusterName, ecsServiceName, ecsTaskDefinitionARN, desiredUnits,
		loadBalancers, ecsServiceRoleName)
	if err != nil {
		return fmt.Errorf("Failed to create ECS service [%s/%s]: %s", ecsClusterName, ecsServiceName, err.Error())
	}

	console.Println("ECS service created", conv.S(ecsService.ServiceName), conv.S(ecsService.ServiceArn))

	return nil
}

func (dc *DeployCommand) prepareELBLoadBalancer(lbName string, ecsServiceRoleName, ecsTaskContainerName string, ecsTaskContainerPort uint16) (*ecs.LoadBalancer, error) {
	elbTargetGroupName := lbName

	elbDesc, err := dc.awsClient.ELB().RetrieveLoadBalancer(lbName)
	if err != nil {
		fmt.Errorf("Failed to retrieve ELB load balancer [%s]: %s", lbName, err.Error())
	}
	if elbDesc != nil {
		elbTargetGroup, err := dc.awsClient.ELB().RetrieveTargetGroup(
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
		elbTargetGroupARN, err := dc.createELBTargetGroup(lbName)
		if err != nil {
			return nil, err
		}

		// load balancer port
		lbPort := uint16(80)

		securityGroupName := fmt.Sprintf("elb-%s", lbName)
		securityGroupID, err := dc.awsClient.EC2().CreateSecurityGroup(securityGroupName, securityGroupName, conv.S(dc.deployFlags.VPCID))
		if err != nil {
			return nil, fmt.Errorf("Failed to create security group [%s]: %s", securityGroupName, err.Error())
		}
		err = dc.awsClient.EC2().AddInboundToSecurityGroup(securityGroupID, ec2.SecurityGroupProtocolTCP, lbPort, lbPort, "0.0.0.0/0")
		if err != nil {
			return nil, fmt.Errorf("Failed to add incoming rule to security group [%s]: %s", securityGroupID, err.Error())
		}

		// create ELB load balancer
		elbLoadBalancerARN, err := dc.createELBLoadBalancer(lbName, securityGroupID)
		if err != nil {
			return nil, err
		}

		err = dc.awsClient.ELB().CreateListener(elbLoadBalancerARN, elbTargetGroupARN, lbPort, "HTTP")
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

func (dc *DeployCommand) createELBTargetGroup(targetGroupName string) (string, error) {
	healthCheck := &elb.HealthCheckParams{
		CheckIntervalSeconds:    30,
		CheckPath:               conv.S(dc.deployFlags.HealthCheckPath),
		Protocol:                "HTTP",
		ExpectedHTTPStatusCodes: "200-299",
		CheckTimeoutSeconds:     10,
		HealthyThresholdCount:   5,
		UnhealthyThresholdCount: 2,
	}

	vpcID := conv.S(dc.deployFlags.VPCID)

	targetGroup, err := dc.awsClient.ELB().CreateTargetGroup(targetGroupName, 80, "HTTP", vpcID, healthCheck)
	if err != nil {
		return "", fmt.Errorf("Failed to create ELB target group [%s]: %s", targetGroupName, err.Error())
	}

	return conv.S(targetGroup.TargetGroupArn), nil
}

func (dc *DeployCommand) createELBLoadBalancer(loadBalancerName, securityGroupID string) (string, error) {
	vpcID := conv.S(dc.deployFlags.VPCID)
	subnetIDs, err := dc.awsClient.EC2().ListVPCSubnets(vpcID)
	if err != nil {
		return "", fmt.Errorf("Failed to list subnets: %s", err.Error())
	}

	lb, err := dc.awsClient.ELB().CreateLoadBalancer(loadBalancerName, true, []string{securityGroupID}, subnetIDs)
	if err != nil {
		return "", fmt.Errorf("Failed to create ELB load balancer [%s]: %s", loadBalancerName, err.Error())
	}

	return conv.S(lb.LoadBalancerArn), nil
}

func (dc *DeployCommand) updateECSService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN string) error {
	desiredUnits := conv.U16(dc.deployFlags.Units)

	ecsService, err := dc.awsClient.ECS().UpdateService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN, desiredUnits)
	if err != nil {
		return fmt.Errorf("Failed to update ECS service [%s/%s]: %s", ecsClusterName, ecsServiceName, err.Error())
	}

	console.Println("ECS service updated", conv.S(ecsService.ServiceName), conv.S(ecsService.ServiceArn))

	return nil
}

func (dc *DeployCommand) populateEnvVars() (map[string]string, error) {
	// load from envs file (JSON)
	envs := make(map[string]string)
	if !utils.IsBlank(conv.S(dc.deployFlags.EnvsFile)) {
		data, err := ioutil.ReadFile(conv.S(dc.deployFlags.EnvsFile))
		if err != nil {
			return nil, fmt.Errorf("Failed to read envs file [%s]: %s", conv.S(dc.deployFlags.EnvsFile), err.Error())
		}

		if err := json.Unmarshal(data, &envs); err != nil {
			return nil, fmt.Errorf("Failed to parse envs file: %s", err.Error())
		}
	}

	// override from CLI params
	if dc.deployFlags.Envs != nil {
		for ek, ev := range *dc.deployFlags.Envs {
			envs[ek] = ev
		}
	}

	return envs, nil
}
