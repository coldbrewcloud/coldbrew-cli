package autoscaling

import (
	"strings"

	"fmt"

	_aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	_autoscaling "github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/coldbrewcloud/coldbrew-cli/utils"
	"github.com/coldbrewcloud/coldbrew-cli/utils/conv"
)

type Client struct {
	svc *_autoscaling.AutoScaling
}

func New(session *session.Session, config *_aws.Config) *Client {
	return &Client{
		svc: _autoscaling.New(session, config),
	}
}

func (c *Client) CreateLaunchConfiguration(launchConfigurationName, instanceType, imageID string, securityGroupIDs []string, keyPairName, iamInstanceProfileNameOrARN, userData string) error {
	params := &_autoscaling.CreateLaunchConfigurationInput{
		IamInstanceProfile:      _aws.String(iamInstanceProfileNameOrARN),
		ImageId:                 _aws.String(imageID),
		InstanceType:            _aws.String(instanceType),
		LaunchConfigurationName: _aws.String(launchConfigurationName),
		SecurityGroups:          _aws.StringSlice(securityGroupIDs),
		UserData:                _aws.String(userData),
		InstanceMonitoring:      &_autoscaling.InstanceMonitoring{Enabled: _aws.Bool(false)},
	}

	if !utils.IsBlank(keyPairName) {
		params.KeyName = _aws.String(keyPairName)
	}

	_, err := c.svc.CreateLaunchConfiguration(params)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RetrieveLaunchConfiguration(launchConfigurationName string) (*_autoscaling.LaunchConfiguration, error) {
	params := &_autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: _aws.StringSlice([]string{launchConfigurationName}),
	}

	res, err := c.svc.DescribeLaunchConfigurations(params)
	if err != nil {
		return nil, err
	}

	if res != nil && len(res.LaunchConfigurations) > 0 {
		return res.LaunchConfigurations[0], nil
	}

	return nil, nil
}

func (c *Client) DeleteLaunchConfiguration(launchConfigurationName string) error {
	params := &_autoscaling.DeleteLaunchConfigurationInput{
		LaunchConfigurationName: _aws.String(launchConfigurationName),
	}

	_, err := c.svc.DeleteLaunchConfiguration(params)

	return err
}

func (c *Client) CreateAutoScalingGroup(autoScalingGroupName, launchConfigurationName string, subnetIDs []string, minCapacity, maxCapacity, initialCapacity uint16) error {
	params := &_autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName:    _aws.String(autoScalingGroupName),
		DesiredCapacity:         _aws.Int64(int64(initialCapacity)),
		LaunchConfigurationName: _aws.String(launchConfigurationName),
		MaxSize:                 _aws.Int64(int64(maxCapacity)),
		MinSize:                 _aws.Int64(int64(minCapacity)),
		VPCZoneIdentifier:       _aws.String(strings.Join(subnetIDs, ",")),
	}

	_, err := c.svc.CreateAutoScalingGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RetrieveAutoScalingGroup(autoScalingGroupName string) (*_autoscaling.Group, error) {
	params := &_autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: _aws.StringSlice([]string{autoScalingGroupName}),
	}

	res, err := c.svc.DescribeAutoScalingGroups(params)
	if err != nil {
		return nil, err
	}

	if res != nil && len(res.AutoScalingGroups) > 0 {
		return res.AutoScalingGroups[0], nil
	}

	return nil, nil
}

func (c *Client) UpdateAutoScalingGroupCapacity(autoScalingGroupName string, minCapacity, maxCapacity, desiredCapacity uint16) error {
	params := &_autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: _aws.String(autoScalingGroupName),
		DesiredCapacity:      _aws.Int64(int64(desiredCapacity)),
		MaxSize:              _aws.Int64(int64(maxCapacity)),
		MinSize:              _aws.Int64(int64(minCapacity)),
	}

	_, err := c.svc.UpdateAutoScalingGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) SetAutoScalingGroupDesiredCapacity(autoScalingGroupName string, desiredCapacity uint16) error {
	params := &_autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: _aws.String(autoScalingGroupName),
		DesiredCapacity:      _aws.Int64(int64(desiredCapacity)),
	}

	_, err := c.svc.SetDesiredCapacity(params)

	return err
}

func (c *Client) DeleteAutoScalingGroup(autoScalingGroupName string, forceDelete bool) error {
	params := &_autoscaling.DeleteAutoScalingGroupInput{
		AutoScalingGroupName: _aws.String(autoScalingGroupName),
		ForceDelete:          _aws.Bool(forceDelete),
	}

	_, err := c.svc.DeleteAutoScalingGroup(params)

	return err
}

func (c *Client) AddTagsToAutoScalingGroup(autoScalingGroupName string, tags map[string]string, tagNewInstances bool) error {
	params := &_autoscaling.CreateOrUpdateTagsInput{}

	for tk, tv := range tags {
		params.Tags = append(params.Tags, &_autoscaling.Tag{
			ResourceId:        _aws.String(autoScalingGroupName),
			ResourceType:      _aws.String("auto-scaling-group"),
			Key:               _aws.String(tk),
			Value:             _aws.String(tv),
			PropagateAtLaunch: _aws.Bool(tagNewInstances),
		})
	}

	_, err := c.svc.CreateOrUpdateTags(params)

	return err
}

func (c *Client) RetrieveTagsForAutoScalingGroup(autoScalingGroupName string) (map[string]string, error) {
	params := &_autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: _aws.StringSlice([]string{autoScalingGroupName}),
	}

	res, err := c.svc.DescribeAutoScalingGroups(params)
	if err != nil {
		return nil, err
	}

	if len(res.AutoScalingGroups) == 0 {
		return nil, fmt.Errorf("EC2 Auto Scaling Group [%s] was not found.", autoScalingGroupName)
	}

	tags := map[string]string{}
	for _, t := range res.AutoScalingGroups[0].Tags {
		tags[conv.S(t.Key)] = conv.S(t.Value)
	}

	return tags, nil
}
