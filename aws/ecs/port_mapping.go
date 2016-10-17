package ecs

type PortMapping struct {
	ContainerPort uint16 `json:"container_port"`
	Protocol      string `json:"protocol"`
}
