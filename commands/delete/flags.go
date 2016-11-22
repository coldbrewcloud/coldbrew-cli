package delete

import "gopkg.in/alecthomas/kingpin.v2"

type Flags struct {
	AppName         *string
	ClusterName     *string
	NoConfirm       *bool
	ContinueOnError *bool
}

func NewFlags(kc *kingpin.CmdClause) *Flags {
	return &Flags{
		AppName:         kc.Flag("app-name", "App name").Default("").String(),
		ClusterName:     kc.Flag("cluster-name", "App name").Default("").String(),
		NoConfirm:       kc.Flag("yes", "Delete all resources with no confirmation").Short('y').Default("false").Bool(),
		ContinueOnError: kc.Flag("continue", "Continue deleting resources on error").Default("false").Bool(),
	}
}
