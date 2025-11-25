package models

type Status int32

const (
	STATUS_ACTIVE Status = iota
	STATUS_INACTIVE
)

const (
	StatusActive   = "active"
	StatusInActive = "inactive"
)

var (
	MapStatus = map[string]Status{
		StatusActive:   STATUS_ACTIVE,
		StatusInActive: STATUS_INACTIVE,
	}
)
