package config

import (
	"testing"

	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"github.com/stretchr/testify/assert"
)

const testConfigYAML = `
name: echo
cluster: cluster1
port: 8080                          
cpu: 1.0
memory: 200m
units: 4

env:                                
  key1: value1
  key2: value2

logging:
  type: default                     

load_balancer:
  name: lb1                         
  https: no                         
  port: 80                          
  security_group: sg-123456789
                                    
  health_check:
    interval: 30s                   
    path: "/ping"                   
    status: "200-299"               
    timeout: 5s                     
    healthy_limit: 5
    unhealthy_limit: 2

aws:
  elb_lb_name: echo_lb
  elb_target_name: echo_target
  elb_security_group: echo_lb_sg
  ecr_repo_name: echo_repo

docker:
  bin: "/usr/local/bin/docker"      
`

const testConfigJSON = `
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
		"name": "lb1",
		"https": false,
		"port": 80,
		"security_group": "sg-123456789",
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
		"elb_lb_name": "echo_lb",
		"elb_target_name": "echo_target",
		"elb_security_group": "echo_lb_sg",
		"ecr_repo_name": "echo_repo"
	},
	"docker": {
		"bin": "/usr/local/bin/docker"
	}
}`

var testRefConfig = &Config{
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
	LoadBalancer: &ConfigLoadBalancer{
		IsHTTPS: conv.BP(false),
		Port:    conv.U16P(80),
		HealthCheck: &ConfigLoadBalancerHealthCheck{
			Interval:       conv.SP("30s"),
			Path:           conv.SP("/ping"),
			Status:         conv.SP("200-299"),
			Timeout:        conv.SP("5s"),
			HealthyLimit:   conv.U16P(5),
			UnhealthyLimit: conv.U16P(2),
		},
	},
	AWS: &ConfigAWS{
		ELBLoadBalancerName: conv.SP("echo_lb"),
		ELBTargetGroupName:  conv.SP("echo_target"),
		ELBSecurityGroup:    conv.SP("echo_lb_sg"),
		ECRRepositoryName:   conv.SP("echo_repo"),
	},
	Docker: &ConfigDocker{
		Bin: conv.SP("/usr/local/bin/docker"),
	},
}

func TestConfig_FromYAML(t *testing.T) {
	testConfig := &Config{}
	err := testConfig.FromYAML([]byte(testConfigYAML))
	assert.Nil(t, err)
	assert.Equal(t, testRefConfig, testConfig)
}

func TestConfig_FromJSON(t *testing.T) {
	testConfig := &Config{}
	err := testConfig.FromJSON([]byte(testConfigJSON))
	assert.Nil(t, err)
	assert.Equal(t, testRefConfig, testConfig)
}

func TestConfig_ToYAML(t *testing.T) {
	data, err := testRefConfig.ToYAML()
	assert.Nil(t, err)
	assert.NotNil(t, data)

	testConfig := &Config{}
	err = testConfig.FromYAML(data)
	assert.Nil(t, err)
	assert.Equal(t, testRefConfig, testConfig)
}

func TestConfig_ToJSON(t *testing.T) {
	data, err := testRefConfig.ToJSON()
	assert.Nil(t, err)
	assert.NotNil(t, data)

	testConfig := &Config{}
	err = testConfig.FromJSON(data)
	assert.Nil(t, err)
	assert.Equal(t, testRefConfig, testConfig)
}

func TestConfig_ToJSONWithIndent(t *testing.T) {
	data, err := testRefConfig.ToJSONWithIndent()
	assert.Nil(t, err)
	assert.NotNil(t, data)

	testConfig := &Config{}
	err = testConfig.FromJSON(data)
	assert.Nil(t, err)
	assert.Equal(t, testRefConfig, testConfig)
}

func TestConfig_YAMLJSON(t *testing.T) {
	jsonConfig := &Config{}
	err := jsonConfig.FromJSON([]byte(testConfigJSON))
	assert.Nil(t, err)
	assert.Equal(t, testRefConfig, jsonConfig)

	yamlData, err := jsonConfig.ToYAML()
	assert.Nil(t, err)
	assert.NotNil(t, yamlData)

	yamlConfig := &Config{}
	err = yamlConfig.FromYAML(yamlData)
	assert.Nil(t, err)
	assert.Equal(t, jsonConfig, yamlConfig)

	jsonData, err := yamlConfig.ToJSON()
	assert.Nil(t, err)
	assert.NotNil(t, jsonData)

	jsonConfig2 := &Config{}
	err = jsonConfig2.FromJSON(jsonData)
	assert.Nil(t, err)
	assert.Equal(t, jsonConfig, jsonConfig2)
}

func TestLoad(t *testing.T) {
	jsonConfig, err := Load([]byte(testConfigJSON))
	assert.Nil(t, err)
	assert.Equal(t, testRefConfig, jsonConfig)

	yamlConfig, err := Load([]byte(testConfigYAML))
	assert.Nil(t, err)
	assert.Equal(t, testRefConfig, yamlConfig)
}
