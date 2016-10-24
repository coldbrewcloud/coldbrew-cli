package clusterscale

import "gopkg.in/alecthomas/kingpin.v2"

type Flags struct {
}

func NewFlags(kc *kingpin.CmdClause) *Flags {
	return &Flags{}
}
