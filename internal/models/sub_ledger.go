package models

import (
	"encoding/base64"
	"fmt"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/money"

	"github.com/shopspring/decimal"
)

const KindSubLedger = "subLedger"

type GetAccountBalancePeriodStartEndOut struct {
	BalancePeriodStart string
	BalancePeriodEnd   string
}

type GetSubLedgerRequest struct {
	AccountNumber string `json:"accountNumber" query:"accountNumber" validate:"required" example:"22201000000008"`
	StartDate     string `json:"startDate" query:"startDate" validate:"required" example:"2023-01-01"`
	EndDate       string `json:"endDate" query:"endDate" validate:"required" example:"2023-01-07"`
	Email         string `json:"email" query:"email" validate:"omitempty,email" example:"tono@amartha.com"`
	Limit         int    `query:"limit" example:"10"`
	NextCursor    string `query:"nextCursor" example:"abc"`
	PrevCursor    string `query:"prevCursor" example:"cba"`
}

type SubLedgerFilterOptions struct {
	AccountNumber, Email string
	StartDate            time.Time
	EndDate              time.Time
	Limit                int
	Offset               int
	NextCursor           string
	PrevCursor           string
	AscendingOrder       bool
	AfterCreatedAt       *time.Time
	BeforeCreatedAt      *time.Time
}

