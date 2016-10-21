package config

import "github.com/coldbrewcloud/coldbrew-cli/utils/conv"

func (c *Config) Merge(defaultConfig *Config) {
	if c.Name == nil {
		c.Name = defaultConfig.Name
	}

	if c.ClusterName == nil {
		c.ClusterName = conv.S(defaultConfig.ClusterName)
	}

	if c.Port == nil {
		c.Port = conv.U16()
	}
}
