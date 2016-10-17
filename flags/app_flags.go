package flags

import "gopkg.in/alecthomas/kingpin.v2"

type AppFlags struct {
	Debug          *bool   `json:"debug,omitempty"`
	DebugLogPrefix *string `json:"debug-log-prefix,omitempty"`
	AWSRegion      *string `json:"aws-region,omitempty"`
	AWSAccessKey   *string `json:"aws-access-key,omitempty"`
	AWSSecretKey   *string `json:"aws-secret-key,omitempty"`
}

func NewAppFlags(ka *kingpin.Application) *AppFlags {
	return &AppFlags{
		Debug:          ka.Flag("debug", "Enable debug mode").Short('D').Default("false").Bool(),
		DebugLogPrefix: ka.Flag("debug-log-prefix", "Debug output prefix").Default("").String(),
		AWSRegion:      ka.Flag("aws-region", "AWS region name (default: $AWS_REGION)").Envar("AWS_REGION").Default("us-east-1").String(),
		AWSAccessKey:   ka.Flag("aws-access-key", "AWS access key (default: $AWS_ACCESS_KEY_ID)").Envar("AWS_ACCESS_KEY_ID").Default("").String(),
		AWSSecretKey:   ka.Flag("aws-secret-key", "AWS secret key (default: $AWS_SECRET_ACCESS_KEY)").Envar("AWS_SECRET_ACCESS_KEY").Default("").String(),
	}
}
