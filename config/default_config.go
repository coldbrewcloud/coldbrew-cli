package config

import (
	"fmt"

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
	conf.LoadBalancer.HealthCheck.Interval = conv.SP("30s")
	conf.LoadBalancer.HealthCheck.Timeout = conv.SP("10s")
	conf.LoadBalancer.HealthCheck.HealthyLimit = conv.U16P(5)
	conf.LoadBalancer.HealthCheck.UnhealthyLimit = conv.U16P(5)

	// AWS
	{
		// ELB name: cannot exceed 32 chars
		elbLoadBalancerName := ""
		if len(appName) > 28 {
			elbLoadBalancerName = fmt.Sprintf("%s-elb", appName[:28])
		} else {
			elbLoadBalancerName = fmt.Sprintf("%s-elb", appName)
		}
		conf.AWS.ELBLoadBalancerName = conv.SP(elbLoadBalancerName)

		// ELB target group name: cannot exceed 32 chars
		elbLoadBalancerTargetGroupName := ""
		if len(appName) > 25 {
			elbLoadBalancerTargetGroupName = fmt.Sprintf("%s-elb-tg", appName[:25])
		} else {
			elbLoadBalancerTargetGroupName = fmt.Sprintf("%s-elb-tg", appName)
		}
		conf.AWS.ELBTargetGroupName = conv.SP(elbLoadBalancerTargetGroupName)

		// ELB security group
		conf.AWS.ELBSecurityGroup = nil

		// ECR Repository name
		conf.AWS.ECRRepositoryName = conv.SP(fmt.Sprintf("coldbrew/%s", conv.S(conf.Name)))
	}

	// Docker
	conf.Docker.Bin = conv.SP("docker")

	return conf
}
