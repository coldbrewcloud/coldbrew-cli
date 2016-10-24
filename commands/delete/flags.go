package delete

import "gopkg.in/alecthomas/kingpin.v2"

type Flags struct {
	NoConfirm       *bool
	ContinueOnError *bool
}

func NewFlags(kc *kingpin.CmdClause) *Flags {
	return &Flags{
		NoConfirm:       kc.Flag("yes", "Delete all resources with no confirmation").Short('y').Default("false").Bool(),
		ContinueOnError: kc.Flag("continue", "Continue deleting resources on error").Default("false").Bool(),
	}
}
