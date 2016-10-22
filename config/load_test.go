package config

import (
	"testing"

	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"github.com/stretchr/testify/assert"
)

type testObject struct {
	String  *string
	Bool    *bool
	Uint16  *uint16
	Float64 *float64
}

func TestLoad(t *testing.T) {
	// loading empty data
	conf, err := Load([]byte(""), flags.GlobalFlagsConfigFileFormatYAML, "app1")
	assert.Nil(t, err)
	assert.NotNil(t, conf)
	assert.Equal(t, "app1", conv.S(conf.Name))

	// empty data and empty app name
	conf, err = Load([]byte(""), flags.GlobalFlagsConfigFileFormatYAML, "")
	assert.NotNil(t, err)

	// loading "name" only data
	conf, err = Load([]byte("name: app2"), flags.GlobalFlagsConfigFileFormatYAML, "app3")
	assert.Nil(t, err)
	assert.NotNil(t, conf)
	assert.Equal(t, "app2", conv.S(conf.Name))

	// reference config data (YAML)
	conf, err = Load([]byte(refConfigYAML), flags.GlobalFlagsConfigFileFormatYAML, "app4")
	assert.Nil(t, err)
	assert.NotNil(t, conf)
	assert.Equal(t, conv.S(refConfig.Name), conv.S(conf.Name))
	assert.Equal(t, refConfig, conf)

	// reference config data (JSON)
	conf, err = Load([]byte(refConfigJSON), flags.GlobalFlagsConfigFileFormatJSON, "app5")
	assert.Nil(t, err)
	assert.NotNil(t, conf)
	assert.Equal(t, conv.S(refConfig.Name), conv.S(conf.Name))
	assert.Equal(t, refConfig, conf)

	// partial config data (YAML)
	conf, err = Load([]byte(partialConfigYAML), flags.GlobalFlagsConfigFileFormatYAML, "app6")
	assert.Nil(t, err)
	assert.NotNil(t, conf)
	assert.Equal(t, partialConfig.Name, conf.Name)
	assert.Equal(t, partialConfig.Port, conf.Port)
	assert.Equal(t, partialConfig.CPU, conf.CPU)
	assert.Equal(t, partialConfig.Memory, conf.Memory)
	assert.Equal(t, partialConfig.LoadBalancer.Enabled, conf.LoadBalancer.Enabled)
	defConf := DefaultConfig("app6")
	assert.Equal(t, defConf.ClusterName, conf.ClusterName)
	assert.Equal(t, defConf.Units, conf.Units)
	assert.Equal(t, defConf.AWS, conf.AWS)
	assert.Equal(t, defConf.Docker, conf.Docker)
}

