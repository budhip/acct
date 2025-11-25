package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/dddnotification"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/file"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/money"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/hashicorp/go-multierror"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

func (as *accounting) GetGeneralLedger(ctx context.Context, opts models.SubLedgerFilterOptions) (account models.SubLedgerAccountResponse, ledgers []models.GetSubLedgerOut, total int, err error) {
	defer func() {
		logService(ctx, err)
	}()

	account, err = as.getAccount(ctx, &opts)
	if err != nil {
		return
	}

	total, err = as.getSubLedgerCount(ctx, opts)
	if err != nil {
		return
	}

	if total > as.srv.conf.AccountConfig.LimitAccountSubLedger {
		// send csv to email
		opts.Limit = 0
		as.SendSubLedgerCSVToEmail(ctx, opts)

		err = models.GetErrMap(models.ErrKeyDataGlisExceedsTheLimit)
		return
	}

	ledgers, err = as.getSubLedgerProcess(ctx, opts)
	if err != nil {
		return
	}

	balancePeriodStart, err := as.getBalancePeriodStart(ctx, opts)
	if err != nil {
		return
	}
	account.BalancePeriodStart = money.FormatAmountToIDR(balancePeriodStart)

	return
}

func (as *accounting) GetSubLedgerAccounts(ctx context.Context, opts models.SubLedgerAccountsFilterOptions) (subLedgerAccounts []models.GetSubLedgerAccountsOut, total int, err error) {
	defer func() {
		logService(ctx, err)
	}()
	acct := as.srv.mySqlRepo.GetAccountingRepository()
	opts.GuestMode = as.srv.flagger.IsEnabled(models.FlagGuestModePayment.String())
	subLedgerAccounts, err = acct.GetSubLedgerAccounts(ctx, opts)
	if err != nil {
		return
	}
	if len(subLedgerAccounts) == 0 {
		err = models.GetErrMap(models.ErrKeyDataNotFound)
		return
	}

	if opts.Search != "" && !opts.StartDate.IsZero() && !opts.EndDate.IsZero() {
		total = len(subLedgerAccounts)
		return
	}
	total, err = as.getSubLedgerAccountsCount(ctx, opts)
	if err != nil {
		return
	}

	return
}

func (as *accounting) GetSubLedger(ctx context.Context, opts models.SubLedgerFilterOptions) (account models.SubLedgerAccountResponse, ledgers []models.GetSubLedgerOut, total int, err error) {
	defer func() {
		logService(ctx, err)
	}()

	account, err = as.getAccount(ctx, &opts)
	if err != nil {
		return
	}

	total, err = as.getSubLedgerCount(ctx, opts)
	if err != nil {
		return
	}
	if total > as.srv.conf.AccountConfig.LimitAccountSubLedger {
		err = models.GetErrMap(models.ErrKeyDataSubLedgerIsExceedsTheLimit)
		return
	}

	ledgers, err = as.getSubLedgerProcess(ctx, opts)
	if err != nil {
		return
	}

	balancePeriodStart, err := as.getBalancePeriodStart(ctx, opts)
	if err != nil {
		return
	}

	account.BalancePeriodStart = money.FormatAmountToIDR(balancePeriodStart)

	return
}

func (as *accounting) GetSubLedgerCount(ctx context.Context, opts models.SubLedgerFilterOptions) (models.GetSubLedgerCountOut, error) {
	total, err := as.getSubLedgerCount(ctx, opts)
	if err != nil {
		return models.GetSubLedgerCountOut{}, err
	}

	return models.GetSubLedgerCountOut{
		Total:             total,
		IsExceedsTheLimit: total > as.srv.conf.AccountConfig.LimitAccountSubLedger,
	}, nil
}

func (as *accounting) DownloadSubLedgerCSV(ctx context.Context, opts models.SubLedgerFilterOptions) (b *bytes.Buffer, filename string, err error) {
	defer func() {
		logService(ctx, err)
	}()

	account, err := as.getAccount(ctx, &opts)
	if err != nil {
		return
	}

	total, err := as.getSubLedgerCount(ctx, opts)
	if err != nil {
		return
	}

	if total > as.srv.conf.AccountConfig.LimitAccountSubLedger {
		err = models.GetErrMap(models.ErrKeyDataSubLedgerIsExceedsTheLimit)
		return
	}

	ledgers, err := as.getSubLedgerProcess(ctx, opts)
	if err != nil {
		return
	}

	balancePeriodStart, err := as.getBalancePeriodStart(ctx, opts)
	if err != nil {
		return
	}

	b, filename, err = as.csvProcessSubLedger(ctx, opts, account, ledgers, balancePeriodStart)
	if err != nil {
		return
	}

	return
}

