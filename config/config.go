package config

type Config struct {
	Name         string              `json:"name" yaml:"name"`
	Port         uint16              `json:"port" yaml:"port"`
	Resources    *ConfigResources    `json:"resources" yaml:"resources"`
	Env          map[string]string   `json:"env" yaml:"env"`
	Logging      *ConfigLogging      `json:"logging" yaml:"logging"`
	LoadBalancer *ConfigLoadBalancer `json:"load_balancer" yaml:"load_balancer"`
	AWS          *ConfigAWS          `json:"aws" yaml:"aws"`
	Docker       *ConfigDocker       `json:"docker" yaml:"docker"`
}

type ConfigResources struct {
	CPU    float64 `json:"cpu" yaml:"cpu"`
	Memory string  `json:"memory" yaml:"memory"`
	Units  uint16  `json:"units" yaml:"units"`
}

type ConfigLogging struct {
	Type string `json:"type" yaml:"type"`
}

type ConfigLoadBalancer struct {
	Name          string                         `json:"name" yaml:"name"`
	IsHTTPS       bool                           `json:"https" yaml:"https"`
	Port          uint16                         `json:"port" yaml:"port"`
	SecurityGroup string                         `json:"security_group" yaml:"security_group"`
	HealthCheck   *ConfigLoadBalancerHealthCheck `json:"health_check" yaml:"health_check"`
}

type ConfigLoadBalancerHealthCheck struct {
	Interval       string `json:"interval" yaml:"interval"`
	Path           string `json:"path" yaml:"path"`
	Status         string `json:"status" yaml:"status"`
	Timeout        string `json:"timeout" yaml:"timeout"`
	HealthyLimit   uint16 `json:"healthy_limit" yaml:"healthy_limit"`
	UnhealthyLimit uint16 `json:"unhealthy_limit" yaml:"unhealthy_limit"`
}

type ConfigAWS struct {
	VPC                string                       `json:"vpc" yaml:"vpc"`
	ClusterName        string                       `json:"cluster_name" yaml:"cluster_name"`
	ServiceRole        string                       `json:"service_role" yaml:"service_role"`
	ECRNamespace       string                       `json:"ecr_namespace" yaml:"ecr_namespace"`
	ContainerInstances *ConfigAWSContainerInstances `json:"container_instances" yaml:"container_instances"`
}

type ConfigAWSContainerInstances struct {
	SecurityGroup string `json:"security_group" yaml:"security_group"`
	ImageID       string `json:"image_id" yaml:"image_id"`
	KeyPair       string `json:"keypair" yaml:"keypair"`
	InstanceType  string `json:"instance_type" yaml:"instance_type"`
}

type ConfigDocker struct {
	Bin string `json:"bin" yaml:"bin"`
}
