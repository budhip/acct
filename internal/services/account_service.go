package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/acuanclient"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/gocustomer"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/cache"

	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
)

type AccountService interface {
	Create(ctx context.Context, in models.CreateAccount) (out models.CreateAccount, err error)
	CreateLoanAccount(ctx context.Context, in models.CreateAccount) (err error)
	CreateLenderInstitutionAccount(ctx context.Context, in models.CreateAccount) (err error)
	CreateBranchPointAccount(ctx context.Context, in models.CreateAccount) (err error)
	Update(ctx context.Context, in models.UpdateAccount) (out models.UpdateAccount, err error)
	UpdateAccountEntity(ctx context.Context, in models.UpdateAccountEntity) (out models.UpdateAccountEntity, err error)
	GetOneByAccountNumber(ctx context.Context, accountNumber string) (out models.GetAccountOut, err error)
	GetOneByLegacyID(ctx context.Context, legacyID string) (out models.GetAccountOut, err error)
	GetAccountList(ctx context.Context, opts models.AccountFilterOptions) (accounts []models.GetAccountOut, total int, err error)
	DownloadCSVGetAccountList(ctx context.Context, in []models.GetAccountOut) (b *bytes.Buffer, filename string, err error)
	CheckAltIdIsExist(ctx context.Context, altId string) (err error)
	ProcessUploadAccounts(ctx context.Context, file *multipart.FileHeader) (err error)
	GetAllAccountNumbersByParam(ctx context.Context, in models.GetAllAccountNumbersByParamIn) (out []models.GetAllAccountNumbersByParamOut, err error)
	GetAllCategoryCodeSeq(ctx context.Context) ([]models.DoGetAllCategoryCodeSeqResponse, error)
	UpdateCategoryCodeSeq(ctx context.Context, in models.DoUpdateCategoryCodeSeqRequest) (err error)
	GetLenderAccountByCIHAccountNumber(ctx context.Context, accountNumber string) (out models.AccountLender, err error)
	CreateCategoryCodeSeq(ctx context.Context, in models.DoCreateCategoryCodeSeqRequest) (err error)
	ConsumerCreateAccountMigration(ctx context.Context, in models.CreateAccount) (err error)
	GetLoanAdvanceAccountByLoanAccount(ctx context.Context, loanAccountNumber string) (out models.AccountLoan, err error)
	UpdateAccountByCustomerData(ctx context.Context, in gocustomer.CustomerEventPayload) (err error)

	CreateLoanPartnerAccount(ctx context.Context, in models.CreateAccountLoanPartner) (out models.AccountsLoanPartner, err error)
}

type account service

var _ AccountService = (*account)(nil)

func (as *account) Update(ctx context.Context, in models.UpdateAccount) (out models.UpdateAccount, err error) {
	defer func() {
		logService(ctx, err)
	}()

	var (
		legacyId models.LegacyID
		keys     []string
	)

	act, err := as.srv.mySqlRepo.GetAccountRepository().GetOneByAccountNumber(ctx, in.AccountNumber)
	if err != nil {
		err = checkDatabaseError(err, models.ErrKeyAccountNumberNotFound)
		return
	}

	/*
	 1. If in.LegacyId is nil, replace value for in.LegacyId with act.LegacyId.
	 2. If in.LegacyId has the same value as act.LegacyId, you can choose which value to use.
	    So, for points 1 & 2, the legacy Id value in database will never change

	 3. If in.LegacyId does not have the same value as act.LegacyId:
	    first check the database whether this value has become another account's legacy ID or not.
	    if true, you are not allowed to update this account
	    if false, you are allowed to update this account
	*/
	if in.LegacyId == nil || reflect.DeepEqual(in.LegacyId, act.LegacyId) {
		in.LegacyId = act.LegacyId
		legacy, legacyErr := in.LegacyId.Value()
		if legacyErr != nil {
			err = models.GetErrMap(models.ErrKeyFailedMarshal, legacyErr.Error())
			return out, err
		}
		if err = json.Unmarshal(legacy.([]byte), &legacyId); err != nil {
			err = models.GetErrMap(models.ErrKeyFailedUnmarshal, err.Error())
			return out, err
		}
	} else {
		value, ok := (*in.LegacyId)["t24AccountNumber"]
		if ok {
			isExistLegacyId, errGetLegacyID := as.srv.mySqlRepo.GetAccountRepository().CheckLegacyIdIsExist(ctx, value.(string))
			if errGetLegacyID != nil {
				err = checkDatabaseError(errGetLegacyID, models.ErrKeyDatabaseError)
				return
			}
			if isExistLegacyId {
				xlog.Info(ctx, "[UPDATE-ACCOUNT]", xlog.String("status", "legacy id is exist, you are not allowed to update this account"), xlog.Any("message", in))
				err = models.GetErrMap(models.ErrKeyLegacyIdalreadyExists)
				return
			}
		}
	}

	in.Name = removeSpecialChars(in.Name)

	if err = as.srv.mySqlRepo.GetAccountRepository().Update(ctx, in); err != nil {
		err = checkDatabaseError(err)
		return
	}

	as.srv.acuanClient.PublishAccount(ctx, acuanclient.PublishAccountData{
		Type: models.TypeAccountUpdated,

		// updated fields
		Name:     in.Name,
		OwnerId:  in.OwnerID,
		LegacyId: in.LegacyId,
		AltId:    in.AltID,

		AccountNumber:   act.AccountNumber,
		ProductTypeName: act.ProductTypeName,
		CategoryCode:    act.CategoryCode,
		SubCategoryCode: act.SubCategoryCode,
		EntityCode:      act.EntityCode,
		Currency:        act.Currency,
		Status:          act.Status,
		Metadata:        act.Metadata,
	})

	keys = append(keys, pasAccountKey(in.AccountNumber))
	if legacyId.T24AccountNumber != "" && legacyId.T24AccountNumber != "0" {
		keys = append(keys, pasAccountLegacyKey(legacyId.T24AccountNumber))
	}
	as.deleteCaching(ctx, keys)

	return in, err
}