func (as *accounting) SendSubLedgerCSVToEmail(ctx context.Context, opts models.SubLedgerFilterOptions) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	account, err := as.getAccount(ctx, &opts)
	if err != nil {
		return
	}

	entity, err := as.srv.mySqlRepo.GetEntityRepository().GetByCode(ctx, account.EntityCode)
	if err != nil {
		err = checkDatabaseError(err, models.ErrKeyEntityCodeNotFound)
		return
	}
	if entity == nil {
		err = models.GetErrMap(models.ErrKeyEntityCodeNotFound)
		return
	}

	balancePeriodStart, err := as.getBalancePeriodStart(ctx, opts)
	if err != nil {
		return
	}

	if err = as.csvProcessAndSendEmail(ctx, opts, account, *entity, balancePeriodStart); err != nil {
		return
	}

	return
}

type (
	calculateMethod interface {
		calculateClosingBalance(
			closingBalance, debit, credit decimal.Decimal,
		) decimal.Decimal
	}
	calculateClosingBalanceAsset     struct{}
	calculateClosingBalanceLiability struct{}
)

func (as calculateClosingBalanceAsset) calculateClosingBalance(closingBalance, debit, credit decimal.Decimal) decimal.Decimal {
	closingBalance = closingBalance.Add(debit).Sub(credit)
	return closingBalance
}

func (as calculateClosingBalanceLiability) calculateClosingBalance(closingBalance, debit, credit decimal.Decimal) decimal.Decimal {
	closingBalance = closingBalance.Add(credit).Sub(debit)
	return closingBalance
}

func (as *accounting) csvProcessSubLedger(ctx context.Context,
	opts models.SubLedgerFilterOptions,
	account models.SubLedgerAccountResponse,
	ledgers []models.GetSubLedgerOut,
	balancePeriodStart decimal.Decimal,
) (b *bytes.Buffer, filename string, err error) {

	modCalculate := make(map[string]calculateMethod)
	modCalculate[models.COATypeAsset] = calculateClosingBalanceAsset{}
	modCalculate[models.COATypeLiability] = calculateClosingBalanceLiability{}

	mapOrderTypeCodeName, mapTrxTypeCodeName, err := as.getAllMapOrderTrxType(ctx)
	if err != nil {
		return
	}

	filename = fmt.Sprintf("SubLedger_%s_%s_%s_%s.csv",
		account.AccountNumber,
		account.AccountName,
		opts.StartDate.Format(atime.DateFormatYYYYMMDDWithoutDash),
		opts.EndDate.Format(atime.DateFormatYYYYMMDDWithoutDash),
	)

	b = &bytes.Buffer{}
	as.srv.file.NewCSVWriter(b)

	if err = as.srv.file.CSVWriteAll(ctx, [][]string{
		{"Sub Ledger"},
		{"Account Number", account.AccountNumber},
		{"Account Name", account.AccountName},
		{"Subcategory", account.SubCategoryName},
		{"Alternate ID", account.AltId},
		{"Currency", account.Currency},
		{"Entity", account.EntityName},
		{""},
	}); err != nil {
		return
	}

	header := []string{"No", "Transaction Date", "Transaction ID", "Reference Number", "Order Type Code", "Order Type Name", "Transaction Type Code", "Transaction Type Name", "Narrative", "Metadata", "Debit", "Credit", "Closing Balance"}
	if err = as.srv.file.CSVWriteHeader(ctx, header); err != nil {
		err = fmt.Errorf("failed to write header: %w", err)
		return
	}

	account.BalancePeriodStart = balancePeriodStart.String()
	if err = as.srv.file.CSVWriteHeader(ctx, []string{"Balance at Period Start", "", "", "", "", "", "", "", "", "", "", "", account.BalancePeriodStart}); err != nil {
		err = fmt.Errorf("failed to write header: %w", err)
		return
	}

	closingBalance := balancePeriodStart
	for i, l := range ledgers {

		metadataStr := ""
		if l.Metadata != nil {
			metadataBytes, errMarshal := json.Marshal(l.Metadata)
			if errMarshal != nil {
				err = fmt.Errorf("failed to marshal metadata: %w", errMarshal)
				return
			}
			metadataStr = string(metadataBytes)
		}

		closingBalance = modCalculate[account.COATypeCode].calculateClosingBalance(closingBalance, l.Debit, l.Credit)
		if err = as.srv.file.CSVWriteBody(ctx, []string{
			fmt.Sprint(i + 1),
			l.TransactionDate.In(atime.GetLocation()).Format(atime.DateFormatDDMMYYYYWithSlash),
			l.TransactionID,
			l.ReferenceNumber,
			l.OrderType,
			mapOrderTypeCodeName[l.OrderType],
			l.TransactionType,
			mapTrxTypeCodeName[l.TransactionType],
			l.Narrative,
			metadataStr,
			l.Debit.String(),
			l.Credit.String(),
			closingBalance.String(),
		}); err != nil {
			err = fmt.Errorf("failed to write body: %w", err)
			return
		}
	}

	if err = as.srv.file.CSVWriteHeader(ctx, []string{"Balance at Period End", "", "", "", "", "", "", "", "", "", "", "", closingBalance.String()}); err != nil {
		err = fmt.Errorf("failed to write header: %w", err)
		return
	}

	if err = as.srv.file.CSVProcessWrite(ctx); err != nil {
		return
	}

	return
}

