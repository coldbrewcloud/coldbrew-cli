package status

import (
	"io/ioutil"

	"strings"

	"fmt"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/config"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
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

	cmd := ka.Command("status", "(status description goes here)")
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
	{
		// ECS cluster
		ecsClusterName := core.DefaultECSClusterName(clusterName)
		ecsCluster, err := c.awsClient.ECS().RetrieveCluster(ecsClusterName)
		if err != nil {
			return console.ExitWithErrorString("Failed to retrieve ECS Cluster [%s]: %s", ecsClusterName, err.Error())
		}
		if ecsCluster == nil || conv.S(ecsCluster.Status) != "ACTIVE" {
			ecsCluster = nil
			console.DetailWithResourceNote("ECS Cluster", ecsClusterName, "(not found)", true)
		} else {
			console.DetailWithResource("ECS Cluster", ecsClusterName)
		}

		// ECS Task Definition
		ecsTaskDefinitionName := core.DefaultECSTaskDefinitionName(appName)
		ecsTaskDefinition, err := c.awsClient.ECS().RetrieveTaskDefinition(ecsTaskDefinitionName)
		if err != nil {
			return console.ExitWithErrorString("Failed to retrieve ECS Task Definition [%s]: %s", ecsTaskDefinitionName, err.Error())
		}
		if ecsTaskDefinition == nil {
			console.DetailWithResourceNote("ECS Task Definition", ecsTaskDefinitionName, "(not found)", true)
		} else {
			console.DetailWithResource("ECS Task Definition",
				fmt.Sprintf("%s:%d", conv.S(ecsTaskDefinition.Family), conv.I64(ecsTaskDefinition.Revision)))
		}

		// ECS Service
		if ecsCluster != nil {
			ecsServiceName := core.DefaultECSServiceName(appName)
			ecsService, err := c.awsClient.ECS().RetrieveService(ecsClusterName, ecsServiceName)
			if err != nil {
				return console.ExitWithErrorString("Failed to retrieve ECS Service [%s]: %s", ecsServiceName, err.Error())
			}
			if ecsService == nil {
				console.DetailWithResourceNote("ECS Service", ecsServiceName, "(not found)", true)
			} else if conv.S(ecsService.Status) == "ACTIVE" {
				console.DetailWithResource("ECS Service", ecsServiceName)
			} else {
				ecsService = nil
				console.DetailWithResourceNote("ECS Service", ecsServiceName, conv.S(ecsService.Status), true)
			}

			// ECS Service Details
			if ecsService != nil {
				for _, lb := range ecsService.LoadBalancers {
					elbTargetGroup, err := c.awsClient.ELB().RetrieveTargetGroup(conv.S(lb.TargetGroupArn))
					if err != nil {
						return console.ExitWithErrorString("Failed to retrieve ELB Target Group [%s]: %s", conv.S(lb.TargetGroupArn), err.Error())
					}

					console.DetailWithResource("ELB Target Group", fmt.Sprintf("%s:%d",
						conv.S(elbTargetGroup.TargetGroupName),
						conv.I64(lb.ContainerPort)))
				}

				console.DetailWithResource("Tasks (current/desired/pending)", fmt.Sprintf("%d/%d/%d",
					conv.I64(ecsService.RunningCount),
					conv.I64(ecsService.DesiredCount),
					conv.I64(ecsService.PendingCount)))

			}
		}

		// Container Definition
		if ecsTaskDefinition != nil {
			for _, containerDefinition := range ecsTaskDefinition.ContainerDefinitions {
				console.Info("Container")

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
		}

	}

	return nil
}
