package clustercreate

import (
	"github.com/coldbrewcloud/coldbrew-cli/core"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Flags struct {
	InstanceType         *string `json:"instance_type"`
	InitialCapacity      *uint16 `json:"initial_capacity"`
	NoKeyPair            *bool   `json:"no-keypair"`
	KeyPairName          *string `json:"keypair_name"`
	InstanceProfile      *string `json:"instance_profile"`
	InstanceImageID      *string `json:"instance_image_id"`
	InstanceUserDataFile *string `json:"instance_user_data_file"`
	ForceCreate          *bool   `json:"force"`
}

func NewFlags(kc *kingpin.CmdClause) *Flags {
	return &Flags{
		InstanceType:         kc.Flag("instance-type", "Container instance type").Default(core.DefaultContainerInstanceType()).String(),
		InitialCapacity:      kc.Flag("instance-count", "Initial number of container instances").Default("1").Uint16(),
		NoKeyPair:            kc.Flag("disable-keypair", "Do not assign EC2 keypairs").Bool(),
		KeyPairName:          kc.Flag("key", "EC2 keypair name").Default("").String(),
		InstanceProfile:      kc.Flag("instance-profile", "IAM instance profile name for container instances").Default("").String(),
		InstanceImageID:      kc.Flag("instance-image", "EC2 Image (AMI) ID for ECS Container Instances").Default("").String(),
		InstanceUserDataFile: kc.Flag("instance-userdata", "File path that contains userdata for ECS Container Instances").Default("").String(),
		ForceCreate:          kc.Flag("yes", "Create all resource with no confirmation").Short('y').Default("false").Bool(),
	}
}
