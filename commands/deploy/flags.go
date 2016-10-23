package deploy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/coldbrewcloud/coldbrew-cli/config"
	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Flags struct {
	DockerImage    *string            `json:"docker-image,omitempty"`
	DockerfilePath *string            `json:"dockerfile,omitempty"`
	Units          *int64             `json:"units,omitempty"`
	CPU            *float64           `json:"cpu,omitempty"`
	Memory         *string            `json:"memory,omitempty"`
	Envs           *map[string]string `json:"env,omitempty"`
}

func NewFlags(kc *kingpin.CmdClause) *Flags {
	return &Flags{
		DockerfilePath: kc.Flag("dockerfile", "Dockerfile path").Default("./Dockerfile").String(),
		DockerImage:    kc.Flag("docker-image", "Docker image (should include image tag)").String(),
		Units:          kc.Flag("units", "Desired count").Default("-1").Int64(),
		CPU:            kc.Flag("cpu", "Docker CPU resource (1 unit: 1024)").Default("-1").Float64(),
		Memory:         kc.Flag("memory", "Docker memory resource").Default("").String(),
		Envs:           kc.Flag("env", "App environment variable (\"key=value\")").Short('E').StringMap(),
	}
}

func (c *Command) mergeFlagsIntoConfiguration(conf *config.Config, flags *Flags) *config.Config {
	if conv.I64(flags.Units) >= 0 {
		conf.Units = conv.U16P(uint16(conv.I64(flags.Units)))
	}

	if conv.F64(flags.CPU) >= 0 {
		conf.CPU = conv.F64P(conv.F64(flags.CPU))
	}

	if !utils.IsBlank(conv.S(flags.Memory)) {
		conf.Memory = conv.SP(conv.S(flags.Memory))
	}

	// envs
	for ek, ev := range *flags.Envs {
		conf.Env[ek] = ev
	}

	return conf
}

func (c *Command) validateFlags(flags *Flags) error {
	if !utils.IsBlank(conv.S(flags.DockerImage)) && !core.DockerImageURIRE.MatchString(conv.S(flags.DockerImage)) {
		return fmt.Errorf("Invalid Docker image [%s]", conv.S(flags.DockerImage))
	}

	if err := c.validatePath(conv.S(flags.DockerfilePath)); err != nil {
		return fmt.Errorf("Invalid Dockerfile path [%s]", conv.S(flags.DockerfilePath))
	}

	if conv.I64(flags.Units) >= 0 && uint16(conv.I64(flags.Units)) > core.MaxAppUnits {
		return fmt.Errorf("Units [%d] cannot exceed %d", conv.I64(flags.Units), core.MaxAppUnits)
	}

	if conv.F64(flags.CPU) > core.MaxAppCPU {
		return fmt.Errorf("CPU [%.2f] cannot exceed %d", conv.F64(flags.CPU), core.MaxAppCPU)
	}

	if !utils.IsBlank(conv.S(flags.Memory)) && !core.SizeExpressionRE.MatchString(conv.S(flags.Memory)) {
		return fmt.Errorf("Invalid app memory [%s]", conv.S(flags.Memory))
	}

	return nil
}

func (c *Command) validatePath(path string) error {
	if utils.IsBlank(path) {
		return fmt.Errorf("Path [%s] is blank", path)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("Failed to determine absolute path [%s]", path)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return errors.New("Path [%s] does not exist")
	}

	return nil
}
