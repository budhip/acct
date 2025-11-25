package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/dddnotification"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/file"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/godbledger"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/money"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/cache"
	xlog "bitbucket.org/Amartha/go-x/log"
	"github.com/shopspring/decimal"
)

var CurrencyIDR = godbledger.CurrencyIDR

type AccountingService interface {
	// Trial Balance
	GetTrialBalance(ctx context.Context, opts models.TrialBalanceFilterOptions) (resp models.GetTrialBalanceResponses, err error)
	GetTrialBalanceBySubCategoryCode(ctx context.Context, opts models.TrialBalanceFilterOptions) (resp models.TrialBalanceBySubCategoryOut, err error)
	GetTrialBalanceDetails(ctx context.Context, opts models.TrialBalanceDetailsFilterOptions) (accounts []models.TrialBalanceDetailOut, total int, err error)
	GetTrialBalanceFromGCS(ctx context.Context, opts models.TrialBalanceDetailsFilterOptions) (tba []models.TrialBalanceDetailOut, summary models.TrialBalanceBySubCategoryOut, err error)

	DownloadTrialBalance(ctx context.Context, opts models.TrialBalanceFilterOptions) (err error)
	SendToEmailGetTrialBalanceDetails(ctx context.Context, opts models.TrialBalanceDetailsFilterOptions) (err error)
	SendEmailTrialBalanceDetails(ctx context.Context, opts models.TrialBalanceDetailsFilterOptions) (err error)
	SendEmailTrialBalanceSummary(ctx context.Context, opts models.TrialBalanceFilterOptions) (err error)

	// Ledger
	GetGeneralLedger(ctx context.Context, opts models.SubLedgerFilterOptions) (account models.SubLedgerAccountResponse, ledgers []models.GetSubLedgerOut, total int, err error)
	GetSubLedger(ctx context.Context, opts models.SubLedgerFilterOptions) (account models.SubLedgerAccountResponse, ledgers []models.GetSubLedgerOut, total int, err error)
	GetSubLedgerCount(ctx context.Context, opts models.SubLedgerFilterOptions) (models.GetSubLedgerCountOut, error)
	SendSubLedgerCSVToEmail(ctx context.Context, opts models.SubLedgerFilterOptions) (err error)
	DownloadSubLedgerCSV(ctx context.Context, opts models.SubLedgerFilterOptions) (b *bytes.Buffer, filename string, err error)
	GetSubLedgerAccounts(ctx context.Context, opts models.SubLedgerAccountsFilterOptions) (subLedgerAccounts []models.GetSubLedgerAccountsOut, total int, err error)

	// Balance Sheet
	GetBalanceSheet(ctx context.Context, opts models.BalanceSheetFilterOptions) (resp models.GetBalanceSheetResponse, err error)
	DownloadCSVGetBalanceSheet(ctx context.Context, opts models.BalanceSheetFilterOptions, resp models.GetBalanceSheetResponse) (b *bytes.Buffer, filename string, err error)

	// Job
	GenerateTrialBalanceBigQuery(ctx context.Context, date time.Time, isAdjustment bool) (err error)
	GenerateAdjustmentTrialBalanceBigQuery(ctx context.Context, in models.AdjustmentTrialBalanceFilter) (err error)

	GenerateAccountDailyBalance(ctx context.Context, date time.Time) (err error)
	GenerateAccountTrialBalanceDaily(ctx context.Context, date time.Time) (err error)
}

type accounting service

var _ AccountingService = (*accounting)(nil)

/*
1. validate entity code
2. query to database
3. loop result from database
- calculate total per category
- calculate total to asset per coa type
- calculate catch total asset - liability
= transform to response
*/
func (as *accounting) GetTrialBalance(ctx context.Context, opts models.TrialBalanceFilterOptions) (resp models.GetTrialBalanceResponses, err error) {
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

	var (
		coaCategories    map[string][]models.TBCOACategory
		coaSubCategories map[models.TBCOACategory][]models.TBSubCategory
		mapSubCategory   map[string]models.CategorySubCategoryCOAType
	)

	switch {
	case as.srv.flagger.IsEnabled(models.FlagGetTrialBalanceGCS.String()):
		if opts.Period.IsZero() {
			return resp, fmt.Errorf("period is required for the new trial balance flow")
		}
		payload := models.CloudStoragePayload{
			Filename: fmt.Sprintf("%s/%s/summaries/%s.csv",
				opts.Period.Format("2006"),
				opts.EntityCode,
				opts.Period.Format("01"),
			),
			Path: fmt.Sprintf("%s", models.TrialBalanceDir),
		}

		reader, errSigned := as.srv.cloudStorageRepo.NewReader(ctx, &payload)
		if errSigned != nil {
			return
		}

		csvData, errCSV := as.parseTrialBalanceCSV(ctx, reader)
		if errCSV != nil {
			return
		}

		_, _, mapSubCategory, err = as.srv.mySqlRepo.GetAllCategorySubCategoryCOAType(ctx)
		if err != nil {
			return resp, fmt.Errorf("failed to get coa mapping: %w", err)
		}

		coaCategories, coaSubCategories = aggregateTrialBalance(csvData, mapSubCategory)

		//for get trial balance status
		trialBalanceData, err := as.srv.mySqlRepo.GetTrialBalanceRepository().GetByPeriod(ctx, opts.Period.Format(atime.DateFormatYYYYMM), opts.EntityCode)
		if err != nil {
			err = checkDatabaseError(err, models.ErrKeyClosedPeriodNotFound)
			return resp, err
		}

		resp = as.transformToTrialBalanceGCS(coaCategories, coaSubCategories)
		resp.EntityCode = entity.Code
		resp.EntityName = entity.Name
		resp.Status = trialBalanceData.Status
		resp.ClosingDate = opts.Period.Format(atime.DateFormatYYYYMM)

	case as.srv.flagger.IsEnabled(models.FlagGetTrialBalanceV2.String()):
		coaCategories, coaSubCategories, err = as.srv.mySqlRepo.GetAccountingRepository().GetTrialBalanceV2(ctx, opts)
		if err != nil {
			return
		}
		resp = as.transformToTrialBalance(coaCategories, coaSubCategories)
		resp.EntityCode = entity.Code
		resp.EntityName = entity.Name
		resp.ClosingDate = opts.EndDate.Format(atime.DateFormatYYYYMMDD)

	default:
		coaCategories, coaSubCategories, err = as.srv.mySqlRepo.GetAccountingRepository().GetTrialBalance(ctx, opts)
		if err != nil {
			return
		}
		resp = as.transformToTrialBalance(coaCategories, coaSubCategories)
		resp.EntityCode = entity.Code
		resp.EntityName = entity.Name
		resp.ClosingDate = opts.EndDate.Format(atime.DateFormatYYYYMMDD)
	}

	return
}

