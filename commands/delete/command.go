package delete

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	_ec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/aws/ec2"
	"github.com/coldbrewcloud/coldbrew-cli/config"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Command struct {
	globalFlags  *flags.GlobalFlags
	commandFlags *Flags
	awsClient    *aws.Client
}

func (c *Command) Init(ka *kingpin.Application, globalFlags *flags.GlobalFlags) *kingpin.CmdClause {
	c.globalFlags = globalFlags

	cmd := ka.Command(
		"delete",
		"See: "+console.ColorFnHelpLink("https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-delete"))
	c.commandFlags = NewFlags(cmd)

	return cmd
}

func (c *Command) Run() error {
	c.awsClient = c.globalFlags.GetAWSClient()

	appName := ""
	clusterName := ""

	// app configuration
	configFilePath, err := c.globalFlags.GetConfigFile()
	if err != nil {
		return console.ExitWithError(err)
	}
	if utils.FileExists(configFilePath) {
		configData, err := ioutil.ReadFile(configFilePath)
		if err != nil {
			return console.ExitWithErrorString("Failed to read configuration file [%s]: %s", configFilePath, err.Error())
		}
		conf, err := config.Load(configData, conv.S(c.globalFlags.ConfigFileFormat), core.DefaultAppName(configFilePath))
		if err != nil {
			return console.ExitWithError(err)
		}

		appName = conv.S(conf.Name)
		clusterName = conv.S(conf.ClusterName)
	}

	// app/cluster name from CLI will override configuration file
	if !utils.IsBlank(conv.S(c.commandFlags.AppName)) {
		appName = conv.S(c.commandFlags.AppName)
	}
	if !utils.IsBlank(conv.S(c.commandFlags.ClusterName)) {
		clusterName = conv.S(c.commandFlags.ClusterName)
	}

	if utils.IsBlank(appName) {
		return console.ExitWithErrorString("App name is required.")
	}
	if utils.IsBlank(clusterName) {
		return console.ExitWithErrorString("Cluster name is required.")
	}

	console.Info("Determining AWS resources that need to be deleted...")

	// ECS Service
	var ecsServiceToDelete *ecs.Service
	ecsClusterName := core.DefaultECSClusterName(clusterName)
	ecsServiceName := core.DefaultECSServiceName(appName)
	ecsServiceToDelete, err = c.awsClient.ECS().RetrieveService(ecsClusterName, ecsServiceName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve ECS Service [%s/%s]: %s", ecsClusterName, ecsServiceName, err.Error())
	}
	if ecsServiceToDelete != nil && conv.S(ecsServiceToDelete.Status) == "ACTIVE" {
		console.DetailWithResource("ECS Service", ecsServiceName)
	} else {
		return console.ExitWithErrorString("ECS Service [%s/%s] was not found.", ecsClusterName, ecsServiceName)
	}

	// identify ECR resources to delete
	ecrRepositoryNameToDelete, err := c.identifyECRResourcesToDelete(appName, ecsServiceToDelete)
	if err != nil {
		return console.ExitWithError(err)
	}

	// identify ELB resources to delete
	elbLoadBalancersToDelete, elbTargetGroupsToDelete, elbLoadBalancerSecurityGroupsToDelete, err := c.identifyELBResourcesToDelete(ecsServiceToDelete)
	if err != nil {
		return console.ExitWithError(err)
	}

	if ecsServiceToDelete == nil &&
		len(elbLoadBalancersToDelete) == 0 &&
		len(elbTargetGroupsToDelete) == 0 &&
		len(elbLoadBalancerSecurityGroupsToDelete) == 0 &&
		utils.IsBlank(ecrRepositoryNameToDelete) {
		console.Info("Looks like everything's already cleaned up.")
		return nil
	}

	console.Blank()

	// confirmation
	if !conv.B(c.commandFlags.NoConfirm) && !console.AskConfirm("Do you want to delete these resources?", false) {
		return nil
	}

	console.Blank()

	// update ECS service (desired units => 0)
	console.UpdatingResource("Updating ECS Service to stop all tasks", conv.S(ecsServiceToDelete.ServiceName), false)
	_, err = c.awsClient.ECS().UpdateService(ecsClusterName, ecsServiceName, conv.S(ecsServiceToDelete.TaskDefinition), 0)
	if err != nil {
		// cannot continue with this error
		return console.ExitWithError(err)
	}

	// delete ELB Load Balancer
	for _, elbLoadBalancerToDelete := range elbLoadBalancersToDelete {
		console.RemovingResource("Deleting ELB Load Balancer", conv.S(elbLoadBalancerToDelete.LoadBalancerName), false)

		if err := c.awsClient.ELB().DeleteLoadBalancer(conv.S(elbLoadBalancerToDelete.LoadBalancerArn)); err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Error(err.Error())
			} else {
				return console.ExitWithError(err)
			}
		}
	}

	// delete ELB Target Group {
	for _, elbTargetGroupToDelete := range elbTargetGroupsToDelete {
		console.RemovingResource("Deleting ELB Target Group", conv.S(elbTargetGroupToDelete.TargetGroupName), true)

		err := utils.RetryOnAWSErrorCode(func() error {
			return c.awsClient.ELB().DeleteTargetGroup(conv.S(elbTargetGroupToDelete.TargetGroupArn))
		}, []string{"ResourceInUse"}, time.Second, 1*time.Minute)

		if err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Error(err.Error())
			} else {
				return console.ExitWithError(err)
			}
		}
	}

	// delete ELB Load Balancer Security Group
	for _, elbLoadBalancerSecurityGroupToDelete := range elbLoadBalancerSecurityGroupsToDelete {
		ecsInstancesSecurityGroupName := core.DefaultInstanceSecurityGroupName(clusterName)
		ecsInstancesSecurityGroup, err := c.awsClient.EC2().RetrieveSecurityGroupByName(ecsInstancesSecurityGroupName)
		if err != nil {
			return console.ExitWithErrorString("Failed to retrieve EC2 Security Group [%s]: %s", ecsInstancesSecurityGroupName, err.Error())
		}

		console.RemovingResource(fmt.Sprintf("Removing inbound rule [%s:%d:%s] from EC2 Security Group",
			ec2.SecurityGroupProtocolTCP, 0, conv.S(elbLoadBalancerSecurityGroupToDelete.GroupId)),
			ecsInstancesSecurityGroupName, false)
		err = c.awsClient.EC2().RemoveInboundToSecurityGroup(
			conv.S(ecsInstancesSecurityGroup.GroupId),
			ec2.SecurityGroupProtocolTCP,
			0, 65535, conv.S(elbLoadBalancerSecurityGroupToDelete.GroupId))
		if err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Error(err.Error())
			} else {
				return console.ExitWithError(err)
			}
		} else {
			console.RemovingResource("Deleting EC2 Security Group for ELB Load Balancer", conv.S(elbLoadBalancerSecurityGroupToDelete.GroupName), true)
			err = utils.RetryOnAWSErrorCode(func() error {
				return c.awsClient.EC2().DeleteSecurityGroup(conv.S(elbLoadBalancerSecurityGroupToDelete.GroupId))
			}, []string{"DependencyViolation", "ResourceInUse"}, time.Second, 5*time.Minute)
			if err != nil {
				if conv.B(c.commandFlags.ContinueOnError) {
					console.Error(err.Error())
				} else {
					return console.ExitWithError(err)
				}
			}
		}
	}

	// delete ECR Repository
	if !utils.IsBlank(ecrRepositoryNameToDelete) {
		console.RemovingResource("Deleting ECR Repository", ecrRepositoryNameToDelete, false)

		if err := c.awsClient.ECR().DeleteRepository(ecrRepositoryNameToDelete); err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Error(err.Error())
			} else {
				return console.ExitWithError(err)
			}
		}
	}

	// Delete ECS Service
	console.RemovingResource("Deleting (and draining) ECS Service", conv.S(ecsServiceToDelete.ServiceName), true)
	if err := c.awsClient.ECS().DeleteService(ecsClusterName, conv.S(ecsServiceToDelete.ServiceName)); err != nil {
		if conv.B(c.commandFlags.ContinueOnError) {
			console.Error(err.Error())
		} else {
			return console.ExitWithError(err)
		}
	}

	// wait until it becomes fully inactive (from draining status)
	utils.Retry(func() (bool, error) {
		service, err := c.awsClient.ECS().RetrieveService(ecsClusterName, conv.S(ecsServiceToDelete.ServiceName))
		if err != nil {
			return false, err
		}
		if service == nil || conv.S(service.Status) == "INACTIVE" {
			return false, nil
		}
		return true, nil
	}, time.Second, 5*time.Minute)

	return nil
}

