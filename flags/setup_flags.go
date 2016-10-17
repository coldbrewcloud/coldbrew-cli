package flags

import "gopkg.in/alecthomas/kingpin.v2"

type SetupFlags struct {
	AppName                     *string `json:"app-name,omitempty"`
	LoadBalancerName            *string `json:"load-balancer,omitempty"`
	LoadBalancerTargetGroupName *string `json:"load-balancer-target-group,omitempty"`
	HealthCheckPath             *string `json:"health-check-path"`
	ECSClusterName              *string `json:"cluster-name,omitempty"`
	ECSServiceRoleName          *string `json:"service-role-name,omitempty"`
	ECRNamespace                *string `json:"ecr-namespace"`
	VPCID                       *string `json:"vpc-id"`
}

func NewSetupFlags(kc *kingpin.CmdClause) *SetupFlags {
	return &SetupFlags{
		AppName:                     kc.Flag("app-name", "App name").Default("app1").String(),
		LoadBalancerName:            kc.Flag("load-balancer", "Load balancer name").String(),
		LoadBalancerTargetGroupName: kc.Flag("load-balancer-target-group", "Load balancer target group name").String(),
		HealthCheckPath:             kc.Flag("health-check-path", "Health check URL").Default("/").String(),
		ECRNamespace:                kc.Flag("ecr-namespace", "ECR repository namespace").Default("coldbrew").String(),
		ECSClusterName:              kc.Flag("cluster-name", "ECS cluster name").Default("coldbrew").String(),
		ECSServiceRoleName:          kc.Flag("service-role-name", "ECS service role name").Default("ecsServiceRole").String(),
		VPCID:                       kc.Flag("vpc", "VPC ID").String(),
	}
}
