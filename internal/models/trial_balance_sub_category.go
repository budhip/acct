package models

import (
	"github.com/shopspring/decimal"

	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/money"
)

type DoGetTrialBalanceBySubCategoryRequest struct {
	SubCategoryCode string `json:"-" param:"subCategoryCode" validate:"required" example:"21108"`
	EntityCode      string `json:"entityCode" query:"entityCode" example:"001"`
	StartDate       string `json:"startDate" query:"startDate" example:"2023-01-01"`
	EndDate         string `json:"endDate" query:"endDate" example:"2023-01-07"`
}

func (req DoGetTrialBalanceBySubCategoryRequest) ToFilterOpts() (*TrialBalanceFilterOptions, error) {
	opts := &TrialBalanceFilterOptions{
		EntityCode:      req.EntityCode,
		SubCategoryCode: req.SubCategoryCode,
	}

	if req.StartDate == "" && req.EndDate == "" {
		start, end := atime.PrevMonth(atime.Now())
		opts.StartDate = start
		opts.EndDate = end

	} else if req.StartDate == "" && req.EndDate != "" || req.StartDate != "" && req.EndDate == "" {
		return nil, GetErrMap(ErrKeyStartDateAndEndDateRequiredIfOneIsFilled)
	} else {
		start, end, err := serializeRangeDate(req.StartDate, req.EndDate)
		if err != nil {
			return nil, err
		}

		opts.StartDate = start
		opts.EndDate = end
	}

	return opts, nil
}

type TrialBalanceBySubCategoryOut struct {
	SubCategoryCode string
	SubCategoryName string
	OpeningBalance  decimal.Decimal
	DebitMovement   decimal.Decimal
	CreditMovement  decimal.Decimal
	ClosingBalance  decimal.Decimal
}

type GetTrialBalanceBySubCategoryOut struct {
	Kind                    string `json:"kind" example:"trialBalanceSubCategory"`
	SubCategoryCode         string `json:"subCategoryCode" example:"21108"`
	SubCategoryName         string `json:"subCategoryName" example:"Lender Earn"`
	IDRFormatOpeningBalance string `json:"openingBalance" example:"1000000"`
	IDRFormatDebitMovement  string `json:"debitMovement" example:"500000"`
	IDRFormatCreditMovement string `json:"creditMovement" example:"5000000"`
	IDRFormatClosingBalance string `json:"closingBalance" example:"2000000"`
}

func (tbs *TrialBalanceBySubCategoryOut) ToResponse() GetTrialBalanceBySubCategoryOut {
	return GetTrialBalanceBySubCategoryOut{
		Kind:                    "trialBalanceSubCategory",
		SubCategoryCode:         tbs.SubCategoryCode,
		SubCategoryName:         tbs.SubCategoryName,
		IDRFormatOpeningBalance: money.FormatAmountToIDR(tbs.OpeningBalance),
		IDRFormatDebitMovement:  money.FormatAmountToIDR(tbs.DebitMovement),
		IDRFormatCreditMovement: money.FormatAmountToIDR(tbs.CreditMovement),
		IDRFormatClosingBalance: money.FormatAmountToIDR(tbs.ClosingBalance),
	}
}
