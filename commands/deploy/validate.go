package deploy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

var (
	appVersionRE = regexp.MustCompile(`^[\w\-.]{1,32}$`)
)

func (c *Command) validateFlags() error {
	if !appVersionRE.MatchString(conv.S(c.commandFlags.AppVersion)) {
		return fmt.Errorf("Invalid app version [%s]", conv.S(c.commandFlags.AppVersion))
	}

	if err := c.validatePath(conv.S(c.commandFlags.DockerfilePath)); err != nil {
		return fmt.Errorf("Invalid Dockerfile path [%s]", conv.S(c.commandFlags.DockerfilePath))
	}

	if utils.IsBlank(conv.S(c.commandFlags.DockerImage)) {
		return fmt.Errorf("Invalid docker image [%s]", conv.S(c.commandFlags.DockerImage))
	}

	if *c.commandFlags.Units > core.MaxAppUnits {
		return fmt.Errorf("Units cannot exceed %d", core.MaxAppUnits)
	}

	if *c.commandFlags.CPU > core.MaxAppCPU {
		return fmt.Errorf("CPU cannot exceed %d", core.MaxAppCPU)
	}

	if *c.commandFlags.Memory > core.MaxAppMemory {
		return fmt.Errorf("Memory cannot exceed %d", core.MaxAppMemory)
	}

	if err := c.validatePath(conv.S(c.commandFlags.EnvsFile)); err != nil {
		return fmt.Errorf("Invalid envs file [%s]", conv.S(c.commandFlags.EnvsFile))
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
