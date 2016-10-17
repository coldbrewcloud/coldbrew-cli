package autoscaling

import (
	_aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	_autoscaling "github.com/aws/aws-sdk-go/service/autoscaling"
)

type Client struct {
	svc *_autoscaling.AutoScaling
}

func New(session *session.Session, config *_aws.Config) *Client {
	return &Client{
		svc: _autoscaling.New(session, config),
	}
}

func (c *Client) CreateLaunchConfiguration(name, instanceType, imageID string, instanceCount uint16, securityGroupIDs []string, keyPairName, iamInstanceProfileNameOrARN, userData string) error {
	params := &_autoscaling.CreateLaunchConfigurationInput{
		IamInstanceProfile:      _aws.String(iamInstanceProfileNameOrARN),
		ImageId:                 _aws.String(imageID),
		KeyName:                 _aws.String(keyPairName),
		LaunchConfigurationName: _aws.String(name),
		SecurityGroups:          _aws.StringSlice(securityGroupIDs),
		UserData:                _aws.String(userData),
	}

	_, err := c.svc.CreateLaunchConfiguration(params)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CreateAutoScalingGroup(autoScalingGroupName, launchConfigurationName string, subnetIDs []string, minCapacity, maxCapacity, initialCapacity uint16) error {
	params := &_autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName:    _aws.String(autoScalingGroupName),
		DesiredCapacity:         _aws.Int64(int64(initialCapacity)),
		LaunchConfigurationName: _aws.String(launchConfigurationName),
		MaxSize:                 _aws.Int64(int64(maxCapacity)),
		MinSize:                 _aws.Int64(int64(minCapacity)),
		VPCZoneIdentifier:       _aws.StringSlice(subnetIDs),
	}

	_, err := c.svc.CreateAutoScalingGroup(params)
	if err != nil {
		return err
	}

	return nil
}
