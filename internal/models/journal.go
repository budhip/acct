package models

import (
	"time"

	"github.com/shopspring/decimal"
)

const KindJournal = "journal"

type (
	JournalRequest struct {
		ReferenceNumber string        `json:"referenceNumber" validate:"required" example:"123456"`
		TransactionId   string        `json:"transactionId" validate:"required" example:"6de11650-dbee-4f67-9ade-ececc7a02571"`
		OrderType       string        `json:"orderType" validate:"required" example:"DSB"`
		TransactionDate string        `json:"transactionDate" validate:"required,datetime" example:"2023-12-29 17:31:21"`
		ProcessingDate  string        `json:"processingDate" validate:"required,datetime" example:"2024-01-29 17:31:21"`
		Currency        string        `json:"currency" validate:"required" example:"IDR"`
		Transactions    []Transaction `json:"transactions" validate:"required,dive,required"`
		Metadata        *Metadata     `json:"metadata" swaggertype:"object,string" example:"t24AccountNumber:1234567890,t24ArrangementId:1234567890"`
	}
	Transaction struct {
		TransactionType     string          `json:"transactionType" validate:"required" example:"DSBAB"`
		TransactionTypeName string          `json:"transactionTypeName" validate:"omitempty,required" example:"DSBAB"`
		Account             string          `json:"account" validate:"required" example:"121001000000003"`
		Narrative           string          `json:"narrative" validate:"omitempty,required" example:"Credit"`
		Amount              decimal.Decimal `json:"amount" validate:"required" example:"5000001"`
		IsDebit             bool            `json:"isDebit" validate:"omitempty,required" example:"false"`
	}

	JournalError struct {
		JournalRequest
		ErrCauser interface{} `json:"errorCauser"`
	}

	JournalResponse struct {
		Kind string `json:"kind" example:"journal"`
		JournalRequest
	}

	JournalEntryCreatedRequest struct {
		TransactionID   string    `json:"transaction_id"`
		ReferenceNumber string    `json:"reference_number"`
		JournalID       string    `json:"journal_id"`
		AccountNumber   string    `json:"account_number"`
		AccountName     string    `json:"account_name"`
		Amount          int64     `json:"amount"`
		IsDebit         bool      `json:"is_debit"`
		OrderType       string    `json:"order_type"`       //
		TransactionDate string    `json:"transaction_date"` // "2024-01-15" (YYYY-MM-DD)
		TransactionType string    `json:"transaction_type"`
		EntityCode      string    `json:"entity_code"`
		CategoryCode    string    `json:"category_code"`
		SubCategoryCode string    `json:"sub_category_code"`
		NormalBalance   string    `json:"normal_balance"` // ASSET, LIABILITY, EQUITY
		CreatedAt       time.Time `json:"created_at"`     // "2024-01-15T10:30:00Z" (RFC3339)
	}
)

func (j *JournalRequest) ToResponse() JournalResponse {
	return JournalResponse{
		Kind:           KindJournal,
		JournalRequest: *j,
	}
}
