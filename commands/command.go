package commands

import (
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Command interface {
	Init(app *kingpin.Application, appFlags *flags.GlobalFlags) *kingpin.CmdClause

	// Run should return error only for critical issue. All other errors should be handled inside Run() function.
	Run() error
}
