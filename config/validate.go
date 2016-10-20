package config

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
)

var (
	appNameRE             = regexp.MustCompile(`^[\w\-]{1,32}$`)
	clusterNameRE         = regexp.MustCompile(`^[\w\-]{1,32}$`)
	elbNameRE             = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]{0,30}(?:[a-zA-Z0-9])?$`)
	elbTargetNameRE       = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]{0,30}(?:[a-zA-Z0-9])?$`)
	ecrNamespaceRE        = regexp.MustCompile(`^[\w\-]{1,32}$`)
	healthCheckIntervalRE = regexp.MustCompile(`^\d+[smhSMH]?$`)
	healthCheckPathRE     = regexp.MustCompile(`^.+$`)                             // TODO: need better matcher
	healthCheckStatusRE   = regexp.MustCompile(`^\d{3}-\d{3}$|^\d{3}(?:,\d{3})*$`) // "200", "200-299", "200,204,201"
	healthCheckTimeoutRE  = regexp.MustCompile(`^\d+[smhSMH]?$`)
)

func (c *Config) Validate() error {
	if !appNameRE.MatchString(c.Name) {
		return fmt.Errorf("Invalid app name [%s]", c.Name)
	}

	if !clusterNameRE.MatchString(c.ClusterName) {
		return fmt.Errorf("Invalid cluster name [%s]", c.ClusterName)
	}

	if c.Units > core.MaxAppUnits {
		return fmt.Errorf("Units cannot exceed %d", core.MaxAppUnits)
	}

	if c.CPU > core.MaxAppCPU {
		return fmt.Errorf("CPU cannot exceed %d", core.MaxAppCPU)
	}

	if c.Memory > core.MaxAppMemory {
		return fmt.Errorf("Memory cannot exceed %d", core.MaxAppMemory)
	}

	if !healthCheckIntervalRE.MatchString(c.LoadBalancer.HealthCheck.Interval) {
		return fmt.Errorf("Invalid health check interval [%s]", c.LoadBalancer.HealthCheck.Interval)
	}

	if !healthCheckPathRE.MatchString(c.LoadBalancer.HealthCheck.Path) {
		return fmt.Errorf("Invalid health check path [%s]", c.LoadBalancer.HealthCheck.Path)
	}

	if !healthCheckStatusRE.MatchString(c.LoadBalancer.HealthCheck.Status) {
		return fmt.Errorf("Invalid health check status [%s]", c.LoadBalancer.HealthCheck.Status)
	}

	if !healthCheckTimeoutRE.MatchString(c.LoadBalancer.HealthCheck.Timeout) {
		return fmt.Errorf("Invalid health check timeout [%s]", c.LoadBalancer.HealthCheck.Timeout)
	}

	if c.LoadBalancer.HealthCheck.HealthyLimit == 0 {
		return errors.New("Health check healthy limit cannot be 0.")
	}

	if c.LoadBalancer.HealthCheck.UnhealthyLimit == 0 {
		return errors.New("Health check unhealthy limit cannot be 0.")
	}

	if !ecrNamespaceRE.MatchString(c.AWS.ECRNamespace) {
		return fmt.Errorf("Invalid ECR namespace [%s]", c.AWS.ECRNamespace)
	}

	if !elbNameRE.MatchString(c.AWS.ELBLoadBalancerName) {
		return fmt.Errorf("Invalid ELB load balancer name [%s]", c.AWS.ELBLoadBalancerName)
	}

	if !elbTargetNameRE.MatchString(c.AWS.ELBTargetName) {
		return fmt.Errorf("Invalid ELB target group name [%s]", c.AWS.ELBTargetName)
	}

	if utils.IsBlank(c.Docker.Bin) {
		return fmt.Errorf("Invalid docker executable path [%s]", c.Docker.Bin)
	}

	return nil
}
