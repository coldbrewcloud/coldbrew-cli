package aws

import (
	_aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/coldbrewcloud/coldbrew-cli/aws/autoscaling"
	"github.com/coldbrewcloud/coldbrew-cli/aws/ec2"
	"github.com/coldbrewcloud/coldbrew-cli/aws/ecr"
	"github.com/coldbrewcloud/coldbrew-cli/aws/ecs"
	"github.com/coldbrewcloud/coldbrew-cli/aws/elb"
	"github.com/coldbrewcloud/coldbrew-cli/aws/iam"
	"github.com/coldbrewcloud/coldbrew-cli/aws/logs"
	"github.com/coldbrewcloud/coldbrew-cli/aws/sns"
)

type Client struct {
	session *session.Session
	config  *_aws.Config

	autoScalingClient *autoscaling.Client
	ec2Client         *ec2.Client
	ecsClient         *ecs.Client
	elbClient         *elb.Client
	ecrClient         *ecr.Client
	iamClient         *iam.Client
	snsClient         *sns.Client
	logsClient        *logs.Client
}

func NewClient(region, accessKey, secretKey string) *Client {
	config := _aws.NewConfig().WithRegion(region)
	if accessKey != "" {
		config = config.WithCredentials(credentials.NewStaticCredentials(accessKey, secretKey, ""))
	}

	return &Client{
		session: session.New(),
		config:  config,
	}
}

func (c *Client) AutoScaling() *autoscaling.Client {
	if c.autoScalingClient == nil {
		c.autoScalingClient = autoscaling.New(c.session, c.config)
	}
	return c.autoScalingClient
}

func (c *Client) EC2() *ec2.Client {
	if c.ec2Client == nil {
		c.ec2Client = ec2.New(c.session, c.config)
	}
	return c.ec2Client
}

func (c *Client) ECS() *ecs.Client {
	if c.ecsClient == nil {
		c.ecsClient = ecs.New(c.session, c.config)
	}
	return c.ecsClient
}

func (c *Client) ELB() *elb.Client {
	if c.elbClient == nil {
		c.elbClient = elb.New(c.session, c.config)
	}
	return c.elbClient
}

func (c *Client) ECR() *ecr.Client {
	if c.ecrClient == nil {
		c.ecrClient = ecr.New(c.session, c.config)
	}
	return c.ecrClient
}

func (c *Client) IAM() *iam.Client {
	if c.iamClient == nil {
		c.iamClient = iam.New(c.session, c.config)
	}
	return c.iamClient
}

func (c *Client) SNS() *sns.Client {
	if c.snsClient == nil {
		c.snsClient = sns.New(c.session, c.config)
	}
	return c.snsClient
}

func (c *Client) CloudWatchLogs() *logs.Client {
	if c.logsClient == nil {
		c.logsClient = logs.New(c.session, c.config)
	}
	return c.logsClient
}
