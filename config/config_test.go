package config

import (
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

const refConfigYAML = `
name: echo
cluster: cluster1
port: 8080
cpu: 1.0
memory: 200m
units: 4

env:
  key1: value1
  key2: value2

load_balancer:
  enabled: true
  port: 80
  https_port: 443

  health_check:
    interval: 30s
    path: "/ping"
    status: "200-299"
    timeout: 5s
    healthy_limit: 5
    unhealthy_limit: 2

aws:
  elb_name: echo-lb
  elb_target_group_name: echo-target
  elb_security_group_name: echo-lb-sg
  elb_certificate_arn: arn:aws:acm:us-west-2:aws-account-id:certificate/certificiate-identifier
  ecr_repo_name: echo-repo

docker:
  bin: "/usr/local/bin/docker"
`

const refConfigJSON = `
{
	"name": "echo",
	"cluster": "cluster1",
	"port": 8080,
	"cpu": 1.0,
	"memory": "200m",
	"units": 4,
	"env": {
		"key1": "value1",
		"key2": "value2"
	},
	"load_balancer": {
		"enabled": true,
		"port": 80,
		"https_port": 443,
		"health_check": {
			"interval": "30s",
			"path": "/ping",
			"status": "200-299",
			"timeout": "5s",
			"healthy_limit": 5,
			"unhealthy_limit": 2
		}
	},
	"aws": {
		"elb_name": "echo-lb",
		"elb_target_group_name": "echo-target",
		"elb_security_group_name": "echo-lb-sg",
		"elb_certificate_arn": "arn:aws:acm:us-west-2:aws-account-id:certificate/certificiate-identifier",
		"ecr_repo_name": "echo-repo"
	},
	"docker": {
		"bin": "/usr/local/bin/docker"
	}
}`

var refConfig = &Config{
	Name:        conv.SP("echo"),
	ClusterName: conv.SP("cluster1"),
	Port:        conv.U16P(8080),
	CPU:         conv.F64P(1.0),
	Memory:      conv.SP("200m"),
	Units:       conv.U16P(4),
	Env: map[string]string{
		"key1": "value1",
		"key2": "value2",
	},
	LoadBalancer: ConfigLoadBalancer{
		Enabled:   conv.BP(true),
		Port:      conv.U16P(80),
		HTTPSPort: conv.U16P(443),
		HealthCheck: ConfigLoadBalancerHealthCheck{
			Interval:       conv.SP("30s"),
			Path:           conv.SP("/ping"),
			Status:         conv.SP("200-299"),
			Timeout:        conv.SP("5s"),
			HealthyLimit:   conv.U16P(5),
			UnhealthyLimit: conv.U16P(2),
		},
	},
	AWS: ConfigAWS{
		ELBLoadBalancerName:  conv.SP("echo-lb"),
		ELBTargetGroupName:   conv.SP("echo-target"),
		ELBSecurityGroupName: conv.SP("echo-lb-sg"),
		ELBCertificateARN:    conv.SP("arn:aws:acm:us-west-2:aws-account-id:certificate/certificiate-identifier"),
		ECRRepositoryName:    conv.SP("echo-repo"),
	},
	Docker: ConfigDocker{
		Bin: conv.SP("/usr/local/bin/docker"),
	},
}

var partialConfigYAML = `
name: hello
port: 0
cpu: 1.0
memory: 512m

load_balancer:
  enabled: false
`

var partialConfig = &Config{
	Name:   conv.SP("hello"),
	Port:   conv.U16P(0),
	CPU:    conv.F64P(1.0),
	Memory: conv.SP("512m"),
	LoadBalancer: ConfigLoadBalancer{
		Enabled: conv.BP(false),
	},
}
