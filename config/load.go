package config

import (
	"fmt"
	"strings"

	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

func Load(data []byte, configFormat string, defaultAppName string) (*Config, error) {
	conf := &Config{}
	configFormat = strings.ToLower(configFormat)
	switch configFormat {
	case flags.GlobalFlagsConfigFileFormatYAML:
		if err := conf.FromYAML(data); err != nil {
			return nil, fmt.Errorf("Failed to read configuration in YAML: %s\n", err.Error())
		}
	case flags.GlobalFlagsConfigFileFormatJSON:
		if err := conf.FromJSON(data); err != nil {
			return nil, fmt.Errorf("Failed to read configuration in JSON: %s\n", err.Error())
		}
	default:
		return nil, fmt.Errorf("Unsupported configuration format [%s]", configFormat)
	}

	// env
	if conf.Env == nil {
		conf.Env = make(map[string]string)
	}

	// merge with defaults: defaultAppName is used only if loaded configuration does not have app name
	appName := conv.S(conf.Name)
	if appName == "" {
		appName = defaultAppName
	}
	conf.Defaults(DefaultConfig(appName))

	// validation
	if err := conf.Validate(); err != nil {
		return nil, core.NewErrorExtraInfo(err, "https://github.com/coldbrewcloud/coldbrew-cli/wiki/Configuration-File")
	}

	return conf, nil
}

func (c *Config) Defaults(source *Config) {
	if source == nil {
		return
	}

	defS(&c.Name, source.Name)
	defS(&c.ClusterName, source.ClusterName)
	defU16(&c.Port, source.Port)
	defF64(&c.CPU, source.CPU)
	defS(&c.Memory, source.Memory)
	defU16(&c.Units, source.Units)

	// envs
	if c.Env == nil {
		c.Env = make(map[string]string)
	}
	for ek, ev := range source.Env {
		c.Env[ek] = ev
	}

	// load balancer
	defB(&c.LoadBalancer.Enabled, source.LoadBalancer.Enabled)
	defU16(&c.LoadBalancer.Port, source.LoadBalancer.Port)
	defU16(&c.LoadBalancer.HTTPSPort, source.LoadBalancer.HTTPSPort)
	defS(&c.LoadBalancer.HealthCheck.Interval, source.LoadBalancer.HealthCheck.Interval)
	defS(&c.LoadBalancer.HealthCheck.Path, source.LoadBalancer.HealthCheck.Path)
	defS(&c.LoadBalancer.HealthCheck.Status, source.LoadBalancer.HealthCheck.Status)
	defS(&c.LoadBalancer.HealthCheck.Timeout, source.LoadBalancer.HealthCheck.Timeout)
	defU16(&c.LoadBalancer.HealthCheck.HealthyLimit, source.LoadBalancer.HealthCheck.HealthyLimit)
	defU16(&c.LoadBalancer.HealthCheck.UnhealthyLimit, source.LoadBalancer.HealthCheck.UnhealthyLimit)

	// AWS
	defS(&c.AWS.ELBLoadBalancerName, source.AWS.ELBLoadBalancerName)
	defS(&c.AWS.ELBTargetGroupName, source.AWS.ELBTargetGroupName)
	defS(&c.AWS.ELBSecurityGroupName, source.AWS.ELBSecurityGroupName)
	defS(&c.AWS.ELBCertificateARN, source.AWS.ELBCertificateARN)
	defS(&c.AWS.ECRRepositoryName, source.AWS.ECRRepositoryName)

	// docker
	defS(&c.Docker.Bin, source.Docker.Bin)
}

func defS(src **string, dest *string) {
	if *src == nil && dest != nil {
		*src = conv.SP(conv.S(dest))
	}
}

func defU16(src **uint16, dest *uint16) {
	if *src == nil && dest != nil {
		*src = conv.U16P(conv.U16(dest))
	}
}

func defB(src **bool, dest *bool) {
	if *src == nil && dest != nil {
		*src = conv.BP(conv.B(dest))
	}
}

func defF64(src **float64, dest *float64) {
	if *src == nil && dest != nil {
		*src = conv.F64P(conv.F64(dest))
	}
}