func (as *accounting) parseTrialBalanceCSV(ctx context.Context, reader io.ReadCloser) ([]models.TrialBalanceCSV, error) {

	var results []models.TrialBalanceCSV
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true
	records, err := csvReader.ReadAll()
	if err != nil {
		xlog.Errorf(ctx, "failed to read CSV: %v", err)
		return results, err
	}

	if len(records) <= 1 {
		return results, fmt.Errorf("no data found in csv")
	}

	for _, row := range records[1:] {
		if len(row) < 7 {
			continue
		}

		debit, err := decimal.NewFromString(row[3])
		if err != nil {
			return results, err
		}

		credit, err := decimal.NewFromString(row[4])
		if err != nil {
			return results, err
		}

		opening, err := decimal.NewFromString(row[5])
		if err != nil {
			return results, err
		}

		closing, err := decimal.NewFromString(row[6])
		if err != nil {
			return results, err
		}

		results = append(results, models.TrialBalanceCSV{
			EntityCode:      row[0],
			CategoryCode:    row[1],
			SubCategoryCode: row[2],
			DebitMovement:   debit,
			CreditMovement:  credit,
			OpeningBalance:  opening,
			ClosingBalance:  closing,
		})
	}
	return results, nil
}

func aggregateTrialBalance(data []models.TrialBalanceCSV, mapSubCategory map[string]models.CategorySubCategoryCOAType) (map[string][]models.TBCOACategory, map[models.TBCOACategory][]models.TBSubCategory) {
	coaCategories := make(map[string][]models.TBCOACategory)
	coaSubCategories := make(map[models.TBCOACategory][]models.TBSubCategory)

	for _, v := range data {
		meta := mapSubCategory[v.SubCategoryCode]
		if meta.CoaTypeName == "" {
			continue
		}

		key := models.TBCOACategory{
			Type:         strings.ToLower(meta.CoaTypeName),
			CoaTypeCode:  meta.CoaTypeCode,
			CoaTypeName:  meta.CoaTypeName,
			CategoryCode: v.CategoryCode,
			CategoryName: meta.CategoryName,
		}

		coaSubCategories[key] = append(coaSubCategories[key], models.TBSubCategory{
			Kind:            models.KindSubCategory,
			SubCategoryCode: v.SubCategoryCode,
			SubCategoryName: meta.SubCategoryName,
			OpeningBalance:  v.OpeningBalance,
			DebitMovement:   v.DebitMovement,
			CreditMovement:  v.CreditMovement,
			ClosingBalance:  v.ClosingBalance,

			IDRFormatOpeningBalance: money.FormatAmountToIDRFromDecimalGCS(v.OpeningBalance),
			IDRFormatDebitMovement:  money.FormatAmountToIDRFromDecimalGCS(v.DebitMovement),
			IDRFormatCreditMovement: money.FormatAmountToIDRFromDecimalGCS(v.CreditMovement),
			IDRFormatClosingBalance: money.FormatAmountToIDRFromDecimalGCS(v.ClosingBalance),
		})

		coaCategories[v.CategoryCode] = append(coaCategories[v.CategoryCode], models.TBCOACategory{
			Type:                meta.CoaTypeName,
			CategoryCode:        v.CategoryCode,
			CategoryName:        meta.CategoryName,
			TotalOpeningBalance: v.OpeningBalance,
			TotalDebitMovement:  v.DebitMovement,
			TotalCreditMovement: v.CreditMovement,
			TotalClosingBalance: v.ClosingBalance,
		})
	}

	return coaCategories, coaSubCategories
}

func (as *accounting) transformToTrialBalanceGCS(coaCategories map[string][]models.TBCOACategory, coaSubCategories map[models.TBCOACategory][]models.TBSubCategory) (trialBalance models.GetTrialBalanceResponses) {
	var (
		asset    models.TrialBalance
		liablity models.TrialBalance
		coaTypes []models.TrialBalance
	)

	for i, val := range coaSubCategories {
		if i.Type == models.TrialBalanceTypeAsset {
			coaTypeGCS(&asset, i, val, coaCategories)
		}
		if i.Type == models.TrialBalanceTypeLiability {
			coaTypeGCS(&liablity, i, val, coaCategories)
		}
	}

	trialBalance = models.GetTrialBalanceResponses{
		Kind:     models.KindTrialBalance,
		COATypes: make([]models.TrialBalance, 0),
	}
	coaTypes = append(coaTypes, asset, liablity)
	trialBalance.COATypes = append(trialBalance.COATypes, coaTypes...)

	for i := range trialBalance.COATypes {
		for j := range trialBalance.COATypes[i].Categories {
			sort.Slice(trialBalance.COATypes[i].Categories[j].SubCategories, func(a, b int) bool {
				return trialBalance.COATypes[i].Categories[j].SubCategories[a].SubCategoryCode <
					trialBalance.COATypes[i].Categories[j].SubCategories[b].SubCategoryCode
			})
		}
	}

	// calculate total asset - total liablity
	trialBalance.CatchAll.CatchAllOpeningBalance = asset.TotalOpeningBalance.Sub(liablity.TotalOpeningBalance)
	trialBalance.CatchAll.CatchAllDebitMovement = asset.TotalDebitMovement.Add(liablity.TotalDebitMovement)
	trialBalance.CatchAll.CatchAllCreditMovement = asset.TotalCreditMovement.Add(liablity.TotalCreditMovement)
	trialBalance.CatchAll.CatchAllClosingBalance = asset.TotalClosingBalance.Sub(liablity.TotalClosingBalance)

	trialBalance.CatchAll.IDRFormatCatchAllOpeningBalance = money.FormatAmountToIDRFromDecimalGCS(trialBalance.CatchAll.CatchAllOpeningBalance)
	trialBalance.CatchAll.IDRFormatCatchAllDebitMovement = money.FormatAmountToIDRFromDecimalGCS(trialBalance.CatchAll.CatchAllDebitMovement)
	trialBalance.CatchAll.IDRFormatCatchAllCreditMovement = money.FormatAmountToIDRFromDecimalGCS(trialBalance.CatchAll.CatchAllCreditMovement)
	trialBalance.CatchAll.IDRFormatCatchAllClosingBalance = money.FormatAmountToIDRFromDecimalGCS(trialBalance.CatchAll.CatchAllClosingBalance)

	return
}

