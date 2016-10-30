package deploy

import (
	"errors"
	"fmt"
	"math"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/aws/ecs"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
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

	// logging
	loggingDriver := conv.S(c.conf.Logging.Driver)
	if c.conf.Logging.Options == nil {
		c.conf.Logging.Options = make(map[string]string)
	}
	switch loggingDriver {
	case aws.ECSTaskDefinitionLogDriverAWSLogs:
		// test if group needs to be created
		awsLogsGroupName, ok := c.conf.Logging.Options["awslogs-group"]
		if !ok || utils.IsBlank(awsLogsGroupName) {
			awsLogsGroupName = core.DefaultCloudWatchLogsGroupName(conv.S(c.conf.Name), conv.S(c.conf.ClusterName))
			c.conf.Logging.Options["awslogs-group"] = awsLogsGroupName
		}
		if err := c.PrepareCloudWatchLogsGroup(awsLogsGroupName); err != nil {
			return "", err
		}

		// assign region if not provided
		awsLogsRegionName, ok := c.conf.Logging.Options["awslogs-region"]
		if !ok || utils.IsBlank(awsLogsRegionName) {
			c.conf.Logging.Options["awslogs-region"] = conv.S(c.globalFlags.AWSRegion)
		}
	}

	console.UpdatingResource("Updating ECS Task Definition", ecsTaskDefinitionName, false)
	ecsTaskDef, err := c.awsClient.ECS().UpdateTaskDefinition(
		ecsTaskDefinitionName,
		dockerImageFullURI,
		ecsTaskContainerName,
		cpu,
		memory,
		c.conf.Env,
		portMappings,
		loggingDriver, c.conf.Logging.Options)
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

			// check if task container port has changed or not
			if conv.I64(ecsService.LoadBalancers[0].ContainerPort) != int64(conv.U16(c.conf.Port)) {
				return core.NewErrorExtraInfo(
					errors.New("App port cannot be changed."),
					"https://github.com/coldbrewcloud/coldbrew-cli/wiki/Configuration-Changes-and-Their-Effects#app-level-changes")
			}
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
	if elbTargetGroupARN != "" {
		if err := c.checkLoadBalancerHealthCheckChanges(elbTargetGroupARN); err != nil {
			return err
		}
	}

	// update ECS service
	console.UpdatingResource("Updating ECS Service", ecsServiceName, false)
	_, err := c.awsClient.ECS().UpdateService(ecsClusterName, ecsServiceName, ecsTaskDefinitionARN, conv.U16(c.conf.Units))
	if err != nil {
		return fmt.Errorf("Failed to update ECS Service [%s]: %s", ecsServiceName, err.Error())
	}

	return nil
}

func (c *Command) PrepareCloudWatchLogsGroup(groupName string) error {
	groups, err := c.awsClient.CloudWatchLogs().ListGroups(groupName)
	if err != nil {
		return fmt.Errorf("Failed to list CloudWatch Logs Group [%s]: %s", groupName, err.Error())
	}

	for _, group := range groups {
		if conv.S(group.LogGroupName) == groupName {
			// log group exists; return with no error
			return nil
		}
	}

	// log group does not exist; create a new group
	console.AddingResource("Creating CloudWatch Logs Group", groupName, false)
	if err := c.awsClient.CloudWatchLogs().CreateGroup(groupName); err != nil {
		return fmt.Errorf("Failed to create CloudWatch Logs Group [%s]: %s", groupName, err.Error())
	}

	return nil
}
