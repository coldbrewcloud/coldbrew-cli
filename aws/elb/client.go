package elb

import (
	"fmt"

	_aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	_elb "github.com/aws/aws-sdk-go/service/elbv2"
)

type Client struct {
	svc *_elb.ELBV2
}

func New(session *session.Session, config *_aws.Config) *Client {
	return &Client{
		svc: _elb.New(session, config),
	}
}

func (c *Client) CreateLoadBalancer(elbName string, internetFacing bool, securityGroupIDs, subnetIDs []string) (*_elb.LoadBalancer, error) {
	params := &_elb.CreateLoadBalancerInput{
		Name:           _aws.String(elbName),
		SecurityGroups: _aws.StringSlice(securityGroupIDs),
		Subnets:        _aws.StringSlice(subnetIDs),
	}

	if internetFacing {
		params.Scheme = _aws.String(_elb.LoadBalancerSchemeEnumInternetFacing)
	} else {
		params.Scheme = _aws.String(_elb.LoadBalancerSchemeEnumInternal)
	}

	res, err := c.svc.CreateLoadBalancer(params)
	if err != nil {
		return nil, err
	}

	return res.LoadBalancers[0], nil
}

func (c *Client) RetrieveLoadBalancer(elbName string) (*_elb.LoadBalancer, error) {
	params := &_elb.DescribeLoadBalancersInput{
		Names: _aws.StringSlice([]string{elbName}),
	}
	res, err := c.svc.DescribeLoadBalancers(params)
	if err != nil {
		return nil, err
	}

	if len(res.LoadBalancers) == 0 {
		return nil, nil
	} else if len(res.LoadBalancers) == 1 {
		return res.LoadBalancers[0], nil
	}

	return nil, fmt.Errorf("Invalid result: %v", res.LoadBalancers)
}

func (c *Client) CreateTargetGroup(name string, port uint16, protocol string, vpcID string, healthCheck *HealthCheckParams) (*_elb.TargetGroup, error) {
	params := &_elb.CreateTargetGroupInput{
		Name:     _aws.String(name),
		Port:     _aws.Int64(int64(port)),
		Protocol: _aws.String(protocol),
		VpcId:    _aws.String(vpcID),
	}

	if healthCheck != nil {
		params.HealthCheckIntervalSeconds = _aws.Int64(int64(healthCheck.CheckIntervalSeconds))
		params.HealthCheckPath = _aws.String(healthCheck.CheckPath)
		if healthCheck.CheckPort != nil {
			params.HealthCheckPort = _aws.String(fmt.Sprintf("%d", *healthCheck.CheckPort))
		}
		params.HealthCheckProtocol = _aws.String(healthCheck.Protocol)
		params.HealthCheckTimeoutSeconds = _aws.Int64(int64(healthCheck.CheckTimeoutSeconds))
		params.HealthyThresholdCount = _aws.Int64(int64(healthCheck.HealthyThresholdCount))
		params.UnhealthyThresholdCount = _aws.Int64(int64(healthCheck.UnhealthyThresholdCount))
		params.Matcher = &_elb.Matcher{HttpCode: _aws.String(healthCheck.ExpectedHTTPStatusCodes)}
	}

	res, err := c.svc.CreateTargetGroup(params)
	if err != nil {
		return nil, err
	}

	return res.TargetGroups[0], nil
}

func (c *Client) RetrieveTargetGroup(elbARN, targetGroupName string) (*_elb.TargetGroup, error) {
	params := &_elb.DescribeTargetGroupsInput{
		LoadBalancerArn: _aws.String(elbARN),
	}
	res, err := c.svc.DescribeTargetGroups(params)
	if err != nil {
		return nil, err
	}

	for _, tg := range res.TargetGroups {
		if *tg.TargetGroupName == targetGroupName {
			return tg, nil
		}
	}

	return nil, nil
}

func (c *Client) CreateListener(loadBalancerARN, targetGroupARN string, port uint16, protocol string) error {
	params := &_elb.CreateListenerInput{
		DefaultActions: []*_elb.Action{
			{
				TargetGroupArn: _aws.String(targetGroupARN),
				Type:           _aws.String(_elb.ActionTypeEnumForward),
			},
		},
		LoadBalancerArn: _aws.String(loadBalancerARN),
		Port:            _aws.Int64(int64(port)),
		Protocol:        _aws.String(protocol),
	}

	_, err := c.svc.CreateListener(params)
	if err != nil {
		return err
	}

	return nil
}
