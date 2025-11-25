package models

import (
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
)

var (
	KindLoanAccountPartner = "loanPartnerAccount"
)

type (
	LoanPartnerAccount struct {
		PartnerId           string
		LoanKind            string
		AccountNumber       string
		AccountType         string
		EntityCode          string
		LoanSubCategoryCode string
		CreatedAt           time.Time
		UpdatedAt           time.Time
	}
	DoCreateLoanPartnerAccountRequest struct {
		PartnerId           string `json:"partnerId" validate:"required" example:"efishery"`
		LoanKind            string `json:"loanKind" validate:"required" example:"EFISHERY_LOAN"`
		AccountNumber       string `json:"accountNumber" validate:"required" example:"22100100000001"`
		AccountType         string `json:"accountType" validate:"required,oneof=CASH_IN_TRANSIT_DISBURSE CASH_IN_TRANSIT_REPAYMENT INTERNAL_ACCOUNTS_REVENUE_AMARTHA INTERNAL_ACCOUNTS_ADMIN_FEE_AMARTHA INTERNAL_ACCOUNTS_PPH_AMARTHA INTERNAL_ACCOUNTS_PPN_AMARTHA" example:"INTERNAL_ACCOUNTS_REVENUE_AMARTHA"`
		LoanSubCategoryCode string `json:"loanSubCategoryCode" validate:"required,numeric,min=5,max=5" example:"13101"`
	}
	DoCreateLoanPartnerAccountResponse struct {
		Kind                string `json:"kind" example:"loanPartnerAccount"`
		PartnerId           string `json:"partnerId" example:"efishery"`
		LoanKind            string `json:"loanKind" example:"EFISHERY_LOAN"`
		AccountNumber       string `json:"accountNumber" example:"22100100000001"`
		AccountType         string `json:"accountType" example:"INTERNAL_ACCOUNTS_REVENUE_AMARTHA"`
		LoanSubCategoryCode string `json:"loanSubCategoryCode" example:"13101"`
	}
)

func (a *LoanPartnerAccount) ToCreateResponse() DoCreateLoanPartnerAccountResponse {
	return DoCreateLoanPartnerAccountResponse{
		Kind:                KindLoanAccountPartner,
		PartnerId:           a.PartnerId,
		LoanKind:            a.LoanKind,
		AccountNumber:       a.AccountNumber,
		AccountType:         a.AccountType,
		LoanSubCategoryCode: a.LoanSubCategoryCode,
	}
}

type (
	UpdateLoanPartnerAccount struct {
		PartnerId           string
		LoanKind            string
		AccountNumber       string
		AccountType         string
		LoanSubCategoryCode string
	}
	DoUpdateLoanPartnerAccountRequest struct {
		PartnerId           string `json:"partnerId" validate:"required_without_all=LoanKind AccountType LoanSubCategoryCode,omitempty" example:"efishery"`
		LoanKind            string `json:"loanKind" validate:"required_without_all=PartnerId AccountType LoanSubCategoryCode,omitempty" example:"EFISHERY_LOAN"`
		AccountNumber       string `param:"accountNumber" json:"accountNumber" validate:"required" example:"22100100000001"`
		AccountType         string `json:"accountType" validate:"required_without_all=LoanKind PartnerId LoanSubCategoryCode,omitempty,oneof=CASH_IN_TRANSIT_DISBURSE CASH_IN_TRANSIT_REPAYMENT INTERNAL_ACCOUNTS_REVENUE_AMARTHA INTERNAL_ACCOUNTS_ADMIN_FEE_AMARTHA INTERNAL_ACCOUNTS_PPH_AMARTHA INTERNAL_ACCOUNTS_PPN_AMARTHA" example:"INTERNAL_ACCOUNTS_REVENUE_AMARTHA"`
		LoanSubCategoryCode string `json:"loanSubCategoryCode" validate:"required_without_all=LoanKind PartnerId AccountType,omitempty" example:"13101"`
	}
	DoUpdateLoanPartnerAccountResponse struct {
		Kind                string `json:"kind" example:"loanPartnerAccount"`
		PartnerId           string `json:"partnerId,omitempty" example:"efishery"`
		LoanKind            string `json:"loanKind,omitempty" example:"EFISHERY_LOAN"`
		AccountNumber       string `json:"accountNumber" example:"22100100000001"`
		AccountType         string `json:"accountType,omitempty" example:"INTERNAL_ACCOUNTS_REVENUE_AMARTHA"`
		LoanSubCategoryCode string `json:"loanSubCategoryCode,omitempty" example:"13101"`
	}
)

