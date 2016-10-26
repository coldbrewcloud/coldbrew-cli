package core

import (
	"fmt"
	"time"
)

const (
	AWSTagNameResourceName     = "Name"
	AWSTagNameCreatedTimestamp = "coldbrew_cli_created"
)

func DefaultTagsForAWSResources(resourceName string) map[string]string {
	return map[string]string{
		AWSTagNameResourceName:     resourceName,
		AWSTagNameCreatedTimestamp: fmt.Sprintf("%d", time.Now().Unix()),
	}
}
