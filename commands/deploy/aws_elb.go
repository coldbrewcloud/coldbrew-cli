package deploy

import (
	"fmt"

	"github.com/coldbrewcloud/coldbrew-cli/aws/ec2"
	"github.com/coldbrewcloud/coldbrew-cli/aws/ecs"
	"github.com/coldbrewcloud/coldbrew-cli/aws/elb"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

func (c *Command) prepareELBLoadBalancer(ecsServiceRoleName, ecsTaskContainerName string, ecsTaskContainerPort uint16) (*ecs.LoadBalancer, error) {
	elbLoadBalancerName := conv.S(c.conf.AWS.ELBLoadBalancerName)
	elbTargetGroupName := conv.S(c.conf.AWS.ELBTargetGroupName)
	elbPort := conv.U16(c.conf.LoadBalancer.Port)

	_, vpcID, err := c.globalFlags.GetAWSRegionAndVPCID()
	if err != nil {
		return nil, err
	}

	// Check if specified ELB Load Balancer exists or not.
	elbLoadBalancer, err := c.awsClient.ELB().RetrieveLoadBalancer(elbLoadBalancerName)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve ELB Load Balancer [%s]: %s", elbLoadBalancerName, err.Error())
	}

	if elbLoadBalancer != nil {
		// ELB Load Balancer exists
		// Check if specified ELB Target Group also exists
		//  -> if exists, make sure ELB Load Balancer and ELB Target Group actually has a listener between them
		//  -> if not exists, create new ELB Target Group and a new listener.

		elbTargetGroup, err := c.awsClient.ELB().RetrieveTargetGroupByName(elbTargetGroupName)
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve ELB Target Group [%s]: %s", elbTargetGroupName, err.Error())
		}
		if elbTargetGroup != nil {
			// ELB Target Group also exists.
			// Test if listener between ELB Load Balancer and ELB Target Group actually exists
			listenerExists := false
			listeners, err := c.awsClient.ELB().RetrieveLoadBalancerListeners(conv.S(elbLoadBalancer.LoadBalancerArn))
			if err != nil {
				return nil, fmt.Errorf("Failed to retrieve listeners for ELB Load Balanacer [%s]: %s", elbLoadBalancerName, err.Error())
			}
		loop1:
			for _, l := range listeners {
				for _, a := range l.DefaultActions {
					if conv.S(a.TargetGroupArn) == conv.S(elbTargetGroup.TargetGroupArn) {
						listenerExists = true
						break loop1
					}
				}
			}
			if !listenerExists {
				return nil, fmt.Errorf("ELB Load Balancer [%s] does not have a listener to ELB Target Group [%s].", elbLoadBalancerName, elbTargetGroupName)
			}

			return &ecs.LoadBalancer{
				ELBTargetGroupARN: conv.S(elbTargetGroup.TargetGroupArn),
				TaskContainerName: ecsTaskContainerName,
				TaskContainerPort: ecsTaskContainerPort,
			}, nil
		} else {
			// ELB Target Group does not exist; Create a new one.
			console.AddingResource("Creating ELB Target Group", elbTargetGroupName, false)
			elbTargetGroupARN, err := c.createELBTargetGroup(elbTargetGroupName)
			if err != nil {
				return nil, err
			}

			// create a new listen between ELB Load Balancer and ELB Target Group
			console.AddingResource("Adding a listenr for ELB Load Balancer", elbLoadBalancerName, false)
			err = c.awsClient.ELB().CreateListener(
				conv.S(elbLoadBalancer.LoadBalancerArn),
				elbTargetGroupARN,
				elbPort, "HTTP")
			if err != nil {
				return nil, fmt.Errorf("Failed to create ELB Listener: %s", err.Error())
			}

			return &ecs.LoadBalancer{
				ELBTargetGroupARN: elbTargetGroupARN,
				TaskContainerName: ecsTaskContainerName,
				TaskContainerPort: ecsTaskContainerPort,
			}, nil
		}
	} else {
		// ELB Load Balancer does not exist.
		// Create a new EC2 Security Group for ELB Load Balancer.
		// Create a new ELB Target Group
		// Create a new ELB Load Balancer.
		// Create a new listener between ELB Load Balanacer and ELB Target Group.

		// create ELB target group
		console.AddingResource("Creating ELB Target Group", elbTargetGroupName, false)
		elbTargetGroupARN, err := c.createELBTargetGroup(elbTargetGroupName)
		if err != nil {
			return nil, err
		}

		// create security group for ELB (if needed)
		elbSecurityGroupID, err := c.prepareLoadBalancerSecurityGroup(vpcID)
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
		return "", fmt.Errorf("Failed to create ELB Target Group [%s]: %s", targetGroupName, err.Error())
	}
	if err := c.awsClient.ELB().CreateTags(conv.S(targetGroup.TargetGroupArn), core.DefaultTagsForAWSResources()); err != nil {
		return "", fmt.Errorf("Failed to tag ELB Target Group [%s]: %s", targetGroupName, err.Error())
	}

	return conv.S(targetGroup.TargetGroupArn), nil
}

