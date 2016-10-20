package flags

import (
	"testing"

	"gopkg.in/alecthomas/kingpin.v2"
)

func TestNewGlobalFlags(t *testing.T) {
	app := kingpin.New("app", "")
	app.Writer(&nullWriter{})
	gf := NewGlobalFlags(app)

	testStringFlag(t, app, &gf.ConfigFile, nil, "config", testBytePtr('C'), nil, nil)
	testStringFlag(t, app, &gf.ConfigFileFormat, nil, "config-format", nil, testSptr(GlobalFlagsConfigFileFormatYAML), nil)
	testStringFlag(t, app, &gf.AppDirectory, nil, "app-dir", testBytePtr('D'), testSptr("."), nil)
	testBoolFlag(t, app, &gf.Verbose, nil, "verbose", testBytePtr('V'), testBptr(false))
	testStringFlag(t, app, &gf.AWSAccessKey, nil, "aws-access-key", nil, nil, testSptr("AWS_ACCESS_KEY_ID"))
	testStringFlag(t, app, &gf.AWSSecretKey, nil, "aws-secret-key", nil, nil, testSptr("AWS_SECRET_ACCESS_KEY"))
	testStringFlag(t, app, &gf.AWSRegion, nil, "aws-region", nil, testSptr("us-west-2"), testSptr("AWS_REGION"))
	testStringFlag(t, app, &gf.AWSVPC, nil, "aws-vpc", nil, nil, testSptr("AWS_VPC"))
}

func TestGlobalFlags_ResolveAppDirectory(t *testing.T) {
	// TODO: implement
}

func TestGlobalFlags_ResolveConfigFile(t *testing.T) {
	// TODO: implement
}
