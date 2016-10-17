package deploy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

const (
	maxUnits  = 1000
	maxCPU    = 1024 * 16
	maxMemory = 1024 * 16
)

var (
	appNameRE                     = regexp.MustCompile(`^[\w\-]{1,32}$`)
	appVersionRE                  = regexp.MustCompile(`^[\w\-.]{1,32}$`)
	loadBalancerNameRE            = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]{0,30}(?:[a-zA-Z0-9])?$`)
	loadBalancerTargetGroupNameRE = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]{0,30}(?:[a-zA-Z0-9])?$`)
	ecrNamespaceRE                = regexp.MustCompile(`^[\w\-]{1,32}$`)
	ecsClusterNameRE              = regexp.MustCompile(`^[\w\-]{1,32}$`)
	ecsServiceRoleNameRE          = regexp.MustCompile(`^[\w\-]{1,32}$`)
)

func (dc *DeployCommand) validateFlags() error {
	if !appNameRE.MatchString(conv.S(dc.deployFlags.AppName)) {
		return fmt.Errorf("Invalid app name [%s]", conv.S(dc.deployFlags.AppName))
	}

	if !appVersionRE.MatchString(conv.S(dc.deployFlags.AppVersion)) {
		return fmt.Errorf("Invalid app version [%s]", conv.S(dc.deployFlags.AppVersion))
	}

	if err := dc.validatePath(conv.S(dc.deployFlags.AppPath)); err != nil {
		return fmt.Errorf("Invalid app path [%s]", err.Error())
	}

	if !loadBalancerNameRE.MatchString(conv.S(dc.deployFlags.LoadBalancerName)) {
		return fmt.Errorf("Invalid load balancer name [%s]", conv.S(dc.deployFlags.LoadBalancerName))
	}

	if !utils.IsBlank(conv.S(dc.deployFlags.LoadBalancerName)) && conv.U16(dc.deployFlags.ContainerPort) == 0 {
		return errors.New("Load balancer cannot be set if container port is 0")
	}

	if !loadBalancerTargetGroupNameRE.MatchString(conv.S(dc.deployFlags.LoadBalancerTargetGroupName)) {
		return fmt.Errorf("Invalid load balancer target group name [%s]", conv.S(dc.deployFlags.LoadBalancerTargetGroupName))
	}

	if utils.IsBlank(conv.S(dc.deployFlags.DockerBinPath)) {
		return fmt.Errorf("Invalid docker bin path [%s]", conv.S(dc.deployFlags.DockerBinPath))
	}

	if err := dc.validatePath(conv.S(dc.deployFlags.DockerfilePath)); err != nil {
		return fmt.Errorf("Invalid Dockerfile path [%s]", conv.S(dc.deployFlags.DockerfilePath))
	}

	if utils.IsBlank(conv.S(dc.deployFlags.DockerImage)) {
		return fmt.Errorf("Invalid docker image [%s]", conv.S(dc.deployFlags.DockerImage))
	}

	if *dc.deployFlags.Units > maxUnits {
		return fmt.Errorf("Units cannot exceed %d", maxUnits)
	}

	if *dc.deployFlags.CPU > maxCPU {
		return fmt.Errorf("CPU cannot exceed %d", maxCPU)
	}

	if *dc.deployFlags.Memory > maxMemory {
		return fmt.Errorf("Memory cannot exceed %d", maxMemory)
	}

	if err := dc.validatePath(conv.S(dc.deployFlags.EnvsFile)); err != nil {
		return fmt.Errorf("Invalid envs file [%s]", conv.S(dc.deployFlags.EnvsFile))
	}

	if !ecrNamespaceRE.MatchString(conv.S(dc.deployFlags.ECRNamespace)) {
		return fmt.Errorf("Invalid ECR namespace [%s]", conv.S(dc.deployFlags.ECRNamespace))
	}

	if !ecsClusterNameRE.MatchString(conv.S(dc.deployFlags.ECSClusterName)) {
		return fmt.Errorf("Invalid ECS cluster name [%s]", conv.S(dc.deployFlags.ECSClusterName))
	}

	if !ecsServiceRoleNameRE.MatchString(conv.S(dc.deployFlags.ECSServiceRoleName)) {
		return fmt.Errorf("Invalid ECS service role name [%s]", conv.S(dc.deployFlags.ECSServiceRoleName))
	}

	return nil
}

func (dc *DeployCommand) validatePath(path string) error {
	if utils.IsBlank(path) {
		return fmt.Errorf("Path [%s] is blank", path)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("Failed to determine absolute path [%s]", path)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return errors.New("Path [%s] does not exist")
	}

	return nil
}
