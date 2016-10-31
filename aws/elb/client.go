package elb

import (
	"fmt"

	_aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	_elb "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
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

func (c *Client) RetrieveLoadBalancer(elbARN string) (*_elb.LoadBalancer, error) {
	params := &_elb.DescribeLoadBalancersInput{
		LoadBalancerArns: _aws.StringSlice([]string{elbARN}),
	}
	res, err := c.svc.DescribeLoadBalancers(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "LoadBalancerNotFound" {
				return nil, nil
			}
		}
		return nil, err
	}

	if len(res.LoadBalancers) == 0 {
		return nil, nil
	} else if len(res.LoadBalancers) == 1 {
		return res.LoadBalancers[0], nil
	}

	return nil, fmt.Errorf("Invalid result: %v", res.LoadBalancers)
}

func (c *Client) RetrieveLoadBalancerByName(elbName string) (*_elb.LoadBalancer, error) {
	params := &_elb.DescribeLoadBalancersInput{
		Names: _aws.StringSlice([]string{elbName}),
	}
	res, err := c.svc.DescribeLoadBalancers(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "LoadBalancerNotFound" {
				return nil, nil
			}
		}
		return nil, err
	}

	if len(res.LoadBalancers) == 0 {
		return nil, nil
	} else if len(res.LoadBalancers) == 1 {
		return res.LoadBalancers[0], nil
	}

	return nil, fmt.Errorf("Invalid result: %v", res.LoadBalancers)
}

func (c *Client) RetrieveLoadBalancerListeners(loadBalancerARN string) ([]*_elb.Listener, error) {
	listeners := []*_elb.Listener{}
	var marker *string

	for {
		params := &_elb.DescribeListenersInput{
			Marker:          marker,
			LoadBalancerArn: _aws.String(loadBalancerARN),
		}

		res, err := c.svc.DescribeListeners(params)
		if err != nil {
			return nil, err
		}

		for _, p := range res.Listeners {
			listeners = append(listeners, p)
		}

		if utils.IsBlank(conv.S(res.NextMarker)) {
			break
		}

		marker = res.NextMarker
	}

	return listeners, nil
}

func (c *Client) DeleteLoadBalancer(loadBalancerARN string) error {
	params := &_elb.DeleteLoadBalancerInput{
		LoadBalancerArn: _aws.String(loadBalancerARN),
	}

	_, err := c.svc.DeleteLoadBalancer(params)

	return err
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

func (c *Client) RetrieveTargetGroup(targetGroupARN string) (*_elb.TargetGroup, error) {
	params := &_elb.DescribeTargetGroupsInput{
		TargetGroupArns: _aws.StringSlice([]string{targetGroupARN}),
	}
	res, err := c.svc.DescribeTargetGroups(params)
	if err != nil {
		return nil, err
	}

	if len(res.TargetGroups) > 0 {
		return res.TargetGroups[0], nil
	}

	return nil, nil
}

func (c *Client) UpdateTargetGroupHealthCheck(targetGroupARN string, healthCheck *HealthCheckParams) error {
	params := &_elb.ModifyTargetGroupInput{
		TargetGroupArn:             _aws.String(targetGroupARN),
		HealthCheckIntervalSeconds: _aws.Int64(int64(healthCheck.CheckIntervalSeconds)),
		HealthCheckPath:            _aws.String(healthCheck.CheckPath),
		HealthCheckProtocol:        _aws.String(healthCheck.Protocol),
		HealthCheckTimeoutSeconds:  _aws.Int64(int64(healthCheck.CheckTimeoutSeconds)),
		HealthyThresholdCount:      _aws.Int64(int64(healthCheck.HealthyThresholdCount)),
		UnhealthyThresholdCount:    _aws.Int64(int64(healthCheck.UnhealthyThresholdCount)),
		Matcher:                    &_elb.Matcher{HttpCode: _aws.String(healthCheck.ExpectedHTTPStatusCodes)},
	}

	_, err := c.svc.ModifyTargetGroup(params)

	return err
}

func (c *Client) RetrieveTargetGroupByName(targetGroupName string) (*_elb.TargetGroup, error) {
	params := &_elb.DescribeTargetGroupsInput{
		Names: _aws.StringSlice([]string{targetGroupName}),
	}
	res, err := c.svc.DescribeTargetGroups(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "TargetGroupNotFound" {
				return nil, nil
			}
		}
		return nil, err
	}

	if len(res.TargetGroups) > 0 {
		return res.TargetGroups[0], nil
	}

	return nil, nil
}

func (c *Client) DeleteTargetGroup(targetGroupARN string) error {
	params := &_elb.DeleteTargetGroupInput{
		TargetGroupArn: _aws.String(targetGroupARN),
	}

	_, err := c.svc.DeleteTargetGroup(params)

	return err
}

func (c *Client) CreateListener(loadBalancerARN, targetGroupARN string, port uint16, protocol, certificateARN string) error {
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
	if certificateARN != "" {
		params.Certificates = []*_elb.Certificate{{CertificateArn: _aws.String(certificateARN)}}
	}

	_, err := c.svc.CreateListener(params)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CreateTags(resourceARN string, tags map[string]string) error {
	params := &_elb.AddTagsInput{
		ResourceArns: _aws.StringSlice([]string{resourceARN}),
		Tags:         []*_elb.Tag{},
	}

	for tk, tv := range tags {
		params.Tags = append(params.Tags, &_elb.Tag{
			Key:   _aws.String(tk),
			Value: _aws.String(tv),
		})
	}

	_, err := c.svc.AddTags(params)

	return err
}

func (c *Client) RetrieveTags(resourceARN string) (map[string]string, error) {
	params := &_elb.DescribeTagsInput{
		ResourceArns: _aws.StringSlice([]string{resourceARN}),
	}

	res, err := c.svc.DescribeTags(params)
	if err != nil {
		return nil, err
	}

	tags := map[string]string{}
	if len(res.TagDescriptions) == 0 {
		return tags, nil
	}
	for _, t := range res.TagDescriptions[0].Tags {
		tags[conv.S(t.Key)] = conv.S(t.Value)
	}

	return tags, nil
}
