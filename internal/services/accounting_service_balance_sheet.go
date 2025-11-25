package services

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/money"
	"github.com/hashicorp/go-multierror"
)

func (as *accounting) GetBalanceSheet(ctx context.Context, opts models.BalanceSheetFilterOptions) (resp models.GetBalanceSheetResponse, err error) {
	defer func() {
		logService(ctx, err)
	}()

	entity, err := as.srv.mySqlRepo.GetEntityRepository().GetByCode(ctx, opts.EntityCode)
	if err != nil {
		return
	}

	if entity == nil {
		err = models.GetErrMap(models.ErrKeyEntityCodeNotFound)
		return
	}

	out, err := as.srv.mySqlRepo.GetAccountingRepository().GetBalanceSheet(ctx, opts)
	if err != nil {
		return
	}

	resp = models.GetBalanceSheetResponse{
		Kind:             models.KindBalanceSheet,
		EntityCode:       opts.EntityCode,
		EntityName:       entity.Name,
		EntityDesc:       entity.Description,
		BalanceSheetDate: opts.BalanceSheetDate.Format(atime.DateFormatYYYYMMDD),
		BalanceSheet: models.BalanceSheetData{
			Kind:          models.KindBalanceSheetData,
			Assets:        out.BalanceSheet[models.COATypeAsset],
			TotalAsset:    money.FormatAmountToIDRFromDecimal(out.TotalPerCOAType[models.COATypeAsset]),
			DecTotalAsset: out.TotalPerCOAType[models.COATypeAsset],

			Liabilities:       out.BalanceSheet[models.COATypeLiability],
			TotalLiability:    money.FormatAmountToIDRFromDecimal(out.TotalPerCOAType[models.COATypeLiability]),
			DecTotalLiability: out.TotalPerCOAType[models.COATypeLiability],

			CatchAll:    money.FormatAmountToIDRFromDecimal(out.TotalPerCOAType[models.COATypeAsset].Sub(out.TotalPerCOAType[models.COATypeLiability])),
			DecCatchAll: out.TotalPerCOAType[models.COATypeAsset].Sub(out.TotalPerCOAType[models.COATypeLiability]),
		},
	}

	return
}

func (as *accounting) DownloadCSVGetBalanceSheet(ctx context.Context, opts models.BalanceSheetFilterOptions, resp models.GetBalanceSheetResponse) (b *bytes.Buffer, filename string, err error) {
	b = &bytes.Buffer{}
	as.srv.file.NewCSVWriter(b)

	if err = as.srv.file.CSVWriteBody(ctx, []string{resp.EntityDesc}); err != nil {
		return
	}

	if err = as.srv.file.CSVWriteBody(ctx, []string{"BALANCE SHEET REPORT"}); err != nil {
		return
	}

	if err = as.srv.file.CSVWriteBody(ctx, []string{fmt.Sprintf("AS AT CLOSE OF %s", strings.ToUpper(opts.BalanceSheetDate.Format(atime.DateFormatDDMMMYYYYWithSpace)))}); err != nil {
		return
	}

	if err = as.srv.file.CSVWriteBody(ctx, []string{}); err != nil {
		return
	}

	var errs *multierror.Error
	if err = as.srv.file.CSVWriteBody(ctx, []string{
		"ASSETS",
		"Rp",
	}); err != nil {
		errs = multierror.Append(errs, err)
	}
	for _, t := range resp.BalanceSheet.Assets {
		if err = as.srv.file.CSVWriteBody(ctx, []string{
			fmt.Sprintf("A.%s %s", t.CategoryCode, t.CategoryName),
			money.FormatBigIntToAmount(t.DecAmount, CurrencyIDR.Decimals).String(),
		}); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	if errs.ErrorOrNil() != nil {
		err = errs.ErrorOrNil()
		return
	}

	if err = as.srv.file.CSVWriteBody(ctx, []string{"TOTAL ASSETS", money.FormatBigIntToAmount(resp.BalanceSheet.DecTotalAsset, CurrencyIDR.Decimals).String()}); err != nil {
		err = fmt.Errorf("failed to write body: %w", err)
		return
	}

	if err = as.srv.file.CSVWriteBody(ctx, []string{}); err != nil {
		err = fmt.Errorf("failed to write body: %w", err)
		return
	}

	if err = as.srv.file.CSVWriteBody(ctx, []string{
		"LIABILITIES",
		"Rp",
	}); err != nil {
		errs = multierror.Append(errs, err)
	}
	for _, t := range resp.BalanceSheet.Liabilities {
		if err = as.srv.file.CSVWriteBody(ctx, []string{
			fmt.Sprintf("B.%s %s", t.CategoryCode, t.CategoryName),
			money.FormatBigIntToAmount(t.DecAmount, CurrencyIDR.Decimals).String(),
		}); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	if errs.ErrorOrNil() != nil {
		err = errs.ErrorOrNil()
		return
	}

	if err = as.srv.file.CSVWriteBody(ctx, []string{"TOTAL LIABILITIES", money.FormatBigIntToAmount(resp.BalanceSheet.DecTotalLiability, CurrencyIDR.Decimals).String()}); err != nil {
		err = fmt.Errorf("failed to write body: %w", err)
		return
	}

	if err = as.srv.file.CSVWriteBody(ctx, []string{}); err != nil {
		err = fmt.Errorf("failed to write body: %w", err)
		return
	}

	if err = as.srv.file.CSVWriteBody(ctx, []string{"Catch all (Balancing)", money.FormatBigIntToAmount(resp.BalanceSheet.DecCatchAll, CurrencyIDR.Decimals).String()}); err != nil {
		err = fmt.Errorf("failed to write body: %w", err)
		return
	}

	if err = as.srv.file.CSVProcessWrite(ctx); err != nil {
		return
	}

	filename = fmt.Sprintf("Balance-Sheets-%s-%s.csv", resp.EntityName, opts.BalanceSheetDate.Format(atime.DateFormatYYYYMMDDWithoutDash))

	return
}
