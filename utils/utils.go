package utils

import (
	"encoding/json"
	"os"
	"regexp"
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
	data, err := json.Marshal(v)
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