func coaTypeGCS(tb *models.TrialBalance, i models.TBCOACategory, val []models.TBSubCategory, coaCategories map[string][]models.TBCOACategory) {
	tb.Kind = models.KindCOAType
	tb.CoaTypeCode = i.CoaTypeCode
	tb.CoaTypeName = i.CoaTypeName
	categorySubCategory := models.TBCategorySubCategory{
		Kind:          models.KindCategory,
		CategoryCode:  i.CategoryCode,
		CategoryName:  i.CategoryName,
		SubCategories: val,
	}

	// calculate total per category
	categories := coaCategories[i.CategoryCode]
	for _, v := range categories {
		categorySubCategory.TotalOpeningBalance = categorySubCategory.TotalOpeningBalance.Add(v.TotalOpeningBalance)
		categorySubCategory.TotalDebitMovement = categorySubCategory.TotalDebitMovement.Add(v.TotalDebitMovement)
		categorySubCategory.TotalCreditMovement = categorySubCategory.TotalCreditMovement.Add(v.TotalCreditMovement)
		categorySubCategory.TotalClosingBalance = categorySubCategory.TotalClosingBalance.Add(v.TotalClosingBalance)

		// calculate total to asset
		tb.TotalOpeningBalance = tb.TotalOpeningBalance.Add(v.TotalOpeningBalance)
		tb.TotalDebitMovement = tb.TotalDebitMovement.Add(v.TotalDebitMovement)
		tb.TotalCreditMovement = tb.TotalCreditMovement.Add(v.TotalCreditMovement)
		tb.TotalClosingBalance = tb.TotalClosingBalance.Add(v.TotalClosingBalance)
	}
	categorySubCategory.IDRFormatTotalOpeningBalance = money.FormatAmountToIDRFromDecimalGCS(categorySubCategory.TotalOpeningBalance)
	categorySubCategory.IDRFormatTotalDebitMovement = money.FormatAmountToIDRFromDecimalGCS(categorySubCategory.TotalDebitMovement)
	categorySubCategory.IDRFormatTotalCreditMovement = money.FormatAmountToIDRFromDecimalGCS(categorySubCategory.TotalCreditMovement)
	categorySubCategory.IDRFormatTotalClosingBalance = money.FormatAmountToIDRFromDecimalGCS(categorySubCategory.TotalClosingBalance)

	tb.IDRFormatTotalOpeningBalance = money.FormatAmountToIDRFromDecimalGCS(tb.TotalOpeningBalance)
	tb.IDRFormatTotalDebitMovement = money.FormatAmountToIDRFromDecimalGCS(tb.TotalDebitMovement)
	tb.IDRFormatTotalCreditMovement = money.FormatAmountToIDRFromDecimalGCS(tb.TotalCreditMovement)
	tb.IDRFormatTotalClosingBalance = money.FormatAmountToIDRFromDecimalGCS(tb.TotalClosingBalance)

	tb.Categories = append(tb.Categories, categorySubCategory)
	sort.Slice(tb.Categories, func(i, j int) bool {
		return tb.Categories[i].CategoryCode < tb.Categories[j].CategoryCode
	})
}

func (as *accounting) transformToTrialBalance(coaCategories map[string][]models.TBCOACategory, coaSubCategories map[models.TBCOACategory][]models.TBSubCategory) (trialBalance models.GetTrialBalanceResponses) {
	var (
		asset    models.TrialBalance
		liablity models.TrialBalance
		coaTypes []models.TrialBalance
	)

	for i, val := range coaSubCategories {
		if i.Type == models.TrialBalanceTypeAsset {
			coaType(&asset, i, val, coaCategories)
		}
		if i.Type == models.TrialBalanceTypeLiability {
			coaType(&liablity, i, val, coaCategories)
		}
	}

	trialBalance = models.GetTrialBalanceResponses{
		Kind:     models.KindTrialBalance,
		COATypes: make([]models.TrialBalance, 0),
	}
	coaTypes = append(coaTypes, asset, liablity)
	trialBalance.COATypes = append(trialBalance.COATypes, coaTypes...)

	// calculate total asset - total liablity
	trialBalance.CatchAll.CatchAllOpeningBalance = asset.TotalOpeningBalance.Sub(liablity.TotalOpeningBalance)
	trialBalance.CatchAll.CatchAllDebitMovement = asset.TotalDebitMovement.Add(liablity.TotalDebitMovement)
	trialBalance.CatchAll.CatchAllCreditMovement = asset.TotalCreditMovement.Add(liablity.TotalCreditMovement)
	trialBalance.CatchAll.CatchAllClosingBalance = asset.TotalClosingBalance.Sub(liablity.TotalClosingBalance)

	trialBalance.CatchAll.IDRFormatCatchAllOpeningBalance = money.FormatAmountToIDRFromDecimal(trialBalance.CatchAll.CatchAllOpeningBalance)
	trialBalance.CatchAll.IDRFormatCatchAllDebitMovement = money.FormatAmountToIDRFromDecimal(trialBalance.CatchAll.CatchAllDebitMovement)
	trialBalance.CatchAll.IDRFormatCatchAllCreditMovement = money.FormatAmountToIDRFromDecimal(trialBalance.CatchAll.CatchAllCreditMovement)
	trialBalance.CatchAll.IDRFormatCatchAllClosingBalance = money.FormatAmountToIDRFromDecimal(trialBalance.CatchAll.CatchAllClosingBalance)

	return
}

