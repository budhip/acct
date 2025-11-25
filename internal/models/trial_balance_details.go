package models

import (
	"encoding/base64"
	"fmt"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/money"

	"github.com/shopspring/decimal"
)

type DoGetTrialBalanceDetailsRequest struct {
	EntityCode      string `json:"entityCode" query:"entityCode" validate:"required" example:"001"`
	SubCategoryCode string `json:"subCategoryCode" query:"subCategoryCode" validate:"required" example:"11101"`
	StartDate       string `json:"startDate" query:"startDate" example:"2023-01-01"`
	EndDate         string `json:"endDate" query:"endDate" example:"2023-01-07"`
	Search          string `json:"search" query:"search" example:"123456"`
	Limit           int    `query:"limit" example:"10"`
	NextCursor      string `query:"nextCursor" example:"abc"`
	PrevCursor      string `query:"prevCursor" example:"cba"`
	Email           string `json:"email" query:"email" validate:"omitempty,email" example:"tono@amartha.com"`
}

// to support trial balance details v2
type DoGetTrialBalanceDetailsByPeriodRequest struct {
	EntityCode      string `json:"entityCode" query:"entityCode" validate:"required" example:"001"`
	SubCategoryCode string `json:"subCategoryCode" query:"subCategoryCode" validate:"required" example:"11101"`
	Period          string `json:"period" query:"period" validate:"required" example:"2025-10"`
	Email           string `json:"email" query:"email" validate:"omitempty,email" example:"tono@amartha.com"`
}

type TrialBalanceDetailsFilterOptions struct {
	SubCategoryCode string
	EntityCode      string
	StartDate       time.Time
	EndDate         time.Time
	Search          string
	Email           string
	SubCategoryName string
	EntityDesc      string
	Month           string
	Year            string

	// CursorValue as accountNumber(string)
	CursorValue *string
	IsBackward  bool
	Limit       int
}

