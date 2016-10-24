package aws

import "strings"

func GetIAMInstanceProfileNameFromARN(arn string) string {
	// format: "arn:aws:iam::865092420289:instance-profile/coldbrew_cluster1_instance_profile"
	tokens := strings.Split(arn, "/")
	if len(tokens) == 0 {
		return ""
	}
	return tokens[len(tokens)-1]
}

func GetECSTaskDefinitionFamilyAndRevisionFromARN(arn string) string {
	// format: "arn:aws:ecs:us-west-2:865092420289:task-definition/echo:112"
	tokens := strings.Split(arn, "/")
	if len(tokens) == 0 {
		return ""
	}
	return tokens[len(tokens)-1]
}

func GetECSContainerInstanceIDFromARN(arn string) string {
	// format: "arn:aws:ecs:us-west-2:865092420289:container-instance/72b93c91-0572-4d9d-b3d6-6e5cc5a0d2be"
	tokens := strings.Split(arn, "/")
	if len(tokens) == 0 {
		return ""
	}
	return tokens[len(tokens)-1]
}