func coaType(tb *models.TrialBalance, i models.TBCOACategory, val []models.TBSubCategory, coaCategories map[string][]models.TBCOACategory) {
	tb.Kind = models.KindCOAType
	tb.CoaTypeCode = i.CoaTypeCode
	tb.CoaTypeName = i.CoaTypeName
	categorySubCategory := models.TBCategorySubCategory{
		Kind:          models.KindCategory,
		CategoryCode:  i.CategoryCode,
		CategoryName:  i.CategoryName,
		SubCategories: val,
	}

	// calculate total per category
	categories := coaCategories[i.CategoryCode]
	for _, v := range categories {
		categorySubCategory.TotalOpeningBalance = categorySubCategory.TotalOpeningBalance.Add(v.TotalOpeningBalance)
		categorySubCategory.TotalDebitMovement = categorySubCategory.TotalDebitMovement.Add(v.TotalDebitMovement)
		categorySubCategory.TotalCreditMovement = categorySubCategory.TotalCreditMovement.Add(v.TotalCreditMovement)
		categorySubCategory.TotalClosingBalance = categorySubCategory.TotalClosingBalance.Add(v.TotalClosingBalance)

		// calculate total to asset
		tb.TotalOpeningBalance = tb.TotalOpeningBalance.Add(v.TotalOpeningBalance)
		tb.TotalDebitMovement = tb.TotalDebitMovement.Add(v.TotalDebitMovement)
		tb.TotalCreditMovement = tb.TotalCreditMovement.Add(v.TotalCreditMovement)
		tb.TotalClosingBalance = tb.TotalClosingBalance.Add(v.TotalClosingBalance)
	}
	categorySubCategory.IDRFormatTotalOpeningBalance = money.FormatAmountToIDRFromDecimal(categorySubCategory.TotalOpeningBalance)
	categorySubCategory.IDRFormatTotalDebitMovement = money.FormatAmountToIDRFromDecimal(categorySubCategory.TotalDebitMovement)
	categorySubCategory.IDRFormatTotalCreditMovement = money.FormatAmountToIDRFromDecimal(categorySubCategory.TotalCreditMovement)
	categorySubCategory.IDRFormatTotalClosingBalance = money.FormatAmountToIDRFromDecimal(categorySubCategory.TotalClosingBalance)

	tb.IDRFormatTotalOpeningBalance = money.FormatAmountToIDRFromDecimal(tb.TotalOpeningBalance)
	tb.IDRFormatTotalDebitMovement = money.FormatAmountToIDRFromDecimal(tb.TotalDebitMovement)
	tb.IDRFormatTotalCreditMovement = money.FormatAmountToIDRFromDecimal(tb.TotalCreditMovement)
	tb.IDRFormatTotalClosingBalance = money.FormatAmountToIDRFromDecimal(tb.TotalClosingBalance)

	tb.Categories = append(tb.Categories, categorySubCategory)
	sort.Slice(tb.Categories, func(i, j int) bool {
		return tb.Categories[i].CategoryCode < tb.Categories[j].CategoryCode
	})
}

func (as *accounting) DownloadTrialBalance(ctx context.Context, opts models.TrialBalanceFilterOptions) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	result, err := as.GetTrialBalance(ctx, opts)
	if err != nil {
		return
	}

	if err = as.trialBalanceCSVProcessAndSendEmail(ctx, models.DownloadTrialBalanceRequest{
		Opts:   opts,
		Result: result,
	}); err != nil {
		return
	}

	return
}