func (req DoGetTrialBalanceDetailsRequest) ToFilterOpts() (*TrialBalanceDetailsFilterOptions, error) {
	opts := &TrialBalanceDetailsFilterOptions{
		EntityCode:      req.EntityCode,
		SubCategoryCode: req.SubCategoryCode,
		Search:          req.Search,
		Limit:           req.Limit,
	}

	if req.Limit < 0 {
		return nil, GetErrMap(ErrKeyLimitMustBeGreaterThanZero)
	}

	if req.Limit == 0 {
		// default limit
		opts.Limit = 10
	}

	// use over-fetch limit for check next page exists or not
	opts.Limit += 1

	// forward pagination
	if req.NextCursor != "" {
		cursorValue, err := decodeCursorTrailBalanceDetail(req.NextCursor)
		if err != nil {
			return nil, err
		}

		opts.CursorValue = &cursorValue
	}

	// backward pagination
	if req.NextCursor == "" && req.PrevCursor != "" {
		cursorValue, err := decodeCursorTrailBalanceDetail(req.PrevCursor)
		if err != nil {
			return nil, err
		}

		opts.CursorValue = &cursorValue

		// reverse order
		opts.IsBackward = true
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

func (req DoGetTrialBalanceDetailsByPeriodRequest) ToFilterOpts() (*TrialBalanceDetailsFilterOptions, error) {
	opts := &TrialBalanceDetailsFilterOptions{
		EntityCode:      req.EntityCode,
		SubCategoryCode: req.SubCategoryCode,
	}

	t, err := time.Parse(atime.DateFormatYYYYMM, req.Period)
	if err != nil {
		err = GetErrMap(ErrKeyInvalidFormatDate, fmt.Sprintf("date %s format must be YYYY-MM", req.Period))
		return opts, err
	}

	year := fmt.Sprintf("%04d", t.Year())
	month := fmt.Sprintf("%02d", int(t.Month()))

	opts.Year = year
	opts.Month = month

	return opts, nil
}

type TrialBalanceDetailOut struct {
	AccountNumber  string
	AccountName    string
	OpeningBalance decimal.Decimal
	DebitMovement  decimal.Decimal
	CreditMovement decimal.Decimal
	ClosingBalance decimal.Decimal
}

type GetTrialBalanceDetailOut struct {
	Kind                    string `json:"kind" example:"trialBalanceDetail"`
	AccountNumber           string `json:"accountNumber" example:"123456"`
	AccountName             string `json:"accountName" example:"Shinji Takeru"`
	IDRFormatOpeningBalance string `json:"openingBalance" example:"1000000"`
	IDRFormatDebitMovement  string `json:"debitMovement" example:"500000"`
	IDRFormatCreditMovement string `json:"creditMovement" example:"5000000"`
	IDRFormatClosingBalance string `json:"closingBalance" example:"2000000"`
}

func (tba TrialBalanceDetailOut) GetCursor() string {
	offsetBytes := []byte(tba.AccountNumber)
	return base64.StdEncoding.EncodeToString(offsetBytes)
}

func decodeCursorTrailBalanceDetail(cursor string) (decodedAccountNumber string, err error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return decodedAccountNumber, fmt.Errorf("failed to parse offset string: %w", err)
	}

	return string(decodedBytes), nil
}

func (tba TrialBalanceDetailOut) ToModelResponse() GetTrialBalanceDetailOut {
	return GetTrialBalanceDetailOut{
		Kind:                    "trialBalanceDetail",
		AccountNumber:           tba.AccountNumber,
		AccountName:             tba.AccountName,
		IDRFormatOpeningBalance: money.FormatAmountToIDR(tba.OpeningBalance),
		IDRFormatDebitMovement:  money.FormatAmountToIDR(tba.DebitMovement),
		IDRFormatCreditMovement: money.FormatAmountToIDR(tba.CreditMovement),
		IDRFormatClosingBalance: money.FormatAmountToIDR(tba.ClosingBalance),
	}
}

type SendEmailTrialBalanceAccountListRequest struct {
	Opts   TrialBalanceDetailsFilterOptions
	Result []TrialBalanceDetailOut
}

type DownloadTrialBalanceDetailsRequest struct {
	EntityCode      string `json:"entityCode" query:"entityCode" validate:"required" example:"001"`
	SubCategoryCode string `json:"subCategoryCode" query:"subCategoryCode" validate:"required" example:"11101"`
	StartDate       string `json:"startDate" query:"startDate" example:"2023-01-01"`
	EndDate         string `json:"endDate" query:"endDate" example:"2023-01-07"`
	Email           string `json:"email" query:"email" validate:"required,email" example:"tono@amartha.com"`
	Period          string `json:"period" query:"period"`
}

func (req DownloadTrialBalanceDetailsRequest) ToFilterOpts() (*TrialBalanceDetailsFilterOptions, error) {
	opts := &TrialBalanceDetailsFilterOptions{
		EntityCode:      req.EntityCode,
		SubCategoryCode: req.SubCategoryCode,
		Email:           req.Email,
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

type DownloadTrialBalanceDetailsByPeriodRequest struct {
	EntityCode      string `json:"entityCode" query:"entityCode" validate:"required" example:"001"`
	SubCategoryCode string `json:"subCategoryCode" query:"subCategoryCode" validate:"required" example:"11101"`
	Email           string `json:"email" query:"email" validate:"required,email" example:"tono@amartha.com"`
	Period          string `json:"period" query:"period" validate:"required"`
}

func (req DownloadTrialBalanceDetailsByPeriodRequest) ToFilterOpts() (*TrialBalanceDetailsFilterOptions, error) {
	opts := &TrialBalanceDetailsFilterOptions{
		EntityCode:      req.EntityCode,
		SubCategoryCode: req.SubCategoryCode,
		Email:           req.Email,
	}
	t, err := time.Parse(atime.DateFormatYYYYMM, req.Period)
	if err != nil {
		err = GetErrMap(ErrKeyInvalidFormatDate, fmt.Sprintf("date %s format must be YYYY-MM", req.Period))
		return nil, err
	}

	year := fmt.Sprintf("%04d", t.Year())
	month := fmt.Sprintf("%02d", int(t.Month()))

	opts.Year = year
	opts.Month = month

	return opts, nil
}
