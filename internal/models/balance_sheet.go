package models

import (
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"

	"github.com/shopspring/decimal"
)

const (
	KindBalanceSheet     = "balanceSheet"
	EntityCodeAMF        = "001"
	KindBalanceSheetData = "balanceSheetData"
)

type GetBalanceSheetRequest struct {
	EntityCode       string `json:"entityCode" query:"entityCode" example:"001"`
	BalanceSheetDate string `json:"balanceSheetDate" query:"balanceSheetDate" example:"2023-12-01"`
}

type BalanceSheetFilterOptions struct {
	EntityCode       string
	BalanceSheetDate time.Time
}

func (req GetBalanceSheetRequest) ToFilterOpts() (opts *BalanceSheetFilterOptions, err error) {
	opts = &BalanceSheetFilterOptions{
		EntityCode: req.EntityCode,
	}

	if req.EntityCode == "" {
		opts.EntityCode = EntityCodeAMF
	}

	if req.BalanceSheetDate != "" {
		opts.BalanceSheetDate, err = atime.ParseStringToDatetime(atime.DateFormatYYYYMMDD, req.BalanceSheetDate)
		if err != nil {
			return nil, err
		}
		if atime.DateEqualToday(opts.BalanceSheetDate) || atime.Now().Before(opts.BalanceSheetDate) {
			return nil, GetErrMap(ErrKeyBalanceSheetDateIsTodayOrLater)
		}
	} else {
		// default is end of previous month
		_, opts.BalanceSheetDate = atime.PrevMonth(atime.Now())
	}
	return opts, nil
}

type GetBalanceSheetResponse struct {
	Kind       string `json:"kind" example:"balanceSheet"`
	EntityCode string `json:"entityCode" example:"001"`
	EntityName string `json:"entityName" example:"AMF"`
	EntityDesc string `json:"entityDesc" example:"PT. Amartha Mikro Fintek"`

	BalanceSheetDate string `json:"balanceSheetDate" example:"2023-01-01"`

	BalanceSheet BalanceSheetData `json:"balanceSheet"`
}

type BalanceSheetData struct {
	Kind string `json:"kind" example:"balanceSheetData"`

	Assets        []BalanceCategory `json:"assets"`
	TotalAsset    string            `json:"totalAsset" example:"1000000"`
	DecTotalAsset decimal.Decimal   `json:"-"`

	Liabilities       []BalanceCategory `json:"liabilities"`
	TotalLiability    string            `json:"totalLiability" example:"1000000"`
	DecTotalLiability decimal.Decimal   `json:"-"`

	CatchAll    string          `json:"catchAll" example:"1000000"`
	DecCatchAll decimal.Decimal `json:"-"`
}

type BalanceCategory struct {
	CategoryCode string          `json:"categoryCode" example:"1001"`
	CategoryName string          `json:"categoryName" example:"Cash"`
	Amount       string          `json:"amount" example:"1000000"`
	DecAmount    decimal.Decimal `json:"-"`
}

type BalancePerCategoryOut struct {
	COAType           string
	CategoryCode      string
	CategoryName      string
	SumClosingBalance decimal.Decimal
}

type BalanceSheetOut struct {
	BalanceSheet    map[string][]BalanceCategory
	TotalPerCOAType map[string]decimal.Decimal
}
