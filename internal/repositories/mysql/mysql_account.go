package mysql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
)

type AccountRepository interface {
	Create(ctx context.Context, in models.CreateAccount) (err error)
	CreateLenderAccount(ctx context.Context, in models.CreateLenderAccount) (err error)
	Update(ctx context.Context, in models.UpdateAccount) (err error)
	UpdateEntity(ctx context.Context, in models.UpdateAccountEntity) (err error)
	UpdateLegacyId(ctx context.Context, in models.UpdateLegacyId) (err error)
	UpdateAltId(ctx context.Context, in models.UpdateAltId) (err error)
	UpdateBySubCategory(ctx context.Context, in models.UpdateBySubCategory) (err error)
	GetOneByAccountNumber(ctx context.Context, accountNumber string) (out models.GetAccountOut, err error)
	GetOneByLegacyID(ctx context.Context, legacyID string) (out models.GetAccountOut, err error)
	GetAccountList(ctx context.Context, opts models.AccountFilterOptions) ([]models.GetAccountOut, error)
	GetAccountListCount(ctx context.Context, opts models.AccountFilterOptions) (total int, err error)
	CheckExistByParam(ctx context.Context, param models.AccountFilterOptions) (exist bool, err error)
	BulkInsertAccount(ctx context.Context, in []models.CreateAccount) (err error)
	BulkInsertAcctAccount(ctx context.Context, in []models.CreateAccount) (err error)
	GetAllAccountNumbersByParam(ctx context.Context, params models.GetAllAccountNumbersByParamIn) (out []models.GetAllAccountNumbersByParamOut, err error)
	GetLenderAccountByCIHAccountNumber(ctx context.Context, accountNumber string) (out models.AccountLender, err error)
	CreateLoanAccount(ctx context.Context, in models.CreateLoanAccount) (err error)
	CheckLegacyIdIsExist(ctx context.Context, legacyId string) (exist bool, err error)
	CheckAccountNumberIsExist(ctx context.Context, accountNumber string) (out *models.CheckAccountNumberIsExist, err error)
	GetAccountNumberByLegacyId(ctx context.Context, t24AccountNumber string) (accountNumber string, err error)
	GetLoanAdvanceAccountByLoanAccount(ctx context.Context, loanAccountNumber string) (out models.AccountLoan, err error)
	GetAllAccountNumber(ctx context.Context, entities []string, subCategories *[]models.SubCategory) <-chan models.StreamResult[models.GetAccountOut]
}

type accountRepository sqlRepo

var _ AccountRepository = (*accountRepository)(nil)

func (ar *accountRepository) Create(ctx context.Context, in models.CreateAccount) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	args, err := getFieldValues(in)
	if err != nil {
		return
	}

	res, err := db.ExecContext(ctx, queryAccountCreate, args...)
	if err != nil {
		return
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		return
	}

	if affectedRows == 0 {
		err = models.ErrNoRowsAffected
		return
	}

	return
}

func (ar *accountRepository) CreateLenderAccount(ctx context.Context, in models.CreateLenderAccount) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	args, err := getFieldValues(in)
	if err != nil {
		err = databaseError(err)
		return
	}

	res, err := db.ExecContext(ctx, queryCreateLenderAccount, args...)
	if err != nil {
		err = databaseError(err)
		return
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		err = databaseError(err)
		return
	}

	if affectedRows == 0 {
		err = databaseError(models.ErrNoRowsAffected)
		return
	}

	return
}

func (ar *accountRepository) Update(ctx context.Context, in models.UpdateAccount) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	query, args, err := buildUpdateAccountQuery(in)
	if err != nil {
		return
	}

	_, err = db.ExecContext(ctx, query, args...)
	return
}

func (ar *accountRepository) UpdateEntity(ctx context.Context, in models.UpdateAccountEntity) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)
	args, err := getFieldValues(in)
	if err != nil {
		return
	}
	_, err = db.ExecContext(ctx, queryUpdateAccountEntity, args...)

	return
}

func (ar *accountRepository) UpdateLegacyId(ctx context.Context, in models.UpdateLegacyId) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)
	args, err := getFieldValues(in)
	if err != nil {
		return
	}
	_, err = db.ExecContext(ctx, queryAccountUpdateLegacyId, args...)

	return
}

func (ar *accountRepository) UpdateAltId(ctx context.Context, in models.UpdateAltId) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)
	args, err := getFieldValues(in)
	if err != nil {
		return
	}
	_, err = db.ExecContext(ctx, queryAccountUpdateAltId, args...)

	return
}