func (as *accounting) csvProcessAndSendEmail(ctx context.Context,
	opts models.SubLedgerFilterOptions,
	account models.SubLedgerAccountResponse,
	entity models.Entity,
	balancePeriodStart decimal.Decimal,
) (err error) {
	modCalculate := make(map[string]calculateMethod)
	modCalculate[models.COATypeAsset] = calculateClosingBalanceAsset{}
	modCalculate[models.COATypeLiability] = calculateClosingBalanceLiability{}

	mapOrderTypeCodeName, mapTrxTypeCodeName, err := as.getAllMapOrderTrxType(ctx)
	if err != nil {
		return
	}

	b := &bytes.Buffer{}
	as.srv.file.NewCSVWriter(b)

	account.BalancePeriodStart = balancePeriodStart.String()
	header1 := [][]string{
		{"Sub Ledger"},
		{"Account Number", account.AccountNumber},
		{"Account Name", account.AccountName},
		{"Subcategory", account.SubCategoryName},
		{"Alternate ID", account.AltId},
		{"Currency", account.Currency},
		{"Entity", account.EntityName},
		{""},
	}
	header2 := []string{"No", "Transaction Date", "Transaction ID", "Reference Number", "Order Type Code", "Order Type Name", "Transaction Type Code", "Transaction Type Name", "Narrative", "Metadata", "Debit", "Credit", "Closing Balance"}
	header3 := []string{"Balance at Period Start", "", "", "", "", "", "", "", "", "", "", "", account.BalancePeriodStart}

	var filepath []models.CloudStoragePayload
	limit := file.EXCEL_MAX_ROWS
	part := 1
	closingBalance := balancePeriodStart

	chanAccountTransaction := as.srv.mySqlRepo.GetAccountingRepository().GetSubLedgerStream(ctx, opts)
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err = as.srv.file.CSVWriteAll(egCtx, header1); err != nil {
			return err
		}

		if err = as.srv.file.CSVWriteHeader(egCtx, header2); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}

		if err = as.srv.file.CSVWriteHeader(egCtx, header3); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}

		i := 1
		for v := range chanAccountTransaction {
			select {
			case <-egCtx.Done():
				return egCtx.Err()
			default:
				if v.Err != nil {
					return v.Err
				}

				l := v.Data
				closingBalance = modCalculate[account.COATypeCode].calculateClosingBalance(closingBalance, l.Debit, l.Credit)

				metadataStr := ""
				if l.Metadata != nil {
					metadataBytes, errMarshal := json.Marshal(l.Metadata)
					if errMarshal != nil {
						err = fmt.Errorf("failed to marshal metadata: %w", errMarshal)
						return err
					}
					metadataStr = string(metadataBytes)
				}

				if err = as.srv.file.CSVWriteBody(egCtx, []string{
					fmt.Sprint(i),
					l.TransactionDate.In(atime.GetLocation()).Format(atime.DateFormatDDMMYYYYWithSlash),
					l.TransactionID,
					l.ReferenceNumber,
					l.OrderType,
					mapOrderTypeCodeName[l.OrderType],
					l.TransactionType,
					mapTrxTypeCodeName[l.TransactionType],
					l.Narrative,
					metadataStr,
					l.Debit.String(),
					l.Credit.String(),
					closingBalance.String(),
				}); err != nil {
					return fmt.Errorf("failed to write body: %w", err)
				}

				if (i)%limit == 0 { // check if limit is reached
					if err = as.srv.file.CSVProcessWrite(egCtx); err != nil {
						return err
					}
					fp, err := as.writeCSVSubledgerToGCS(egCtx, opts, account, part, b)
					if err != nil {
						return err
					}
					filepath = append(filepath, fp)
					part++

					b = &bytes.Buffer{}
					as.srv.file.NewCSVWriter(b)

					if err = as.srv.file.CSVWriteAll(egCtx, header1); err != nil {
						return err
					}

					if err = as.srv.file.CSVWriteHeader(egCtx, header2); err != nil {
						return fmt.Errorf("failed to write header: %w", err)
					}
				}
				i++
			}
		}

		if err = as.srv.file.CSVWriteHeader(ctx, []string{"Balance at Period End", "", "", "", "", "", "", "", "", "", "", "", closingBalance.String()}); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}

		if err = as.srv.file.CSVProcessWrite(ctx); err != nil {
			return err
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		return
	}

	if b.Len() == 0 {
		err = fmt.Errorf("failed get data sub ledger %s", account.AccountNumber)
		return err
	}

	fp, err := as.writeCSVSubledgerToGCS(ctx, opts, account, part, b)
	if err != nil {
		return
	}
	filepath = append(filepath, fp)

	var urls []string
	for i, v := range filepath {
		url, errSignedUrl := as.srv.cloudStorageRepo.GetSignedURL(v, as.srv.conf.CloudStorageConfig.SubLedgerURLDuration)
		if errSignedUrl != nil {
			err = errSignedUrl
			return
		}
		urls = append(urls, fmt.Sprintf("<a href=\"%s\">Part %d</a>", url, i+1))
	}

	requestEmail := dddnotification.RequestEmail{
		From:     "noreply@amartha.com",
		FromName: entity.Description,
		To:       opts.Email,
		Subject: fmt.Sprintf("Sub Ledger - %s - %s ",
			account.AccountNumber,
			account.AccountName,
		),
		Subs: []interface{}{
			map[string]interface{}{
				"ACCOUNT_NUMBER": account.AccountNumber,
				"ACCOUNT_NAME":   account.AccountName,
				"START_DATE":     opts.StartDate.Format(atime.DateFormatYYYYMMDD),
				"END_DATE":       opts.EndDate.Format(atime.DateFormatYYYYMMDD),
				"ENTITY":         entity.Description,
				"URL":            strings.Join(urls, "<br>"),
			},
		},
		Template: as.srv.conf.DDDNotification.EmailTemplateSubLedger,
	}
	if err = as.srv.dddNotificationClient.SendEmail(ctx, requestEmail); err != nil {
		return
	}

	return
}

