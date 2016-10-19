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
