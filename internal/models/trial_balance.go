package models

import (
	"fmt"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"github.com/shopspring/decimal"
)

const (
	TrialBalanceTypeAsset     = "asset"
	TrialBalanceTypeLiability = "liability"
)

const (
	KindTrialBalance = "trialBalance"
)

type DoGetTrialBalanceRequest struct {
	EntityCode string `json:"entityCode" query:"entityCode" validate:"required" example:"AMF"`
	StartDate  string `json:"startDate" query:"startDate" example:"2023-01-01"`
	EndDate    string `json:"endDate" query:"endDate" example:"2023-01-07"`
	Email      string `json:"email" query:"email" validate:"omitempty,email" example:"tono@amartha.com"`
	Period     string `json:"period" query:"period" validate:"omitempty" example:"2023-01-07"`
}

type TrialBalanceFilterOptions struct {
	EntityCode      string
	SubCategoryCode string
	StartDate       time.Time
	EndDate         time.Time
	Email           string
	Period          time.Time
}

func serializeRangeDate(inputStartDate, inputEndDate string) (startDate time.Time, endDate time.Time, err error) {
	startDate, err = atime.ParseStringToDatetime(atime.DateFormatYYYYMMDD, inputStartDate)
	if err != nil {
		err = GetErrMap(ErrKeyInvalidFormatDate, fmt.Sprintf("date %s format must be YYYY-MM-DD", inputStartDate))
		return
	}

	endDate, err = atime.ParseStringToDatetime(atime.DateFormatYYYYMMDD, inputEndDate)
	if err != nil {
		err = GetErrMap(ErrKeyInvalidFormatDate, fmt.Sprintf("date %s format must be YYYY-MM-DD", inputEndDate))
		return
	}

	if atime.DateEqualToday(startDate) || atime.DateEqualToday(endDate) {
		err = GetErrMap(ErrKeyStartDateAndEndDateEqualToday)
		return
	}

	if startDate.After(endDate) {
		err = GetErrMap(ErrKeyStartDateIsAfterEndDate)
		return
	}

	if atime.GetTotalDiffDayBetweenTwoDate(startDate, endDate) > 30 {
		err = GetErrMap(ErrKeyDateRangeMax31Days)
		return
	}

	if endDate.After(atime.Now()) {
		err = GetErrMap(ErrKeyEndDateIsAfterToday)
		return
	}

	startDate, endDate = atime.StartDateEndDate(startDate, endDate)
	return
}

