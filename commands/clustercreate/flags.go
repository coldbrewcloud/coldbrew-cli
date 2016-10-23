package clustercreate

import "gopkg.in/alecthomas/kingpin.v2"

type Flags struct {
	InstanceType    *string `json:"instance_type"`
	InitialCapacity *uint16 `json:"initial_capacity"`
	KeyPairName     *string `json:"keypair_name"`
	InstanceProfile *string `json:"instance_profile"`
	ForceCreate     *bool   `json:"force"`
}

func NewFlags(kc *kingpin.CmdClause) *Flags {
	return &Flags{
		InstanceType:    kc.Flag("instance-type", "Container instance type").String(),
		InitialCapacity: kc.Flag("instance-count", "Initial number of container instances").Default("1").Uint16(),
		KeyPairName:     kc.Flag("key", "EC2 keypair name").String(),
		InstanceProfile: kc.Flag("instance-profile", "IAM instance profile name for container instances").String(),
		ForceCreate:     kc.Flag("force", "Create all resource with no confirmation").Short('y').Default("false").Bool(),
	}
}
