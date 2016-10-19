package clusterstatus

import (
	"errors"
	"fmt"
	"strings"

	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

func (c *Command) getAWSNetwork() (string, string, []string, error) {
	regionName := strings.TrimSpace(conv.S(c.globalFlags.AWSRegion))

	// VPC ID
	vpcID := strings.TrimSpace(conv.S(c.commandFlags.VPC))
	if vpcID == "" {
		// find/use default VPC for the account
		defaultVPC, err := c.awsClient.EC2().RetrieveDefaultVPC()
		if err != nil {
			return "", "", nil, fmt.Errorf("Failed to retrieve default VPC: %s", err.Error())
		} else if defaultVPC == nil {
			return "", "", nil, errors.New("No default VPC configured")
		}

		vpcID = conv.S(defaultVPC.VpcId)
	} else {
		vpc, err := c.awsClient.EC2().RetrieveVPC(vpcID)
		if err != nil {
			return "", "", nil, fmt.Errorf("Failed to retrieve VPC [%s] info: %s", vpcID, err.Error())
		}
		if vpc == nil {
			return "", "", nil, fmt.Errorf("VPC [%s] not found", vpcID)
		}
	}

	// Subnet IDs
	subnetIDs, err := c.awsClient.EC2().ListVPCSubnets(vpcID)
	if err != nil {
		return "", "", nil, fmt.Errorf("Failed to list subnets of VPC [%s]: %s", vpcID, err.Error())
	}

	return regionName, vpcID, subnetIDs, nil
}
