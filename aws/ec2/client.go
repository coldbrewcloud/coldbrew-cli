package ec2

import (
	"fmt"
	"regexp"
	"strings"

	_aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	_ec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

const (
	SecurityGroupProtocolTCP  = "tcp"
	SecurityGroupProtocolUDP  = "udp"
	SecurityGroupProtocolICMP = "icmp"
	SecurityGroupProtocolAll  = "all"
)

var (
	cidrRE = regexp.MustCompile(`^[\d/.]+$`) // loose matcher
)

type Client struct {
	svc *_ec2.EC2
}

func New(session *session.Session, config *_aws.Config) *Client {
	return &Client{
		svc: _ec2.New(session, config),
	}
}

func (c *Client) CreateSecurityGroup(name, description, vpcID string) (string, error) {
	params := &_ec2.CreateSecurityGroupInput{
		GroupName:   _aws.String(name),
		Description: _aws.String(description),
		VpcId:       _aws.String(vpcID),
	}

	res, err := c.svc.CreateSecurityGroup(params)
	if err != nil {
		return "", err
	}

	return conv.S(res.GroupId), nil
}

func (c *Client) AddInboundToSecurityGroup(securityGroupID, protocol string, portRangeFrom, portRangeTo uint16, source string) error {
	params := &_ec2.AuthorizeSecurityGroupIngressInput{
		IpProtocol: _aws.String(protocol),
		GroupId:    _aws.String(securityGroupID),
		FromPort:   _aws.Int64(int64(portRangeFrom)),
		ToPort:     _aws.Int64(int64(portRangeTo)),
	}

	if strings.HasPrefix(source, "sg-") {
		// Source: other security group
		params.IpPermissions = []*_ec2.IpPermission{
			{
				UserIdGroupPairs: []*_ec2.UserIdGroupPair{
					{GroupId: _aws.String(source)},
				},
				IpProtocol: _aws.String(protocol),
				FromPort:   _aws.Int64(int64(portRangeFrom)),
				ToPort:     _aws.Int64(int64(portRangeTo)),
			},
		}
	} else if cidrRE.MatchString(source) {
		// Source: IP CIDR
		params.CidrIp = _aws.String(source)
		params.IpProtocol = _aws.String(protocol)
		params.FromPort = _aws.Int64(int64(portRangeFrom))
		params.ToPort = _aws.Int64(int64(portRangeTo))
	} else {
		return fmt.Errorf("Invalid source [%s]", source)
	}

	_, err := c.svc.AuthorizeSecurityGroupIngress(params)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RemoveInboundToSecurityGroup(securityGroupID, protocol string, portRangeFrom, portRangeTo uint16, source string) error {
	params := &_ec2.RevokeSecurityGroupIngressInput{
		IpProtocol: _aws.String(protocol),
		GroupId:    _aws.String(securityGroupID),
		FromPort:   _aws.Int64(int64(portRangeFrom)),
		ToPort:     _aws.Int64(int64(portRangeTo)),
	}

	if strings.HasPrefix(source, "sg-") {
		// Source: other security group
		params.IpPermissions = []*_ec2.IpPermission{
			{
				UserIdGroupPairs: []*_ec2.UserIdGroupPair{
					{GroupId: _aws.String(source)},
				},
				IpProtocol: _aws.String(protocol),
				FromPort:   _aws.Int64(int64(portRangeFrom)),
				ToPort:     _aws.Int64(int64(portRangeTo)),
			},
		}
	} else if cidrRE.MatchString(source) {
		// Source: IP CIDR
		params.CidrIp = _aws.String(source)
		params.IpProtocol = _aws.String(protocol)
		params.FromPort = _aws.Int64(int64(portRangeFrom))
		params.ToPort = _aws.Int64(int64(portRangeTo))
	} else {
		return fmt.Errorf("Invalid source [%s]", source)
	}

	_, err := c.svc.RevokeSecurityGroupIngress(params)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RetrieveSecurityGroup(id string) (*_ec2.SecurityGroup, error) {
	// NOTE: used Filter instead of GroupIds attribute because GroupIds
	// returns error when it cannot find the matching security groups.
	params := &_ec2.DescribeSecurityGroupsInput{
		Filters: []*_ec2.Filter{
			{
				Name:   _aws.String("group-id"),
				Values: _aws.StringSlice([]string{id}),
			},
		},
	}

	res, err := c.svc.DescribeSecurityGroups(params)
	if err != nil {
		return nil, err
	}

	if len(res.SecurityGroups) > 0 {
		return res.SecurityGroups[0], nil
	} else {
		return nil, nil
	}
}

func (c *Client) RetrieveSecurityGroups(securityGroupIDs []string) ([]*_ec2.SecurityGroup, error) {
	// NOTE: used Filter instead of GroupIds attribute because GroupIds
	// returns error when it cannot find the matching security groups.
	params := &_ec2.DescribeSecurityGroupsInput{
		Filters: []*_ec2.Filter{
			{
				Name:   _aws.String("group-id"),
				Values: _aws.StringSlice(securityGroupIDs),
			},
		},
	}

	res, err := c.svc.DescribeSecurityGroups(params)
	if err != nil {
		return nil, err
	}

	return res.SecurityGroups, nil
}

func (c *Client) RetrieveSecurityGroupByName(name string) (*_ec2.SecurityGroup, error) {
	// NOTE: used Filter instead of GroupNames attribute because GroupNames
	// returns error when it cannot find the matching security groups.
	params := &_ec2.DescribeSecurityGroupsInput{
		Filters: []*_ec2.Filter{
			{
				Name:   _aws.String("group-name"),
				Values: _aws.StringSlice([]string{name}),
			},
		},
	}

	res, err := c.svc.DescribeSecurityGroups(params)
	if err != nil {
		return nil, err
	}

	if len(res.SecurityGroups) > 0 {
		return res.SecurityGroups[0], nil
	} else {
		return nil, nil
	}
}

func (c *Client) DeleteSecurityGroup(securityGroupID string) error {
	params := &_ec2.DeleteSecurityGroupInput{
		GroupId: _aws.String(securityGroupID),
	}

	_, err := c.svc.DeleteSecurityGroup(params)

	return err
}

func (c *Client) CreateInstances(instanceType, imageID string, instanceCount uint16, securityGroupIDs []string, keyPairName, subnetID, iamInstanceProfileName, userData string) ([]*_ec2.Instance, error) {
	params := &_ec2.RunInstancesInput{
		EbsOptimized:       _aws.Bool(false),
		IamInstanceProfile: &_ec2.IamInstanceProfileSpecification{Name: _aws.String(iamInstanceProfileName)},
		ImageId:            _aws.String(imageID),
		InstanceType:       _aws.String(instanceType),
		KeyName:            _aws.String(keyPairName),
		MaxCount:           _aws.Int64(int64(instanceCount)),
		MinCount:           _aws.Int64(int64(instanceCount)),
		SecurityGroupIds:   _aws.StringSlice(securityGroupIDs),
		SubnetId:           _aws.String(subnetID),
		UserData:           _aws.String(userData),
	}

	res, err := c.svc.RunInstances(params)
	if err != nil {
		return nil, err
	}

	return res.Instances, nil
}

func (c *Client) RetrieveVPC(vpcID string) (*_ec2.Vpc, error) {
	params := &_ec2.DescribeVpcsInput{
		VpcIds: _aws.StringSlice([]string{vpcID}),
	}

	res, err := c.svc.DescribeVpcs(params)
	if err != nil {
		return nil, err
	}

	if res.Vpcs != nil && len(res.Vpcs) > 0 {
		return res.Vpcs[0], nil
	}

	return nil, nil
}

func (c *Client) RetrieveDefaultVPC() (*_ec2.Vpc, error) {
	params := &_ec2.DescribeVpcsInput{
		Filters: []*_ec2.Filter{
			{
				Name:   _aws.String("isDefault"),
				Values: _aws.StringSlice([]string{"true"}),
			},
		},
	}

	res, err := c.svc.DescribeVpcs(params)
	if err != nil {
		return nil, err
	}

	if res.Vpcs != nil && len(res.Vpcs) > 0 {
		return res.Vpcs[0], nil
	}

	return nil, nil
}

func (c *Client) ListVPCs() ([]string, error) {
	params := &_ec2.DescribeVpcsInput{}

	res, err := c.svc.DescribeVpcs(params)
	if err != nil {
		return nil, err
	}

	vpcIDs := []string{}
	for _, v := range res.Vpcs {
		vpcIDs = append(vpcIDs, conv.S(v.VpcId))
	}

	return vpcIDs, nil
}

func (c *Client) ListVPCSubnets(vpcID string) ([]string, error) {
	params := &_ec2.DescribeSubnetsInput{
		Filters: []*_ec2.Filter{
			{
				Name:   _aws.String("vpc-id"),
				Values: _aws.StringSlice([]string{vpcID}),
			},
		},
	}

	res, err := c.svc.DescribeSubnets(params)
	if err != nil {
		return nil, err
	}

	subnetIDs := []string{}
	for _, s := range res.Subnets {
		subnetIDs = append(subnetIDs, conv.S(s.SubnetId))
	}

	return subnetIDs, nil
}

func (c *Client) RetrieveKeyPair(keyPairName string) (*_ec2.KeyPairInfo, error) {
	params := &_ec2.DescribeKeyPairsInput{
		KeyNames: _aws.StringSlice([]string{keyPairName}),
	}

	res, err := c.svc.DescribeKeyPairs(params)
	if err != nil {
		return nil, err
	}

	if res.KeyPairs != nil && len(res.KeyPairs) > 0 {
		return res.KeyPairs[0], nil
	}

	return nil, nil
}
