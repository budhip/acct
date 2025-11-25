package models

type AccountStatus int32

const (
	ACCOUNT_STATUS_ACTIVE AccountStatus = iota
	ACCOUNT_STATUS_INACTIVE
)

const (
	AccountStatusActive   = "active"
	AccountStatusInActive = "inactive"
)

var (
	MapAccountStatus = map[AccountStatus]string{
		ACCOUNT_STATUS_ACTIVE:   AccountStatusActive,
		ACCOUNT_STATUS_INACTIVE: AccountStatusInActive,
	}
)
