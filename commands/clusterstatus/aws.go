package clusterstatus

import "fmt"

func (c *Command) getAWSInfo() (string, string, []string, error) {
	regionName, vpcID, err := c.globalFlags.GetAWSRegionAndVPCID()
	if err != nil {
		return "", "", nil, err
	}

	// Subnet IDs
	subnetIDs, err := c.awsClient.EC2().ListVPCSubnets(vpcID)
	if err != nil {
		return "", "", nil, fmt.Errorf("Failed to list subnets of VPC [%s]: %s", vpcID, err.Error())
	}
	if len(subnetIDs) == 0 {
		return "", "", nil, fmt.Errorf("VPC [%s] does not have any subnets.", vpcID)
	}

	return regionName, vpcID, subnetIDs, nil
}
