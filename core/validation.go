package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	MaxAppUnits  = uint16(1000)
	MaxAppCPU    = float64(1024 * 16)
	MaxAppMemory = uint64(1024 * 16)
)

var (
	AppNameRE           = regexp.MustCompile(`^[\w\-]{1,32}$`)
	ClusterNameRE       = regexp.MustCompile(`^[\w\-]{1,32}$`)
	ELBNameRE           = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]{0,30}(?:[a-zA-Z0-9])?$`)
	ELBTargetNameRE     = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]{0,30}(?:[a-zA-Z0-9])?$`)
	ECRRepoNameRE       = regexp.MustCompile(`^.{1,256}$`)                       // TODO: need better matcher
	HealthCheckPathRE   = regexp.MustCompile(`^.+$`)                             // TODO: need better matcher
	HealthCheckStatusRE = regexp.MustCompile(`^\d{3}-\d{3}$|^\d{3}(?:,\d{3})*$`) // "200", "200-299", "200,204,201"
	DockerImageURIRE    = regexp.MustCompile(`^([^:]+)(?::([^:]+))?$`)

	SizeExpressionRE = regexp.MustCompile(`^(\d+)(?:([kmgtKMGT])([bB])?)?$`)
	TimeExpressionRE = regexp.MustCompile(`^(\d+)([smhSMH])?$`)
)

func ParseSizeExpression(expression string) (uint64, error) {
	m := SizeExpressionRE.FindAllStringSubmatch(expression, -1)
	if len(m) != 1 || len(m[0]) < 2 {
		return 0, fmt.Errorf("Invalid size expression [%s]", expression)
	}

	multiplier := uint64(1)
	switch strings.ToLower(m[0][2]) {
	case "k":
		multiplier = uint64(1000)
	case "m":
		multiplier = uint64(1000 * 1000)
	case "g":
		multiplier = uint64(1000 * 1000 * 1000)
	case "t":
		multiplier = uint64(1000 * 1000 * 1000 * 1000)
	}

	parsed, err := strconv.ParseUint(m[0][1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid size expression [%s]: %s", expression, err.Error())
	}

	return parsed * multiplier, nil
}

func ParseTimeExpression(expression string) (uint64, error) {
	m := TimeExpressionRE.FindAllStringSubmatch(expression, -1)
	if len(m) != 1 || len(m[0]) < 1 {
		return 0, fmt.Errorf("Invalid time expression [%s]", expression)
	}

	multiplier := uint64(1)
	switch strings.ToLower(m[0][2]) {
	case "m":
		multiplier = uint64(60)
	case "h":
		multiplier = uint64(60 * 60)
	}

	parsed, err := strconv.ParseUint(m[0][1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid time expression [%s]: %s", expression, err.Error())
	}

	return parsed * multiplier, nil
}
