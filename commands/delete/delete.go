package delete

import (
	"io/ioutil"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/config"
	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"github.com/d5/cc"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Command struct {
	globalFlags  *flags.GlobalFlags
	commandFlags *Flags
	awsClient    *aws.Client
}

func (c *Command) Init(ka *kingpin.Application, globalFlags *flags.GlobalFlags) *kingpin.CmdClause {
	c.globalFlags = globalFlags

	cmd := ka.Command("delete", "(delete description goes here)")
	c.commandFlags = NewFlags(cmd)

	return cmd
}

func (c *Command) Run() error {
	c.awsClient = c.globalFlags.GetAWSClient()

	// configuration
	configFilePath, err := c.globalFlags.GetConfigFile()
	if err != nil {
		return core.ExitWithError(err)
	}
	configData, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return core.ExitWithErrorString("Failed to read configuration file [%s]: %s", configFilePath, err.Error())
	}
	conf, err := config.Load(configData, conv.S(c.globalFlags.ConfigFileFormat))
	if err != nil {
		return core.ExitWithError(err)
	}

	console.Println("Identifying resources to delete...")
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
		return core.ExitWithErrorString("Failed to retrieve ECS Service [%s/%s]: %s", ecsClusterName, ecsServiceName, err.Error())
	}
	if ecsService != nil && conv.S(ecsService.Status) == "ACTIVE" {
		deleteECSService = true
		console.Println(" ", cc.BlackH("ECS Service"), cc.Green(ecsServiceName))
	}

	// ELB Load Balancer
	elbLoadBalancerName := conv.S(conf.AWS.ELBLoadBalancerName)
	elbLoadBalancer, err := c.awsClient.ELB().RetrieveLoadBalancer(elbLoadBalancerName)
	if err != nil {
		return core.ExitWithErrorString("Failed to retrieve ELB Load Balancer [%s]: %s", elbLoadBalancerName, err.Error())
	}
	if elbLoadBalancer != nil {
		deleteELBLoadBalancer = true
		console.Println(" ", cc.BlackH("ELB Load Balancer"), cc.Green(elbLoadBalancerName))
	}

	// ELB Target Group
	elbTargetGroupName := conv.S(conf.AWS.ELBTargetGroupName)
	elbTargetGroup, err := c.awsClient.ELB().RetrieveTargetGroupByName(elbTargetGroupName)
	if err != nil {
		return core.ExitWithErrorString("Failed to retrieve ELB Target Group [%s]: %s", elbTargetGroupName, err.Error())
	}
	if elbTargetGroup != nil {
		deleteELBTargetGroup = true
		console.Println(" ", cc.BlackH("ELB Target Group"), cc.Green(elbTargetGroupName))
	}

	// ELB Load Balancer Security Group
	elbLoadBalancerSecurityGroupName := conv.S(conf.AWS.ELBSecurityGroup)
	elbLoadBalancerSecurityGroup, err := c.awsClient.EC2().RetrieveSecurityGroupByName(elbLoadBalancerSecurityGroupName)
	if err != nil {
		return core.ExitWithErrorString("Failed to retrieve EC2 Security Group [%s]: %s", elbLoadBalancerSecurityGroupName, err.Error())
	}
	if elbLoadBalancerSecurityGroup != nil {
		deleteELBLoadBalancerSecurityGroup = true
		console.Println(" ", cc.BlackH("ELB Load Balancer Security Group"), cc.Green(elbLoadBalancerSecurityGroupName))
	}

	// ECR Repository
	ecrRepositoryName := conv.S(conf.AWS.ECRRepositoryName)
	ecrRepository, err := c.awsClient.ECR().RetrieveRepository(ecrRepositoryName)
	if err != nil {
		return core.ExitWithErrorString("Failed to retrieve ECR Repository [%s]: %s", ecrRepositoryName, err.Error())
	}
	if ecrRepository != nil {
		deleteECRRepository = true
		console.Println(" ", cc.BlackH("ECR Repository"), cc.Green(ecrRepositoryName))
	}

	if !deleteECSService && !deleteELBLoadBalancerSecurityGroup && !deleteELBLoadBalancer &&
		!deleteELBTargetGroup && !deleteECRRepository {
		console.Println("Looks like everything's already cleaned up.")
		return nil
	}

	// confirmation
	if !conv.B(c.commandFlags.ForceDelete) && !console.AskConfirm("Do you want to delete these resources?", false) {
		return nil
	}

	// Delete ECS Service
	if deleteECSService {
		console.Printf("Deleting ECS Service [%s]...\n", cc.Red(ecsServiceName))

		// update ECS Service units = 0
		_, err := c.awsClient.ECS().UpdateService(ecsClusterName, ecsServiceName, conv.S(ecsService.TaskDefinition), 0)
		if err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(cc.Red("Error:"), err.Error())
			} else {
				return core.ExitWithError(err)
			}
		} else {
			// delete ECS Service
			if err := c.awsClient.ECS().DeleteService(ecsServiceName, ecsServiceName); err != nil {
				if conv.B(c.commandFlags.ContinueOnError) {
					console.Errorln(cc.Red("Error:"), err.Error())
				} else {
					return core.ExitWithError(err)
				}
			}
		}
	}

	// delete ELB Load Balancer
	if deleteELBLoadBalancer {
		console.Printf("Deleting ELB Load Balancer [%s]...\n", cc.Red(elbLoadBalancerName))

		if err := c.awsClient.ELB().DeleteLoadBalancer(conv.S(elbLoadBalancer.LoadBalancerArn)); err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(cc.Red("Error:"), err.Error())
			} else {
				return core.ExitWithError(err)
			}
		}
	}

	// delete ELB Target Group {
	if deleteELBTargetGroup {
		console.Printf("Deleting ELB Target Group [%s]...\n", cc.Red(elbTargetGroupName))

		if err := c.awsClient.ELB().DeleteTargetGroup(conv.S(elbTargetGroup.TargetGroupArn)); err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(cc.Red("Error:"), err.Error())
			} else {
				return core.ExitWithError(err)
			}
		}
	}

	// delete ELB Load Balancer Security Group
	if deleteELBLoadBalancerSecurityGroup {
		console.Printf("Deleting ELB Load Balancer Security Group [%s]...\n", cc.Red(elbLoadBalancerSecurityGroupName))

		if err := c.awsClient.EC2().DeleteSecurityGroup(conv.S(elbLoadBalancerSecurityGroup.GroupId)); err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(cc.Red("Error:"), err.Error())
			} else {
				return core.ExitWithError(err)
			}
		}
	}

	// delete ECR Repository
	if deleteECRRepository {
		console.Printf("Deleting ECR Repository [%s]...\n", cc.Red(ecrRepositoryName))

		if err := c.awsClient.ECR().DeleteRepository(ecrRepositoryName); err != nil {
			if conv.B(c.commandFlags.ContinueOnError) {
				console.Errorln(cc.Red("Error:"), err.Error())
			} else {
				return core.ExitWithError(err)
			}
		}
	}

	return nil
}
