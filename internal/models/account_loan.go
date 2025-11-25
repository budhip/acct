package models

import "time"

type CreateLoanAccount struct {
	LoanAccountNumber               string `json:"loanAccountNumber"`
	LoanAdvancePaymentAccountNumber string `json:"loanAdvancePaymentAccountNumber"`
}

type AccountLoan struct {
	LoanAccountNumber               string `json:"loanAccountNumber"`
	LoanAdvancePaymentAccountNumber string `json:"loanAdvancePaymentAccountNumber"`
	CreatedAt                       *time.Time
	UpdatedAt                       *time.Time
}

type DoGetLoanAccountResponse struct {
	Kind                            string `json:"kind" example:"account"`
	LoanAccountNumber               string `json:"loanAccountNumber" example:"211001000381110"`
	LoanAdvancePaymentAccountNumber string `json:"loanAdvancePaymentAccountNumber" example:"211001000381110"`
}

func (a *AccountLoan) ToResponse() DoGetLoanAccountResponse {
	return DoGetLoanAccountResponse{
		Kind:                            KindAccount,
		LoanAccountNumber:               a.LoanAccountNumber,
		LoanAdvancePaymentAccountNumber: a.LoanAdvancePaymentAccountNumber,
	}
}
