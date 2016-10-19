package clusterstatus

import "gopkg.in/alecthomas/kingpin.v2"

type Flags struct {
	VPC *string `json:"vpc"`
}

func NewFlags(kc *kingpin.CmdClause) *Flags {
	return &Flags{
		VPC: kc.Flag("vpc", "VPC ID").String(),
	}
}
