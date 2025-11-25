package models

import "time"

type CreateSplit struct {
	TransactionID string
	SplitID       string
	SplitDate     time.Time
	Description   string
	Currency      string
	Amount        int64 // BIGINT
}
