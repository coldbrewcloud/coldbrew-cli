package conv

import "strconv"

func SP(v string) *string {
	p := v
	return &p
}

func S(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func U16P(v uint16) *uint16 {
	p := v
	return &p
}

func U16(p *uint16) uint16 {
	if p == nil {
		return 0
	}
	return *p
}

func U64P(v uint64) *uint64 {
	p := v
	return &p
}

func U64(p *uint64) uint64 {
	if p == nil {
		return 0
	}
	return *p
}

func F64(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func F64P(v float64) *float64 {
	p := v
	return &p
}

func B(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

func BP(v bool) *bool {
	p := v
	return &p
}

func I64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

func I64S(v int64) string {
	return strconv.FormatInt(v, 10)
}
