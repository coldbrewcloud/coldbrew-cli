package logs

import (
	_aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	_logs "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type Client struct {
	svc       *_logs.CloudWatchLogs
	awsRegion string
}

func New(session *session.Session, config *_aws.Config) *Client {
	return &Client{
		awsRegion: *config.Region,
		svc:       _logs.New(session, config),
	}
}

func (c *Client) CreateGroup(groupName string) error {
	params := &_logs.CreateLogGroupInput{
		LogGroupName: _aws.String(groupName),
	}

	_, err := c.svc.CreateLogGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ListGroups(groupNamePrefix string) ([]*_logs.LogGroup, error) {
	var nextToken *string
	groups := []*_logs.LogGroup{}

	for {
		params := &_logs.DescribeLogGroupsInput{
			LogGroupNamePrefix: _aws.String(groupNamePrefix),
			NextToken:          nextToken,
		}

		res, err := c.svc.DescribeLogGroups(params)
		if err != nil {
			return nil, err
		}

		groups = append(groups, res.LogGroups...)

		if res.NextToken == nil {
			break
		} else {
			nextToken = res.NextToken
		}
	}

	return groups, nil
}
