package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
)

var KindAccount = "account"

const (
	TypeAccountCreated = "account_created"
	TypeAccountUpdated = "account_updated"
)

var ExcludedSubCategoryForGuestMode = []string{
	"11101", "11201", "11301", "11401", "11402", "11403", "11404",
	"11405", "11406", "12101", "13101", "13102", "13106", "14101",
	"14201", "14301", "14401", "14501", "21101", "21103", "21105",
	"21302", "22101", "22102", "22104", "23101", "24101", "24102",
	"25101", "26101", "27101", "28101",
}

type CreateAccount struct {
	AccountNumber   string           `json:"accountNumber"`
	OwnerID         string           `json:"ownerId"`
	AccountType     string           `json:"accountType"`
	ProductTypeCode string           `json:"productTypeCode"`
	EntityCode      string           `json:"entityCode"`
	CategoryCode    string           `json:"categoryCode"`
	SubCategoryCode string           `json:"subCategoryCode"`
	Currency        string           `json:"currency"`
	Status          string           `json:"status"`
	Name            string           `json:"name"`
	AltId           string           `json:"altId"`
	LegacyId        *AccountLegacyId `json:"legacyId"`
	Metadata        *Metadata        `json:"metadata"`

	ProductTypeName string `json:"-"`
}

type CreateLenderAccount struct {
	CIHAccountNumber         string `json:"cih_account_number"`
	InvestedAccountNumber    string `json:"invested_account_number"`
	ReceivablesAccountNumber string `json:"receivables_account_number"`
}

func (a *CreateAccount) ToCreateAccountResponse() *DoCreateAccountResponse {
	return &DoCreateAccountResponse{
		Kind:            KindAccount,
		AccountNumber:   a.AccountNumber,
		OwnerID:         a.OwnerID,
		ProductTypeCode: a.ProductTypeCode,
		EntityCode:      a.EntityCode,
		CategoryCode:    a.CategoryCode,
		SubCategoryCode: a.SubCategoryCode,
		Currency:        a.Currency,
		Status:          a.Status,
		Name:            a.Name,
		AltId:           a.AltId,
		LegacyId:        a.LegacyId,
		Metadata:        a.Metadata,
	}
}

type (
	DoCreateAccountRequest struct {
		Name            string           `json:"name" validate:"omitempty,max=100" example:"Lender Yang Baik"`
		OwnerID         string           `json:"ownerId" validate:"required,alphanum,min=1,max=15" example:"12345"`
		AccountType     string           `json:"accountType" example:"LENDER_INSTITUTIONAL"`
		ProductTypeCode string           `json:"productTypeCode" validate:"omitempty,numeric,min=3,max=5" example:"1001"`
		EntityCode      string           `json:"entityCode" validate:"omitempty,min=3,max=5,numeric" example:"001"`
		CategoryCode    string           `json:"categoryCode" validate:"omitempty,numeric,min=3,max=3" example:"211"`
		SubCategoryCode string           `json:"subCategoryCode" validate:"omitempty,numeric,min=5,max=5" example:"10000"`
		Currency        string           `json:"currency" validate:"omitempty,alpha,min=3,max=3" example:"IDR"`
		AltId           string           `json:"altId" validate:"omitempty,alphanumDashUscore" example:"534534534555353523523423423"`
		LegacyId        *AccountLegacyId `json:"legacyId" swaggertype:"object,string" example:"t24AccountNumber:1234567890,t24ArrangementId:1234567890"`
		Metadata        *Metadata        `json:"metadata" swaggertype:"object,string" example:"t24AccountNumber:1234567890,t24ArrangementId:1234567890"`
	}
	DoCreateAccountResponse struct {
		Kind            string           `json:"kind" example:"account"`
		Name            string           `json:"name,omitempty" example:"Lender Yang Baik"`
		AccountNumber   string           `json:"accountNumber" example:"21100100000001"`
		OwnerID         string           `json:"ownerId" example:"12345"`
		ProductTypeCode string           `json:"productTypeCode,omitempty" example:"1001"`
		EntityCode      string           `json:"entityCode" example:"001"`
		CategoryCode    string           `json:"categoryCode" example:"211"`
		SubCategoryCode string           `json:"subCategoryCode" example:"10000"`
		Currency        string           `json:"currency" example:"IDR"`
		Status          string           `json:"status" example:"active"`
		AltId           string           `json:"altId,omitempty" example:"534534534555353523523423423"`
		LegacyId        *AccountLegacyId `json:"legacyId,omitempty" swaggertype:"object,string" example:"t24AccountNumber:1234567890,t24ArrangementId:1234567890"`
		Metadata        *Metadata        `json:"metadata,omitempty" swaggertype:"object,string" example:"t24AccountNumber:1234567890,t24ArrangementId:1234567890"`
	}
)