func (as *accounting) trialBalanceCSVProcessAndSendEmail(ctx context.Context, in models.DownloadTrialBalanceRequest) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	entityData, err := as.srv.mySqlRepo.GetEntityRepository().GetByCode(ctx, in.Opts.EntityCode)
	if err != nil {
		return
	}
	if entityData == nil {
		err = models.GetErrMap(models.ErrKeyEntityCodeNotFound)
		return
	}

	b := &bytes.Buffer{}
	as.srv.file.NewCSVWriter(b)

	// write title
	if err = as.srv.file.CSVWriteBody(ctx, []string{entityData.Description}); err != nil {
		err = fmt.Errorf("failed to write body: %w", err)
		return
	}

	if err = as.srv.file.CSVWriteBody(ctx, []string{"TRIAL BALANCE REPORT"}); err != nil {
		err = fmt.Errorf("failed to write body: %w", err)
		return
	}

	if err = as.srv.file.CSVWriteBody(ctx, []string{fmt.Sprintf("AS AT CLOSE OF %s", strings.ToUpper(in.Opts.EndDate.Format(atime.DateFormatDDMMMYYYYWithSpace)))}); err != nil {
		err = fmt.Errorf("failed to write body: %w", err)
		return
	}

	if err = as.srv.file.CSVWriteBody(ctx, []string{}); err != nil {
		err = fmt.Errorf("failed to write body: %w", err)
		return
	}

	// write contents
	header := []string{"LINE", "SUBCATEGORY", "OPENING BALANCE", "DEBIT MOVEMENTS", "CREDIT MOVEMENTS", "CLOSING BALANCE"}
	if err = as.srv.file.CSVWriteHeader(ctx, header); err != nil {
		err = fmt.Errorf("failed to write header: %w", err)
		return
	}

	res := in.Result
	for i, t := range res.COATypes {
		if err = as.srv.file.CSVWriteBody(ctx, []string{
			fmt.Sprint(i + 1),
			t.CoaTypeName,
		}); err != nil {
			err = fmt.Errorf("failed to write body: %w", err)
			return
		}

		for _, c := range t.Categories {
			// write categories
			if err = as.srv.file.CSVWriteBody(ctx, []string{
				fmt.Sprintf("%s.%s", c.CategoryCode, c.CategoryName),
				"",
				money.FormatBigIntToAmount(c.TotalOpeningBalance, CurrencyIDR.Decimals).String(),
				money.FormatBigIntToAmount(c.TotalDebitMovement, CurrencyIDR.Decimals).String(),
				money.FormatBigIntToAmount(c.TotalCreditMovement, CurrencyIDR.Decimals).String(),
				money.FormatBigIntToAmount(c.TotalClosingBalance, CurrencyIDR.Decimals).String(),
			}); err != nil {
				err = fmt.Errorf("failed to write body: %w", err)
				return
			}

			// write sub categories
			for _, sc := range c.SubCategories {
				if err = as.srv.file.CSVWriteBody(ctx, []string{
					"",
					fmt.Sprintf("%s.%s", sc.SubCategoryCode, sc.SubCategoryName),
					money.FormatBigIntToAmount(sc.OpeningBalance, CurrencyIDR.Decimals).String(),
					money.FormatBigIntToAmount(sc.DebitMovement, CurrencyIDR.Decimals).String(),
					money.FormatBigIntToAmount(sc.CreditMovement, CurrencyIDR.Decimals).String(),
					money.FormatBigIntToAmount(sc.ClosingBalance, CurrencyIDR.Decimals).String(),
				}); err != nil {
					err = fmt.Errorf("failed to write body: %w", err)
					return
				}
			}
		}

		if err = as.srv.file.CSVWriteBody(ctx, []string{
			t.CoaTypeCode,
			fmt.Sprintf("Total %s", t.CoaTypeName),
			money.FormatBigIntToAmount(t.TotalOpeningBalance, CurrencyIDR.Decimals).String(),
			money.FormatBigIntToAmount(t.TotalDebitMovement, CurrencyIDR.Decimals).String(),
			money.FormatBigIntToAmount(t.TotalCreditMovement, CurrencyIDR.Decimals).String(),
			money.FormatBigIntToAmount(t.TotalClosingBalance, CurrencyIDR.Decimals).String(),
		}); err != nil {
			err = fmt.Errorf("failed to write body: %w", err)
			return
		}

		if err = as.srv.file.CSVWriteBody(ctx, []string{}); err != nil {
			err = fmt.Errorf("failed to write body: %w", err)
			return
		}
	}

	// write catch all
	if err = as.srv.file.CSVWriteBody(ctx, []string{
		"",
		"Catch All",
		money.FormatBigIntToAmount(res.CatchAll.CatchAllOpeningBalance, CurrencyIDR.Decimals).String(),
		money.FormatBigIntToAmount(res.CatchAll.CatchAllDebitMovement, CurrencyIDR.Decimals).String(),
		money.FormatBigIntToAmount(res.CatchAll.CatchAllCreditMovement, CurrencyIDR.Decimals).String(),
		money.FormatBigIntToAmount(res.CatchAll.CatchAllClosingBalance, CurrencyIDR.Decimals).String(),
	}); err != nil {
		err = fmt.Errorf("failed to write body: %w", err)
		return
	}

	if err = as.srv.file.CSVProcessWrite(ctx); err != nil {
		return
	}

	entityDesc := entityData.Description
	startDate := in.Opts.StartDate.Format(atime.DateFormatYYYYMMDD)
	endDate := in.Opts.EndDate.Format(atime.DateFormatYYYYMMDD)

	requestEmail := dddnotification.RequestEmail{
		From:     "noreply@amartha.com",
		FromName: entityDesc,
		To:       in.Opts.Email,
		Subject: fmt.Sprintf("Trial Balance - %s - %s to %s",
			entityDesc,
			startDate,
			endDate,
		),
		Attachments: []dddnotification.Attachment{
			{
				Type:    "text/csv",
				Name:    fmt.Sprintf("trialBalance-%s-%s.csv", startDate, endDate),
				Content: base64.StdEncoding.EncodeToString(b.Bytes()),
			},
		},
		Subs: []interface{}{
			map[string]interface{}{
				"ENTITY":     entityDesc,
				"START_DATE": startDate,
				"END_DATE":   endDate,
			},
		},
		Template: as.srv.conf.DDDNotification.EmailTemplateTrialBalance,
	}
	if err = as.srv.dddNotificationClient.SendEmail(ctx, requestEmail); err != nil {
		return
	}

	return
}

func (as *accounting) GetTrialBalanceDetails(ctx context.Context, opts models.TrialBalanceDetailsFilterOptions) (tba []models.TrialBalanceDetailOut, total int, err error) {
	defer func() {
		logService(ctx, err)
	}()

	ent, err := as.srv.mySqlRepo.GetEntityRepository().GetByCode(ctx, opts.EntityCode)
	if err != nil {
		return
	}

	if ent == nil {
		err = models.GetErrMap(models.ErrKeyEntityCodeNotFound)
		return
	}

	sc, err := as.srv.mySqlRepo.GetSubCategoryRepository().GetByCode(ctx, opts.SubCategoryCode)
	if err != nil {
		return
	}

	if sc == nil {
		err = models.GetErrMap(models.ErrKeySubCategoryCodeNotFound)
		return
	}

	tba, err = as.srv.mySqlRepo.GetAccountingRepository().GetTrialBalanceDetails(ctx, opts)
	if err != nil {
		return
	}

	total, err = cache.GetOrSet(as.srv.cacheRepo, models.GetOrSetCacheOpts[int]{
		Ctx: ctx,
		Key: models.GenerateTrialBalanceDetailsCountCacheKey(opts),
		TTL: as.srv.conf.CacheTTL.GetTrialBalanceDetailsCount,
		Callback: func() (int, error) {
			return as.srv.mySqlRepo.GetAccountRepository().GetAccountListCount(ctx, models.AccountFilterOptions{
				EntityCode:      opts.EntityCode,
				SubCategoryCode: opts.SubCategoryCode,
				Search:          opts.Search,
				SearchBy:        "account_number",
			})
		},
	})
	if err != nil {
		return
	}

	return
}

func (as *accounting) GetTrialBalanceFromGCS(ctx context.Context, opts models.TrialBalanceDetailsFilterOptions) (tbd []models.TrialBalanceDetailOut, summary models.TrialBalanceBySubCategoryOut, err error) {
	defer func() {
		logService(ctx, err)
	}()

	path := fmt.Sprintf("%s/%s/%s/details/%s/%s", models.TrialBalanceDir, opts.Year, opts.EntityCode, opts.Month, opts.SubCategoryCode)

	fp := models.CloudStoragePayload{
		Path: path,
	}
	totalRows, err := as.getTotalRowsFromGCS(ctx, fp)
	if err != nil {
		return nil, summary, err
	}
	if totalRows > as.srv.conf.AccountConfig.LimitAccountTrialBalance {
		err = models.GetErrMap(models.ErrKeyDataTrialBalanceDetailIsExceedsTheLimit)
		return nil, summary, err
	}

	tbd, err = as.loadTrialBalanceDetailsFromGCS(ctx, opts, fp)
	if err != nil {
		return nil, summary, err
	}

	subCategory, err := as.srv.mySqlRepo.GetSubCategoryRepository().GetByCode(ctx, opts.SubCategoryCode)
	if err != nil {
		return nil, summary, err
	}

	summary = as.calculateTrialBalanceSummary(tbd, subCategory)

	return
}

