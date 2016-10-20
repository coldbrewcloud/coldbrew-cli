package deploy

import "gopkg.in/alecthomas/kingpin.v2"

type Flags struct {
	AppVersion     *string            `json:"app-version"`
	DockerfilePath *string            `json:"dockerfile,omitempty"`
	DockerImage    *string            `json:"docker-image,omitempty"`
	Units          *uint16            `json:"units,omitempty"`
	CPU            *uint64            `json:"cpu,omitempty"`
	Memory         *uint64            `json:"memory,omitempty"`
	Envs           *map[string]string `json:"env,omitempty"`
	EnvsFile       *string            `json:"env-file,omitempty"`
}

func NewFlags(kc *kingpin.CmdClause) *Flags {
	return &Flags{
		AppVersion:     kc.Flag("app-version", "App version").Default("1.0.0").String(),
		DockerfilePath: kc.Flag("dockerfile", "Dockerfile path").Default("./Dockerfile").String(),
		DockerImage:    kc.Flag("docker-image", "Docker image (should include image tag)").String(),
		Units:          kc.Flag("units", "Desired count").Default("1").Uint16(),
		CPU:            kc.Flag("cpu", "Docker CPU resource (1 unit: 1024)").Default("128").Uint64(),
		Memory:         kc.Flag("memory", "Docker memory resource (in MB)").Default("128").Uint64(),
		Envs:           kc.Flag("env", "App environment variable (\"key=value\")").Short('E').StringMap(),
		EnvsFile:       kc.Flag("env-file", "App environment variable file (JSON)").String(),
	}
}
