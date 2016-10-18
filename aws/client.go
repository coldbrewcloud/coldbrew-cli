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
	"github.com/coldbrewcloud/coldbrew-cli/aws/sns"
)

type Client struct {
	session *session.Session
	config  *_aws.Config
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
	return autoscaling.New(c.session, c.config)
}

func (c *Client) EC2() *ec2.Client {
	return ec2.New(c.session, c.config)
}

func (c *Client) ECS() *ecs.Client {
	return ecs.New(c.session, c.config)
}

func (c *Client) ELB() *elb.Client {
	return elb.New(c.session, c.config)
}

func (c *Client) ECR() *ecr.Client {
	return ecr.New(c.session, c.config)
}

func (c *Client) IAM() *iam.Client {
	return iam.New(c.session, c.config)
}

func (c *Client) SNS() *sns.Client {
	return sns.New(c.session, c.config)
}