func (as *accounting) loadTrialBalanceDetailsFromGCS(ctx context.Context, opts models.TrialBalanceDetailsFilterOptions, payload models.CloudStoragePayload) (tbd []models.TrialBalanceDetailOut, err error) {

	filenames, err := as.srv.cloudStorageRepo.ListFiles(ctx, &payload)
	if err != nil {
		return nil, err
	}

	for _, file := range filenames {
		if strings.Contains(file, "total_rows") {
			continue
		}
		payload.Filename = file
		data, err := as.readTrialBalanceDetailCSV(ctx, payload)
		if err != nil {
			return nil, err
		}
		tbd = append(tbd, data...)
	}

	return
}

func (as *accounting) getTotalRowsFromGCS(ctx context.Context, payload models.CloudStoragePayload) (total int, err error) {
	payload.Filename = "total_rows.json"
	reader, err := as.srv.cloudStorageRepo.NewReader(ctx, &payload)
	if err != nil {
		return 0, err
	}

	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return 0, err
	}

	var summary struct {
		TotalRows string `json:"total_rows"`
	}
	if err := json.Unmarshal(data, &summary); err != nil {
		return 0, err
	}

	total, err = strconv.Atoi(summary.TotalRows)
	if err != nil {
		return 0, err
	}

	return
}

func (as *accounting) readTrialBalanceDetailCSV(ctx context.Context, payload models.CloudStoragePayload) (tbd []models.TrialBalanceDetailOut, err error) {
	rc, err := as.srv.cloudStorageRepo.NewReader(ctx, &payload)
	if err != nil {
		return nil, err
	}

	defer rc.Close()

	reader := csv.NewReader(rc)
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	for i, row := range records {
		if i == 0 {
			continue // skip header
		}

		opening, _ := decimal.NewFromString(row[5])
		debit, _ := decimal.NewFromString(row[6])
		credit, _ := decimal.NewFromString(row[7])
		closing, _ := decimal.NewFromString(row[8])

		item := models.TrialBalanceDetailOut{
			AccountNumber:  row[0],
			AccountName:    row[1],
			OpeningBalance: opening,
			DebitMovement:  debit,
			CreditMovement: credit,
			ClosingBalance: closing,
		}

		tbd = append(tbd, item)
	}

	return
}

func (as *accounting) calculateTrialBalanceSummary(data []models.TrialBalanceDetailOut, subCategory *models.SubCategory) (summary models.TrialBalanceBySubCategoryOut) {
	var totalOpeningBalance, totalDebitMovement, totalCreditMovement, totalClosingBalance decimal.Decimal
	for _, v := range data {
		totalOpeningBalance = totalOpeningBalance.Add(v.OpeningBalance)
		totalDebitMovement = totalDebitMovement.Add(v.DebitMovement)
		totalCreditMovement = totalCreditMovement.Add(v.CreditMovement)
		totalClosingBalance = totalClosingBalance.Add(v.ClosingBalance)
	}

	summary = models.TrialBalanceBySubCategoryOut{
		SubCategoryCode: subCategory.Code,
		SubCategoryName: subCategory.Name,
		OpeningBalance:  totalOpeningBalance,
		DebitMovement:   totalDebitMovement,
		CreditMovement:  totalCreditMovement,
		ClosingBalance:  totalClosingBalance,
	}

	return
}

func (as *accounting) SendToEmailGetTrialBalanceDetails(ctx context.Context, opts models.TrialBalanceDetailsFilterOptions) (err error) {
	defer func() {
		logService(ctx, err)
	}()
	entityData, err := as.srv.mySqlRepo.GetEntityRepository().GetByCode(ctx, opts.EntityCode)
	if err != nil {
		return
	}

	if entityData == nil {
		err = models.GetErrMap(models.ErrKeyEntityCodeNotFound)
		return
	}

	subCategoryData, err := as.srv.mySqlRepo.GetSubCategoryRepository().GetByCode(ctx, opts.SubCategoryCode)
	if err != nil {
		return
	}

	if subCategoryData == nil {
		err = models.GetErrMap(models.ErrKeySubCategoryCodeNotFound)
		return
	}

	result, err := as.srv.mySqlRepo.GetAccountingRepository().GetTrialBalanceDetails(ctx, opts)
	if err != nil {
		return
	}

	opts.EntityDesc = entityData.Description
	opts.SubCategoryName = subCategoryData.Name

	return as.trialBalanceDetailsProcessCSVAndSendEmail(ctx, models.SendEmailTrialBalanceAccountListRequest{
		Opts:   opts,
		Result: result,
	})
}

