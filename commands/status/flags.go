package status

import "gopkg.in/alecthomas/kingpin.v2"

type Flags struct {
	AppName     *string
	ClusterName *string
}

func NewFlags(kc *kingpin.CmdClause) *Flags {
	return &Flags{
		AppName:     kc.Flag("app-name", "App name").Default("").String(),
		ClusterName: kc.Flag("cluster-name", "App name").Default("").String(),
	}
}