type (
	AccountLegacyId map[string]interface{}
	Metadata        map[string]interface{}
)

func (al *AccountLegacyId) Value() (driver.Value, error) {
	// Convert AccountLegacyId to a JSON string representation.
	jsonValue, err := json.Marshal(al)
	if err != nil {
		return nil, err
	}
	return jsonValue, nil
}

func (al *AccountLegacyId) Scan(value interface{}) error {
	// Ensure the input value is of []byte type.
	jsonValue, ok := value.([]byte)
	if !ok {
		return errors.New("invalid JSON data")
	}

	// Unmarshal the JSON data into the AccountLegacyId struct.
	if err := json.Unmarshal(jsonValue, al); err != nil {
		return err
	}
	return nil
}

func (al *Metadata) Value() (driver.Value, error) {
	// Convert Metadata to a JSON string representation.
	jsonValue, err := json.Marshal(al)
	if err != nil {
		return nil, err
	}
	return jsonValue, nil
}

func (al *Metadata) Scan(value interface{}) error {
	// Ensure the input value is of []byte type.
	jsonValue, ok := value.([]byte)
	if !ok {
		return errors.New("invalid JSON data")
	}

	// Unmarshal the JSON data into the Metadata struct.
	if err := json.Unmarshal(jsonValue, al); err != nil {
		return err
	}
	return nil
}

type (
	DoUpdateAccountRequest struct {
		AccountNumber string           `param:"accountNumber" json:"accountNumber" validate:"required" example:"21100100000001"`
		Name          string           `json:"name" validate:"max=100" example:"Lender Yang Baik"`
		OwnerID       string           `json:"ownerId" validate:"required,alphanum,min=1,max=15" example:"12345"`
		AltID         string           `json:"altId" validate:"omitempty,alphanumDashUscore,max=100" example:"534534534555353523523423423"`
		LegacyID      *AccountLegacyId `json:"legacyId"`
	}
	DoUpdateAccountResponse struct {
		Kind    string `json:"kind" example:"account"`
		Name    string `json:"name" example:"Lender Yang Baik"`
		OwnerID string `json:"ownerId" example:"12345"`
		AltID   string `json:"altId" example:"534534534555353523523423423"`
	}
	UpdateAccount struct {
		Name          string
		OwnerID       string
		AltID         string
		LegacyId      *AccountLegacyId
		AccountNumber string
	}

	UpdateLegacyId struct {
		LegacyId      *AccountLegacyId
		AccountNumber string
	}

	UpdateAltId struct {
		AltId         string
		AccountNumber string
	}

	UpdateBySubCategory struct {
		ProductTypeCode *string
		Currency        *string
		Code            string
	}
)

func (a *UpdateAccount) ToUpdateAccountResponse() *DoUpdateAccountResponse {
	return &DoUpdateAccountResponse{
		Kind:    KindAccount,
		Name:    a.Name,
		OwnerID: a.OwnerID,
		AltID:   a.AltID,
	}
}

type DoGetAccountRequest struct {
	AccountNumber string `params:"accountNumber" validate:"required" example:"21100100000001"`
}