func (a *UpdateLoanPartnerAccount) ToUpdateResponse() DoUpdateLoanPartnerAccountResponse {
	return DoUpdateLoanPartnerAccountResponse{
		Kind:                KindLoanAccountPartner,
		PartnerId:           a.PartnerId,
		LoanKind:            a.LoanKind,
		AccountNumber:       a.AccountNumber,
		AccountType:         a.AccountType,
		LoanSubCategoryCode: a.LoanSubCategoryCode,
	}
}

type (
	GetLoanPartnerAccountByParamsIn struct {
		PartnerId           string
		LoanKind            string
		AccountNumber       string
		AccountType         string
		EntityCode          string
		LoanSubCategoryCode string
		LoanAccountNumber   string
	}
	DoGetLoanPartnerAccountByParamsRequest struct {
		PartnerId           string `query:"partnerId" json:"partnerId" example:"efishery"`
		LoanKind            string `query:"loanKind" json:"loanKind" example:"EFISHERY_LOAN"`
		AccountNumber       string `query:"accountNumber" json:"accountNumber" example:"22100100000001"`
		AccountType         string `query:"accountType" json:"accountType" example:"INTERNAL_ACCOUNTS_REVENUE_AMARTHA"`
		EntityCode          string `query:"entityCode" json:"entityCode" validate:"omitempty,min=3,max=5,numeric" example:"001"`
		LoanSubCategoryCode string `query:"loanSubCategoryCode" json:"loanSubCategoryCode" example:"13101"`
		LoanAccountNumber   string `query:"loanAccountNumber" json:"loanAccountNumber" example:"22100100000001"`
	}
	DoGetLoanPartnerAccountByParamsResponse struct {
		Kind                string `json:"kind" example:"loanPartnerAccount"`
		PartnerId           string `json:"partnerId" example:"efishery"`
		LoanKind            string `json:"loanKind" example:"EFISHERY_LOAN"`
		AccountNumber       string `json:"accountNumber" example:"22100100000001"`
		AccountType         string `json:"accountType" example:"INTERNAL_ACCOUNTS_REVENUE_AMARTHA"`
		EntityCode          string `json:"entityCode" example:"001"`
		LoanSubCategoryCode string `json:"loanSubCategoryCode" example:"13101"`
		CreatedAt           string `json:"createdAt" example:"2006-01-02 15:04:05"`
		UpdatedAt           string `json:"updatedAt" example:"2006-01-02 15:04:05"`
	}
)

func (a *LoanPartnerAccount) ToGetResponse() DoGetLoanPartnerAccountByParamsResponse {
	return DoGetLoanPartnerAccountByParamsResponse{
		Kind:                KindLoanAccountPartner,
		PartnerId:           a.PartnerId,
		LoanKind:            a.LoanKind,
		AccountNumber:       a.AccountNumber,
		AccountType:         a.AccountType,
		EntityCode:          a.EntityCode,
		LoanSubCategoryCode: a.LoanSubCategoryCode,
		CreatedAt:           a.CreatedAt.In(atime.GetLocation()).Format(atime.DateFormatYYYYMMDDWithTime),
		UpdatedAt:           a.UpdatedAt.In(atime.GetLocation()).Format(atime.DateFormatYYYYMMDDWithTime),
	}
}