func (as *accounting) getSubLedgerProcess(
	ctx context.Context,
	opts models.SubLedgerFilterOptions,
) (ledgers []models.GetSubLedgerOut, err error) {
	ledgers, err = as.srv.mySqlRepo.GetAccountingRepository().GetSubLedger(ctx, opts)
	if err != nil {
		return
	}
	return
}

func (as *accounting) getBalancePeriodStart(
	ctx context.Context,
	opts models.SubLedgerFilterOptions,
) (balancePeriodStart decimal.Decimal, err error) {
	acc := as.srv.mySqlRepo.GetAccountingRepository()
	if as.srv.flagger.IsEnabled(models.FlagGetOpeningBalanceV2.String()) {
		balancePeriodStart, err = as.getOpeningBalance(ctx, opts.AccountNumber, opts.StartDate.AddDate(0, 0, -1))
		if err != nil {
			err = checkDatabaseError(err)
			return
		}
	} else {
		balancePeriodStart, err = acc.GetAccountBalancePeriodStart(ctx, opts.AccountNumber, opts.StartDate)
		if err != nil {
			return
		}
	}
	balancePeriodStart = money.FormatBigIntToAmount(balancePeriodStart, CurrencyIDR.Decimals)

	return
}

func (as *accounting) getAccount(ctx context.Context,
	opts *models.SubLedgerFilterOptions,
) (account models.SubLedgerAccountResponse, err error) {
	acc, err := as.srv.mySqlRepo.GetAccountRepository().GetOneByAccountNumber(ctx, opts.AccountNumber)
	if err != nil {
		err = checkDatabaseError(err, models.ErrKeyAccountNumberNotFound)
		return account, err
	}
	account = models.SubLedgerAccountResponse{
		Kind:            models.KindAccount,
		AccountNumber:   acc.AccountNumber,
		AccountName:     acc.AccountName,
		AltId:           acc.AltID,
		COATypeCode:     acc.CoaTypeCode,
		EntityCode:      acc.EntityCode,
		EntityName:      acc.EntityName,
		ProductTypeCode: acc.ProductTypeCode,
		ProductTypeName: acc.ProductTypeName,
		SubCategoryCode: acc.SubCategoryCode,
		SubCategoryName: acc.SubCategoryName,
		Currency:        acc.Currency,
	}
	opts.AccountNumber = acc.AccountNumber

	return
}

