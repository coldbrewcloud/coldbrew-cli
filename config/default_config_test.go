package config

import (
	"testing"

	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	defConf := DefaultConfig("app1")
	assert.Equal(t, "app1", conv.S(defConf.Name))
	err := defConf.Validate()
	assert.Nil(t, err)

	// max app name: 32 chars
	defConf = DefaultConfig("12345678901234567890123456789012")
	assert.Equal(t, "12345678901234567890123456789012", conv.S(defConf.Name))
	err = defConf.Validate()
	assert.Nil(t, err)
	assert.Len(t, conv.S(defConf.AWS.ELBLoadBalancerName), 32)
	assert.Len(t, conv.S(defConf.AWS.ELBTargetGroupName), 32)

	// app name's too long
	defConf = DefaultConfig("123456789012345678901234567890123")
	err = defConf.Validate()
	assert.NotNil(t, err)
}
