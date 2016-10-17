package flags

import (
	"testing"

	"gopkg.in/alecthomas/kingpin.v2"
)

func TestNewAppFlags(t *testing.T) {
	app := kingpin.New("app", "")
	app.Writer(&nullWriter{})
	appFlags := NewAppFlags(app)

	testBoolFlag(t, app, &appFlags.Debug, nil, "debug", testBytePtr('D'), testBptr(false))
	testStringFlag(t, app, &appFlags.DebugLogPrefix, nil, "debug-log-prefix", nil, testSptr(""), nil)
	testStringFlag(t, app, &appFlags.AWSRegion, nil, "aws-region", nil, testSptr("us-east-1"), testSptr("AWS_REGION"))
	testStringFlag(t, app, &appFlags.AWSAccessKey, nil, "aws-access-key", nil, nil, testSptr("AWS_ACCESS_KEY_ID"))
	testStringFlag(t, app, &appFlags.AWSSecretKey, nil, "aws-secret-key", nil, nil, testSptr("AWS_SECRET_ACCESS_KEY"))
}