type DoGetAccountResponse struct {
	Kind            string           `json:"kind" example:"account"`
	AccountNumber   string           `json:"accountNumber" example:"21100100000001"`
	AccountName     string           `json:"accountName" example:"Lender Yang Baik"`
	CoaTypeCode     string           `json:"coaTypeCode" example:"LIA"`
	CoaTypeName     string           `json:"coaTypeName" example:"Liability"`
	CategoryCode    string           `json:"categoryCode" example:"211"`
	CategoryName    string           `json:"categoryName" example:"Marketplace Payable (Lender Balance)"`
	SubCategoryCode string           `json:"subCategoryCode" example:"21101"`
	SubCategoryName string           `json:"subCategoryName" example:"Lender Balance - Individual Non RDL"`
	AltID           string           `json:"altID" example:"535235235235325"`
	EntityCode      string           `json:"entityCode" example:"001"`
	EntityName      string           `json:"entityName" example:"PT. Amartha Mikro Fintek (AMF)"`
	ProductTypeCode string           `json:"productTypeCode" example:"1001"`
	ProductTypeName string           `json:"productTypeName" example:"Group Loan"`
	OwnerID         string           `json:"ownerID" example:"211000000412"`
	Status          string           `json:"status" example:"active"`
	Currency        string           `json:"currency" example:"IDR"`
	AccountType     string           `json:"accountType" example:"LENDER_RETAIL"`
	LegacyId        *AccountLegacyId `json:"legacyId,omitempty" swaggertype:"object,string" example:"t24AccountNumber:1234567890,t24ArrangementId:1234567890"`
	Metadata        *Metadata        `json:"metadata,omitempty" swaggertype:"object,string" example:"t24AccountNumber:1234567890,t24ArrangementId:1234567890"`
	CreatedAt       string           `json:"createdAt" example:"2006-01-02 15:04:05"`
	UpdatedAt       string           `json:"updatedAt" example:"2006-01-02 15:04:05"`
}

type GetAccountOut struct {
	AccountNumber    string
	AccountName      string
	CoaTypeCode      string
	CoaTypeName      string
	CategoryCode     string
	CategoryName     string
	SubCategoryCode  string
	SubCategoryName  string
	EntityCode       string
	EntityName       string
	ProductTypeName  string
	ProductTypeCode  string
	AltID            string
	OwnerID          string
	Status           string
	Currency         string
	AccountType      string
	LegacyId         *AccountLegacyId
	Metadata         *Metadata
	CreatedAt        time.Time
	UpdatedAt        time.Time
	T24AccountNumber string
}

func (a *GetAccountOut) ToGetAccountResponse() DoGetAccountResponse {
	return DoGetAccountResponse{
		Kind:            KindAccount,
		AccountNumber:   a.AccountNumber,
		AccountName:     strings.TrimSpace(a.AccountName),
		CoaTypeCode:     a.CoaTypeCode,
		CoaTypeName:     a.CoaTypeName,
		CategoryCode:    a.CategoryCode,
		CategoryName:    a.CategoryName,
		SubCategoryCode: a.SubCategoryCode,
		SubCategoryName: a.SubCategoryName,
		AltID:           a.AltID,
		EntityCode:      a.EntityCode,
		EntityName:      a.EntityName,
		ProductTypeCode: a.ProductTypeCode,
		ProductTypeName: a.ProductTypeName,
		OwnerID:         a.OwnerID,
		Status:          a.Status,
		Currency:        a.Currency,
		AccountType:     a.AccountType,
		LegacyId:        a.LegacyId,
		Metadata:        a.Metadata,
		CreatedAt:       a.CreatedAt.In(atime.GetLocation()).Format(atime.DateFormatYYYYMMDDWithTime),
		UpdatedAt:       a.UpdatedAt.In(atime.GetLocation()).Format(atime.DateFormatYYYYMMDDWithTime),
	}
}

