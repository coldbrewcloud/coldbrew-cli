package config

import (
	"errors"
	"fmt"

	"github.com/coldbrewcloud/coldbrew-cli/aws"
	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

func (c *Config) Validate() error {
	if !core.AppNameRE.MatchString(conv.S(c.Name)) {
		return fmt.Errorf("Invalid app name [%s]", conv.S(c.Name))
	}

	if !core.ClusterNameRE.MatchString(conv.S(c.ClusterName)) {
		return fmt.Errorf("Invalid cluster name [%s]", conv.S(c.ClusterName))
	}

	if conv.U16(c.Units) > core.MaxAppUnits {
		return fmt.Errorf("Units cannot exceed %d", core.MaxAppUnits)
	}

	if conv.F64(c.CPU) == 0 {
		return errors.New("CPU cannot be 0")
	}
	if conv.F64(c.CPU) > core.MaxAppCPU {
		return fmt.Errorf("CPU cannot exceed %d", core.MaxAppCPU)
	}

	if !core.SizeExpressionRE.MatchString(conv.S(c.Memory)) {
		return fmt.Errorf("Invalid app memory [%s] (1)", conv.S(c.Memory))
	} else {
		sizeInBytes, err := core.ParseSizeExpression((conv.S(c.Memory)))
		if err != nil {
			return fmt.Errorf("Invalid app memory: %s", err.Error())
		}
		if sizeInBytes > core.MaxAppMemoryInMB*1000*1000 {
			return fmt.Errorf("App memory cannot exceed %dM", core.MaxAppMemoryInMB)
		}
	}

	if conv.U16(c.LoadBalancer.HTTPSPort) == 0 &&
		conv.U16(c.LoadBalancer.Port) == 0 {
		return errors.New("Load balancer ort number is required.")
	}

	if !core.TimeExpressionRE.MatchString(conv.S(c.LoadBalancer.HealthCheck.Interval)) {
		return fmt.Errorf("Invalid health check interval [%s]", conv.S(c.LoadBalancer.HealthCheck.Interval))
	}

	if !core.HealthCheckPathRE.MatchString(conv.S(c.LoadBalancer.HealthCheck.Path)) {
		return fmt.Errorf("Invalid health check path [%s]", conv.S(c.LoadBalancer.HealthCheck.Path))
	}

	if !core.HealthCheckStatusRE.MatchString(conv.S(c.LoadBalancer.HealthCheck.Status)) {
		return fmt.Errorf("Invalid health check status [%s]", conv.S(c.LoadBalancer.HealthCheck.Status))
	}

	if !core.TimeExpressionRE.MatchString(conv.S(c.LoadBalancer.HealthCheck.Timeout)) {
		return fmt.Errorf("Invalid health check timeout [%s]", conv.S(c.LoadBalancer.HealthCheck.Timeout))
	}

	if conv.U16(c.LoadBalancer.HealthCheck.HealthyLimit) == 0 {
		return errors.New("Health check healthy limit cannot be 0.")
	}

	if conv.U16(c.LoadBalancer.HealthCheck.UnhealthyLimit) == 0 {
		return errors.New("Health check unhealthy limit cannot be 0.")
	}

	if !core.ECRRepoNameRE.MatchString(conv.S(c.AWS.ECRRepositoryName)) {
		return fmt.Errorf("Invalid ECR Resitory name [%s]", conv.S(c.AWS.ECRRepositoryName))
	}

	if !core.ELBNameRE.MatchString(conv.S(c.AWS.ELBLoadBalancerName)) {
		return fmt.Errorf("Invalid ELB Load Balancer name [%s]", conv.S(c.AWS.ELBLoadBalancerName))
	}

	if !core.ELBTargetGroupNameRE.MatchString(conv.S(c.AWS.ELBTargetGroupName)) {
		return fmt.Errorf("Invalid ELB Target Group name [%s]", conv.S(c.AWS.ELBTargetGroupName))
	}

	if conv.U16(c.LoadBalancer.HTTPSPort) > 0 && utils.IsBlank(conv.S(c.AWS.ELBCertificateARN)) {
		return errors.New("Certificate ARN required to enable HTTPS.")
	}

	if !core.ELBSecurityGroupNameRE.MatchString(conv.S(c.AWS.ELBSecurityGroupName)) {
		return fmt.Errorf("Invalid ELB Security Group name [%s]", conv.S(c.AWS.ELBSecurityGroupName))
	}

	switch conv.S(c.Logging.Driver) {
	case "",
		aws.ECSTaskDefinitionLogDriverAWSLogs,
		aws.ECSTaskDefinitionLogDriverJSONFile,
		aws.ECSTaskDefinitionLogDriverSyslog,
		aws.ECSTaskDefinitionLogDriverFluentd,
		aws.ECSTaskDefinitionLogDriverGelf,
		aws.ECSTaskDefinitionLogDriverJournald,
		aws.ECSTaskDefinitionLogDriverSplunk:
		// need more validation for other driver types
	default:
		return fmt.Errorf("Log driver [%s] not supported.", conv.S(c.Logging.Driver))
	}

	if utils.IsBlank(conv.S(c.Docker.Bin)) {
		return fmt.Errorf("Invalid docker executable path [%s]", conv.S(c.Docker.Bin))
	}

	return nil
}
