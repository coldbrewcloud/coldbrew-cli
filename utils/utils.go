package utils

import (
	"encoding/json"
	"os"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

var (
	blankRE = regexp.MustCompile(`^\s*$`)
)

func AsMap(v interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	asMap := make(map[string]interface{})
	if err := json.Unmarshal(data, &asMap); err != nil {
		return nil, err
	}

	return asMap, nil
}

func ToJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "(error) " + err.Error()
	}
	return string(data)
}

func IsBlank(s string) bool {
	return blankRE.MatchString(s)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func IsDirectory(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}

func RetryOnAWSErrorCode(fn func() error, retryErrorCodes []string, interval, timeout time.Duration) error {
	return Retry(func() (bool, error) {
		err := fn()
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				for _, rec := range retryErrorCodes {
					if awsErr.Code() == rec {
						return true, err
					}
				}
			}
		}
		return false, err
	}, interval, timeout)
}

func Retry(fn func() (bool, error), interval, timeout time.Duration) error {
	startTime := time.Now()
	endTime := startTime.Add(timeout)

	var cont bool
	var lastErr error

	for time.Now().Before(endTime) {
		cont, lastErr = fn()
		if !cont {
			break
		}

		time.Sleep(interval)
	}

	return lastErr
}
