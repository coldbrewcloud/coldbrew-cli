package deploy

import (
	"errors"
	"fmt"
	"math"

	"github.com/coldbrewcloud/coldbrew-cli/aws/ecs"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

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

	console.UpdatingResource("Updating ECS Task Definition", ecsTaskDefinitionName, false)
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
		return "", fmt.Errorf("Failed to update ECS Task Definition [%s]: %s", ecsTaskDefinitionName, err.Error())
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
		elbLoadBalancerName := ""
		elbTargetGroupARN := ""
		if ecsService.LoadBalancers != nil && len(ecsService.LoadBalancers) > 0 {
			elbLoadBalancerName = conv.S(ecsService.LoadBalancers[0].LoadBalancerName)
			elbTargetGroupARN = conv.S(ecsService.LoadBalancers[0].TargetGroupArn)
		}

		if err := c.updateECSService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN, elbLoadBalancerName, elbTargetGroupARN); err != nil {
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
	if conv.B(c.conf.LoadBalancer.Enabled) {
		if conv.U16(c.conf.Port) == 0 {
			return errors.New("App port must be specified to enable load balancer.")
		}

		loadBalancer, err := c.prepareELBLoadBalancer(
			ecsServiceRoleName,
			ecsTaskContainerName,
			ecsTaskContainerPort)
		if err != nil {
			return err
		}

		loadBalancers = []*ecs.LoadBalancer{loadBalancer}
	}

	console.AddingResource("Creating ECS Service", ecsServiceName, false)
	_, err := c.awsClient.ECS().CreateService(
		ecsClusterName, ecsServiceName, ecsTaskDefinitionARN, conv.U16(c.conf.Units),
		loadBalancers, ecsServiceRoleName)
	if err != nil {
		return fmt.Errorf("Failed to create ECS Service [%s]: %s", ecsServiceName, err.Error())
	}

	return nil
}

func (c *Command) updateECSService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN, elbLoadBalancerName, elbTargetGroupARN string) error {

	// check if ELB Target Group health check needs to be updated
	if err := c.checkLoadBalancerHealthCheckChanges(elbTargetGroupARN); err != nil {
		return err
	}

	// update ECS service
	console.UpdatingResource("Updating ECS Service", ecsServiceName, false)
	_, err := c.awsClient.ECS().UpdateService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN, conv.U16(c.conf.Units))
	if err != nil {
		return fmt.Errorf("Failed to update ECS Service [%s]: %s", ecsServiceName, err.Error())
	}

	return nil
}
