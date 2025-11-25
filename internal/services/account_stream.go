package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-igate"
	xlog "bitbucket.org/Amartha/go-x/log"
)

const (
	accountTypeLoan   = "loan"
	accountTypeBranch = "branch"
	accountTypeLender = "lender"
)

func (as *account) ConsumerAccountStream(ctx context.Context, in models.CreateAccount) (err error) {
	defer func() {
		if err != nil {
			xlog.Warn(ctx, "[ACCOUNT-STREAM]", xlog.Any("message", in), xlog.Err(err))
			as.srv.Account.publishToAccountStreamDLQ(ctx, in, err)
		}
	}()

	switch strings.ToLower(in.AccountType) {
	case accountTypeLoan: // from rgate
		if isProceed := as.srv.conf.AccountConfig.IsT24CreateAccountPAS[accountTypeLoan]; isProceed {
			err = as.CreateLoanAccount(ctx, in)
		}
	case accountTypeBranch: // from rgate
		if isProceed := as.srv.conf.AccountConfig.IsT24CreateAccountPAS[accountTypeBranch]; isProceed {
			err = as.CreateBranchPointAccount(ctx, in)
		}
	case accountTypeLender: // from rgate
		if isProceed := as.srv.conf.AccountConfig.IsT24CreateAccountPAS[accountTypeLender]; isProceed {
			err = as.CreateLenderInstitutionAccount(ctx, in)
		}
	default: // general
		if in.OwnerID == "" {
			err = models.GetErrMap(models.ErrKeyOwnerIdRequired)
			return
		}
		_, err = as.Create(ctx, in)
	}

	xlog.Info(ctx, "[ACCOUNT-STREAM]",
		xlog.String("account-number", in.AccountNumber),
		xlog.Any("data", in),
	)

	return
}

// CreateLoanAccount will create the main loan account and payable account.
// This func initially used for migration, so maybe remove it in the future if not used.
func (as *account) CreateLoanAccount(ctx context.Context, in models.CreateAccount) (err error) {
	if in.Metadata == nil {
		err = errors.New("metadata is nil")
		return
	}

	metadata := *in.Metadata
	product, ok := metadata["product"].(string)
	if !ok {
		err = fmt.Errorf("product not available on meta: %v", metadata)
		return
	}

	// Mapping accountType
	mapProductToAccountType := map[string]string{
		// "CONSUMER.LOAN.CLAIM":    "LOAN_ACCOUNT_CREDIT_CLAIM",
		// "CONSUMER.LOAN.WRITEOFF": "LOAN_ACCOUNT_WRITE_OFF",
		"NORMAL.KONVEN.LOAN":   "LOAN_ACCOUNT_NORMAL",
		"GRADUATION.LOAN":      "LOAN_ACCOUNT_GRADUATION",
		"PARTNERSHIP.PAYLATER": "LOAN_ACCOUNT_EFISHERY",
	}
	accountType, ok := mapProductToAccountType[product]
	if !ok {
		err = fmt.Errorf("accountType for loanProduct not found: %s", product)
		return
	}

	// Get loan account
	loanAccount, err := as.srv.igateClient.GetLoanAccount(ctx, igate.LoanAccountGetOptions{AccountNumber: in.AccountNumber})
	if err != nil {
		err = fmt.Errorf("unable to GetLoanAccount %s: %w", in.AccountNumber, err)
		return err
	}

	// Complete account detail
	in.AccountType = accountType
	in.Name = *loanAccount.OwnerName
	// in.AltId -> Loan ID ? Skipped in this migration

	// This column filled by rgate: AccountNumber, OwnerID
	// This column will be filled later in createAccountMigration: EntityCode, ProductTypeCode, CategoryCode, SubCategoryCode, Currency, Status

	_, err = as.createAccountMigration(ctx, in)

	return err
}