func TestConfig_Defaults(t *testing.T) {
	// defaulting with nil: should not change anything
	conf := testClone(refConfig)
	conf.Defaults(nil)
	assert.Equal(t, refConfig, conf)

	// test envs (other attributes test covered by def* tests)
	conf1 := &Config{}
	conf2 := &Config{Env: map[string]string{
		"key1": "value1",
		"key2": "value2",
	}}
	conf3 := &Config{Env: map[string]string{
		"key2": "value2-2",
		"key3": "value3",
	}}
	conf1.Defaults(conf2)
	assert.Len(t, conf1.Env, 2)
	assert.Equal(t, "value1", conf1.Env["key1"])
	assert.Equal(t, "value2", conf1.Env["key2"])
	conf2.Defaults(conf3)
	assert.Len(t, conf2.Env, 3)
	assert.Equal(t, "value1", conf2.Env["key1"])
	assert.Equal(t, "value2-2", conf2.Env["key2"])
	assert.Equal(t, "value3", conf2.Env["key3"])

	// test defS()
	obj1 := &testObject{}
	obj2 := &testObject{String: conv.SP("foo")}
	obj3 := &testObject{String: conv.SP("bar")}
	defS(&obj1.String, obj2.String)
	assert.Equal(t, "foo", conv.S(obj1.String))
	assert.Equal(t, "foo", conv.S(obj2.String))
	defS(&obj1.String, obj2.String)
	assert.Equal(t, "foo", conv.S(obj2.String))
	assert.Equal(t, "bar", conv.S(obj3.String))
	obj1 = &testObject{}
	obj2 = &testObject{}
	defS(&obj1.String, obj2.String)
	assert.Nil(t, obj1.String)
	assert.Nil(t, obj2.String)
	obj1 = &testObject{String: conv.SP("foo")}
	obj2 = &testObject{}
	defS(&obj1.String, obj2.String)
	assert.Equal(t, "foo", conv.S(obj1.String))
	assert.Nil(t, obj2.String)

	// test defB()
	obj1 = &testObject{}
	obj2 = &testObject{Bool: conv.BP(true)}
	obj3 = &testObject{Bool: conv.BP(false)}
	defB(&obj1.Bool, obj2.Bool)
	assert.Equal(t, true, conv.B(obj1.Bool))
	assert.Equal(t, true, conv.B(obj2.Bool))
	defB(&obj1.Bool, obj2.Bool)
	assert.Equal(t, true, conv.B(obj2.Bool))
	assert.Equal(t, false, conv.B(obj3.Bool))
	obj1 = &testObject{}
	obj2 = &testObject{}
	defB(&obj1.Bool, obj2.Bool)
	assert.Nil(t, obj1.Bool)
	assert.Nil(t, obj2.Bool)
	obj1 = &testObject{Bool: conv.BP(true)}
	obj2 = &testObject{}
	defB(&obj1.Bool, obj2.Bool)
	assert.Equal(t, true, conv.B(obj1.Bool))
	assert.Nil(t, obj2.Bool)

	// test defU16()
	obj1 = &testObject{}
	obj2 = &testObject{Uint16: conv.U16P(39)}
	obj3 = &testObject{Uint16: conv.U16P(42)}
	defU16(&obj1.Uint16, obj2.Uint16)
	assert.Equal(t, uint16(39), conv.U16(obj1.Uint16))
	assert.Equal(t, uint16(39), conv.U16(obj2.Uint16))
	defU16(&obj1.Uint16, obj2.Uint16)
	assert.Equal(t, uint16(39), conv.U16(obj2.Uint16))
	assert.Equal(t, uint16(42), conv.U16(obj3.Uint16))
	obj1 = &testObject{}
	obj2 = &testObject{}
	defU16(&obj1.Uint16, obj2.Uint16)
	assert.Nil(t, obj1.Uint16)
	assert.Nil(t, obj2.Uint16)
	obj1 = &testObject{Uint16: conv.U16P(39)}
	obj2 = &testObject{}
	defU16(&obj1.Uint16, obj2.Uint16)
	assert.Equal(t, uint16(39), conv.U16(obj1.Uint16))
	assert.Nil(t, obj2.Uint16)

	// test defF64()
	obj1 = &testObject{}
	obj2 = &testObject{Float64: conv.F64P(52.64)}
	obj3 = &testObject{Float64: conv.F64P(-20.22)}
	defF64(&obj1.Float64, obj2.Float64)
	assert.Equal(t, 52.64, conv.F64(obj1.Float64))
	assert.Equal(t, 52.64, conv.F64(obj2.Float64))
	defF64(&obj1.Float64, obj2.Float64)
	assert.Equal(t, 52.64, conv.F64(obj2.Float64))
	assert.Equal(t, -20.22, conv.F64(obj3.Float64))
	obj1 = &testObject{}
	obj2 = &testObject{}
	defF64(&obj1.Float64, obj2.Float64)
	assert.Nil(t, obj1.Float64)
	assert.Nil(t, obj2.Float64)
	obj1 = &testObject{Float64: conv.F64P(52.64)}
	obj2 = &testObject{}
	defF64(&obj1.Float64, obj2.Float64)
	assert.Equal(t, 52.64, conv.F64(obj1.Float64))
	assert.Nil(t, obj2.Float64)
}

func testClone(src *Config) *Config {
	yaml, err := src.ToYAML()
	if err != nil {
		panic(err)
	}
	dest := &Config{}
	if err := dest.FromYAML(yaml); err != nil {
		panic(err)
	}
	return dest
}
