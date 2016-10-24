package delete

import (
	"fmt"
	"io/ioutil"
	"time"

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

	// app configuration
	configFilePath, err := c.globalFlags.GetConfigFile()
	if err != nil {
		return console.ExitWithError(err)
	}
	configData, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return console.ExitWithErrorString("Failed to read configuration file [%s]: %s", configFilePath, err.Error())
	}
	conf, err := config.Load(configData, conv.S(c.globalFlags.ConfigFileFormat), core.DefaultAppName(configFilePath))
	if err != nil {
		return console.ExitWithError(err)
	}

	console.Info("Determining AWS resources that need to be deleted...")
	//deleteECSTaskDefinition := false // TODO: should delete ECS task definitions
	deleteECSService := false
	deleteELBLoadBalancer := false
	deleteELBTargetGroup := false
	deleteELBLoadBalancerSecurityGroup := false
	deleteECRRepository := false

	// ECS Service
	ecsClusterName := core.DefaultECSClusterName(conv.S(conf.ClusterName))
	ecsServiceName := core.DefaultECSServiceName(conv.S(conf.Name))
	ecsService, err := c.awsClient.ECS().RetrieveService(ecsClusterName, ecsServiceName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve ECS Service [%s/%s]: %s", ecsClusterName, ecsServiceName, err.Error())
	}
	if ecsService != nil && conv.S(ecsService.Status) == "ACTIVE" {
		deleteECSService = true
		console.DetailWithResource("ECS Service", ecsServiceName)
	}

	// ELB Load Balancer
	elbLoadBalancerName := conv.S(conf.AWS.ELBLoadBalancerName)
	elbLoadBalancer, err := c.awsClient.ELB().RetrieveLoadBalancer(elbLoadBalancerName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve ELB Load Balancer [%s]: %s", elbLoadBalancerName, err.Error())
	}
	if elbLoadBalancer != nil {
		tags, err := c.awsClient.ELB().RetrieveTags(conv.S(elbLoadBalancer.LoadBalancerArn))
		if err != nil {
			return console.ExitWithErrorString("Failed to retrieve tags for ELB Load Balancer [%s]: %s", elbLoadBalancerName, err.Error())
		}
		if _, ok := tags[core.AWSTagNameCreatedTimestamp]; ok {
			deleteELBLoadBalancer = true
			console.DetailWithResource("ELB Load Balancer", elbLoadBalancerName)
		}
	}

	// ELB Target Group
	elbTargetGroupName := conv.S(conf.AWS.ELBTargetGroupName)
	elbTargetGroup, err := c.awsClient.ELB().RetrieveTargetGroupByName(elbTargetGroupName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve ELB Target Group [%s]: %s", elbTargetGroupName, err.Error())
	}
	if elbTargetGroup != nil {
		tags, err := c.awsClient.ELB().RetrieveTags(conv.S(elbTargetGroup.TargetGroupArn))
		if err != nil {
			return console.ExitWithErrorString("Failed to retrieve tags for ELB Target Group [%s]: %s", elbTargetGroupName, err.Error())
		}
		if _, ok := tags[core.AWSTagNameCreatedTimestamp]; ok {
			deleteELBTargetGroup = true
			console.DetailWithResource("ELB Target Group", elbTargetGroupName)
		}
	}

	// ELB Load Balancer Security Group
	elbLoadBalancerSecurityGroupName := conv.S(conf.AWS.ELBSecurityGroupName)
	elbLoadBalancerSecurityGroup, err := c.awsClient.EC2().RetrieveSecurityGroupByName(elbLoadBalancerSecurityGroupName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve EC2 Security Group [%s]: %s", elbLoadBalancerSecurityGroupName, err.Error())
	}
	if elbLoadBalancerSecurityGroup != nil {
		tags, err := c.awsClient.EC2().RetrieveTags(conv.S(elbLoadBalancerSecurityGroup.GroupId))
		if err != nil {
			return console.ExitWithErrorString("Failed to retrieve tags for EC2 Security Group [%s]: %s", elbLoadBalancerSecurityGroupName, err.Error())
		}
		if _, ok := tags[core.AWSTagNameCreatedTimestamp]; ok {
			deleteELBLoadBalancerSecurityGroup = true
			console.DetailWithResource("EC2 Security Group for ELB Load Balancer", elbLoadBalancerSecurityGroupName)
		}
	}

	// ECR Repository
	ecrRepositoryName := conv.S(conf.AWS.ECRRepositoryName)
	ecrRepository, err := c.awsClient.ECR().RetrieveRepository(ecrRepositoryName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve ECR Repository [%s]: %s", ecrRepositoryName, err.Error())
	}
	if ecrRepository != nil {
		deleteECRRepository = true
		console.DetailWithResource("ECR Repository", ecrRepositoryName)
	}

	if !deleteECSService && !deleteELBLoadBalancerSecurityGroup && !deleteELBLoadBalancer &&
		!deleteELBTargetGroup && !deleteECRRepository {
		console.Info("Looks like everything's already cleaned up.")
		return nil
	}

	console.Blank()

	// confirmation
	if !conv.B(c.commandFlags.NoConfirm) && !console.AskConfirm("Do you want to delete these resources?", false) {
		return nil
	}

	console.Blank()

	// Delete ECS Service
	if deleteECSService {
		// update ECS Service units = 0
		console.UpdatingResource("Updating ECS Service to stop all tasks", ecsServiceName, false)
		_, err := c.awsClient.ECS().UpdateService(ecsClusterName, ecsServiceName, conv.S(ecsService.TaskDefinition), 0)
		if err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Error(err.Error())
			} else {
				return console.ExitWithError(err)
			}
		} else {
			console.RemovingResource("Deleting (and draining) ECS Service", ecsServiceName, true)

			// delete ECS Service
			if err := c.awsClient.ECS().DeleteService(ecsClusterName, ecsServiceName); err != nil {
				if conv.B(c.commandFlags.ContinueOnError) {
					console.Error(err.Error())
				} else {
					return console.ExitWithError(err)
				}
			}

			// wait until it becomes fully inactive (from draining status)
			utils.Retry(func() (bool, error) {
				service, err := c.awsClient.ECS().RetrieveService(ecsClusterName, ecsServiceName)
				if err != nil {
					return false, err
				}
				if service == nil || conv.S(service.Status) == "INACTIVE" {
					return false, nil
				}
				return true, nil
			}, time.Second, 5*time.Minute)
		}
	}

	// delete ELB Load Balancer
	if deleteELBLoadBalancer {
		console.RemovingResource("Deleting ELB Load Balancer", elbLoadBalancerName, false)

		if err := c.awsClient.ELB().DeleteLoadBalancer(conv.S(elbLoadBalancer.LoadBalancerArn)); err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Error(err.Error())
			} else {
				return console.ExitWithError(err)
			}
		}
	}

	// delete ELB Target Group {
	if deleteELBTargetGroup {
		console.RemovingResource("Deleting ELB Target Group", elbTargetGroupName, true)

		err := utils.RetryOnAWSErrorCode(func() error {
			return c.awsClient.ELB().DeleteTargetGroup(conv.S(elbTargetGroup.TargetGroupArn))
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
	if deleteELBLoadBalancerSecurityGroup {
		ecsInstancesSecurityGroupName := core.DefaultInstanceSecurityGroupName(conv.S(conf.ClusterName))
		ecsInstancesSecurityGroup, err := c.awsClient.EC2().RetrieveSecurityGroupByName(ecsInstancesSecurityGroupName)
		if err != nil {
			return console.ExitWithErrorString("Failed to retrieve EC2 Security Group [%s]: %s", ecsInstancesSecurityGroupName, err.Error())
		}

		console.RemovingResource(fmt.Sprintf("Removing inbound rule [%s:%d:%s] from EC2 Security Group",
			ec2.SecurityGroupProtocolTCP, 0, conv.S(elbLoadBalancerSecurityGroup.GroupId)),
			ecsInstancesSecurityGroupName, false)
		err = c.awsClient.EC2().RemoveInboundToSecurityGroup(
			conv.S(ecsInstancesSecurityGroup.GroupId),
			ec2.SecurityGroupProtocolTCP,
			0, 65535, conv.S(elbLoadBalancerSecurityGroup.GroupId))
		if err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Error(err.Error())
			} else {
				return console.ExitWithError(err)
			}
		} else {
			console.RemovingResource("Deleting EC2 Security Group for ELB Load Balancer", elbLoadBalancerSecurityGroupName, true)
			err = utils.RetryOnAWSErrorCode(func() error {
				return c.awsClient.EC2().DeleteSecurityGroup(conv.S(elbLoadBalancerSecurityGroup.GroupId))
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
	if deleteECRRepository {
		console.RemovingResource("Deleting ECR Repository", ecrRepositoryName, false)

		if err := c.awsClient.ECR().DeleteRepository(ecrRepositoryName); err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Error(err.Error())
			} else {
				return console.ExitWithError(err)
			}
		}
	}

	return nil
}
