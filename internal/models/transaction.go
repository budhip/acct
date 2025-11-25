package models

import "time"

type CreateTransaction struct {
	TransactionID string
	Postdate      time.Time
	PosterUserID  string
}
