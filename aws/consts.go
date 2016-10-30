package aws

const (
	AWSRegionUSEast1      = "us-east-1"
	AWSRegionUSEast2      = "us-east-2"
	AWSRegionUSWest1      = "us-west-1"
	AWSRegionUSWest2      = "us-west-2"
	AWSRegionEUWest1      = "eu-west-1"
	AWSRegionEUCentral1   = "eu-central-1"
	AWSRegionAPNorthEast1 = "ap-northeast-1"
	AWSRegionAPSouthEast1 = "ap-southeast-1"
	AWSRegionAPSouthEast2 = "ap-southeast-2"
	AWSRegionSAEast1      = "sa-east-1"

	ECSTaskDefinitionLogDriverJSONFile = "json-file"
	ECSTaskDefinitionLogDriverAWSLogs  = "awslogs"
	ECSTaskDefinitionLogDriverSyslog   = "syslog"
	ECSTaskDefinitionLogDriverJournald = "journald"
	ECSTaskDefinitionLogDriverGelf     = "gelf"
	ECSTaskDefinitionLogDriverFluentd  = "fluentd"
	ECSTaskDefinitionLogDriverSplunk   = "splunk"
)
