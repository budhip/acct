package models

import "time"

type AccountLender struct {
	CIHAccountNumber         string `json:"cihAccountNumber"`
	InvestedAccountNumber    string `json:"investedAccountNumber"`
	ReceivablesAccountNumber string `json:"receivablesAccountNumber"`
	CreatedAt                *time.Time
	UpdatedAt                *time.Time
}

type DoGetInvestedAccountResponse struct {
	Kind                     string `json:"kind" example:"account"`
	CIHAccountNumber         string `json:"cihAccountNumber" example:"211001000381110"`
	InvestedAccountNumber    string `json:"investedAccountNumber" example:"21100100000001"`
	ReceivablesAccountNumber string `json:"receivablesAccountNumber" example:"142001000000001"`
}

func (a *AccountLender) ToResponse() DoGetInvestedAccountResponse {
	return DoGetInvestedAccountResponse{
		Kind:                     KindAccount,
		CIHAccountNumber:         a.CIHAccountNumber,
		InvestedAccountNumber:    a.InvestedAccountNumber,
		ReceivablesAccountNumber: a.ReceivablesAccountNumber,
	}
}