func (as *accounting) getAllMapOrderTrxType(ctx context.Context) (orderTypeCodeName map[string]string, trxTypeCodeName map[string]string, err error) {
	orderTypes, err := as.srv.fpTransactionClient.GetAllOrderTypes(ctx, "", "")
	if err != nil {
		return
	}

	orderTypeCodeName = make(map[string]string)
	trxTypeCodeName = make(map[string]string)
	for _, o := range orderTypes.Contents {
		orderTypeCodeName[o.OrderTypeCode] = o.OrderTypeName
		for _, t := range o.TransactionTypes {
			trxTypeCodeName[t.TransactionTypeCode] = t.TransactionTypeName
		}
	}

	return
}

func (as *accounting) getSubLedgerAccountsCount(ctx context.Context, opts models.SubLedgerAccountsFilterOptions) (total int, err error) {
	var errs *multierror.Error

	defer func() {
		logService(ctx, err)
	}()

	key := fmt.Sprintf("pas_sub_ledger_accounts_total_%s", opts.EntityCode)
	val, err := as.srv.cacheRepo.Get(ctx, key)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	if val != "" {
		total, err = strconv.Atoi(val)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	if errs.ErrorOrNil() != nil {
		total, err = as.srv.mySqlRepo.GetAccountingRepository().GetSubLedgerAccountsCount(ctx, opts)
		if err != nil {
			return total, err
		}
		if errSet := as.srv.cacheRepo.Set(ctx, key, total, as.srv.conf.CacheTTL.GetSubLedgerAccountsCount); errSet != nil {
			xlog.Warn(ctx, logMessageService, xlog.Err(errSet))
		}
	}

	return total, err
}

func (as *accounting) writeCSVSubledgerToGCS(
	ctx context.Context,
	opts models.SubLedgerFilterOptions,
	account models.SubLedgerAccountResponse,
	part int,
	b *bytes.Buffer,
) (fp models.CloudStoragePayload, err error) {
	defer func() {
		logService(ctx, err)
	}()

	fp = models.CloudStoragePayload{
		Filename: strings.TrimSpace(fmt.Sprintf("SubLedger-%s-%s-%s-%s-%d.csv",
			account.AccountNumber,
			account.AccountName,
			opts.StartDate.Format(atime.DateFormatYYYYMMDDWithoutDash),
			opts.EndDate.Format(atime.DateFormatYYYYMMDDWithoutDash),
			part,
		)),
		Path: string(models.SubLedgerDir),
	}

	r := as.srv.cloudStorageRepo.NewWriter(ctx, &fp)
	_, err = r.Write(b.Bytes())
	if err != nil {
		if err = r.Close(); err != nil {
			return
		}
		return
	}

	if err = r.Close(); err != nil {
		return
	}

	return
}

func (as *accounting) getSubLedgerCount(ctx context.Context, opts models.SubLedgerFilterOptions) (total int, err error) {
	total, err = as.srv.mySqlRepo.GetAccountingRepository().GetSubLedgerCount(ctx, models.SubLedgerFilterOptions{
		AccountNumber: opts.AccountNumber,
		StartDate:     opts.StartDate,
		EndDate:       opts.EndDate,
		Limit:         as.srv.conf.AccountConfig.LimitAccountSubLedger + 1,
	})
	return
}
