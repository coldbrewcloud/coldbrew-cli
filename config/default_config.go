package config

import (
	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

func DefaultConfig(appName string) *Config {
	conf := new(Config)

	conf.Name = conv.SP(appName)
	conf.ClusterName = conv.SP("cluster1")
	conf.Port = conv.U16P(80)
	conf.CPU = conv.F64P(0.5)
	conf.Memory = conv.SP("500m")
	conf.Units = conv.U16P(1)

	// Environment variables
	conf.Env = make(map[string]string)

	// load balancer
	conf.LoadBalancer.Enabled = conv.BP(false)
	conf.LoadBalancer.IsHTTPS = conv.BP(false)
	conf.LoadBalancer.Port = conv.U16P(80)

	// health check
	conf.LoadBalancer.HealthCheck.Path = conv.SP("/")
	conf.LoadBalancer.HealthCheck.Status = conv.SP("200-299")
	conf.LoadBalancer.HealthCheck.Interval = conv.SP("15s")
	conf.LoadBalancer.HealthCheck.Timeout = conv.SP("10s")
	conf.LoadBalancer.HealthCheck.HealthyLimit = conv.U16P(3)
	conf.LoadBalancer.HealthCheck.UnhealthyLimit = conv.U16P(3)

	// AWS
	{
		// ELB name: cannot exceed 32 chars
		elbLoadBalancerName := ""
		if len(appName) > 28 {
			elbLoadBalancerName = core.DefaultELBLoadBalancerName(appName[:28])
		} else {
			elbLoadBalancerName = core.DefaultELBLoadBalancerName(appName)
		}
		conf.AWS.ELBLoadBalancerName = conv.SP(elbLoadBalancerName)

		// ELB target group name: cannot exceed 32 chars
		elbLoadBalancerTargetGroupName := ""
		if len(appName) > 25 {
			elbLoadBalancerTargetGroupName = core.DefaultELBTargetGroupName(appName[:25])
		} else {
			elbLoadBalancerTargetGroupName = core.DefaultELBTargetGroupName(appName)
		}
		conf.AWS.ELBTargetGroupName = conv.SP(elbLoadBalancerTargetGroupName)

		// ELB security group
		elbSecurityGroupName := ""
		if len(appName) > 25 {
			elbSecurityGroupName = core.DefaultELBLoadBalancerSecurityGroupName(appName[:25])
		} else {
			elbSecurityGroupName = core.DefaultELBLoadBalancerSecurityGroupName(appName)
		}
		conf.AWS.ELBSecurityGroupName = conv.SP(elbSecurityGroupName)

		// ECR Repository name
		conf.AWS.ECRRepositoryName = conv.SP(core.DefaultECRRepository(appName))
	}

	// Docker
	conf.Docker.Bin = conv.SP("docker")

	return conf
}
