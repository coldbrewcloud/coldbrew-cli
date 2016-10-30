package config

import (
	"testing"

	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	// empty config: fail
	conf := &Config{}
	err := conf.Validate()
	assert.NotNil(t, err)

	// default config: pass
	conf = DefaultConfig("app1")
	err = conf.Validate()
	assert.Nil(t, err)

	// App Name
	conf = DefaultConfig("app1")
	conf.Name = nil // nil
	err = conf.Validate()
	assert.NotNil(t, err)
	conf.Name = conv.SP("") // empty
	err = conf.Validate()
	assert.NotNil(t, err)
	conf.Name = conv.SP("app!") // invalid character
	err = conf.Validate()
	assert.NotNil(t, err)
	conf.Name = conv.SP("12345678901234567890123456789012") // max chars (32)
	err = conf.Validate()
	assert.Nil(t, err)
	conf.Name = conv.SP("123456789012345678901234567890123") // too long (> 32)
	err = conf.Validate()
	assert.NotNil(t, err)

	// Cluster Name
	conf = DefaultConfig("app1")
	conf.ClusterName = nil // nil
	err = conf.Validate()
	assert.NotNil(t, err)
	conf.ClusterName = conv.SP("") // empty
	err = conf.Validate()
	assert.NotNil(t, err)
	conf.ClusterName = conv.SP("cluster!") // invalid character
	err = conf.Validate()
	assert.NotNil(t, err)
	conf.ClusterName = conv.SP("12345678901234567890123456789012") // max chars (32)
	err = conf.Validate()
	assert.Nil(t, err)
	conf.ClusterName = conv.SP("123456789012345678901234567890123") // too long (> 32)
	err = conf.Validate()
	assert.NotNil(t, err)

	// Units
	conf = DefaultConfig("app1")
	conf.Units = nil // nil (OK; considered as zero)
	err = conf.Validate()
	assert.Nil(t, err)
	conf.Units = conv.U16P(0) // zero (OK)
	err = conf.Validate()
	assert.Nil(t, err)
	conf.Units = conv.U16P(core.MaxAppUnits) // max (OK)
	err = conf.Validate()
	assert.Nil(t, err)
	conf.Units = conv.U16P(core.MaxAppUnits + 1) // too large
	err = conf.Validate()
	assert.NotNil(t, err)

	// Port
	conf = DefaultConfig("app1")
	conf.Port = nil // nil; (OK: == zero)
	err = conf.Validate()
	assert.Nil(t, err)
	conf.Port = conv.U16P(0) // zero (OK)
	err = conf.Validate()
	assert.Nil(t, err)

	// CPU
	conf = DefaultConfig("app1")
	conf.CPU = nil // nil
	err = conf.Validate()
	assert.NotNil(t, err)
	conf.CPU = conv.F64P(0) // zero
	err = conf.Validate()
	assert.NotNil(t, err)
	conf.CPU = conv.F64P(core.MaxAppCPU) // max CPU (ok)
	err = conf.Validate()
	assert.Nil(t, err)
	conf.CPU = conv.F64P(core.MaxAppCPU + 0.1) // too large
	err = conf.Validate()
	assert.NotNil(t, err)

	// Memory
	conf = DefaultConfig("app1")
	conf.Memory = nil // nil
	err = conf.Validate()
	assert.NotNil(t, err)
	conf.Memory = conv.SP("") // empty
	err = conf.Validate()
	assert.NotNil(t, err)
	conf.Memory = conv.SP("100") // 100
	err = conf.Validate()
	assert.Nil(t, err)
	conf.Memory = conv.SP("100m") // 100m
	err = conf.Validate()
	assert.Nil(t, err)
	conf.Memory = conv.SP("100mb") // 100mb
	err = conf.Validate()
	assert.Nil(t, err)
	conf.Memory = conv.SP("16g") // 16g < max
	err = conf.Validate()
	assert.Nil(t, err)
	conf.Memory = conv.SP("100g") // 100g > max
	err = conf.Validate()
	assert.NotNil(t, err)
	conf.Memory = conv.SP("100gb") // 100gb > max
	err = conf.Validate()
	assert.NotNil(t, err)
	conf.Memory = conv.SP("-100") // -100: no negative
	err = conf.Validate()
	assert.NotNil(t, err)
	conf.Memory = conv.SP("100b") // 100b: invalid
	err = conf.Validate()
	assert.NotNil(t, err)

	// Env
	conf = DefaultConfig("app1")
	conf.Env = nil // nil
	err = conf.Validate()
	assert.Nil(t, err)
	conf.Env = map[string]string{} // empty
	err = conf.Validate()
	assert.Nil(t, err)

	// Load Balancer Enable
	conf = DefaultConfig("app1")
	conf.LoadBalancer.Enabled = nil // nil
	assert.Nil(t, conf.Validate())

	// Load Balancer Ports
	conf = DefaultConfig("app1")
	conf.AWS.ELBCertificateARN = conv.SP("certificate") // to tests HTTPS ports
	conf.LoadBalancer.Port = conv.U16P(80)
	conf.LoadBalancer.HTTPSPort = conv.U16P(443)
	assert.Nil(t, conf.Validate())
	conf.LoadBalancer.Port = conv.U16P(0)
	conf.LoadBalancer.HTTPSPort = conv.U16P(443)
	assert.Nil(t, conf.Validate())
	conf.LoadBalancer.Port = conv.U16P(80)
	conf.LoadBalancer.HTTPSPort = conv.U16P(0)
	assert.Nil(t, conf.Validate())
	conf.LoadBalancer.Port = conv.U16P(0)
	conf.LoadBalancer.HTTPSPort = conv.U16P(0)
	assert.NotNil(t, conf.Validate()) // both cannot be zero

	// Health Check Interval
	conf = DefaultConfig("app1")
	conf.LoadBalancer.HealthCheck.Interval = nil
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Interval = conv.SP("")
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Interval = conv.SP("10")
	assert.Nil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Interval = conv.SP("10s")
	assert.Nil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Interval = conv.SP("10m")
	assert.Nil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Interval = conv.SP("10h")
	assert.Nil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Interval = conv.SP("-10")
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Interval = conv.SP("10x")
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Interval = conv.SP("10x")
	assert.NotNil(t, conf.Validate())

	// Health Check Path
	conf = DefaultConfig("app1")
	conf.LoadBalancer.HealthCheck.Path = nil
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Path = conv.SP("")
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Path = conv.SP("/ping")
	assert.Nil(t, conf.Validate())

	// Health Check Status
	conf = DefaultConfig("app1")
	conf.LoadBalancer.HealthCheck.Status = nil
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Status = conv.SP("")
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Status = conv.SP("200")
	assert.Nil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Status = conv.SP("200,300")
	assert.Nil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Status = conv.SP("200,300,400")
	assert.Nil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Status = conv.SP("200-400")
	assert.Nil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Status = conv.SP("20")
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Status = conv.SP("2000")
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Status = conv.SP("200,300-400")
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Status = conv.SP("-200")
	assert.NotNil(t, conf.Validate())

	// Health Check Timeout
	conf = DefaultConfig("app1")
	conf.LoadBalancer.HealthCheck.Timeout = nil
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Timeout = conv.SP("")
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Timeout = conv.SP("10")
	assert.Nil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Timeout = conv.SP("10s")
	assert.Nil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Timeout = conv.SP("10m")
	assert.Nil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Timeout = conv.SP("10h")
	assert.Nil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Timeout = conv.SP("-10")
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Timeout = conv.SP("10x")
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.Timeout = conv.SP("10x")
	assert.NotNil(t, conf.Validate())

	// Health Check Healthy Limit
	conf = DefaultConfig("app1")
	conf.LoadBalancer.HealthCheck.HealthyLimit = nil
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.HealthyLimit = conv.U16P(0)
	assert.NotNil(t, conf.Validate())

	// Health Check Unhealthy Limit
	conf = DefaultConfig("app1")
	conf.LoadBalancer.HealthCheck.UnhealthyLimit = nil
	assert.NotNil(t, conf.Validate())
	conf.LoadBalancer.HealthCheck.UnhealthyLimit = conv.U16P(0)
	assert.NotNil(t, conf.Validate())

	// Logging Driver
	conf = DefaultConfig("app1")
	conf.Logging.Driver = nil
	assert.Nil(t, conf.Validate())
	conf.Logging.Driver = conv.SP("")
	assert.Nil(t, conf.Validate())
	conf.Logging.Driver = conv.SP("json-file")
	assert.Nil(t, conf.Validate())
	conf.Logging.Driver = conv.SP("syslog")
	assert.Nil(t, conf.Validate())
	conf.Logging.Driver = conv.SP("journald")
	assert.Nil(t, conf.Validate())
	conf.Logging.Driver = conv.SP("gelf")
	assert.Nil(t, conf.Validate())
	conf.Logging.Driver = conv.SP("fluentd")
	assert.Nil(t, conf.Validate())
	conf.Logging.Driver = conv.SP("splunk")
	assert.Nil(t, conf.Validate())
	conf.Logging.Driver = conv.SP("awslogs")
	assert.Nil(t, conf.Validate())
	conf.Logging.Driver = conv.SP("unknowndriver")
	assert.NotNil(t, conf.Validate())

	// AWS ECR Repository Name
	conf = DefaultConfig("app1")
	conf.AWS.ECRRepositoryName = nil
	assert.NotNil(t, conf.Validate())
	conf.AWS.ECRRepositoryName = conv.SP("")
	assert.NotNil(t, conf.Validate())

	// AWS ELB Load Balancer Name
	conf = DefaultConfig("app1")
	conf.AWS.ELBLoadBalancerName = nil
	assert.NotNil(t, conf.Validate())
	conf.AWS.ELBLoadBalancerName = conv.SP("")
	assert.NotNil(t, conf.Validate())
	conf.AWS.ELBLoadBalancerName = conv.SP("-name") // start with dash
	assert.NotNil(t, conf.Validate())
	conf.AWS.ELBLoadBalancerName = conv.SP("name-") // end with dash
	assert.NotNil(t, conf.Validate())
	conf.AWS.ELBLoadBalancerName = conv.SP("na-me")
	assert.Nil(t, conf.Validate())
	conf.AWS.ELBLoadBalancerName = conv.SP("123456789012345678901234567890121") // too long
	assert.NotNil(t, conf.Validate())

	// AWS ELB Target Group Name
	conf = DefaultConfig("app1")
	conf.AWS.ELBTargetGroupName = nil
	assert.NotNil(t, conf.Validate())
	conf.AWS.ELBTargetGroupName = conv.SP("")
	assert.NotNil(t, conf.Validate())
	conf.AWS.ELBTargetGroupName = conv.SP("-name") // start with dash
	assert.NotNil(t, conf.Validate())
	conf.AWS.ELBTargetGroupName = conv.SP("name-") // end with dash
	assert.NotNil(t, conf.Validate())
	conf.AWS.ELBTargetGroupName = conv.SP("na-me")
	assert.Nil(t, conf.Validate())
	conf.AWS.ELBTargetGroupName = conv.SP("123456789012345678901234567890121") // too long
	assert.NotNil(t, conf.Validate())

	// AWS ELB LB Security Gorup Name
	conf = DefaultConfig("app1")
	conf.AWS.ELBSecurityGroupName = nil
	assert.NotNil(t, conf.Validate())
	conf.AWS.ELBSecurityGroupName = conv.SP("")
	assert.NotNil(t, conf.Validate())
	conf.AWS.ELBSecurityGroupName = conv.SP("-name") // start with dash
	assert.NotNil(t, conf.Validate())
	conf.AWS.ELBSecurityGroupName = conv.SP("name-") // end with dash
	assert.NotNil(t, conf.Validate())
	conf.AWS.ELBSecurityGroupName = conv.SP("na-me")
	assert.Nil(t, conf.Validate())
	conf.AWS.ELBSecurityGroupName = conv.SP("123456789012345678901234567890121") // too long
	assert.NotNil(t, conf.Validate())

	// AWS ELB Certificate ARN
	conf = DefaultConfig("app1")
	conf.AWS.ELBCertificateARN = nil
	assert.Nil(t, conf.Validate())
	conf.AWS.ELBCertificateARN = conv.SP("")
	assert.Nil(t, conf.Validate())
	conf.AWS.ELBCertificateARN = conv.SP("")
	conf.LoadBalancer.HTTPSPort = conv.U16P(443) // HTTPS port enabled
	conf.AWS.ELBCertificateARN = nil
	assert.NotNil(t, conf.Validate())
	conf.AWS.ELBCertificateARN = conv.SP("")
	assert.NotNil(t, conf.Validate())
	conf.AWS.ELBCertificateARN = conv.SP("")

	// Docker Bin
	conf = DefaultConfig("app1")
	conf.Docker.Bin = nil
	assert.NotNil(t, conf.Validate())
	conf.Docker.Bin = conv.SP("")
	assert.NotNil(t, conf.Validate())
}
