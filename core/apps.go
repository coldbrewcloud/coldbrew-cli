package core

import (
	"path/filepath"

	"fmt"

	"github.com/coldbrewcloud/coldbrew-cli/utils"
)

func DefaultECSTaskDefinitionName(appName string) string {
	return appName
}

func DefaultECSServiceName(appName string) string {
	return appName
}

func DefaultECSTaskMainContainerName(appName string) string {
	return appName
}

func DefaultAppName(appDirectoryOrConfigFile string) string {
	isDir, err := utils.IsDirectory(appDirectoryOrConfigFile)
	if err != nil {
		return "app1"
	}
	if !isDir {
		appDirectoryOrConfigFile = filepath.Dir(appDirectoryOrConfigFile)
	}

	base := filepath.Base(appDirectoryOrConfigFile)
	if base == "/" {
		return "app1"
	}

	// validation check
	if !AppNameRE.MatchString(base) {
		// TODO: probably better to strip unacceptable characters instead of "app1"
		return "app1"
	}

	return base
}

func DefaultELBLoadBalancerName(appName string) string {
	return fmt.Sprintf("%s-elb", appName)
}

func DefaultELBTargetGroupName(appName string) string {
	return fmt.Sprintf("%s-elb-tg", appName)
}

func DefaultELBLoadBalancerSecurityGroupName(appName string) string {
	return fmt.Sprintf("%s-elb-sg", appName)
}

func DefaultECRRepository(appName string) string {
	return fmt.Sprintf("coldbrew/%s", appName)
}
