package deploy

import "gopkg.in/alecthomas/kingpin.v2"

type Flags struct {
	DockerImage    *string            `json:"docker-image,omitempty"`
	DockerfilePath *string            `json:"dockerfile,omitempty"`
	Units          *int64             `json:"units,omitempty"`
	CPU            *float64           `json:"cpu,omitempty"`
	Memory         *string            `json:"memory,omitempty"`
	Envs           *map[string]string `json:"env,omitempty"`
	//EnvsFile       *string            `json:"env-file,omitempty"`
}

func NewFlags(kc *kingpin.CmdClause) *Flags {
	return &Flags{
		DockerfilePath: kc.Flag("dockerfile", "Dockerfile path").Default("./Dockerfile").String(),
		DockerImage:    kc.Flag("docker-image", "Docker image (should include image tag)").String(),
		Units:          kc.Flag("units", "Desired count").Default("-1").Int64(),
		CPU:            kc.Flag("cpu", "Docker CPU resource (1 unit: 1024)").Default("-1").Float64(),
		Memory:         kc.Flag("memory", "Docker memory resource").Default("").String(),
		Envs:           kc.Flag("env", "App environment variable (\"key=value\")").Short('E').StringMap(),
		//EnvsFile:       kc.Flag("env-file", "App environment variable file (JSON)").String(),
	}
}
