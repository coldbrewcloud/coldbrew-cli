package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_FromYAML(t *testing.T) {
	testConfig := &Config{}
	err := testConfig.FromYAML([]byte(refConfigYAML))
	assert.Nil(t, err)
	assert.Equal(t, refConfig, testConfig)
}

func TestConfig_FromJSON(t *testing.T) {
	testConfig := &Config{}
	err := testConfig.FromJSON([]byte(refConfigJSON))
	assert.Nil(t, err)
	assert.Equal(t, refConfig, testConfig)
}

func TestConfig_ToYAML(t *testing.T) {
	data, err := refConfig.ToYAML()
	assert.Nil(t, err)
	assert.NotNil(t, data)

	testConfig := &Config{}
	err = testConfig.FromYAML(data)
	assert.Nil(t, err)
	assert.Equal(t, refConfig, testConfig)
}

func TestConfig_ToJSON(t *testing.T) {
	data, err := refConfig.ToJSON()
	assert.Nil(t, err)
	assert.NotNil(t, data)

	testConfig := &Config{}
	err = testConfig.FromJSON(data)
	assert.Nil(t, err)
	assert.Equal(t, refConfig, testConfig)
}

func TestConfig_ToJSONWithIndent(t *testing.T) {
	data, err := refConfig.ToJSONWithIndent()
	assert.Nil(t, err)
	assert.NotNil(t, data)

	testConfig := &Config{}
	err = testConfig.FromJSON(data)
	assert.Nil(t, err)
	assert.Equal(t, refConfig, testConfig)
}

func TestConfig_YAMLJSON(t *testing.T) {
	jsonConfig := &Config{}
	err := jsonConfig.FromJSON([]byte(refConfigJSON))
	assert.Nil(t, err)
	assert.Equal(t, refConfig, jsonConfig)

	yamlData, err := jsonConfig.ToYAML()
	assert.Nil(t, err)
	assert.NotNil(t, yamlData)

	yamlConfig := &Config{}
	err = yamlConfig.FromYAML(yamlData)
	assert.Nil(t, err)
	assert.Equal(t, jsonConfig, yamlConfig)

	jsonData, err := yamlConfig.ToJSON()
	assert.Nil(t, err)
	assert.NotNil(t, jsonData)

	jsonConfig2 := &Config{}
	err = jsonConfig2.FromJSON(jsonData)
	assert.Nil(t, err)
	assert.Equal(t, jsonConfig, jsonConfig2)
}
