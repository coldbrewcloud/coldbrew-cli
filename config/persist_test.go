package config

import (
	"testing"

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
  vpc: vpc-123456789                

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
		"vpc": "vpc-123456789"
	},
	"docker": {
		"bin": "/usr/local/bin/docker"
	}
}`

var testRefConfig = &Config{
	Name:        "echo",
	ClusterName: "cluster1",
	Port:        8080,
	CPU:         1.0,
	Memory:      "200m",
	Units:       4,
	Env: map[string]string{
		"key1": "value1",
		"key2": "value2",
	},
	LoadBalancer: &ConfigLoadBalancer{
		Name:          "lb1",
		IsHTTPS:       false,
		Port:          80,
		SecurityGroup: "sg-123456789",
		HealthCheck: &ConfigLoadBalancerHealthCheck{
			Interval:       "30s",
			Path:           "/ping",
			Status:         "200-299",
			Timeout:        "5s",
			HealthyLimit:   5,
			UnhealthyLimit: 2,
		},
	},
	AWS: &ConfigAWS{
		VPC: "vpc-123456789",
	},
	Docker: &ConfigDocker{
		Bin: "/usr/local/bin/docker",
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