func (ar *accountRepository) GetOneByAccountNumber(ctx context.Context, accountNumber string) (out models.GetAccountOut, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)
	err = db.QueryRowContext(ctx, queryAccountByNumber, accountNumber, accountNumber).Scan(
		&out.AccountNumber,
		&out.AccountName,
		&out.OwnerID,
		&out.CategoryCode,
		&out.CategoryName,
		&out.CoaTypeCode,
		&out.CoaTypeName,
		&out.SubCategoryCode,
		&out.SubCategoryName,
		&out.EntityCode,
		&out.EntityName,
		&out.ProductTypeCode,
		&out.ProductTypeName,
		&out.Currency,
		&out.Status,
		&out.AltID,
		&out.CreatedAt,
		&out.UpdatedAt,
		&out.LegacyId,
		&out.Metadata,
		&out.AccountType,
	)
	if err != nil {
		return
	}
	return out, err
}

func (ar *accountRepository) GetOneByLegacyID(ctx context.Context, legacyID string) (out models.GetAccountOut, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)
	err = db.QueryRowContext(ctx, queryAccountByLegacyID, legacyID).Scan(
		&out.AccountNumber,
		&out.AccountName,
		&out.OwnerID,
		&out.CategoryCode,
		&out.CategoryName,
		&out.CoaTypeCode,
		&out.CoaTypeName,
		&out.SubCategoryCode,
		&out.SubCategoryName,
		&out.EntityCode,
		&out.EntityName,
		&out.ProductTypeCode,
		&out.ProductTypeName,
		&out.Currency,
		&out.Status,
		&out.AltID,
		&out.CreatedAt,
		&out.UpdatedAt,
		&out.LegacyId,
		&out.Metadata,
		&out.AccountType,
	)
	if err != nil {
		return
	}
	return out, err
}

func (ar *accountRepository) GetAccountList(ctx context.Context, opts models.AccountFilterOptions) ([]models.GetAccountOut, error) {
	var (
		start = atime.Now()
		err   error
	)

	defer func() {
		logSQL(ctx, err, start)
	}()

	db := ar.r.extractTx(ctx)

	_, _, mapSubCategory, err := ar.r.GetAllCategorySubCategoryCOAType(ctx)
	if err != nil {
		return nil, err
	}

	query, args, err := buildAccountListQuery(opts)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		err = databaseError(err)
		return nil, err
	}
	defer rows.Close()

	var result []models.GetAccountOut
	for rows.Next() {
		var out = models.GetAccountOut{}
		var err = rows.Scan(
			&out.AccountNumber,
			&out.AccountName,
			&out.CategoryCode,
			&out.SubCategoryCode,
			&out.EntityCode,
			&out.EntityName,
			&out.ProductTypeCode,
			&out.ProductTypeName,
			&out.AltID,
			&out.OwnerID,
			&out.Status,
			&out.LegacyId,
			&out.CreatedAt,
			&out.UpdatedAt,
			&out.T24AccountNumber,
		)
		if err != nil {
			return nil, err
		}
		subCategoryCode := out.SubCategoryCode
		out.CoaTypeCode = mapSubCategory[subCategoryCode].CoaTypeCode
		out.CoaTypeName = mapSubCategory[subCategoryCode].CoaTypeName
		out.CategoryName = mapSubCategory[subCategoryCode].CategoryName
		out.SubCategoryName = mapSubCategory[subCategoryCode].SubCategoryName
		result = append(result, out)
	}
	if rows.Err() != nil {
		err = databaseError(rows.Err())
		return result, err
	}

	return result, nil
}

func (ar *accountRepository) GetAccountListCount(ctx context.Context, opts models.AccountFilterOptions) (total int, err error) {
	var (
		start = atime.Now()
	)
	defer func() {
		logSQL(ctx, err, start)
	}()

	db := ar.r.extractTx(ctx)

	query, args, err := buildCountAccountListQuery(opts)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return total, err
	}

	if err = db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		err = databaseError(err)
		return
	}

	return
}

func (ar *accountRepository) CheckExistByParam(ctx context.Context, param models.AccountFilterOptions) (exist bool, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	query, args, err := buildQueryCheckExistByParam(param)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return
	}

	var id string
	if err = db.QueryRowContext(ctx, query, args...).Scan(
		&id,
	); err != nil {
		if errors.Is(err, models.ErrNoRows) {
			err = nil
			return
		}
		err = databaseError(err)
		return
	}
	exist = true

	return
}