func (req GetSubLedgerRequest) ToFilterOpts() (*SubLedgerFilterOptions, error) {
	opts := &SubLedgerFilterOptions{
		AccountNumber: req.AccountNumber,
		Email:         req.Email,
		Limit:         req.Limit,
	}

	if req.Limit < 0 {
		return nil, GetErrMap(ErrKeyLimitMustBeGreaterThanZero)
	}

	if req.Limit == 0 {
		// default limit
		opts.Limit = 10
	}

	start, end, err := toFilterDateOpts(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	opts.StartDate = start
	opts.EndDate = end

	// use over-fetch limit for check next page exists or not
	opts.Limit += 1

	// forward pagination
	if req.NextCursor != "" {
		afterTime, err := decodeCursor(req.NextCursor)
		if err != nil {
			return nil, err
		}
		opts.BeforeCreatedAt = &afterTime
	}

	// backward pagination
	if req.NextCursor == "" && req.PrevCursor != "" {
		prevTime, err := decodeCursor(req.PrevCursor)
		if err != nil {
			return nil, err
		}
		opts.AfterCreatedAt = &prevTime

		// reverse order
		opts.AscendingOrder = true
	}

	return opts, nil
}

type SubLedgerAccountResponse struct {
	Kind               string `json:"kind" example:"account"`
	AccountNumber      string `json:"accountNumber" example:"AMF"`
	AccountName        string `json:"accountName" example:"PT AMARTHA MIKRO FINTEK"`
	AltId              string `json:"altId" example:"12345"`
	COATypeCode        string `json:"coaTypeCode" example:"AST"`
	EntityCode         string `json:"entityCode" example:"AMF"`
	EntityName         string `json:"entityName" example:"PT AMARTHA MIKRO FINTEK"`
	ProductTypeCode    string `json:"productTypeCode" example:"1001"`
	ProductTypeName    string `json:"productTypeName" example:"Group Loan"`
	SubCategoryCode    string `json:"subCategoryCode" example:"10000"`
	SubCategoryName    string `json:"subCategoryName" example:"RETAIL"`
	Currency           string `json:"currency" example:"PT AMARTHA MIKRO FINTEK"`
	BalancePeriodStart string `json:"balancePeriodStart" example:"500.000,00"`
}

type GetSubLedgerResponse struct {
	Kind                string    `json:"kind" example:"subLedger"`
	TransactionID       string    `json:"transactionId" example:"3c8c389b-abbc-4d17-a718-7bd84721b40f"`
	ReferenceNumber     string    `json:"referenceNumber" example:"DSB_1465621"`
	TransactionDate     string    `json:"transactionDate" example:"14/10/2023"`
	TransactionType     string    `json:"transactionType" example:"DSBAC"`
	TransactionTypeName string    `json:"transactionTypeName" example:"Admin Fee Partner Loan Deduction"`
	Narrative           string    `json:"narrative" example:"Invest To Loan"`
	Metadata            *Metadata `json:"metadata" swaggertype:"object,string" example:"billingId:203562195,loanId:7406883"`
	Debit               string    `json:"debit" example:"30.000,00"`
	Credit              string    `json:"credit" example:"0,00"`
}

type GetSubLedgerOut struct {
	TransactionID       string
	ReferenceNumber     string
	TransactionDate     time.Time
	OrderType           string
	TransactionType     string
	TransactionTypeName string
	Narrative           string
	Metadata            *Metadata
	Debit               decimal.Decimal
	Credit              decimal.Decimal
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (l GetSubLedgerOut) ToModelResponse() GetSubLedgerResponse {
	return GetSubLedgerResponse{
		Kind:                KindSubLedger,
		TransactionID:       l.TransactionID,
		ReferenceNumber:     l.ReferenceNumber,
		TransactionDate:     l.TransactionDate.In(atime.GetLocation()).Format(atime.DateFormatDDMMMYYYYTimeWithSpace),
		TransactionType:     l.TransactionType,
		TransactionTypeName: l.TransactionTypeName,
		Narrative:           l.Narrative,
		Metadata:            l.Metadata,
		Debit:               money.FormatAmountToIDR(l.Debit),
		Credit:              money.FormatAmountToIDR(l.Credit),
	}
}

func (l GetSubLedgerOut) GetCursor() string {
	offsetBytes := []byte(l.TransactionDate.In(atime.GetLocation()).Format(time.RFC3339Nano))
	return base64.StdEncoding.EncodeToString(offsetBytes)
}

type DownloadCSVSubLedgerRequest struct {
	AccountNumber string `json:"accountNumber" query:"accountNumber" validate:"required" example:"22201000000008"`
	StartDate     string `json:"startDate" query:"startDate" validate:"required" example:"2023-01-01"`
	EndDate       string `json:"endDate" query:"endDate" validate:"required" example:"2023-01-07"`
}

func (req DownloadCSVSubLedgerRequest) ToFilterOpts() (*SubLedgerFilterOptions, error) {
	opts := &SubLedgerFilterOptions{
		AccountNumber: req.AccountNumber,
	}

	start, end, err := toFilterDateOpts(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	opts.StartDate = start
	opts.EndDate = end

	return opts, nil
}

type DownloadCSVToEmailSubLedgerRequest struct {
	AccountNumber string `json:"accountNumber" query:"accountNumber" validate:"required" example:"22201000000008"`
	StartDate     string `json:"startDate" query:"startDate" validate:"required" example:"2023-01-01"`
	EndDate       string `json:"endDate" query:"endDate" validate:"required" example:"2023-01-07"`
	Email         string `json:"email" query:"email" validate:"required,email" example:"tono@amartha.com"`
}

func (req DownloadCSVToEmailSubLedgerRequest) ToFilterOpts() (*SubLedgerFilterOptions, error) {
	opts := &SubLedgerFilterOptions{
		AccountNumber: req.AccountNumber,
		Email:         req.Email,
	}

	start, end, err := toFilterDateOpts(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	opts.StartDate = start
	opts.EndDate = end

	return opts, nil
}

func toFilterDateOpts(startDate, endDate string) (start, end time.Time, err error) {
	startDateTime, err := atime.ParseStringToDatetime(atime.DateFormatYYYYMMDD, startDate)
	if err != nil {
		return start, end, GetErrMap(ErrKeyInvalidFormatDate, fmt.Sprintf("date %s format must be YYYY-MM-DD", startDate))
	}

	endDateTime, err := atime.ParseStringToDatetime(atime.DateFormatYYYYMMDD, endDate)
	if err != nil {
		return start, end, GetErrMap(ErrKeyInvalidFormatDate, fmt.Sprintf("date %s format must be YYYY-MM-DD", endDate))
	}

	if startDateTime.After(endDateTime) {
		return start, end, GetErrMap(ErrKeyStartDateIsAfterEndDate)
	}
	if atime.GetTotalDiffDayBetweenTwoDate(startDateTime, endDateTime) > 30 {
		return start, end, GetErrMap(ErrKeyDateRangeMax31Days)
	}
	if endDateTime.After(atime.Now()) {
		return start, end, GetErrMap(ErrKeyEndDateIsAfterToday)
	}

	start, end = atime.StartDateEndDate(startDateTime, endDateTime)
	return
}

type (
	GetSubLedgerCountOut struct {
		Total             int
		IsExceedsTheLimit bool
	}
	GetSubLedgerCountResponse struct {
		Kind              string `json:"kind" example:"subLedger"`
		Total             int    `json:"total" example:"10"`
		IsExceedsTheLimit bool   `json:"isExceedsTheLimit" example:"true"`
	}
)

func (l GetSubLedgerCountOut) ToModelResponse() GetSubLedgerCountResponse {
	return GetSubLedgerCountResponse{
		Kind:              KindSubLedger,
		Total:             l.Total,
		IsExceedsTheLimit: l.IsExceedsTheLimit,
	}
}
