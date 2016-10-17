package commands

import (
	"github.com/coldbrewcloud/coldbrew-cli/flags"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Command interface {
	Init(app *kingpin.Application, appFlags *flags.AppFlags) *kingpin.CmdClause
	Run() error
}
