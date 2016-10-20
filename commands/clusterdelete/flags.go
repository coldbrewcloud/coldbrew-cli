package clusterdelete

import "gopkg.in/alecthomas/kingpin.v2"

type Flags struct {
	ForceDelete     *bool `json:"force"`
	ContinueOnError *bool `json:"continue"`
}

func NewFlags(kc *kingpin.CmdClause) *Flags {
	return &Flags{
		ForceDelete:     kc.Flag("force", "Delete all resources with no confirmation").Short('F').Default("false").Bool(),
		ContinueOnError: kc.Flag("continue", "Continue deleting resources on error").Default("false").Bool(),
	}
}
