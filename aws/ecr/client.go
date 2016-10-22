package ecr

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	_aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	_ecr "github.com/aws/aws-sdk-go/service/ecr"
)

type Client struct {
	svc *_ecr.ECR
}

func New(session *session.Session, config *_aws.Config) *Client {
	return &Client{
		svc: _ecr.New(session, config),
	}
}

func (c *Client) RetrieveRepository(repoName string) (*_ecr.Repository, error) {
	if repoName == "" {
		return nil, errors.New("repoName is empty")
	}

	params := &_ecr.DescribeRepositoriesInput{
		RepositoryNames: _aws.StringSlice([]string{repoName}),
	}

	res, err := c.svc.DescribeRepositories(params)
	if err != nil {
		if reqFail, ok := err.(awserr.RequestFailure); ok {
			if reqFail.StatusCode() == http.StatusBadRequest {
				return nil, nil
			}
		}
		return nil, err
	}

	if len(res.Repositories) != 1 {
		return nil, fmt.Errorf("Invali result: %v", res.Repositories)
	}

	return res.Repositories[0], nil
}

func (c *Client) CreateRepository(repoName string) (*_ecr.Repository, error) {
	if repoName == "" {
		return nil, errors.New("repoName is empty")
	}

	params := &_ecr.CreateRepositoryInput{
		RepositoryName: _aws.String(repoName),
	}

	res, err := c.svc.CreateRepository(params)
	if err != nil {
		return nil, err
	}

	return res.Repository, nil
}

func (c *Client) GetDockerLogin() (string, string, string, error) {
	params := &_ecr.GetAuthorizationTokenInput{}

	res, err := c.svc.GetAuthorizationToken(params)
	if err != nil {
		return "", "", "", err
	}

	data, err := base64.StdEncoding.DecodeString(*res.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return "", "", "", err
	}

	tokens := strings.SplitN(string(data), ":", 2)

	return tokens[0], tokens[1], *res.AuthorizationData[0].ProxyEndpoint, nil
}

func (c *Client) DeleteRepository(repoName string) error {
	params := &_ecr.DeleteRepositoryInput{
		Force:          _aws.Bool(true),
		RepositoryName: _aws.String(repoName),
	}

	_, err := c.svc.DeleteRepository(params)

	return err
}