func (as *account) GetOneByAccountNumber(ctx context.Context, accountNumber string) (out models.GetAccountOut, err error) {
	var errs *multierror.Error

	defer func() {
		logService(ctx, err)
	}()

	key := pasAccountKey(accountNumber)
	val, err := as.srv.cacheRepo.Get(ctx, key)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	if val != "" {
		if err = json.Unmarshal([]byte(val), &out); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	if errs.ErrorOrNil() != nil {
		out, err = as.srv.mySqlRepo.GetAccountRepository().GetOneByAccountNumber(ctx, accountNumber)
		if err != nil {
			err = checkDatabaseError(err, models.ErrKeyAccountNumberNotFound)
			return
		}

		data, errMarshal := json.Marshal(out)
		if errMarshal != nil {
			err = errMarshal
			xlog.Warn(ctx, logMessageService, xlog.Err(errMarshal))
			return
		}
		if errSet := as.srv.cacheRepo.Set(ctx, key, data, as.srv.conf.CacheTTL.GetOneByAccountNumber); errSet != nil {
			err = errSet
			xlog.Warn(ctx, logMessageService, xlog.Err(errSet))
			return
		}
	}

	return out, err
}

func (as *account) GetOneByLegacyID(ctx context.Context, legacyID string) (out models.GetAccountOut, err error) {
	var errs *multierror.Error

	defer func() {
		logService(ctx, err)
	}()

	if legacyID == "0" {
		err = models.GetErrMap(models.ErrKeyLegacyIdNotFound)
		return
	}

	key := pasAccountLegacyKey(legacyID)
	val, err := as.srv.cacheRepo.Get(ctx, key)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	if val != "" {
		if err = json.Unmarshal([]byte(val), &out); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	if errs.ErrorOrNil() != nil {
		out, err = as.srv.mySqlRepo.GetAccountRepository().GetOneByLegacyID(ctx, legacyID)
		if err != nil {
			err = checkDatabaseError(err, models.ErrKeyLegacyIdNotFound)
			return
		}

		data, errMarshal := json.Marshal(out)
		if errMarshal != nil {
			err = errMarshal
			xlog.Warn(ctx, logMessageService, xlog.Err(errMarshal))
			return
		}
		if errSet := as.srv.cacheRepo.Set(ctx, key, data, as.srv.conf.CacheTTL.GetOneByAccountNumber); errSet != nil {
			err = errSet
			xlog.Warn(ctx, logMessageService, xlog.Err(errSet))
			return
		}
	}

	return out, err
}

func (as *account) GetAccountList(ctx context.Context, opts models.AccountFilterOptions) (accounts []models.GetAccountOut, total int, err error) {
	defer func() {
		logService(ctx, err)
	}()

	// if user input "0" when search by t24AccountNumber then return empty
	if opts.SearchBy == "t24AccountNumber" && opts.Search == "0" {
		return accounts, total, nil
	}

	opts.GuestMode = as.srv.flagger.IsEnabled(models.FlagGuestModePayment.String())
	acc := as.srv.mySqlRepo.GetAccountRepository()
	accounts, err = acc.GetAccountList(ctx, opts)
	if err != nil {
		return
	}

	if opts.ExcludeTotalEntries {
		return accounts, 0, nil
	}

	total, err = as.getAccountListCount(ctx, opts)
	if err != nil {
		return
	}

	return accounts, total, nil
}

func (as *account) DownloadCSVGetAccountList(ctx context.Context, in []models.GetAccountOut) (b *bytes.Buffer, filename string, err error) {
	defer func() {
		logService(ctx, err)
	}()

	b = &bytes.Buffer{}
	as.srv.file.NewCSVWriter(b)

	if err = as.srv.file.CSVWriteHeader(ctx,
		[]string{
			"Account Number",
			"Account Name",
			"Asset/Liabity",
			"Entity",
			"Product",
			"Category",
			"Sub Category",
			"Owner Id",
			"Alternate Id",
			"T24 Account Number",
			"Created At",
			"Updated At",
		},
	); err != nil {
		return
	}

	var errs *multierror.Error
	for _, t := range in {
		if err = as.srv.file.CSVWriteBody(ctx, []string{
			t.AccountNumber,
			t.AccountName,
			t.CoaTypeName,
			t.EntityName,
			t.ProductTypeName,
			t.CategoryName,
			t.SubCategoryName,
			t.OwnerID,
			t.AltID,
			t.T24AccountNumber,
			t.CreatedAt.Format(atime.DateFormatYYYYMMDDWithTime),
			t.UpdatedAt.Format(atime.DateFormatYYYYMMDDWithTime),
		}); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	if errs.ErrorOrNil() != nil {
		err = errs.ErrorOrNil()
		return
	}

	if err = as.srv.file.CSVProcessWrite(ctx); err != nil {
		return
	}

	filename = fmt.Sprintf("accounts-%s.csv", atime.Now().Format(atime.DateFormatYYYYMMDDWithTimeWithoutColon))

	return
}

func (as *account) CheckAltIdIsExist(ctx context.Context, altId string) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	isExist, err := as.srv.mySqlRepo.GetAccountRepository().CheckExistByParam(ctx, models.AccountFilterOptions{AltID: altId})
	if err != nil {
		return
	}
	if isExist {
		err = models.GetErrMap(models.ErrKeyAltIdIsExist)
	}

	return
}

func (as *account) ProcessUploadAccounts(ctx context.Context, file *multipart.FileHeader) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	path := fmt.Sprintf("./%s_%s", atime.Now().Format(atime.DateFormatYYYYMMDDWithTimeWithoutDash), file.Filename)
	src, err := file.Open()
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := as.srv.file.CreateFile(path)
	if err != nil {
		return
	}
	defer dst.Close()

	if err = as.srv.file.CopyFile(dst, src); err != nil {
		return err
	}

	defer as.srv.file.RemoveFile(dst.Name())
	fs, err := as.srv.file.OpenFile(dst.Name())
	if err != nil {
		return
	}

	records, err := as.srv.file.CSVReadAll(fs)
	if err != nil {
		return
	}

	for _, r := range records[1:] {
		account := models.CreateAccount{
			Name:            strings.TrimSpace(r[1]),
			OwnerID:         strings.TrimSpace(r[2]),
			ProductTypeCode: strings.TrimSpace(r[3]),
			EntityCode:      strings.TrimSpace(r[4]),
			CategoryCode:    strings.TrimSpace(r[5]),
			SubCategoryCode: strings.TrimSpace(r[6]),
			Currency:        strings.TrimSpace(r[7]),
			AltId:           strings.TrimSpace(r[8]),
			Metadata: &models.Metadata{
				"remarks":    "manual upload",
				"uploadDate": atime.Now().String(),
			},
		}

		legacyId := strings.TrimSpace(r[0])
		if legacyId != "" {
			account.LegacyId = &models.AccountLegacyId{
				"t24AccountNumber": legacyId,
			}
		}

		if err = as.srv.publisher.PublishSyncWithKeyAndLog(ctx,
			"publish account to pas_account_stream",
			as.srv.conf.Kafka.Publishers.PASAccountStream.Topic,
			uuid.New().String(),
			account,
		); err != nil {
			err = fmt.Errorf("error process data %s caused by - %v ", strings.Join(r, ","), err)
			return
		}
	}

	return
}

func (as *account) GetAllAccountNumbersByParam(ctx context.Context, in models.GetAllAccountNumbersByParamIn) (out []models.GetAllAccountNumbersByParamOut, err error) {
	defer func() {
		logService(ctx, err)
	}()

	out, err = cache.GetOrSet(as.srv.cacheRepo, models.GetOrSetCacheOpts[[]models.GetAllAccountNumbersByParamOut]{
		Ctx: ctx,
		Key: pasAccountsKey(in),
		TTL: as.srv.conf.CacheTTL.GetAccounts,
		Callback: func() ([]models.GetAllAccountNumbersByParamOut, error) {
			return as.srv.mySqlRepo.GetAccountRepository().GetAllAccountNumbersByParam(ctx, in)
		},
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (as *account) GetLenderAccountByCIHAccountNumber(ctx context.Context, accountNumber string) (out models.AccountLender, err error) {
	var errs *multierror.Error

	defer func() {
		logService(ctx, err)
	}()

	val, err := as.srv.cacheRepo.Get(ctx, accountNumber)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	if err = json.Unmarshal([]byte(val), &out); err != nil {
		errs = multierror.Append(errs, err)
	}

	if errs.ErrorOrNil() != nil {
		if len(accountNumber) < 15 {
			var pasAccountNumber string
			pasAccountNumber, err = as.srv.mySqlRepo.GetAccountRepository().GetAccountNumberByLegacyId(ctx, accountNumber)
			if err != nil {
				err = checkDatabaseError(err, models.ErrKeyAccountNumberNotFound)
				return
			}

			if pasAccountNumber != "" {
				accountNumber = pasAccountNumber
			}
		}

		out, err = as.srv.mySqlRepo.GetAccountRepository().GetLenderAccountByCIHAccountNumber(ctx, accountNumber)
		if err != nil {
			err = checkDatabaseError(err, models.ErrKeyAccountNumberNotFound)
			return
		}

		data, errMarshal := json.Marshal(out)
		if errMarshal != nil {
			xlog.Warn(ctx, logMessageService, xlog.Err(errMarshal))
			return
		}
		if errSet := as.srv.cacheRepo.Set(ctx, accountNumber, data, as.srv.conf.CacheTTL.GetLenderAccountByCIHAccountNumber); errSet != nil {
			xlog.Warn(ctx, logMessageService, xlog.Err(errMarshal))
			return
		}
	}

	return out, err

}

func (as *account) deleteCaching(ctx context.Context, keys []string) {
	if err := as.srv.cacheRepo.Del(ctx, keys...); err != nil {
		xlog.Warn(ctx, "[CACHING]", xlog.Any("keys", keys), xlog.Err(err))
	}
}

func (as *account) GetLoanAdvanceAccountByLoanAccount(ctx context.Context, loanAccountNumber string) (out models.AccountLoan, err error) {
	var errs *multierror.Error

	defer func() {
		logService(ctx, err)
	}()

	key := pasLoanAccountKey(loanAccountNumber)

	val, err := as.srv.cacheRepo.Get(ctx, key)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	if err = json.Unmarshal([]byte(val), &out); err != nil {
		errs = multierror.Append(errs, err)
	}

	if errs.ErrorOrNil() != nil {
		if len(loanAccountNumber) < 15 {
			var pasAccountNumber string
			pasAccountNumber, err = as.srv.mySqlRepo.GetAccountRepository().GetAccountNumberByLegacyId(ctx, loanAccountNumber)
			if err != nil {
				err = checkDatabaseError(err, models.ErrKeyAccountNumberNotFound)
				return
			}

			if pasAccountNumber != "" {
				loanAccountNumber = pasAccountNumber
			}
		}

		out, err = as.srv.mySqlRepo.GetAccountRepository().GetLoanAdvanceAccountByLoanAccount(ctx, loanAccountNumber)
		if err != nil {
			err = checkDatabaseError(err, models.ErrKeyAccountNumberNotFound)
			return
		}

		data, errMarshal := json.Marshal(out)
		if errMarshal != nil {
			xlog.Warn(ctx, logMessageService, xlog.Err(errMarshal))
			return
		}
		if errSet := as.srv.cacheRepo.Set(ctx, key, data, as.srv.conf.CacheTTL.GetLoanAdvanceAccountByLoanAccount); errSet != nil {
			xlog.Warn(ctx, logMessageService, xlog.Err(errMarshal))
			return
		}
	}

	return out, err
}

func (as *account) getAccountListCount(ctx context.Context, opts models.AccountFilterOptions) (total int, err error) {
	var errs *multierror.Error

	defer func() {
		logService(ctx, err)
	}()

	value := fmt.Sprintf("%s%s%s%s%s",
		opts.EntityCode,
		opts.CoaTypeCode,
		opts.CategoryCode,
		opts.SubCategoryCode,
		opts.ProductTypeCode,
	)
	if opts.Search != "" {
		value = fmt.Sprintf("%s%s%s",
			value,
			opts.SearchBy,
			opts.Search,
		)
	}
	key := fmt.Sprintf("pas_chart_of_accounts_total_%s", value)

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
		total, err = as.srv.mySqlRepo.GetAccountRepository().GetAccountListCount(ctx, opts)
		if err != nil {
			return
		}
		if errSet := as.srv.cacheRepo.Set(ctx, key, total, as.srv.conf.CacheTTL.GetAccountListCount); errSet != nil {
			xlog.Warn(ctx, logMessageService, xlog.Err(errSet))
		}
	}

	return total, err
}

func (as *account) UpdateAccountByCustomerData(ctx context.Context, in gocustomer.CustomerEventPayload) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	accounts, err := as.srv.mySqlRepo.GetAccountRepository().GetAccountList(ctx, models.AccountFilterOptions{
		Search:   in.CustomerNumber,
		SearchBy: "owner_id",
	})
	if err != nil {
		return
	}

	for _, act := range accounts {
		if act.AccountName != in.Name {
			err = as.srv.mySqlRepo.GetAccountRepository().Update(ctx, models.UpdateAccount{
				AccountNumber: act.AccountNumber,
				OwnerID:       act.OwnerID,
				AltID:         act.AltID,
				LegacyId:      act.LegacyId,
				Name:          in.Name,
			})
			if err != nil {
				err = checkDatabaseError(err)
				return
			}

			as.srv.acuanClient.PublishAccount(ctx, acuanclient.PublishAccountData{
				Type:            models.TypeAccountUpdated,
				Name:            in.Name,
				OwnerId:         act.OwnerID,
				LegacyId:        act.LegacyId,
				AltId:           act.AltID,
				AccountNumber:   act.AccountNumber,
				ProductTypeName: act.ProductTypeName,
				CategoryCode:    act.CategoryCode,
				SubCategoryCode: act.SubCategoryCode,
				EntityCode:      act.EntityCode,
				Currency:        act.Currency,
				Status:          act.Status,
				Metadata:        act.Metadata,
			})
		}
	}

	return nil
}

func (as *account) UpdateAccountEntity(ctx context.Context, in models.UpdateAccountEntity) (out models.UpdateAccountEntity, err error) {
	defer func() {
		logService(ctx, err)
	}()

	var (
		keys    []string
		entity  *models.Entity
		isExist bool
	)

	act, err := as.srv.mySqlRepo.GetAccountRepository().GetOneByAccountNumber(ctx, in.AccountNumber)
	if err != nil {
		err = checkDatabaseError(err, models.ErrKeyAccountNumberNotFound)
		return
	}

	entity, err = as.srv.mySqlRepo.GetEntityRepository().GetByCode(ctx, in.EntityCode)
	if err != nil {
		return
	}
	if entity == nil {
		err = models.GetErrMap(models.ErrKeyEntityCodeNotFound)
		return
	}

	isExist, err = as.srv.mySqlRepo.GetAccountingRepository().GetOneSplitAccount(ctx, in.AccountNumber)
	if err != nil {
		err = checkDatabaseError(err)
		return
	}
	if isExist {
		err = models.GetErrMap(models.ErrKeyJournalAccountIsExist)
		return
	}

	act.EntityCode = in.EntityCode

	if err = as.srv.mySqlRepo.GetAccountRepository().UpdateEntity(ctx, in); err != nil {
		err = checkDatabaseError(err)
		return
	}

	as.srv.acuanClient.PublishAccount(ctx, acuanclient.PublishAccountData{
		Type: models.TypeAccountUpdated,
		// updated fields
		EntityCode: act.EntityCode,

		AccountNumber:   act.AccountNumber,
		Name:            act.AccountName,
		ProductTypeName: act.ProductTypeName,
		OwnerId:         act.OwnerID,
		CategoryCode:    act.CategoryCode,
		SubCategoryCode: act.SubCategoryCode,
		Currency:        act.Currency,
		AltId:           act.AltID,
		Status:          act.Status,
		LegacyId:        act.LegacyId,
		Metadata:        act.Metadata,
	})

	keys = append(keys, pasAccountKey(in.AccountNumber))
	as.deleteCaching(ctx, keys)

	return in, err
}
