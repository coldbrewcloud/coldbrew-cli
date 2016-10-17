package conv

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
