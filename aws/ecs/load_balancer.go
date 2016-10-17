package ecs

type LoadBalancer struct {
	ELBTargetGroupARN string `json:"elb_target_group_arn"`
	TaskContainerName string `json:"task_container_name"`
	TaskContainerPort uint16 `json:"task_container_port"`
}
