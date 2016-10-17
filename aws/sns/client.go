package sns

import (
	_aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	_sns "github.com/aws/aws-sdk-go/service/sns"
)

type Client struct {
	svc *_sns.SNS
}

func New(session *session.Session, config *_aws.Config) *Client {
	return &Client{
		svc: _sns.New(session, config),
	}
}

func (c *Client) PublishToTopic(subject, message, topicARN string) error {
	params := &_sns.PublishInput{
		Message:  _aws.String(message),
		Subject:  _aws.String(subject),
		TopicArn: _aws.String(topicARN),
	}

	_, err := c.svc.Publish(params)
	if err != nil {
		return err
	}

	return nil
}
