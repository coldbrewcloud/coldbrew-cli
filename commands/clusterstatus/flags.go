package clusterstatus

import "gopkg.in/alecthomas/kingpin.v2"

type Flags struct {
	ExcludeContainerInstanceInfos *bool
}

func NewFlags(kc *kingpin.CmdClause) *Flags {
	return &Flags{
		ExcludeContainerInstanceInfos: kc.Flag("exclude-container-instances", "Exclude ECS Container Instance infos").Bool(),
	}
}