type (
	DoCheckAltIdRequest struct {
		AltId string `json:"altId" validate:"required" example:"535235235235325"`
	}
	DoCheckAltIdResponse struct {
		Kind    string `json:"kind" example:"account"`
		AltId   string `json:"altId" example:"535235235235325"`
		IsExist bool   `json:"isExist" example:"false"`
	}
)

func (a *DoCheckAltIdResponse) ToResponse() DoCheckAltIdResponse {
	return DoCheckAltIdResponse{
		Kind:    KindAccount,
		AltId:   a.AltId,
		IsExist: false,
	}
}

type MetadataT24 struct {
	IgateMethod string      `json:"igateMethod" example:"CreateLenderAccountNonRDL"`
	Argument    interface{} `json:"argument"`
}

type MetadataLoan struct {
	PartnerId string `json:"partnerId" example:"1234567890"`
}

type LegacyID struct {
	T24AccountNumber string `json:"t24AccountNumber"`
	T24ArrangementID string `json:"t24ArrangementId"`
}

type AccountError struct {
	CreateAccount
	ErrCauser interface{} `json:"errorCauser"`
}

type (
	DoGetAllAccountNumbersByParamRequest struct {
		OwnerId        string `query:"ownerId" json:"ownerId" validate:"required_without_all=AltId AccountNumbers AccountType SubCategoryCode,omitempty" example:"5000125433"`
		AltId          string `query:"altId" json:"altId" validate:"required_without_all=OwnerId AccountNumbers AccountType SubCategoryCode,omitempty" example:"12345"`
		AccountNumbers string `query:"accountNumbers" json:"accountNumbers" validate:"required_without_all=OwnerId AltId AccountType SubCategoryCode,omitempty" example:"131001000030128"`
		AccountType    string `query:"accountType" json:"accountType" validate:"required_without_all=OwnerId AltId AccountNumbers SubCategoryCode,omitempty"`
		Limit          int    `query:"limit" json:"limit" example:"10"`
	}

	DoGetAllAccountNumbersByParamResponse struct {
		Kind            string           `json:"kind" example:"account"`
		OwnerId         string           `json:"ownerId" example:"12345"`
		AccountNumber   string           `json:"accountNumber" example:"21100100000001"`
		AltId           string           `json:"altId" example:"534534534555353523523423423"`
		Name            string           `json:"name" example:"Lender Yang Baik"`
		AccountType     string           `json:"accountType" example:"LENDER_INSTITUTIONAL"`
		EntityCode      string           `json:"entityCode" example:"001"`
		ProductTypeCode string           `json:"productTypeCode" example:"1001"`
		CategoryCode    string           `json:"categoryCode" example:"211"`
		SubCategoryCode string           `json:"subCategoryCode" example:"10000"`
		Currency        string           `json:"currency" example:"IDR"`
		Status          string           `json:"status" example:"active"`
		LegacyId        *AccountLegacyId `json:"legacyId" swaggertype:"object,string" example:"t24AccountNumber:1234567890,t24ArrangementId:1234567890"`
		Metadata        *Metadata        `json:"metadata" swaggertype:"object,string" example:"t24AccountNumber:1234567890,t24ArrangementId:1234567890"`
	}

	Accounts struct {
		Kind            string `json:"kind" example:"account"`
		OwnerId         string `json:"ownerId" example:"5000125433"`
		AccountNumber   string `json:"accountNumber" example:"131001000030128"`
		AltId           string `json:"altId" example:"12345"`
		AccountType     string `json:"accountType" example:"LOAN_ACCOUNT_NORMAL"`
		SubCategoryCode string `json:"subCategoryCode" example:"13101"`
	}

	GetAllAccountNumbersByParamIn struct {
		OwnerId         string
		AltId           string
		SubCategoryCode string
		AccountNumbers  string
		AccountType     string
		Limit           int
	}
	GetAllAccountNumbersByParamOut struct {
		OwnerId         string
		AccountNumber   string
		AltId           string
		Name            string
		AccountType     string
		EntityCode      string
		ProductTypeCode string
		CategoryCode    string
		SubCategoryCode string
		Currency        string
		Status          string
		LegacyId        *AccountLegacyId
		Metadata        *Metadata
		CreatedAt       time.Time
	}
)

