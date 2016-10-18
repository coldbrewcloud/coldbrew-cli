package flags

import "gopkg.in/alecthomas/kingpin.v2"

type CreateFlags struct {
	OverwriteExisting *bool
}

func NewCreateFlags(kc *kingpin.CmdClause) *CreateFlags {
	return &CreateFlags{
		OverwriteExisting: kc.Flag("overwirte", "Overwrite config file if it exists").Default("false").Bool(),
	}
}
