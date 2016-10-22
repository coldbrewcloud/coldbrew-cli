package core

import "path/filepath"

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
	base := filepath.Base(filepath.Dir(appDirectoryOrConfigFile))
	if base == "/" {
		return "app1"
	}
	return base
}
