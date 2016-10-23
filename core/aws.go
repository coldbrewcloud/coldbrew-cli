package core

import (
	"fmt"
	"time"
)

const (
	AWSTagNameCreatedTimestamp = "coldbrew_cli_created"
)

func DefaultTagsForAWSResources() map[string]string {
	return map[string]string{
		AWSTagNameCreatedTimestamp: fmt.Sprintf("%d", time.Now().Unix()),
	}
}
