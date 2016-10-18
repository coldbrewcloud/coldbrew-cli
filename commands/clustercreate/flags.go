package clustercreate

import "gopkg.in/alecthomas/kingpin.v2"

type Flags struct {
	InstanceType    *string `json:"instance_type"`
	InitialCapacity *uint16 `json:"initial_capacity"`
	KeyPairName     *string `json:"keypair_name"`
	VPC             *string `json:"vpc"`
	InstanceProfile *string `json:"instance_profile"`
	ReuseResources  *bool   `json:"reuse"`
}

func NewFlags(kc *kingpin.CmdClause) *Flags {
	return &Flags{
		InstanceType:    kc.Flag("instance-type", "Container instance type").String(),
		InitialCapacity: kc.Flag("initial-capacity", "Initial number of container instances").Default("1").Uint16(),
		KeyPairName:     kc.Flag("keypair", "Key pair name").String(),
		VPC:             kc.Flag("vpc", "VPC ID").String(),
		InstanceProfile: kc.Flag("instance-profile", "IAM instance profile ARN for container instances").String(),
		ReuseResources:  kc.Flag("reuse", "Re-use existing resources").Default("false").Bool(),
	}
}
