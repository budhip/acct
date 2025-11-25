package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type AccountBalanceDaily struct {
	BalanceDate     time.Time
	AccountNumber   string
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

type CategorySubCategoryCOAType struct {
	CategoryCode    string
	CategoryName    string
	SubCategoryCode string
	SubCategoryName string
	CoaTypeCode     string
	CoaTypeName     string
}

type AccountJournalTransation struct {
	AccountNumber  string
	DebitMovement  decimal.Decimal
	CreditMovement decimal.Decimal
}

type AccountTransation struct {
	AccountNumber string
	Amount        decimal.Decimal
	IsDebit       bool
}
