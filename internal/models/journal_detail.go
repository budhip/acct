package models

import (
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/money"

	"github.com/shopspring/decimal"
)

type CreateJournalDetail struct {
	JournalId           string
	ReferenceNumber     string
	OrderType           string
	TransactionType     string
	TransactionTypeName string
	TransactionDate     time.Time
	IsDebit             bool
	Metadata            *Metadata
}

type (
	DoGetJournalDetailResponse struct {
		Kind            string `json:"kind" example:"journal"`
		TransactionId   string `json:"transactionId" example:"12c5692e-cfbd-4cee-a4ce-86eac1447d48"`
		JournalId       string `json:"journalId" example:"2024060601093678"`
		AccountNumber   string `json:"accountNumber" example:"214003000000194"`
		AccountName     string `json:"accountName" example:"Mails Morales"`
		AltId           string `json:"altId" example:"Mails002"`
		EntityCode      string `json:"entityCode" example:"003"`
		EntityName      string `json:"entityName" example:"AFA"`
		SubCategoryCode string `json:"subCategoryCode" example:"21401"`
		SubCategoryName string `json:"subCategoryName" example:"eWallet User"`
		TransactionType string `json:"transactionType" example:"PAYGL"`
		Amount          string `json:"amount" example:"10000"`
		TransactionDate string `json:"transactionDate" example:"2006-01-02 15:04:05"`
		Narrative       string `json:"narrative" example:"Repayment Group Loan via Poket jindankarasuno"`
		IsDebit         bool   `json:"isDebit" example:"true"`
	}
	GetJournalDetailOut struct {
		TransactionId   string
		JournalId       string
		AccountNumber   string
		AccountName     string
		AltId           string
		EntityCode      string
		EntityName      string
		SubCategoryCode string
		SubCategoryName string
		TransactionType string
		Amount          decimal.Decimal
		TransactionDate time.Time
		Narrative       string
		IsDebit         bool
	}
)

func (j *GetJournalDetailOut) ToResponse() DoGetJournalDetailResponse {
	return DoGetJournalDetailResponse{
		Kind:            KindJournal,
		TransactionId:   j.TransactionId,
		JournalId:       j.JournalId,
		AccountNumber:   j.AccountNumber,
		AccountName:     j.AccountName,
		AltId:           j.AltId,
		EntityCode:      j.EntityCode,
		EntityName:      j.EntityName,
		SubCategoryCode: j.SubCategoryCode,
		SubCategoryName: j.SubCategoryName,
		TransactionType: j.TransactionType,
		Amount:          money.FormatAmountToIDR(j.Amount),
		TransactionDate: j.TransactionDate.In(atime.GetLocation()).Format(atime.DateFormatYYYYMMDDWithTime),
		Narrative:       j.Narrative,
		IsDebit:         j.IsDebit,
	}
}