func (ar *accountRepository) BulkInsertAccount(ctx context.Context, in []models.CreateAccount) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	valueStrings := []string{}
	valueArgs := []interface{}{}
	for _, req := range in {
		valueStrings = append(valueStrings, "(?, ?)")
		valueArgs = append(valueArgs, req.AccountNumber)
		valueArgs = append(valueArgs, req.AccountNumber)
	}

	query := fmt.Sprintf(queryBulkInsertAccount, strings.Join(valueStrings, ","))
	res, err := db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		err = databaseError(err)
		return
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		err = databaseError(err)
		return
	}

	if affectedRows == 0 {
		err = databaseError(models.ErrNoRowsAffected)
		return
	}

	return
}

func (ar *accountRepository) BulkInsertAcctAccount(ctx context.Context, in []models.CreateAccount) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	valueStrings := []string{}
	valueArgs := []interface{}{}
	for _, req := range in {
		valueStrings = append(valueStrings, "(NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, '{}'), NULLIF(?, '{}'), CURRENT_TIMESTAMP(6), CURRENT_TIMESTAMP(6))")
		valueArgs = append(valueArgs, req.AccountNumber)
		valueArgs = append(valueArgs, req.OwnerID)
		valueArgs = append(valueArgs, req.AccountType)
		valueArgs = append(valueArgs, req.ProductTypeCode)
		valueArgs = append(valueArgs, req.EntityCode)
		valueArgs = append(valueArgs, req.CategoryCode)
		valueArgs = append(valueArgs, req.SubCategoryCode)
		valueArgs = append(valueArgs, req.Currency)
		valueArgs = append(valueArgs, req.Status)
		valueArgs = append(valueArgs, req.Name)
		valueArgs = append(valueArgs, req.AltId)
		valueArgs = append(valueArgs, req.LegacyId)
		valueArgs = append(valueArgs, req.Metadata)
	}

	query := fmt.Sprintf(queryBulkInsertAcctAccount, strings.Join(valueStrings, ","))
	res, err := db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		err = databaseError(err)
		return
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		err = databaseError(err)
		return
	}

	if affectedRows == 0 {
		err = databaseError(models.ErrNoRowsAffected)
		return
	}

	return
}

func (ar *accountRepository) GetAllAccountNumbersByParam(ctx context.Context, params models.GetAllAccountNumbersByParamIn) (out []models.GetAllAccountNumbersByParamOut, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	query, args, err := buildQueryGetAllAccountNumbersByParam(params)
	if err != nil {
		err = fmt.Errorf("failed to build query: %w", err)
		return
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		err = databaseError(err)
		return nil, err
	}
	defer rows.Close()

	var result []models.GetAllAccountNumbersByParamOut
	for rows.Next() {
		var value models.GetAllAccountNumbersByParamOut
		var err = rows.Scan(
			&value.OwnerId,
			&value.AccountNumber,
			&value.AltId,
			&value.Name,
			&value.AccountType,
			&value.EntityCode,
			&value.ProductTypeCode,
			&value.CategoryCode,
			&value.SubCategoryCode,
			&value.Currency,
			&value.Status,
			&value.LegacyId,
			&value.Metadata,
			&value.CreatedAt,
		)
		if err != nil {
			err = databaseError(err)
			return nil, err
		}
		result = append(result, value)
	}
	if rows.Err() != nil {
		err = databaseError(rows.Err())
		return nil, err
	}

	return result, nil
}

func (ar *accountRepository) GetLenderAccountByCIHAccountNumber(ctx context.Context, accountNumber string) (out models.AccountLender, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	err = db.QueryRowContext(ctx, queryGetLenderAccountByCIHAccountNumber, accountNumber).Scan(
		&out.CIHAccountNumber,
		&out.InvestedAccountNumber,
		&out.ReceivablesAccountNumber,
	)
	if err != nil {
		return
	}

	return out, err
}

func (ar *accountRepository) CreateLoanAccount(ctx context.Context, in models.CreateLoanAccount) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	args, err := getFieldValues(in)
	if err != nil {
		err = databaseError(err)
		return
	}

	res, err := db.ExecContext(ctx, queryCreateLoanAccount, args...)
	if err != nil {
		err = databaseError(err)
		return
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		err = databaseError(err)
		return
	}

	if affectedRows == 0 {
		err = databaseError(models.ErrNoRowsAffected)
		return
	}

	return
}