func (c *Command) createELBLoadBalancer(name, vpcID, securityGroupID string) (string, error) {
	subnetIDs, err := c.awsClient.EC2().ListVPCSubnets(vpcID)
	if err != nil {
		return "", fmt.Errorf("Failed to list subnets: %s", err.Error())
	}

	console.AddingResource("Creating ELB Load Balanacer", name, false)
	lb, err := c.awsClient.ELB().CreateLoadBalancer(name, true, []string{securityGroupID}, subnetIDs)
	if err != nil {
		return "", fmt.Errorf("Failed to create ELB Load Balancer [%s]: %s", name, err.Error())
	}
	if err := c.awsClient.ELB().CreateTags(conv.S(lb.LoadBalancerArn), core.DefaultTagsForAWSResources()); err != nil {
		return "", fmt.Errorf("Failed to tag ELB Load Balancer [%s]: %s", name, err.Error())
	}

	return conv.S(lb.LoadBalancerArn), nil
}

func (c *Command) prepareLoadBalancerSecurityGroup(vpcID string) (string, error) {
	elbSecurityGroupName := conv.S(c.conf.AWS.ELBSecurityGroupName)
	elbPort := conv.U16(c.conf.LoadBalancer.Port)

	securityGroup, err := c.awsClient.EC2().RetrieveSecurityGroupByNameOrID(elbSecurityGroupName)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve EC2 Security Group [%s]: %s", elbSecurityGroupName, err.Error())
	}
	if securityGroup == nil {
		// create a new one if specified security group does not exists
		return c.createLoadBalancerSecurityGroup(vpcID, elbPort, elbSecurityGroupName)
	}

	return conv.S(securityGroup.GroupId), nil
}

func (c *Command) createLoadBalancerSecurityGroup(vpcID string, elbPort uint16, securityGroupName string) (string, error) {
	console.AddingResource("Creating EC2 Security Group", securityGroupName, false)
	securityGroupID, err := c.awsClient.EC2().CreateSecurityGroup(securityGroupName, securityGroupName, vpcID)
	if err != nil {
		return "", fmt.Errorf("Failed to create EC2 Security Group [%s]: %s", securityGroupName, err.Error())
	}
	if err := c.awsClient.EC2().CreateTags(securityGroupID, core.DefaultTagsForAWSResources()); err != nil {
		return "", fmt.Errorf("Failed to tag EC2 Security Group [%s]: %s", securityGroupName, err.Error())
	}

	// add load balancer inbound rule
	console.UpdatingResource(fmt.Sprintf("Adding inbound rule [%s:%d:%s] to EC2 Security Group",
		ec2.SecurityGroupProtocolTCP, elbPort, "0.0.0.0/0"),
		securityGroupName, false)
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

	console.UpdatingResource(fmt.Sprintf("Adding inbound rule [%s:%d:%s] to EC2 Security Group",
		ec2.SecurityGroupProtocolTCP, 0, securityGroupID),
		ecsInstancesSecurityGroupName, false)
	err = c.awsClient.EC2().AddInboundToSecurityGroup(
		conv.S(ecsInstancesSecurityGroup.GroupId),
		ec2.SecurityGroupProtocolTCP,
		0, 0, securityGroupID)
	if err != nil {
		return "", fmt.Errorf("Failed to add inbound rule to EC2 Security Group [%s]: %s", ecsInstancesSecurityGroupName, err.Error())
	}

	return securityGroupID, nil
}
