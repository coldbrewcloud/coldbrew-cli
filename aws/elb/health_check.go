package elb

type HealthCheckParams struct {
	CheckIntervalSeconds    uint16
	CheckPath               string
	CheckPort               *uint16
	Protocol                string
	ExpectedHTTPStatusCodes string
	CheckTimeoutSeconds     uint16
	HealthyThresholdCount   uint16
	UnhealthyThresholdCount uint16
}