func (c *Command) identifyECRResourcesToDelete(appName string, ecsService *ecs.Service) (string, error) {
	// ECR repository name
	ecsTaskDefinitionARN := conv.S(ecsService.TaskDefinition)
	if !utils.IsBlank(ecsTaskDefinitionARN) {
		ecsTaskDefinition, err := c.awsClient.ECS().RetrieveTaskDefinition(ecsTaskDefinitionARN)
		if err != nil {
			return "", fmt.Errorf("Failed to retrieve ECS Task Definition [%s]: %s", ecsTaskDefinitionARN, err.Error())
		}

		for _, cd := range ecsTaskDefinition.ContainerDefinitions {
			if conv.S(cd.Name) == core.DefaultECSTaskMainContainerName(appName) {
				// extract ECR repository name from Docker image URI
				tokens := strings.SplitN(conv.S(cd.Image), "/", 2)
				if len(tokens) > 1 && !utils.IsBlank(tokens[1]) {
					// strip out docker tag ":tag"
					nameAndTag := strings.SplitN(tokens[1], ":", 2)
					console.DetailWithResource("ECR Repository", nameAndTag[0])
					return nameAndTag[0], nil
				}
			}
		}
	}

	return "", nil
}

func (c *Command) identifyELBResourcesToDelete(ecsService *ecs.Service) ([]*elbv2.LoadBalancer, []*elbv2.TargetGroup, []*_ec2.SecurityGroup, error) {
	elbLoadBalancersToDelete := []*elbv2.LoadBalancer{}
	elbTargetGroupsToDelete := []*elbv2.TargetGroup{}
	elbLoadBalancerSecurityGroupsToDelete := []*_ec2.SecurityGroup{}

	for _, lb := range ecsService.LoadBalancers {
		elbTargetGroupARN := conv.S(lb.TargetGroupArn)
		if !utils.IsBlank(elbTargetGroupARN) {
			elbTargetGroup, err := c.awsClient.ELB().RetrieveTargetGroup(elbTargetGroupARN)
			if err != nil {
				return nil, nil, nil, console.ExitWithErrorString("Failed to retrieve ELB Target Group [%s]: %s", elbTargetGroupARN, err.Error())
			}
			if elbTargetGroup == nil {
				continue
			}

			// check tags
			tags, err := c.awsClient.ELB().RetrieveTags(conv.S(elbTargetGroup.TargetGroupArn))
			if err != nil {
				return nil, nil, nil, console.ExitWithErrorString("Failed to retrieve tags for ELB Target Group [%s]: %s", conv.S(elbTargetGroup.TargetGroupName), err.Error())
			}
			if _, ok := tags[core.AWSTagNameCreatedTimestamp]; ok {
				elbTargetGroupsToDelete = append(elbTargetGroupsToDelete, elbTargetGroup)
				console.DetailWithResource("ELB Target Group", conv.S(elbTargetGroup.TargetGroupName))
			} else {
				continue
			}

			if elbTargetGroup.LoadBalancerArns != nil {
				for _, elbARN := range elbTargetGroup.LoadBalancerArns {
					elbLoadBalancer, err := c.awsClient.ELB().RetrieveLoadBalancer(conv.S(elbARN))
					if err != nil {
						return nil, nil, nil, console.ExitWithErrorString("Failed to retrieve ELB Load Balancer [%s]: %s", elbARN, err.Error())
					}
					if elbLoadBalancer == nil {
						continue
					}

					// check tags
					tags, err := c.awsClient.ELB().RetrieveTags(conv.S(elbARN))
					if err != nil {
						return nil, nil, nil, console.ExitWithErrorString("Failed to retrieve tags for ELB Load Balancer [%s]: %s", conv.S(elbLoadBalancer.LoadBalancerName), err.Error())
					}
					if _, ok := tags[core.AWSTagNameCreatedTimestamp]; ok {
						elbLoadBalancersToDelete = append(elbLoadBalancersToDelete, elbLoadBalancer)
						console.DetailWithResource("ELB Load Balancer", conv.S(elbLoadBalancer.LoadBalancerName))
					} else {
						continue
					}

					// security groups
					for _, securityGroupID := range elbLoadBalancer.SecurityGroups {
						elbLoadBalancerSecurityGroup, err := c.awsClient.EC2().RetrieveSecurityGroup(conv.S(securityGroupID))
						if err != nil {
							return nil, nil, nil, console.ExitWithErrorString("Failed to retrieve EC2 Security Group [%s]: %s", conv.S(securityGroupID), err.Error())
						}
						if elbLoadBalancerSecurityGroup == nil {
							continue
						}

						// check tags
						tags, err := c.awsClient.EC2().RetrieveTags(conv.S(elbLoadBalancerSecurityGroup.GroupId))
						if err != nil {
							return nil, nil, nil, console.ExitWithErrorString("Failed to retrieve tags for EC2 Security Group [%s]: %s", conv.S(elbLoadBalancerSecurityGroup.GroupName), err.Error())
						}
						if _, ok := tags[core.AWSTagNameCreatedTimestamp]; ok {
							elbLoadBalancerSecurityGroupsToDelete = append(elbLoadBalancerSecurityGroupsToDelete, elbLoadBalancerSecurityGroup)
							console.DetailWithResource("EC2 Security Group for ELB Load Balancer", conv.S(elbLoadBalancerSecurityGroup.GroupName))
						}
					}
				}
			}
		}
	}

	return elbLoadBalancersToDelete, elbTargetGroupsToDelete, elbLoadBalancerSecurityGroupsToDelete, nil
}
