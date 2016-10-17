package flags

import "gopkg.in/alecthomas/kingpin.v2"

type DeployFlags struct {
	AppName                     *string            `json:"app-name,omitempty"`
	AppVersion                  *string            `json:"app-version"`
	AppPath                     *string            `json:"app-path,omitempty"`
	ContainerPort               *uint16            `json:"container-port"`
	LoadBalancerName            *string            `json:"load-balancer,omitempty"`
	LoadBalancerTargetGroupName *string            `json:"load-balancer-target-group,omitempty"`
	HealthCheckPath             *string            `json:"health-check-path"`
	DockerBinPath               *string            `json:"docker-bin,omitempty"`
	DockerfilePath              *string            `json:"dockerfile,omitempty"`
	DockerImage                 *string            `json:"docker-image,omitempty"`
	Units                       *uint16            `json:"units,omitempty"`
	CPU                         *uint64            `json:"cpu,omitempty"`
	Memory                      *uint64            `json:"memory,omitempty"`
	Envs                        *map[string]string `json:"env,omitempty"`
	EnvsFile                    *string            `json:"env-file,omitempty"`
	ECSClusterName              *string            `json:"cluster-name,omitempty"`
	ECSServiceRoleName          *string            `json:"service-role-name,omitempty"`
	ECRNamespace                *string            `json:"ecr-namespace"`
	VPCID                       *string            `json:"vpc-id"`
	CloudWatchLogs              *bool              `json:"cloud-watch-logs,omitempty"`
}

func NewDeployFlags(kc *kingpin.CmdClause) *DeployFlags {
	return &DeployFlags{
		AppName:                     kc.Flag("app-name", "App name").Default("app1").String(),
		AppVersion:                  kc.Flag("app-version", "App version").Default("1.0.0").String(),
		AppPath:                     kc.Flag("app-path", "App path").Default(".").String(),
		ContainerPort:               kc.Flag("container-port", "App container port").Default("0").Uint16(),
		LoadBalancerName:            kc.Flag("load-balancer", "Load balancer name").String(),
		LoadBalancerTargetGroupName: kc.Flag("load-balancer-target-group", "Load balancer target group name").String(),
		HealthCheckPath:             kc.Flag("health-check-path", "Health check URL").Default("/").String(),
		DockerBinPath:               kc.Flag("docker-bin", "Docker binary path").Default("docker").String(),
		DockerfilePath:              kc.Flag("dockerfile", "Dockerfile path").Default("./Dockerfile").String(),
		DockerImage:                 kc.Flag("docker-image", "Docker image (should include image tag)").String(),
		Units:                       kc.Flag("units", "Desired count").Default("1").Uint16(),
		CPU:                         kc.Flag("cpu", "Docker CPU resource (1 unit: 1024)").Default("128").Uint64(),
		Memory:                      kc.Flag("memory", "Docker memory resource (in MB)").Default("128").Uint64(),
		Envs:                        kc.Flag("env", "App environment variable (\"key=value\")").Short('E').StringMap(),
		EnvsFile:                    kc.Flag("env-file", "App environment variable file (JSON)").String(),
		ECRNamespace:                kc.Flag("ecr-namespace", "ECR repository namespace").Default("coldbrew").String(),
		ECSClusterName:              kc.Flag("cluster-name", "ECS cluster name").Default("coldbrew").String(),
		ECSServiceRoleName:          kc.Flag("service-role-name", "ECS service role name").Default("ecsServiceRole").String(),
		VPCID:                       kc.Flag("vpc", "VPC ID").String(),
		CloudWatchLogs:              kc.Flag("cloud-watch-logs", "Enable AWS CloudWatch Logs").Bool(),
	}
}