func (as *accounting) trialBalanceDetailsProcessCSVAndSendEmail(ctx context.Context, in models.SendEmailTrialBalanceAccountListRequest) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	var b *bytes.Buffer
	b = &bytes.Buffer{}
	as.srv.file.NewCSVWriter(b)

	header := []string{"Account Number", "Account Name", "Opening Balance (Rp)", "Debit Movement (Rp)", "Credit Movement (Rp)", "Closing Balance (Rp)"}
	if err = as.srv.file.CSVWriteHeader(ctx, header); err != nil {
		err = fmt.Errorf("failed to write header: %w", err)
		return
	}

	limit := file.EXCEL_MAX_ROWS
	part := 1
	subCategoryName := in.Opts.SubCategoryName
	startDate := in.Opts.StartDate.Format(atime.DateFormatYYYYMMDD)
	endDate := in.Opts.EndDate.Format(atime.DateFormatYYYYMMDD)
	var filepath []models.CloudStoragePayload
	for i, t := range in.Result {
		if (i+1)%limit == 0 { // check if limit is reached
			if err = as.srv.file.CSVProcessWrite(ctx); err != nil {
				return
			}

			fp, errWriteGCS := as.writeCSVTrialBalanceDetailsToGCS(ctx, startDate, endDate, subCategoryName, part, b)
			if errWriteGCS != nil {
				return errWriteGCS
			}

			filepath = append(filepath, fp)

			part++

			// create new csv
			b = &bytes.Buffer{}
			as.srv.file.NewCSVWriter(b)
			if err = as.srv.file.CSVWriteHeader(ctx, header); err != nil {
				err = fmt.Errorf("failed to write header: %w", err)
				return
			}
		}

		if err = as.srv.file.CSVWriteBody(ctx, []string{
			t.AccountNumber,
			t.AccountName,
			t.OpeningBalance.String(),
			t.DebitMovement.String(),
			t.CreditMovement.String(),
			t.ClosingBalance.String(),
		}); err != nil {
			err = fmt.Errorf("failed to write body: %w", err)
			return
		}
	}

	//write last csv
	if err = as.srv.file.CSVProcessWrite(ctx); err != nil {
		return
	}

	fp, errWriteGCS := as.writeCSVTrialBalanceDetailsToGCS(ctx, startDate, endDate, subCategoryName, part, b)
	if errWriteGCS != nil {
		return errWriteGCS
	}

	filepath = append(filepath, fp)

	//generate signed url
	var urls []string
	for i, v := range filepath {
		url, errSignedUrl := as.srv.cloudStorageRepo.GetSignedURL(v, as.srv.conf.CloudStorageConfig.TrialBalanceDetailURLDuration)
		if errSignedUrl != nil {
			return errSignedUrl
		}

		urls = append(urls, fmt.Sprintf("<a href=\"%s\">Part %d</a>", url, i+1))
	}

	requestEmail := dddnotification.RequestEmail{
		From:     "noreply@amartha.com",
		FromName: in.Opts.EntityDesc,
		To:       in.Opts.Email,
		Subject: fmt.Sprintf("Trial Balance Detail - %s - %s to %s",
			subCategoryName,
			startDate,
			endDate,
		),
		Subs: []interface{}{
			map[string]interface{}{
				"SUB_CATEGORY": subCategoryName,
				"START_DATE":   startDate,
				"END_DATE":     endDate,
				"URL":          strings.Join(urls, "<br>"),
			},
		},
		Template: as.srv.conf.DDDNotification.EmailTemplateTrialBalanceDetail,
	}
	if err = as.srv.dddNotificationClient.SendEmail(ctx, requestEmail); err != nil {
		return
	}

	return
}

func (as *accounting) writeCSVTrialBalanceDetailsToGCS(ctx context.Context, startDate, endDate, subCategoryName string, part int, b *bytes.Buffer) (fp models.CloudStoragePayload, err error) {
	defer func() {
		logService(ctx, err)
	}()

	fp = models.CloudStoragePayload{
		Filename: fmt.Sprintf("trialBalanceDetail-%s-%s-%s-%d.csv", startDate, endDate, subCategoryName, part),
		Path:     fmt.Sprintf("%s", models.TrialBalanceDetailDir),
	}

	r := as.srv.cloudStorageRepo.NewWriter(ctx, &fp)

	_, errWrite := r.Write(b.Bytes())
	if errWrite != nil {
		if err = r.Close(); err != nil {
			return
		}
		return fp, errWrite
	}

	if err = r.Close(); err != nil {
		return fp, err
	}

	return
}

func (as *accounting) GetTrialBalanceBySubCategoryCode(ctx context.Context, opts models.TrialBalanceFilterOptions) (out models.TrialBalanceBySubCategoryOut, err error) {
	defer func() {
		logService(ctx, err)
	}()
	out, err = as.srv.mySqlRepo.GetAccountingRepository().GetTrialBalanceSubCategory(ctx, opts)
	if err != nil {
		if errors.Is(err, models.ErrNoRows) {
			err = models.GetErrMap(models.ErrKeySubCategoryCodeNotFound)
		}
		return
	}

	return
}

// source data from gcs
func (as *accounting) SendEmailTrialBalanceDetails(ctx context.Context, opts models.TrialBalanceDetailsFilterOptions) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	entityData, err := as.srv.mySqlRepo.GetEntityRepository().GetByCode(ctx, opts.EntityCode)
	if err != nil {
		return
	}

	if entityData == nil {
		err = models.GetErrMap(models.ErrKeyEntityCodeNotFound)
		return
	}

	subCategoryData, err := as.srv.mySqlRepo.GetSubCategoryRepository().GetByCode(ctx, opts.SubCategoryCode)
	if err != nil {
		return
	}

	if subCategoryData == nil {
		err = models.GetErrMap(models.ErrKeySubCategoryCodeNotFound)
		return
	}

	path := fmt.Sprintf("%s/%s/%s/details/%s/%s", models.TrialBalanceDir, opts.Year, opts.EntityCode, opts.Month, opts.SubCategoryCode)

	fp := models.CloudStoragePayload{
		Path: path,
	}

	filenames, err := as.srv.cloudStorageRepo.ListFiles(ctx, &fp)
	if err != nil {
		return err
	}

	var urls []string
	for i, v := range filenames {
		//skip file total rows
		if strings.Contains(v, "total_rows") {
			continue
		}

		fp.Filename = v
		url, errSignedUrl := as.srv.cloudStorageRepo.GetSignedURL(fp, as.srv.conf.CloudStorageConfig.TrialBalanceDetailURLDuration)
		if errSignedUrl != nil {
			return errSignedUrl
		}

		urls = append(urls, fmt.Sprintf("<a href=\"%s\">Part %d</a>", url, i+1))
	}

	startDate, endDate := atime.GenerateStartAndEndDateByMonthYear(opts.Year, opts.Month)
	requestEmail := dddnotification.RequestEmail{
		From:     "noreply@amartha.com",
		FromName: entityData.Description,
		To:       opts.Email,
		Subject: fmt.Sprintf("Trial Balance Detail - %s - %s to %s",
			subCategoryData.Name,
			startDate,
			endDate,
		),
		Subs: []interface{}{
			map[string]interface{}{
				"SUB_CATEGORY": subCategoryData.Name,
				"START_DATE":   startDate,
				"END_DATE":     endDate,
				"URL":          strings.Join(urls, "<br>"),
			},
		},
		Template: as.srv.conf.DDDNotification.EmailTemplateTrialBalanceDetail,
	}
	if err = as.srv.dddNotificationClient.SendEmail(ctx, requestEmail); err != nil {
		return
	}

	return
}

