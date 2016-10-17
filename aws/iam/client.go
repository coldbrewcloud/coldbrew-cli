package iam

import (
	"errors"
	"net/http"

	_aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	_iam "github.com/aws/aws-sdk-go/service/iam"
)

type Client struct {
	svc *_iam.IAM
}

func New(session *session.Session, config *_aws.Config) *Client {
	return &Client{
		svc: _iam.New(session, config),
	}
}

func (c *Client) RetrieveRole(roleName string) (*_iam.Role, error) {
	if roleName == "" {
		return nil, errors.New("roleName is empty")
	}

	params := &_iam.GetRoleInput{
		RoleName: _aws.String(roleName),
	}

	res, err := c.svc.GetRole(params)
	if err != nil {
		if reqFail, ok := err.(awserr.RequestFailure); ok {
			if reqFail.StatusCode() == http.StatusNotFound {
				return nil, nil
			}
		}
		return nil, err
	}

	return res.Role, nil
}

func (c *Client) CreateRole(assumeRolePolicyDocument, roleName string) (*_iam.Role, error) {
	if assumeRolePolicyDocument == "" {
		return nil, errors.New("assumeRolePolicyDocument is empty")
	}
	if roleName == "" {
		return nil, errors.New("roleName is empty")
	}

	params := &_iam.CreateRoleInput{
		AssumeRolePolicyDocument: _aws.String(assumeRolePolicyDocument),
		RoleName:                 _aws.String(roleName),
		Path:                     _aws.String("/"),
	}

	res, err := c.svc.CreateRole(params)
	if err != nil {
		return nil, err
	}

	return res.Role, nil
}

func (c *Client) AttachRolePolicy(policyARN, roleName string) error {
	if policyARN == "" {
		return errors.New("policyARN is empty")
	}
	if roleName == "" {
		return errors.New("roleName is empty")
	}

	params := &_iam.AttachRolePolicyInput{
		PolicyArn: _aws.String(policyARN),
		RoleName:  _aws.String(roleName),
	}

	_, err := c.svc.AttachRolePolicy(params)
	if err != nil {
		return err
	}

	return nil
}
