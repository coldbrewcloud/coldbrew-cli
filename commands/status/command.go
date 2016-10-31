package status

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
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

	cmd := ka.Command("status",
		"See: "+console.ColorFnHelpLink("https://github.com/coldbrewcloud/coldbrew-cli/wiki/CLI-Command:-status"))
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

	appName := conv.S(conf.Name)
	clusterName := conv.S(conf.ClusterName)
	console.Info("Application")
	console.DetailWithResource("Name", appName)
	console.DetailWithResource("Cluster", clusterName)

	// AWS networking
	regionName, vpcID, err := c.globalFlags.GetAWSRegionAndVPCID()
	if err != nil {
		return console.ExitWithError(err)
	}
	subnetIDs, err := c.awsClient.EC2().ListVPCSubnets(vpcID)
	if err != nil {
		return console.ExitWithErrorString("Failed to list subnets for VPC [%s]: %s", vpcID, err.Error())
	}

	// AWS env
	console.Info("AWS")
	console.DetailWithResource("Region", regionName)
	console.DetailWithResource("VPC", vpcID)
	console.DetailWithResource("Subnets", strings.Join(subnetIDs, " "))

	// ECS
	console.Info("ECS")

	// ECS cluster
	ecsClusterName := core.DefaultECSClusterName(clusterName)
	ecsCluster, err := c.awsClient.ECS().RetrieveCluster(ecsClusterName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve ECS Cluster [%s]: %s", ecsClusterName, err.Error())
	}
	if ecsCluster == nil || conv.S(ecsCluster.Status) != "ACTIVE" {
		console.DetailWithResourceNote("ECS Cluster", ecsClusterName, "(not found)", true)
		return nil // stop here
	} else {
		console.DetailWithResource("ECS Cluster", ecsClusterName)
	}

	// ECS Service
	ecsServiceName := core.DefaultECSServiceName(appName)
	ecsService, err := c.awsClient.ECS().RetrieveService(ecsClusterName, ecsServiceName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve ECS Service [%s]: %s", ecsServiceName, err.Error())
	}
	if ecsService == nil {
		console.DetailWithResourceNote("ECS Service", ecsServiceName, "(not found)", true)
		return nil // stop here
	} else if conv.S(ecsService.Status) == "ACTIVE" {
		console.DetailWithResource("ECS Service", ecsServiceName)
	} else {
		console.DetailWithResourceNote("ECS Service", ecsServiceName, fmt.Sprintf("(%s)", conv.S(ecsService.Status)), true)
		return nil // stop here
	}

	// ECS Task Definition
	ecsTaskDefinitionName := conv.S(ecsService.TaskDefinition)
	ecsTaskDefinition, err := c.awsClient.ECS().RetrieveTaskDefinition(ecsTaskDefinitionName)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve ECS Task Definition [%s]: %s", ecsTaskDefinitionName, err.Error())
	}
	if ecsTaskDefinition == nil {
		console.DetailWithResourceNote("ECS Task Definition", ecsTaskDefinitionName, "(not found)", true)
		return nil // stop here
	} else {
		console.DetailWithResource("ECS Task Definition",
			fmt.Sprintf("%s:%d", conv.S(ecsTaskDefinition.Family), conv.I64(ecsTaskDefinition.Revision)))
	}

	// Tasks count / status
	isDeploying := false
	if ecsService.Deployments != nil {
		for _, d := range ecsService.Deployments {
			switch conv.S(d.Status) {
			case "ACTIVE":
				isDeploying = true
			case "PRIMARY":
			}
		}
	}
	if isDeploying {
		console.DetailWithResourceNote("Tasks (current/desired/pending)", fmt.Sprintf("%d/%d/%d",
			conv.I64(ecsService.RunningCount),
			conv.I64(ecsService.DesiredCount),
			conv.I64(ecsService.PendingCount)),
			"(deploying)", true)
	} else {
		console.DetailWithResource("Tasks (current/desired/pending)", fmt.Sprintf("%d/%d/%d",
			conv.I64(ecsService.RunningCount),
			conv.I64(ecsService.DesiredCount),
			conv.I64(ecsService.PendingCount)))
	}

	// Container Definition
	for _, containerDefinition := range ecsTaskDefinition.ContainerDefinitions {
		console.Info("Container Definition")

		console.DetailWithResource("Name", conv.S(containerDefinition.Name))
		console.DetailWithResource("Image", conv.S(containerDefinition.Image))

		cpu := float64(conv.I64(containerDefinition.Cpu)) / 1024.0
		console.DetailWithResource("CPU", fmt.Sprintf("%.2f", cpu))

		memory := conv.I64(containerDefinition.Memory)
		console.DetailWithResource("Memory", fmt.Sprintf("%dm", memory))

		for _, pm := range containerDefinition.PortMappings {
			console.DetailWithResource("Port Mapping (protocol:container:host)", fmt.Sprintf("%s:%d:%d",
				conv.S(pm.Protocol), conv.I64(pm.ContainerPort), conv.I64(pm.HostPort)))
		}

		for _, ev := range containerDefinition.Environment {
			console.DetailWithResource("Env", fmt.Sprintf("%s=%s",
				conv.S(ev.Name), conv.S(ev.Value)))
		}
	}

	// Tasks
	taskARNs, err := c.awsClient.ECS().ListServiceTaskARNs(ecsClusterName, ecsServiceName)
	if err != nil {
		return console.ExitWithErrorString("Failed to list ECS Tasks for ECS Service [%s]: %s", ecsServiceName, err.Error())
	}
	tasks, err := c.awsClient.ECS().RetrieveTasks(ecsClusterName, taskARNs)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve ECS Tasks for ECS Service [%s]: %s", ecsServiceName, err.Error())
	}

	// retrieve container instance info
	containerInstanceARNs := []string{}
	for _, task := range tasks {
		containerInstanceARNs = append(containerInstanceARNs, conv.S(task.ContainerInstanceArn))
	}
	containerInstances, err := c.awsClient.ECS().RetrieveContainerInstances(ecsClusterName, containerInstanceARNs)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve ECS Container Instances: %s", err.Error())
	}

	// retrieve EC2 Instance info
	ec2InstanceIDs := []string{}
	for _, ci := range containerInstances {
		ec2InstanceIDs = append(ec2InstanceIDs, conv.S(ci.Ec2InstanceId))
	}
	ec2Instances, err := c.awsClient.EC2().RetrieveInstances(ec2InstanceIDs)
	if err != nil {
		return console.ExitWithErrorString("Failed to retrieve EC2 Instances: %s", err.Error())
	}

	for _, task := range tasks {
		console.Info("ECS Task")

		taskDefinition := aws.GetECSTaskDefinitionFamilyAndRevisionFromARN(conv.S(task.TaskDefinitionArn))
		console.DetailWithResource("Task Definition", taskDefinition)

		console.DetailWithResource("Status (current/desired)", fmt.Sprintf("%s/%s",
			conv.S(task.LastStatus), conv.S(task.DesiredStatus)))

		for _, ci := range containerInstances {
			if conv.S(task.ContainerInstanceArn) == conv.S(ci.ContainerInstanceArn) {
				console.DetailWithResource("EC2 Instance ID", conv.S(ci.Ec2InstanceId))

				for _, ec2Instance := range ec2Instances {
					if conv.S(ci.Ec2InstanceId) == conv.S(ec2Instance.InstanceId) {
						if !utils.IsBlank(conv.S(ec2Instance.PrivateIpAddress)) {
							console.DetailWithResource("  Private IP", conv.S(ec2Instance.PrivateIpAddress))
						}
						if !utils.IsBlank(conv.S(ec2Instance.PublicIpAddress)) {
							console.DetailWithResource("  Public IP", conv.S(ec2Instance.PublicIpAddress))
						}
						break
					}
				}
				break
			}
		}
	}

	// Load Balancer
	if ecsService.LoadBalancers != nil && len(ecsService.LoadBalancers) > 0 {
		for _, lb := range ecsService.LoadBalancers {
			console.Info("Load Balancer")

			elbTargetGroup, err := c.awsClient.ELB().RetrieveTargetGroup(conv.S(lb.TargetGroupArn))
			if err != nil {
				return console.ExitWithErrorString("Failed to retrieve ELB Target Group [%s]: %s", conv.S(lb.TargetGroupArn), err.Error())
			}

			console.DetailWithResource("Container Port", fmt.Sprintf("%d", conv.I64(lb.ContainerPort)))
			console.DetailWithResource("ELB Target Group", conv.S(elbTargetGroup.TargetGroupName))

			if elbTargetGroup.LoadBalancerArns != nil {
				for _, elbARN := range elbTargetGroup.LoadBalancerArns {
					elbLoadBalancer, err := c.awsClient.ELB().RetrieveLoadBalancer(conv.S(elbARN))
					if err != nil {
						return console.ExitWithErrorString("Failed to retrieve ELB Load Balancer [%s]: %s", elbARN, err.Error())
					}

					console.DetailWithResource("ELB Load Balancer", conv.S(elbLoadBalancer.LoadBalancerName))
					console.DetailWithResource("  Scheme", conv.S(elbLoadBalancer.Scheme))
					//console.DetailWithResource("  DNS", conv.S(elbLoadBalancer.DNSName))
					if elbLoadBalancer.State != nil {
						console.DetailWithResource("  State", fmt.Sprintf("%s %s",
							conv.S(elbLoadBalancer.State.Code),
							conv.S(elbLoadBalancer.State.Reason)))
					}

					// listeners
					listeners, err := c.awsClient.ELB().RetrieveLoadBalancerListeners(conv.S(elbARN))
					if err != nil {
						return console.ExitWithErrorString("Failed to retrieve Listeners for ELB Load Balancer [%s]: %s", elbARN, err.Error())
					}
					for _, listener := range listeners {
						if listener.DefaultActions != nil &&
							len(listener.DefaultActions) > 0 &&
							conv.S(listener.DefaultActions[0].TargetGroupArn) == conv.S(elbTargetGroup.TargetGroupArn) {
							console.DetailWithResource("  Endpoint", fmt.Sprintf("%s://%s:%d",
								strings.ToLower(conv.S(listener.Protocol)),
								conv.S(elbLoadBalancer.DNSName),
								conv.I64(listener.Port)))
						}
					}
				}
			}
		}
	}

	return nil
}
