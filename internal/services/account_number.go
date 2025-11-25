package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"

	"bitbucket.org/Amartha/go-accounting/internal/models"
)

func generateAccountNumber(categoryCode, entityCode string, padWidth, lastSequence int64) (string, error) {
	if entityCode == "" {
		entityCode = "000"
	}
	accountPrefix := fmt.Sprintf("%s%s", categoryCode, entityCode)
	pad := leftZeroPad(lastSequence, padWidth)
	if len(pad) != int(padWidth) {
		return "", fmt.Errorf("lastSequence %v exceed padding width %v", padWidth, lastSequence)
	}
	accountNumber := fmt.Sprintf("%s%s", accountPrefix, pad)
	return accountNumber, nil
}

func leftZeroPad(input, padWidth int64) string {
	return fmt.Sprintf(fmt.Sprintf("%%0%dd", padWidth), input)
}

func pasAccountKey(accountNumber string) string {
	return fmt.Sprintf("%s_%s", "pas_account_key", accountNumber)
}

func pasLoanAccountKey(loanAccount string) string {
	return fmt.Sprintf("%s_%s", "pas_loan_account_key", loanAccount)
}

func pasAccountLegacyKey(legacyId string) string {
	return fmt.Sprintf("%s_%s", "pas_account_legacy_key", legacyId)
}

func pasEntityKey(entity string) string {
	return fmt.Sprintf("%s_%s", "pas_entity_key", entity)
}

func pasAccountsKey(p models.GetAllAccountNumbersByParamIn) string {
	hashURL := hashURL(buildAccountNumbersURL(p))
	return fmt.Sprintf("%s_%s", "pas_accounts_key", hashURL)
}

func deletePasAccountsKey(p models.GetAllAccountNumbersByParamIn) []string {
	var keys []string
	keys = append(keys,
		pasAccountsKey(models.GetAllAccountNumbersByParamIn{OwnerId: p.OwnerId}),
		pasAccountsKey(models.GetAllAccountNumbersByParamIn{AltId: p.AltId}),
		pasAccountsKey(models.GetAllAccountNumbersByParamIn{OwnerId: p.OwnerId, AccountType: p.AccountType}),
		pasAccountsKey(models.GetAllAccountNumbersByParamIn{OwnerId: p.OwnerId, SubCategoryCode: p.SubCategoryCode}),
		pasAccountsKey(models.GetAllAccountNumbersByParamIn{AltId: p.AltId, AccountType: p.AccountType}),
	)
	return keys
}

func hashURL(rawURL string) string {
	h := sha256.Sum256([]byte(rawURL))
	return hex.EncodeToString(h[:])
}

func buildAccountNumbersURL(p models.GetAllAccountNumbersByParamIn) string {
	u := url.URL{Path: "/accounts/account-numbers"}
	q := url.Values{}

	if p.AccountNumbers != "" {
		q.Set("accountNumbers", p.AccountNumbers)
	}
	if p.OwnerId != "" {
		q.Set("ownerId", p.OwnerId)
	}
	if p.AltId != "" {
		q.Set("altId", p.AltId)
	}
	if p.SubCategoryCode != "" {
		q.Set("subCategoryCode", p.SubCategoryCode)
	}
	if p.AccountType != "" {
		q.Set("accountType", p.AccountType)
	}

	u.RawQuery = q.Encode()
	return u.String()
}
