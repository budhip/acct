package models

type DoCreateAccountLoanPartnerRequest struct {
	PartnerName string    `json:"partnerName" validate:"required,minnospace=4,maxnospace=55" example:"BroilerX"`
	PartnerId   string    `json:"partnerId" validate:"required,maxnospace=15" example:"1234567890"`
	LoanKind    string    `json:"loanKind" validate:"required" example:"PartnershipLoan"`
	Metadata    *Metadata `json:"metadata"`
}

type DoCreateAccountLoanPartnerResponse struct {
	Kind           string                      `json:"kind"`
	PartnerName    string                      `json:"partnerName"`
	PartnerId      string                      `json:"partnerId"`
	AccountNumbers []LoanPartnerAccountNumbers `json:"accountNumbers"`
	Metadata       *Metadata                   `json:"metadata"`
}

type AccountsLoanPartner struct {
	PartnerName    string                      `json:"partnerName"`
	PartnerId      string                      `json:"partnerId"`
	AccountNumbers []LoanPartnerAccountNumbers `json:"accountNumbers"`
	Metadata       *Metadata                   `json:"metadata"`
}

type LoanPartnerAccountNumbers struct {
	Entity                         string `json:"entity"`
	CashInTransitDisburseDeduction string `json:"cashInTransitDisburseDeduction"`
	CashInTransitRepayment         string `json:"cashInTransitRepayment"`
	AmarthaRevenue                 string `json:"amarthaRevenue"`
	AdminFee                       string `json:"adminFee"`
	WHT2326                        string `json:"wht23_26"`
	VATOut                         string `json:"vatOut"`
}

type CreateAccountLoanPartner struct {
	PartnerName string    `json:"partnerName"`
	PartnerId   string    `json:"partnerId"`
	LoanKind    string    `json:"loanKind"`
	Metadata    *Metadata `json:"metadata"`
}

func (a *AccountsLoanPartner) ToResponse() DoCreateAccountLoanPartnerResponse {
	return DoCreateAccountLoanPartnerResponse{
		Kind:           KindLoanAccountPartner,
		PartnerName:    a.PartnerName,
		PartnerId:      a.PartnerId,
		AccountNumbers: a.AccountNumbers,
		Metadata:       a.Metadata,
	}
}
