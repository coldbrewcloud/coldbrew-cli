package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/alecthomas/kingpin.v2"
)

func TestNewDeployFlags(t *testing.T) {
	app := kingpin.New("app", "")
	app.Writer(&nullWriter{})
	app.Terminate(nil)
	deployFlags := NewDeployFlags(app.Command("deploy", ""))

	// command
	cmd, err := app.Parse([]string{}) // no command
	assert.NotNil(t, err)
	assert.Empty(t, cmd)
	cmd, err = app.Parse([]string{"deploy"})
	assert.Nil(t, err)
	assert.Equal(t, "deploy", cmd)

	testStringFlag(t, app, &deployFlags.AppName, testSptr("deploy"), "app-name", nil, testSptr("app1"), nil)
	testStringFlag(t, app, &deployFlags.AppVersion, testSptr("deploy"), "app-version", nil, testSptr("1.0.0"), nil)
	testStringFlag(t, app, &deployFlags.AppPath, testSptr("deploy"), "app-path", nil, testSptr("."), nil)
	testUint16Flag(t, app, &deployFlags.ContainerPort, testSptr("deploy"), "container-port", nil, testU16ptr(0), nil)
	testStringFlag(t, app, &deployFlags.LoadBalancerName, testSptr("deploy"), "load-balancer", nil, nil, nil)
	testStringFlag(t, app, &deployFlags.LoadBalancerTargetGroupName, testSptr("deploy"), "load-balancer-target-group", nil, nil, nil)
	testStringFlag(t, app, &deployFlags.DockerBinPath, testSptr("deploy"), "docker-bin", nil, testSptr("docker"), nil)
	testStringFlag(t, app, &deployFlags.DockerfilePath, testSptr("deploy"), "dockerfile", nil, testSptr("./Dockerfile"), nil)
	testStringFlag(t, app, &deployFlags.DockerImage, testSptr("deploy"), "docker-image", nil, nil, nil)
	testUint64Flag(t, app, &deployFlags.Units, testSptr("deploy"), "units", nil, testU16ptr(1), nil)
	testUint64Flag(t, app, &deployFlags.CPU, testSptr("deploy"), "cpu", nil, testU64ptr(128), nil)
	testUint64Flag(t, app, &deployFlags.Memory, testSptr("deploy"), "memory", nil, testU64ptr(128), nil)
	testStringFlag(t, app, &deployFlags.EnvsFile, testSptr("deploy"), "env-file", nil, nil, nil)
	testStringFlag(t, app, &deployFlags.ECSClusterName, testSptr("deploy"), "cluster-name", nil, testSptr("coldbrew"), nil)
	testStringFlag(t, app, &deployFlags.ECSServiceRoleName, testSptr("deploy"), "service-role-name", nil, testSptr("ecsServiceRole"), nil)
	testStringFlag(t, app, &deployFlags.ECRNamespace, testSptr("deploy"), "ecs-namespace", nil, testSptr("coldbrew"), nil)
	testStringFlag(t, app, &deployFlags.VPCID, testSptr("deploy"), "vpc", nil, nil, nil)
	testBoolFlag(t, app, &deployFlags.CloudWatchLogs, testSptr("deploy"), "cloud-watch-logs", nil, nil)

	// envs
	_, err = app.Parse([]string{"deploy"}) // default
	assert.Nil(t, err)
	assert.NotNil(t, deployFlags.Envs)
	assert.Empty(t, *deployFlags.Envs)
	*deployFlags.Envs = make(map[string]string)
	_, err = app.Parse([]string{"deploy"}) // default
	assert.Nil(t, err)
	assert.NotNil(t, deployFlags.Envs)
	assert.Empty(t, *deployFlags.Envs)
	*deployFlags.Envs = make(map[string]string)
	_, err = app.Parse([]string{"deploy", "--env", "key1=value1"}) // 1 pair
	assert.Nil(t, err)
	assert.NotNil(t, deployFlags.Envs)
	assert.Len(t, *deployFlags.Envs, 1)
	assert.Equal(t, "value1", (*deployFlags.Envs)["key1"])
	*deployFlags.Envs = make(map[string]string)
	_, err = app.Parse([]string{"deploy", "--env", "key1=value1", "--env", "key2=value2"}) // 2 pairs
	assert.Nil(t, err)
	assert.NotNil(t, deployFlags.Envs)
	assert.Len(t, *deployFlags.Envs, 2)
	assert.Equal(t, "value1", (*deployFlags.Envs)["key1"])
	assert.Equal(t, "value2", (*deployFlags.Envs)["key2"])
	*deployFlags.Envs = make(map[string]string)
	_, err = app.Parse([]string{"deploy", "-E", "key1=value1"}) // 1 pair (short)
	assert.Nil(t, err)
	assert.NotNil(t, deployFlags.Envs)
	assert.Len(t, *deployFlags.Envs, 1)
	assert.Equal(t, "value1", (*deployFlags.Envs)["key1"])
	*deployFlags.Envs = make(map[string]string)
	_, err = app.Parse([]string{"deploy", "-E", "key1=value1", "-E", "key2=value2"}) // 2 pairs (short)
	assert.Nil(t, err)
	assert.NotNil(t, deployFlags.Envs)
	assert.Len(t, *deployFlags.Envs, 2)
	assert.Equal(t, "value1", (*deployFlags.Envs)["key1"])
	assert.Equal(t, "value2", (*deployFlags.Envs)["key2"])
	*deployFlags.Envs = make(map[string]string)
	_, err = app.Parse([]string{"deploy", "-E", "key1=value1", "-E", "key2=value2", "--env", "key3=value3"}) // mixed
	assert.Nil(t, err)
	assert.NotNil(t, deployFlags.Envs)
	assert.Len(t, *deployFlags.Envs, 3)
	assert.Equal(t, "value1", (*deployFlags.Envs)["key1"])
	assert.Equal(t, "value2", (*deployFlags.Envs)["key2"])
	assert.Equal(t, "value3", (*deployFlags.Envs)["key3"])
}
