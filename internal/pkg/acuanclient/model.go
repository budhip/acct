package acuanclient

import (
	"bitbucket.org/Amartha/go-accounting/internal/models"
)

type PublishAccountData struct {
	Type            string
	AccountNumber   string
	Name            string
	ProductTypeName string
	OwnerId         string
	CategoryCode    string
	SubCategoryCode string
	EntityCode      string
	Currency        string
	AltId           string
	Status          string
	LegacyId        *models.AccountLegacyId
	Metadata        interface{}
}
