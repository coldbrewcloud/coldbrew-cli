package ecs

import (
	"errors"
	"fmt"

	_aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	_ecs "github.com/aws/aws-sdk-go/service/ecs"
)

type Client struct {
	svc       *_ecs.ECS
	awsRegion string
}

func New(session *session.Session, config *_aws.Config) *Client {
	return &Client{
		awsRegion: *config.Region,
		svc:       _ecs.New(session, config),
	}
}

func (c *Client) RetrieveCluster(clusterName string) (*_ecs.Cluster, error) {
	if clusterName == "" {
		return nil, errors.New("clusterName is empty")
	}

	params := &_ecs.DescribeClustersInput{
		Clusters: _aws.StringSlice([]string{clusterName}),
	}
	res, err := c.svc.DescribeClusters(params)
	if err != nil {
		return nil, err
	}

	if len(res.Clusters) == 0 {
		return nil, nil
	} else if len(res.Clusters) == 1 {
		return res.Clusters[0], nil
	}

	return nil, fmt.Errorf("Invalid result: %v", res.Clusters)
}

func (c *Client) CreateCluster(clusterName string) (*_ecs.Cluster, error) {
	if clusterName == "" {
		return nil, errors.New("clusterName is empty")
	}

	params := &_ecs.CreateClusterInput{
		ClusterName: _aws.String(clusterName),
	}

	res, err := c.svc.CreateCluster(params)
	if err != nil {
		return nil, err
	}

	return res.Cluster, nil
}

func (c *Client) UpdateTaskDefinition(taskDefinitionName, image, taskContainerName string, cpu, memory uint64, envs map[string]string, portMappings []PortMapping, cloudWatchLogs bool) (*_ecs.TaskDefinition, error) {
	if taskDefinitionName == "" {
		return nil, errors.New("taskDefinitionName is empty")
	}
	if image == "" {
		return nil, errors.New("image is empty")
	}
	if taskContainerName == "" {
		return nil, errors.New("taskContainerName is empty")
	}

	params := &_ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: []*_ecs.ContainerDefinition{
			{
				Name:             _aws.String(taskContainerName),
				Cpu:              _aws.Int64(int64(cpu)),
				Memory:           _aws.Int64(int64(memory)),
				Essential:        _aws.Bool(true),
				Image:            _aws.String(image),
				LogConfiguration: nil,
			},
		},
		Family: _aws.String(taskDefinitionName),
	}

	if cloudWatchLogs {
		params.ContainerDefinitions[0].LogConfiguration = &_ecs.LogConfiguration{
			LogDriver: _aws.String(_ecs.LogDriverAwslogs),
			Options: _aws.StringMap(map[string]string{
				"awslogs-group":  "coldbrewcloud-deploy-logs",
				"awslogs-region": c.awsRegion,
			}),
		}
	}

	for ek, ev := range envs {
		params.ContainerDefinitions[0].Environment = append(params.ContainerDefinitions[0].Environment, &_ecs.KeyValuePair{
			Name:  _aws.String(ek),
			Value: _aws.String(ev),
		})
	}

	for _, pm := range portMappings {
		params.ContainerDefinitions[0].PortMappings = append(params.ContainerDefinitions[0].PortMappings, &_ecs.PortMapping{
			ContainerPort: _aws.Int64(int64(pm.ContainerPort)),
			HostPort:      _aws.Int64(0),
			Protocol:      _aws.String(pm.Protocol),
		})
	}

	res, err := c.svc.RegisterTaskDefinition(params)
	if err != nil {
		return nil, err
	}

	return res.TaskDefinition, nil
}

func (c *Client) RetrieveService(clusterName, serviceName string) (*_ecs.Service, error) {
	if clusterName == "" {
		return nil, errors.New("clusterName is empty")
	}
	if serviceName == "" {
		return nil, errors.New("serviceName is empty")
	}

	params := &_ecs.DescribeServicesInput{
		Cluster:  _aws.String(clusterName),
		Services: _aws.StringSlice([]string{serviceName}),
	}

	res, err := c.svc.DescribeServices(params)
	if err != nil {
		return nil, err
	}

	if len(res.Services) == 0 {
		return nil, nil
	} else if len(res.Services) == 1 {
		return res.Services[0], nil
	}

	return nil, fmt.Errorf("Invalid result: %v", res.Services)
}

func (c *Client) CreateService(clusterName, serviceName, taskDefARN string, desiredCount uint16, loadBalancers []*LoadBalancer, serviceRole string) (*_ecs.Service, error) {
	if clusterName == "" {
		return nil, errors.New("clusterName is empty")
	}
	if serviceName == "" {
		return nil, errors.New("serviceName is empty")
	}
	if taskDefARN == "" {
		return nil, errors.New("taskDefARN is empty")
	}

	params := &_ecs.CreateServiceInput{
		DesiredCount:   _aws.Int64(int64(desiredCount)),
		ServiceName:    _aws.String(serviceName),
		TaskDefinition: _aws.String(taskDefARN),
		Cluster:        _aws.String(clusterName),
		DeploymentConfiguration: &_ecs.DeploymentConfiguration{
			MaximumPercent:        _aws.Int64(200),
			MinimumHealthyPercent: _aws.Int64(50),
		},
	}

	if loadBalancers != nil && len(loadBalancers) > 0 {
		params.LoadBalancers = []*_ecs.LoadBalancer{}

		for _, lb := range loadBalancers {

			params.LoadBalancers = append(params.LoadBalancers, &_ecs.LoadBalancer{
				ContainerName:  _aws.String(lb.TaskContainerName),
				ContainerPort:  _aws.Int64(int64(lb.TaskContainerPort)),
				TargetGroupArn: _aws.String(lb.ELBTargetGroupARN),
			})
		}

		params.Role = _aws.String(serviceRole)
	}

	res, err := c.svc.CreateService(params)
	if err != nil {
		return nil, err
	}

	return res.Service, nil
}

func (c *Client) UpdateService(clusterName, serviceName, taskDefARN string, desiredCount uint16) (*_ecs.Service, error) {
	if clusterName == "" {
		return nil, errors.New("clusterName is empty")
	}
	if serviceName == "" {
		return nil, errors.New("serviceName is empty")
	}
	if taskDefARN == "" {
		return nil, errors.New("taskDefARN is empty")
	}

	params := &_ecs.UpdateServiceInput{
		Service:        _aws.String(serviceName),
		Cluster:        _aws.String(clusterName),
		DesiredCount:   _aws.Int64(int64(desiredCount)),
		TaskDefinition: _aws.String(taskDefARN),
		DeploymentConfiguration: &_ecs.DeploymentConfiguration{
			MaximumPercent:        _aws.Int64(200),
			MinimumHealthyPercent: _aws.Int64(50),
		},
	}

	res, err := c.svc.UpdateService(params)
	if err != nil {
		return nil, err
	}

	return res.Service, nil
}
