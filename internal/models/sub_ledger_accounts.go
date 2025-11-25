package models

import (
	"encoding/base64"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
)

const KindSubLedgerAccounts = "subLedgerAccount"

type (
	DoGetSubLedgerAccountsRequest struct {
		EntityCode string `json:"entityCode" query:"entityCode" validate:"required" example:"001"`
		Search     string `json:"search" query:"search" validate:"required_with=StartDate EndDate,omitempty" example:"211000000412"`
		SearchBy   string `json:"searchBy" query:"searchBy" validate:"omitempty,oneof=accountNumber altId" example:"accountNumber"`
		StartDate  string `json:"startDate" query:"startDate" validate:"required_with=EndDate Search,omitempty" example:"2024-01-01"`
		EndDate    string `json:"endDate" query:"endDate" validate:"required_with=StartDate Search,omitempty" example:"2024-01-31"`
		Limit      int    `query:"limit" example:"10"`
		NextCursor string `query:"nextCursor" example:"abc"`
		PrevCursor string `query:"prevCursor" example:"cba"`
	}
	DoGetSubLedgerAccountsResponse struct {
		Kind            string `json:"kind" example:"subLedgerAccount"`
		AccountNumber   string `json:"accountNumber" example:"211001000000001"`
		AccountName     string `json:"accountName" example:"Tono"`
		AltId           string `json:"altId" example:"535235235235325"`
		SubCategoryCode string `json:"subCategoryCode" example:"21101"`
		SubCategoryName string `json:"subCategoryName" example:"Lender Balance - Individual Non RDL"`
		TotalRowData    int    `json:"totalRowData" example:"10"`
		CreatedAt       string `json:"createdAt" example:"2006-01-02 15:04:05"`
	}
)

type SubLedgerAccountsFilterOptions struct {
	EntityCode      string
	Search          string
	SearchBy        string
	StartDate       time.Time
	EndDate         time.Time
	Limit           int
	NextCursor      string
	PrevCursor      string
	AscendingOrder  bool
	AfterCreatedAt  *time.Time
	BeforeCreatedAt *time.Time
	GuestMode       bool //guest mode
}

func (req DoGetSubLedgerAccountsRequest) ToFilterOpts() (*SubLedgerAccountsFilterOptions, error) {
	opts := &SubLedgerAccountsFilterOptions{
		EntityCode: req.EntityCode,
		Search:     req.Search,
		Limit:      req.Limit,
	}

	if req.Limit < 0 {
		return nil, GetErrMap(ErrKeyLimitMustBeGreaterThanZero)
	}
	if req.Limit == 0 {
		// default limit
		opts.Limit = 10
	}

	if req.StartDate != "" && req.EndDate != "" {
		start, end, err := toFilterDateOpts(req.StartDate, req.EndDate)
		if err != nil {
			return nil, err
		}
		opts.StartDate = start
		opts.EndDate = end
	}

	if req.SearchBy == "" {
		// default search by
		req.SearchBy = "accountNumber"
	}
	// search by
	searchColumnMap := map[string]string{
		"accountNumber": "account_number",
		"altId":         "alt_id",
	}
	opts.SearchBy = searchColumnMap[req.SearchBy]

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

type GetSubLedgerAccountsOut struct {
	AccountNumber   string
	AccountName     string
	AltId           string
	SubCategoryCode string
	SubCategoryName string
	TotalRowData    int
	CreatedAt       time.Time
}

func (l GetSubLedgerAccountsOut) ToModelResponse() DoGetSubLedgerAccountsResponse {
	return DoGetSubLedgerAccountsResponse{
		Kind:            KindSubLedgerAccounts,
		AccountNumber:   l.AccountNumber,
		AccountName:     l.AccountName,
		AltId:           l.AltId,
		SubCategoryCode: l.SubCategoryCode,
		SubCategoryName: l.SubCategoryName,
		TotalRowData:    l.TotalRowData,
		CreatedAt:       l.CreatedAt.In(atime.GetLocation()).Format(atime.DateFormatYYYYMMDDWithTime),
	}
}

func (l GetSubLedgerAccountsOut) GetCursor() string {
	offsetBytes := []byte(l.CreatedAt.In(atime.GetLocation()).Format(time.RFC3339Nano))
	return base64.StdEncoding.EncodeToString(offsetBytes)
}