func (as *accounting) SendEmailTrialBalanceSummary(ctx context.Context, opts models.TrialBalanceFilterOptions) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	result, err := as.GetTrialBalance(ctx, opts)
	if err != nil {
		return
	}

	if err = as.trialBalanceSummaryCSVFromGCS(ctx, models.DownloadTrialBalanceRequest{
		Opts:   opts,
		Result: result,
	}); err != nil {
		return
	}

	return
}

func (as *accounting) trialBalanceSummaryCSVFromGCS(ctx context.Context, in models.DownloadTrialBalanceRequest) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	entityData, err := as.srv.mySqlRepo.GetEntityRepository().GetByCode(ctx, in.Opts.EntityCode)
	if err != nil {
		return
	}
	if entityData == nil {
		err = models.GetErrMap(models.ErrKeyEntityCodeNotFound)
		return
	}

	b := &bytes.Buffer{}
	as.srv.file.NewCSVWriter(b)

	// write title
	if err = as.srv.file.CSVWriteBody(ctx, []string{entityData.Description}); err != nil {
		err = fmt.Errorf("failed to write body: %w", err)
		return
	}

	if err = as.srv.file.CSVWriteBody(ctx, []string{"TRIAL BALANCE REPORT"}); err != nil {
		err = fmt.Errorf("failed to write body: %w", err)
		return
	}

	if err = as.srv.file.CSVWriteBody(ctx, []string{fmt.Sprintf("AS AT CLOSE OF %s", strings.ToUpper(in.Opts.Period.Format(atime.DateFormatYYYYMM)))}); err != nil {
		err = fmt.Errorf("failed to write body: %w", err)
		return
	}

	if err = as.srv.file.CSVWriteBody(ctx, []string{}); err != nil {
		err = fmt.Errorf("failed to write body: %w", err)
		return
	}

	// write contents
	header := []string{"LINE", "SUBCATEGORY", "OPENING BALANCE", "DEBIT MOVEMENTS", "CREDIT MOVEMENTS", "CLOSING BALANCE"}
	if err = as.srv.file.CSVWriteHeader(ctx, header); err != nil {
		err = fmt.Errorf("failed to write header: %w", err)
		return
	}

	res := in.Result
	for i, t := range res.COATypes {
		if err = as.srv.file.CSVWriteBody(ctx, []string{
			fmt.Sprint(i + 1),
			t.CoaTypeName,
		}); err != nil {
			err = fmt.Errorf("failed to write body: %w", err)
			return
		}

		for _, c := range t.Categories {
			// write categories
			if err = as.srv.file.CSVWriteBody(ctx, []string{
				fmt.Sprintf("%s.%s", c.CategoryCode, c.CategoryName),
				"",
				c.IDRFormatTotalOpeningBalance,
				c.IDRFormatTotalDebitMovement,
				c.IDRFormatTotalCreditMovement,
				c.IDRFormatTotalClosingBalance,
			}); err != nil {
				err = fmt.Errorf("failed to write body: %w", err)
				return
			}

			// write sub categories
			for _, sc := range c.SubCategories {
				if err = as.srv.file.CSVWriteBody(ctx, []string{
					"",
					fmt.Sprintf("%s.%s", sc.SubCategoryCode, sc.SubCategoryName),
					sc.IDRFormatOpeningBalance,
					sc.IDRFormatDebitMovement,
					sc.IDRFormatCreditMovement,
					sc.IDRFormatClosingBalance,
				}); err != nil {
					err = fmt.Errorf("failed to write body: %w", err)
					return
				}
			}
		}

		if err = as.srv.file.CSVWriteBody(ctx, []string{
			t.CoaTypeCode,
			fmt.Sprintf("Total %s", t.CoaTypeName),
			t.IDRFormatTotalOpeningBalance,
			t.IDRFormatTotalDebitMovement,
			t.IDRFormatTotalCreditMovement,
			t.IDRFormatTotalClosingBalance,
		}); err != nil {
			err = fmt.Errorf("failed to write body: %w", err)
			return
		}

		if err = as.srv.file.CSVWriteBody(ctx, []string{}); err != nil {
			err = fmt.Errorf("failed to write body: %w", err)
			return
		}
	}

	// write catch all
	if err = as.srv.file.CSVWriteBody(ctx, []string{
		"",
		"Catch All",
		res.CatchAll.IDRFormatCatchAllOpeningBalance,
		res.CatchAll.IDRFormatCatchAllDebitMovement,
		res.CatchAll.IDRFormatCatchAllCreditMovement,
		res.CatchAll.IDRFormatCatchAllClosingBalance,
	}); err != nil {
		err = fmt.Errorf("failed to write body: %w", err)
		return
	}

	if err = as.srv.file.CSVProcessWrite(ctx); err != nil {
		return
	}

	entityDesc := entityData.Description
	period := in.Opts.Period.Format(atime.DateFormatYYYYMM)

	zipFile, err := file.CompressCSVToZip(fmt.Sprintf("trialBalance-%s.csv", period), b.Bytes())
	if err != nil {
		return fmt.Errorf("failed to compress zip: %w", err)
	}

	requestEmail := dddnotification.RequestEmail{
		From:     "noreply@amartha.com",
		FromName: entityDesc,
		To:       in.Opts.Email,
		Subject: fmt.Sprintf("Trial Balance - %s - %s",
			entityDesc,
			period,
		),
		Attachments: []dddnotification.Attachment{
			{
				Type:    "application/zip",
				Name:    fmt.Sprintf("trialBalance-%s.zip", period),
				Content: base64.StdEncoding.EncodeToString(zipFile),
			},
		},
		Subs: []interface{}{
			map[string]interface{}{
				"ENTITY": entityDesc,
				"PERIOD": period,
			},
		},
		Template: as.srv.conf.DDDNotification.EmailTemplateTrialBalance,
	}
	if err = as.srv.dddNotificationClient.SendEmail(ctx, requestEmail); err != nil {
		return
	}

	return
}
