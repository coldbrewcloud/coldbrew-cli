package deploy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

func (c *Command) validateFlags(flags *Flags) error {
	if !utils.IsBlank(conv.S(flags.DockerImage)) && core.DockerImageURIRE.MatchString(conv.S(flags.DockerImage)) {
		return fmt.Errorf("Invalid docker image [%s]", conv.S(flags.DockerImage))
	}

	if err := c.validatePath(conv.S(flags.DockerfilePath)); err != nil {
		return fmt.Errorf("Invalid Dockerfile path [%s]", conv.S(flags.DockerfilePath))
	}

	if utils.IsBlank(conv.S(flags.DockerImage)) {
		return fmt.Errorf("Invalid docker image [%s]", conv.S(flags.DockerImage))
	}

	if uint16(*flags.Units) > core.MaxAppUnits {
		return fmt.Errorf("Units cannot exceed %d", core.MaxAppUnits)
	}

	if *flags.CPU > core.MaxAppCPU {
		return fmt.Errorf("CPU cannot exceed %d", core.MaxAppCPU)
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
