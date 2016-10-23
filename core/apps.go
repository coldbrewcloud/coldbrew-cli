package core

import (
	"path/filepath"

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

	return base
}