// CreateLenderInstitutionAccount will create the lender institution account.
// if insti account, create an invested and receivable account.
// if non-insti account, reject.
//
// This func initially used for migration, so maybe remove it in the future if not used.
func (as *account) CreateLenderInstitutionAccount(ctx context.Context, in models.CreateAccount) (err error) {
	if in.Metadata == nil {
		err = errors.New("metadata is nil")
		return
	}
	metadata := *in.Metadata

	// inquiry customer
	lender, err := as.srv.igateClient.GetLender(ctx, igate.CustomerGetLenderOptions{CustomerNumber: in.OwnerID})
	if err != nil {
		err = fmt.Errorf("unable to get lender from igate: %v", err)
		return err
	}

	// Map accountType
	if lender.Sector == nil {
		err = fmt.Errorf("sector is nil")
		return
	}

	mapSectorToAccountType := map[string]string{
		"2002": "LENDER_INSTITUSI_NON_RDL", // -> lender corporate
		"2003": "LENDER_INSTITUSI_NON_RDL", // -> foreign institution
		"3002": "LENDER_INSTITUSI_NON_RDL", // -> bank as lender institution
	}
	accountType, ok := mapSectorToAccountType[*lender.Sector]
	if !ok {
		err = fmt.Errorf("accountType for sector not found (not institution): %s", *lender.Sector)
		return
	}

	// Complete account detail
	in.AccountType = accountType
	accountTitle1, ok := metadata["accountTitle1"].(string)
	if ok {
		in.Name = accountTitle1
	}

	// This column filled by rgate: AccountNumber, OwnerID, AltId, LegacyId
	// This column will be filled later in createAccountMigration: EntityCode, ProductTypeCode, CategoryCode, SubCategoryCode, Currency, Status

	_, err = as.createAccountMigration(ctx, in)

	return err
}

// CreateBranchPointAccount will create the branch point account.
//
// This func initially used for migration, so maybe remove it in the future if not used.
func (as *account) CreateBranchPointAccount(ctx context.Context, in models.CreateAccount) (err error) {
	if in.Metadata == nil {
		err = errors.New("metadata is nil")
		return
	}
	metadata := *in.Metadata
	bpID, ok := metadata["id"].(string)
	if !ok {
		err = fmt.Errorf("id not available on meta: %v", metadata)
		return
	}

	// Get IA accounts
	iaAccounts, err := as.srv.igateClient.GetAccountIA(ctx, igate.GetAccountIAOptions{AccountNumber: bpID})
	if err != nil {
		err = fmt.Errorf("unable to get account IA %s: %v", bpID, err)
		return err
	}

	if len(iaAccounts) == 0 {
		err = errors.New("account not found")
		return err
	}

	iaAccount := iaAccounts[0]
	if iaAccount == nil {
		err = errors.New("account is nil")
		return err
	}

	// Complete account detail
	if iaAccount.AccountNumber == nil {
		err = errors.New("accountNumber is nil")
		return err
	}
	in.AccountNumber = *iaAccount.AccountNumber
	in.AccountType = "KAS_BP"
	if iaAccount.AccountTitle != nil {
		in.Name = *iaAccount.AccountTitle
	}
	if iaAccount.AltAcctId != nil {
		altAcctIdInt, err := strconv.Atoi(*iaAccount.AltAcctId)
		if err != nil {
			err = fmt.Errorf("unable to parse AltAcctId to int: %w", err)
			return err
		}
		ownerID := altAcctIdInt - 10000000 // ex: 10003467 -> 3467
		in.OwnerID = fmt.Sprint(ownerID)
	}
	// in.AltId = should be empty

	// This column will be filled later in createAccountMigration: EntityCode, ProductTypeCode, CategoryCode, SubCategoryCode, Currency, Status
	_, err = as.createAccountMigration(ctx, in)

	return err
}

func (as *account) publishToAccountStreamDLQ(ctx context.Context, data models.CreateAccount, err error) error {
	messages := models.AccountError{
		CreateAccount: data,
		ErrCauser:     err,
	}
	return as.srv.publisher.PublishSyncWithKeyAndLog(ctx,
		"publish account to pas_account_stream.dlq",
		as.srv.conf.Kafka.Publishers.PASAccountStreamDLQ.Topic,
		data.AccountNumber,
		messages,
	)
}
