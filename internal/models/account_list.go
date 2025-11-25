package models

import (
	"encoding/base64"
	"fmt"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
)

type (
	DoGetListAccountRequest struct {
		Search              string `json:"search" query:"search" example:"211000000412"`
		SearchBy            string `json:"searchBy" query:"searchBy" validate:"omitempty,oneof=accountNumber altId ownerId t24AccountNumber" example:"accountNumber"`
		CoaTypeCode         string `json:"coaTypeCode" query:"coaTypeCode" example:"AST"`
		EntityCode          string `json:"entityCode" query:"entityCode" validate:"required" example:"001"`
		CategoryCode        string `json:"categoryCode" query:"categoryCode" example:"112"`
		SubCategoryCode     string `json:"subCategoryCode" query:"subCategoryCode" example:"11201"`
		ProductTypeCode     string `json:"productTypeCode" query:"productTypeCode" example:"1001"`
		Limit               int    `query:"limit" example:"10"`
		ExcludeTotalEntries bool   `query:"excludeTotalEntries" example:"true"`
		NextCursor          string `query:"nextCursor" example:"abc"`
		PrevCursor          string `query:"prevCursor" example:"cba"`
	}

	DoDownloadGetListAccountRequest struct {
		Search          string `json:"search" query:"search" example:"211000000412"`
		SearchBy        string `json:"searchBy" query:"searchBy" validate:"omitempty,oneof=accountNumber altId ownerId t24AccountNumber" example:"accountNumber"`
		CoaTypeCode     string `json:"coaTypeCode" query:"coaTypeCode" validate:"required" example:"AST"`
		EntityCode      string `json:"entityCode" query:"entityCode" validate:"required" example:"001"`
		ProductTypeCode string `json:"productTypeCode" query:"productTypeCode" example:"1001"`
		CategoryCode    string `json:"categoryCode" query:"categoryCode" validate:"required" example:"112"`
		SubCategoryCode string `json:"subCategoryCode" query:"subCategoryCode" validate:"required" example:"11201"`
		Limit           int    `query:"limit" example:"10"`
	}

	DoGetListAccountResponse struct {
		Kind             string `json:"kind" example:"transaction"`
		AccountNumber    string `json:"accountNumber" example:"211001000000001"`
		AccountName      string `json:"accountName" example:"Tono"`
		CoaTypeCode      string `json:"coaTypeCode" example:"LIA"`
		CoaTypeName      string `json:"coaTypeName" example:"Liability"`
		CategoryCode     string `json:"categoryCode" example:"211"`
		CategoryName     string `json:"categoryName" example:"Marketplace Payable (Lender Balance)"`
		SubCategoryCode  string `json:"subCategoryCode" example:"21101"`
		SubCategoryName  string `json:"subCategoryName" example:"Lender Balance - Individual Non RDL"`
		EntityCode       string `json:"entityCode" example:"001"`
		EntityName       string `json:"entityName" example:"PT. Amartha Mikro Fintek (AMF)"`
		ProductTypeCode  string `json:"productTypeCode" example:"1001"`
		ProductTypeName  string `json:"productTypeName" example:"Group Loan"`
		AltID            string `json:"altId" example:"535235235235325"`
		OwnerID          string `json:"ownerId" example:"211000000412"`
		Status           string `json:"status" example:"active"`
		CreatedAt        string `json:"createdAt" example:"2006-01-02 15:04:05"`
		UpdatedAt        string `json:"updatedAt" example:"2006-01-02 15:04:05"`
		T24AccountNumber string `json:"t24AccountNumber" example:"211000000412"`
	}
)

type AccountFilterOptions struct {
	Search              string
	SearchBy            string
	CoaTypeCode         string
	EntityCode          string
	CategoryCode        string
	SubCategoryCode     string
	ProductTypeCode     string
	AltID               string
	T24AccountNumber    string
	Limit               int
	ExcludeTotalEntries bool
	AscendingOrder      bool
	AfterCreatedAt      *time.Time
	BeforeCreatedAt     *time.Time
	GuestMode           bool
}

func (req DoGetListAccountRequest) ToFilterOpts() (*AccountFilterOptions, error) {
	opts := &AccountFilterOptions{
		Search:              req.Search,
		CoaTypeCode:         req.CoaTypeCode,
		EntityCode:          req.EntityCode,
		ProductTypeCode:     req.ProductTypeCode,
		CategoryCode:        req.CategoryCode,
		SubCategoryCode:     req.SubCategoryCode,
		ExcludeTotalEntries: req.ExcludeTotalEntries,
		Limit:               req.Limit,
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
		afterTime, err := decodeCursor(req.NextCursor)
		if err != nil {
			return nil, err
		}
		opts.AfterCreatedAt = &afterTime
	}

	// backward pagination
	if req.NextCursor == "" && req.PrevCursor != "" {
		prevTime, err := decodeCursor(req.PrevCursor)
		if err != nil {
			return nil, err
		}
		opts.BeforeCreatedAt = &prevTime

		// reverse order
		opts.AscendingOrder = true
	}

	if req.SearchBy == "" {
		// default search by
		req.SearchBy = "accountNumber"
	}
	// search by
	searchColumnMap := map[string]string{
		"accountNumber":    "account_number",
		"altId":            "alt_id",
		"ownerId":          "owner_id",
		"t24AccountNumber": "t24AccountNumber",
	}
	opts.SearchBy = searchColumnMap[req.SearchBy]

	return opts, nil
}

func (a GetAccountOut) ToModelResponse() DoGetListAccountResponse {
	return DoGetListAccountResponse{
		Kind:             KindAccount,
		AccountNumber:    a.AccountNumber,
		AccountName:      a.AccountName,
		CoaTypeCode:      a.CoaTypeCode,
		CoaTypeName:      a.CoaTypeName,
		CategoryCode:     a.CategoryCode,
		CategoryName:     a.CategoryName,
		SubCategoryCode:  a.SubCategoryCode,
		SubCategoryName:  a.SubCategoryName,
		EntityCode:       a.EntityCode,
		EntityName:       a.EntityName,
		ProductTypeCode:  a.ProductTypeCode,
		ProductTypeName:  a.ProductTypeName,
		AltID:            a.AltID,
		OwnerID:          a.OwnerID,
		Status:           a.Status,
		CreatedAt:        a.CreatedAt.In(atime.GetLocation()).Format(atime.DateFormatYYYYMMDDWithTime),
		UpdatedAt:        a.UpdatedAt.In(atime.GetLocation()).Format(atime.DateFormatYYYYMMDDWithTime),
		T24AccountNumber: a.T24AccountNumber,
	}
}

func (a GetAccountOut) GetCursor() string {
	offsetBytes := []byte(a.CreatedAt.In(atime.GetLocation()).Format(time.RFC3339Nano))
	return base64.StdEncoding.EncodeToString(offsetBytes)
}

func decodeCursor(cursor string) (decodedTime time.Time, err error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return decodedTime, fmt.Errorf("failed to parse offset string: %w", err)
	}

	decodedTime, err = time.Parse(time.RFC3339Nano, string(decodedBytes))
	if err != nil {
		return decodedTime, fmt.Errorf("failed to parse offset time: %w", err)
	}

	return decodedTime, nil
}
