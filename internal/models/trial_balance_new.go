package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type AccountTrialBalance struct {
	ClosingDate     time.Time
	EntityCode      string
	CategoryCode    string
	SubCategoryCode string
	DebitMovement   decimal.Decimal
	CreditMovement  decimal.Decimal
	OpeningBalance  decimal.Decimal
	ClosingBalance  decimal.Decimal
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CalculateTrialBalance struct {
	COATypeCode     string
	EntityCode      string
	SubCategoryCode string
	Date            time.Time
}

type GetTrialBalanceV2Out struct {
	EntityCode      string
	CoaTypeCode     string
	CoaTypeName     string
	CategoryCode    string
	CategoryName    string
	SubCategoryCode string
	SubCategoryName string
	DebitMovement   decimal.Decimal
	CreditMovement  decimal.Decimal
	OpeningBalance  decimal.Decimal
	ClosingBalance  decimal.Decimal
}

type TrialBalanceCSV struct {
	EntityCode      string
	CategoryCode    string
	SubCategoryCode string
	DebitMovement   decimal.Decimal
	CreditMovement  decimal.Decimal
	OpeningBalance  decimal.Decimal
	ClosingBalance  decimal.Decimal
}

type GetTransactionsIn struct {
	TransactionDate time.Time
	CreatedAt       time.Time
}
