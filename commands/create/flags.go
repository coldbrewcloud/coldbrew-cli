package create

import "gopkg.in/alecthomas/kingpin.v2"

type Flags struct {
	Default *bool
}

func NewFlags(kc *kingpin.CmdClause) *Flags {
	return &Flags{
		Default: kc.Flag("default", "Generate default configuration").Bool(),
	}
}