func (ar *accountRepository) CheckLegacyIdIsExist(ctx context.Context, legacyId string) (exist bool, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	var id string
	if err = db.QueryRowContext(ctx, queryCheckLegacyId, legacyId).Scan(
		&id,
	); err != nil {
		if errors.Is(err, models.ErrNoRows) {
			err = nil
			return
		}
		err = databaseError(err)
		return
	}
	exist = true

	return
}

func (ar *accountRepository) CheckAccountNumberIsExist(ctx context.Context, accountNumber string) (out *models.CheckAccountNumberIsExist, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	var val models.CheckAccountNumberIsExist
	if err = db.QueryRowContext(ctx, queryCheckAccountNumber, accountNumber).Scan(
		&val.AccountNumber,
		&val.EntityCode,
	); err != nil {
		if errors.Is(err, models.ErrNoRows) {
			err = nil
			return nil, nil
		}
		err = databaseError(err)
		return
	}

	return &val, nil
}

func (ar *accountRepository) GetAccountNumberByLegacyId(ctx context.Context, t24accountNumber string) (accountNumber string, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	if err = db.QueryRowContext(ctx, queryGetAccountNumberByLegacyId, t24accountNumber).Scan(
		&accountNumber,
	); err != nil {
		if errors.Is(err, models.ErrNoRows) {
			err = nil
			return
		}
		err = databaseError(err)
		return
	}

	return accountNumber, nil
}

func (ar *accountRepository) UpdateBySubCategory(ctx context.Context, in models.UpdateBySubCategory) (err error) {
	defer func() {
		logSQL(ctx, err)
	}()
	query := queryAccountUpdateBySubCategory
	args := []interface{}{}
	if in.ProductTypeCode != nil {
		query += ` product_type_code = ?, `
		args = append(args, *in.ProductTypeCode)
	}
	if in.Currency != nil {
		query += ` currency = ?, `
		args = append(args, *in.Currency)
	}

	query += queryAccountUpdateBySubCategoryWhere
	args = append(args, in.Code)

	db := ar.r.extractTx(ctx)
	_, err = db.ExecContext(ctx, query, args...)

	return
}

func (ar *accountRepository) GetLoanAdvanceAccountByLoanAccount(ctx context.Context, loanAccountNumber string) (out models.AccountLoan, err error) {
	defer func() {
		logSQL(ctx, err)
	}()

	db := ar.r.extractTx(ctx)

	err = db.QueryRowContext(ctx, queryGetLoanAdvanceAccountByLoanAccount, loanAccountNumber).Scan(
		&out.LoanAccountNumber,
		&out.LoanAdvancePaymentAccountNumber,
	)
	if err != nil {
		return
	}

	return out, err
}

func (ar *accountRepository) GetAllAccountNumber(ctx context.Context, entities []string, subCategories *[]models.SubCategory) <-chan models.StreamResult[models.GetAccountOut] {
	db := ar.r.extractTx(ctx)
	ch := make(chan models.StreamResult[models.GetAccountOut], 64)
	wg := new(sync.WaitGroup)
	// sem := make(chan struct{}, 128) // semaphore

	for _, subCategory := range *subCategories {
		wg.Add(1)

		go func(wg *sync.WaitGroup, subCategoryCode string) {
			defer wg.Done()

			entityPlaceholders, entityArgs := buildInClause(entities)
			query := fmt.Sprintf(queryGetAllAccountNumber, entityPlaceholders)
			args := append([]interface{}{subCategoryCode}, entityArgs...)
			rows, err := db.QueryContext(ctx, query, args...)
			if err != nil {
				ch <- models.StreamResult[models.GetAccountOut]{
					Err: fmt.Errorf("query error for  subCategoryCode %s: %w", subCategoryCode, err),
				}
				return
			}
			defer rows.Close()

			for rows.Next() {
				select {
				case <-ctx.Done():
					return
				default:
					var v models.GetAccountOut
					if err := rows.Scan(
						&v.AccountNumber,
						&v.EntityCode,
						&v.CategoryCode,
						&v.SubCategoryCode,
					); err != nil {
						ch <- models.StreamResult[models.GetAccountOut]{
							Err: fmt.Errorf("scan error for subCategoryCode %s: %w", subCategoryCode, err),
						}
						return
					}
					ch <- models.StreamResult[models.GetAccountOut]{Data: v}
				}
			}

			if err := rows.Err(); err != nil {
				ch <- models.StreamResult[models.GetAccountOut]{
					Err: fmt.Errorf("rows error for subCategoryCode %s: %w", subCategoryCode, err),
				}
			}
		}(wg, subCategory.Code)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}
