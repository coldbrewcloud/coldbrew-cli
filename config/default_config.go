package config

import (
	"fmt"
	"path/filepath"

	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

func DefaultConfig(appDirectory string) *Config {
	conf := new(Config)

	conf.Name = conv.SP(filepath.Base(appDirectory))
	conf.ClusterName = conv.SP("cluster1")
	conf.Port = conv.U16P(80)
	conf.CPU = conv.F64P(0.5)
	conf.Memory = conv.SP("500m")
	conf.Units = conv.U16P(1)

	// Environment variables
	conf.Env = make(map[string]string)

	// load balancer
	{
		conf.LoadBalancer = new(ConfigLoadBalancer)
		conf.LoadBalancer.HealthCheck = new(ConfigLoadBalancerHealthCheck)

		conf.LoadBalancer.IsHTTPS = conv.BP(false)
		conf.LoadBalancer.Port = conv.U16P(80)

		// health check
		conf.LoadBalancer.HealthCheck.Path = conv.SP("/")
		conf.LoadBalancer.HealthCheck.Status = conv.SP("200-299")
		conf.LoadBalancer.HealthCheck.Interval = conv.SP("30s")
		conf.LoadBalancer.HealthCheck.Timeout = conv.SP("10s")
		conf.LoadBalancer.HealthCheck.HealthyLimit = conv.U16P(5)
		conf.LoadBalancer.HealthCheck.UnhealthyLimit = conv.U16P(5)
	}

	// AWS
	{
		conf.AWS = new(ConfigAWS)

		conf.AWS.ELBLoadBalancerName = conv.SP(fmt.Sprintf("elb_%s", conf.Name))
		conf.AWS.ELBTargetGroupName = conv.SP(fmt.Sprintf("elb_tg_%s", conf.Name))
		conf.AWS.ELBSecurityGroup = nil
		conf.AWS.ECRRepositoryName = conv.SP(fmt.Sprintf("coldbrew/%s", conf.Name))
	}

	// Docker
	{
		conf.Docker = new(ConfigDocker)

		conf.Docker.Bin = conv.SP("docker")
	}

	return conf
}