func (req DoGetTrialBalanceRequest) ToFilterOpts() (*TrialBalanceFilterOptions, error) {
	opts := &TrialBalanceFilterOptions{
		EntityCode: req.EntityCode,
		Email:      req.Email,
	}

	if req.Period != "" && (req.StartDate != "" || req.EndDate != "") {
		return nil, fmt.Errorf("cannot use both 'period' and 'startDate/endDate' in the same request")
	}

	if req.Period != "" {
		periodDate, err := time.Parse("2006-01", req.Period)
		if err != nil {
			return nil, fmt.Errorf("invalid period format, must be YYYY-MM: %w", err)
		}

		opts.Period = periodDate
		return opts, nil
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

type GetTrialBalanceResponses struct {
	Kind        string `json:"kind" example:"trialBalance"`
	EntityCode  string `json:"entityCode" example:"AMF"`
	EntityName  string `json:"entityName" example:"PT AMARTHA MIKRO FINTEK"`
	ClosingDate string `json:"closingDate" example:"2023-10-31"`
	Status      string `json:"status,omitempty"`
	CatchAll    struct {
		IDRFormatCatchAllOpeningBalance string `json:"catchAllOpeningBalance" example:"1000000"`
		IDRFormatCatchAllDebitMovement  string `json:"catchAllDebitMovement" example:"500000"`
		IDRFormatCatchAllCreditMovement string `json:"catchAllCreditMovement" example:"5000000"`
		IDRFormatCatchAllClosingBalance string `json:"catchAllClosingBalance" example:"2000000"`

		CatchAllOpeningBalance decimal.Decimal `json:"-"`
		CatchAllDebitMovement  decimal.Decimal `json:"-"`
		CatchAllCreditMovement decimal.Decimal `json:"-"`
		CatchAllClosingBalance decimal.Decimal `json:"-"`
	} `json:"catchAll,omitempty"`

	COATypes []TrialBalance `json:"coaTypes"`
}

type (
	TrialBalance struct {
		Kind        string `json:"kind" example:"asset"`
		CoaTypeCode string `json:"coaTypeId" example:"asset"`
		CoaTypeName string `json:"coaTypeName" example:"asset"`

		IDRFormatTotalOpeningBalance string `json:"totalOpeningBalance" example:"1000000"`
		IDRFormatTotalDebitMovement  string `json:"totalDebitMovement" example:"500000"`
		IDRFormatTotalCreditMovement string `json:"totalCreditMovement" example:"5000000"`
		IDRFormatTotalClosingBalance string `json:"totalClosingBalance" example:"2000000"`

		TotalOpeningBalance decimal.Decimal `json:"-"`
		TotalDebitMovement  decimal.Decimal `json:"-"`
		TotalCreditMovement decimal.Decimal `json:"-"`
		TotalClosingBalance decimal.Decimal `json:"-"`

		Categories []TBCategorySubCategory `json:"categories,omitempty"`
	}

	TBCategorySubCategory struct {
		Kind         string `json:"kind" example:"category"`
		CategoryCode string `json:"categoryId" example:"A.111"`
		CategoryName string `json:"categoryName" example:"Kas Teller"`

		IDRFormatTotalOpeningBalance string `json:"totalOpeningBalance" example:"1000000"`
		IDRFormatTotalDebitMovement  string `json:"totalDebitMovement" example:"500000"`
		IDRFormatTotalCreditMovement string `json:"totalCreditMovement" example:"5000000"`
		IDRFormatTotalClosingBalance string `json:"totalClosingBalance" example:"2000000"`

		TotalOpeningBalance decimal.Decimal `json:"-"`
		TotalDebitMovement  decimal.Decimal `json:"-"`
		TotalCreditMovement decimal.Decimal `json:"-"`
		TotalClosingBalance decimal.Decimal `json:"-"`

		SubCategories []TBSubCategory `json:"subCategories,omitempty"`
	}

	TBSubCategory struct {
		Kind            string `json:"kind" example:"subCategory"`
		SubCategoryCode string `json:"subCategoryId" example:"A.111.01"`
		SubCategoryName string `json:"subCategoryName" example:"Kas Teller Point"`

		IDRFormatOpeningBalance string `json:"openingBalance" example:"100000"`
		IDRFormatDebitMovement  string `json:"debitMovement" example:"20000"`
		IDRFormatCreditMovement string `json:"creditMovement" example:"20000"`
		IDRFormatClosingBalance string `json:"closingBalance" example:"140000"`

		OpeningBalance decimal.Decimal `json:"-"`
		DebitMovement  decimal.Decimal `json:"-"`
		CreditMovement decimal.Decimal `json:"-"`
		ClosingBalance decimal.Decimal `json:"-"`
	}
)

type GetTrialBalanceOut struct {
	CoaTypeCode     string
	CoaTypeName     string
	CategoryCode    string
	CategoryName    string
	SubCategoryCode string
	SubCategoryName string
	OpeningBalance  decimal.Decimal
	DebitMovement   decimal.Decimal
	CreditMovement  decimal.Decimal
	ClosingBalance  decimal.Decimal
}

type TBCOACategory struct {
	Type                string
	CoaTypeCode         string
	CoaTypeName         string
	CategoryCode        string
	CategoryName        string
	TotalOpeningBalance decimal.Decimal
	TotalDebitMovement  decimal.Decimal
	TotalCreditMovement decimal.Decimal
	TotalClosingBalance decimal.Decimal
}

type DownloadTrialBalanceRequest struct {
	Opts   TrialBalanceFilterOptions
	Result GetTrialBalanceResponses
}

const (
	TrialBalanceStatusOpen   = "OPEN"
	TrialBalanceStatusClosed = "CLOSED"
)

type CloseTrialBalanceRequest struct {
	Period     string `param:"period" json:"period" validate:"required" example:"2025-01"`
	EntityCode string `json:"entityCode" validate:"required,min=3,max=5,numeric" example:"001"`
	ClosedBy   string `json:"closedBy" validate:"required" example:"tono@amartha.com"`
}

type CloseTrialBalanceResponse struct {
	Kind             string    `json:"kind"`
	Period           string    `json:"period"`
	Status           string    `json:"status"`
	ClosedAt         time.Time `json:"closedAt"`
	ClosedBy         string    `json:"closedBy"`
	TrialBalanceFile string    `json:"trialBalanceFile"`
}

func (c *TrialBalancePeriod) ToCloseTrialBalanceResponse() CloseTrialBalanceResponse {
	return CloseTrialBalanceResponse{
		Kind:             "trialBalance",
		Period:           c.Period,
		Status:           c.Status,
		ClosedAt:         c.CreatedAt,
		ClosedBy:         c.ClosedBy,
		TrialBalanceFile: c.TBFilePath,
	}
}

type TrialBalancePeriod struct {
	ID           int
	Period       string
	EntityCode   string
	TBFilePath   string
	Status       string
	ClosedBy     string
	IsAdjustment bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateTrialBalancePeriod struct {
	Period       string
	EntityCode   string
	TBFilePath   string
	Status       string
	IsAdjustment bool
}
