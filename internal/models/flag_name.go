package models

type FlagEnum int

const (
	FlagGetOpeningBalanceV2 FlagEnum = iota + 1
	FlagGetTrialBalanceV2
	FlagT24CreateBranchAccountPAS
	FlagT24CreateLenderAccountPAS
	FlagT24CreateLoanAccountPAS
	FlagGuestModePayment
	FlagChunkSizeAccountBalance
	FlagGetTrialBalanceGCS
	FlagTrialBalanceAutoAdjustment
	FlagGetOpeningBalanceFromPreviousMonth
)

func (f FlagEnum) String() string {
	switch f {
	case FlagGetOpeningBalanceV2:
		return "PAS-BE-flag-get-opening-balance-v2"
	case FlagGetTrialBalanceV2:
		return "PAS-BE-flag-get-trial-balance-v2"
	case FlagT24CreateBranchAccountPAS:
		return "PAS-BE-flag-t24-create-branch-account-pas"
	case FlagT24CreateLenderAccountPAS:
		return "PAS-BE-flag-t24-create-lender-account-pas"
	case FlagT24CreateLoanAccountPAS:
		return "PAS-BE-flag-t24-create-loan-account-pas"
	case FlagGuestModePayment:
		return "PAYMENT-BE-show-only-amf-transaction-list"
	case FlagChunkSizeAccountBalance:
		return "PAS-BE-chunk-size-account-balance"
	case FlagGetTrialBalanceGCS:
		return "PAS-BE-flag-get-trial-balance-gcs"
	case FlagTrialBalanceAutoAdjustment:
		return "PAS-BE-flag-trial-balance-auto-adjustment"
	case FlagGetOpeningBalanceFromPreviousMonth:
		return "PAS-BE-get-opening-balance-from-previous-month"
	default:
		return "UNKNOWN"
	}
}

func GetAll() []string {
	var flags []string
	for i := 1; FlagEnum(i).String() != "UNKNOWN"; i++ {
		flags = append(flags, FlagEnum(i).String())
	}
	return flags
}

type ChunkSizeVariant struct {
	ChunkSize int `json:"chunkSize"`
}