func (a *GetAllAccountNumbersByParamOut) ToResponse() *DoGetAllAccountNumbersByParamResponse {
	return &DoGetAllAccountNumbersByParamResponse{
		Kind:            KindAccount,
		OwnerId:         a.OwnerId,
		AccountNumber:   a.AccountNumber,
		AltId:           a.AltId,
		Name:            a.Name,
		AccountType:     a.AccountType,
		EntityCode:      a.EntityCode,
		ProductTypeCode: a.ProductTypeCode,
		CategoryCode:    a.CategoryCode,
		SubCategoryCode: a.SubCategoryCode,
		Currency:        a.Currency,
		Status:          a.Status,
		LegacyId:        a.LegacyId,
		Metadata:        a.Metadata,
	}
}

type DoGetAllCategoryCodeSeqResponse struct {
	Kind  string `json:"kind" example:"account"`
	Key   string `json:"key" example:"category_code_212_seq"`
	Value string `json:"value" example:"100"`
}

type (
	DoUpdateCategoryCodeSeqRequest struct {
		Key   string `json:"key" validate:"required" example:"category_code_212_seq"`
		Value int64  `json:"value" validate:"required,number" example:"100"`
	}
	DoUpdateCategoryCodeSeqResponse struct {
		Kind  string `json:"kind" example:"account"`
		Key   string `json:"key" example:"category_code_212_seq"`
		Value int64  `json:"value" example:"100"`
	}

	DoCreateCategoryCodeSeqRequest struct {
		Key   string `json:"key" validate:"required" example:"category_code_212_seq"`
		Value int64  `json:"value" validate:"required,number" example:"100"`
	}
	DoCreateCategoryCodeSeqResponse struct {
		Kind  string `json:"kind" example:"account"`
		Key   string `json:"key" example:"category_code_212_seq"`
		Value int64  `json:"value" example:"100"`
	}
)

func (a *DoUpdateCategoryCodeSeqRequest) ToResponse() DoUpdateCategoryCodeSeqResponse {
	return DoUpdateCategoryCodeSeqResponse{
		Kind:  KindAccount,
		Key:   a.Key,
		Value: a.Value,
	}
}

func (a *DoCreateCategoryCodeSeqRequest) ToResponse() DoCreateCategoryCodeSeqResponse {
	return DoCreateCategoryCodeSeqResponse{
		Kind:  KindAccount,
		Key:   a.Key,
		Value: a.Value,
	}
}

type CreateAccountMigration struct {
	Accounts      []CreateAccount     `json:"account"`
	LenderAccount CreateLenderAccount `json:"lenderAccount,omitempty"`
	LoanAccount   CreateLoanAccount   `json:"loanAccount,omitempty"`
}

type (
	DoUpdateAccountEntityRequest struct {
		AccountNumber string `param:"accountNumber" json:"accountNumber" validate:"required" example:"21100100000001"`
		EntityCode    string `json:"entityCode" validate:"required,min=3,max=5,numeric" example:"001"`
	}
	DoUpdateAccountEntityResponse struct {
		Kind       string `json:"kind" example:"account"`
		EntityCode string `json:"entityCode" xample:"001"`
	}
	UpdateAccountEntity struct {
		EntityCode    string
		AccountNumber string
	}
)

func (a *UpdateAccountEntity) ToResponse() *DoUpdateAccountEntityResponse {
	return &DoUpdateAccountEntityResponse{
		Kind:       KindAccount,
		EntityCode: a.EntityCode,
	}
}

type CheckAccountNumberIsExist struct {
	AccountNumber string
	EntityCode    string
}
