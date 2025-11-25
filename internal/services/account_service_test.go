package services_test

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"os"
	"strings"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/atime"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/gocustomer"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func createMultipartFormData(t *testing.T) *multipart.FileHeader {
	const (
		fileaContents = "This is a test file."
		filebContents = "Another test file."
		textaValue    = "foo"
		textbValue    = "bar"
		boundary      = `MyBoundary`
	)

	const message = `
--MyBoundary
Content-Disposition: form-data; name="filea"; filename="filea.txt"
Content-Type: text/plain

` + fileaContents + `
--MyBoundary
Content-Disposition: form-data; name="fileb"; filename="fileb.txt"
Content-Type: text/plain

` + filebContents + `
--MyBoundary
Content-Disposition: form-data; name="texta"

` + textaValue + `
--MyBoundary
Content-Disposition: form-data; name="textb"

` + textbValue + `
--MyBoundary--
`
	b := strings.NewReader(strings.ReplaceAll(message, "\n", "\r\n"))
	r := multipart.NewReader(b, `MyBoundary`)
	f, err := r.ReadForm(25)
	if err != nil {
		t.Error("ReadForm:", err)
	}
	defer f.RemoveAll()
	if g, e := f.Value["texta"][0], textaValue; g != e {
		t.Errorf("texta value = %q, want %q", g, e)
	}

	return f.File["filea"][0]
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		pwd, _ := os.Getwd()
		fmt.Println("PWD: ", pwd)
	}
	return r
}

func Test_account_Update(t *testing.T) {
	testHelper := serviceTestHelper(t)
	type (
		args struct {
			ctx context.Context
			req models.UpdateAccount
		}
		mockData struct{}
	)
	tests := []struct {
		name     string
		args     args
		mockData mockData
		doMock   func(args args, mockData mockData)
		wantErr  bool
	}{
		{
			name: "success case - update account",
			args: args{
				ctx: context.Background(),
				req: models.UpdateAccount{
					Name:          "Lender Yang Baik",
					OwnerID:       "12345",
					AccountNumber: "22200100000001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockAccRepository.EXPECT().Update(args.ctx, args.req).Return(nil)
				testHelper.mockAcuanClient.EXPECT().PublishAccount(args.ctx, gomock.Any())
				testHelper.mockCacheRepository.EXPECT().Del(args.ctx, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - database error get account",
			args: args{
				ctx: context.Background(),
				req: models.UpdateAccount{
					Name:          "Lender Yang Baik",
					OwnerID:       "12345",
					AltID:         "534534534555353523523423423",
					AccountNumber: "22200100000001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - database error update account",
			args: args{
				ctx: context.Background(),
				req: models.UpdateAccount{
					Name:          "Lender Yang Baik",
					OwnerID:       "12345",
					AltID:         "534534534555353523523423423",
					AccountNumber: "22200100000001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockAccRepository.EXPECT().Update(args.ctx, args.req).Return(models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.mockData)
			}
			_, err := testHelper.accountService.Update(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_account_UpdateEntity(t *testing.T) {
	testHelper := serviceTestHelper(t)
	mockReq := models.UpdateAccountEntity{
		AccountNumber: "22200100000001",
		EntityCode:    "001",
	}
	type (
		args struct {
			ctx context.Context
			req models.UpdateAccountEntity
		}
		mockData struct{}
	)
	tests := []struct {
		name     string
		args     args
		mockData mockData
		doMock   func(args args, mockData mockData)
		wantErr  bool
	}{
		{
			name: "success case - update account entity",
			args: args{
				ctx: context.Background(),
				req: mockReq,
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{
					AccountNumber: args.req.AccountNumber,
					LegacyId: &models.AccountLegacyId{
						"t24AccountNumber": "111000035909",
					},
				}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetOneSplitAccount(args.ctx, args.req.AccountNumber).Return(false, nil)
				testHelper.mockAccRepository.EXPECT().UpdateEntity(args.ctx, args.req).Return(nil)
				testHelper.mockAcuanClient.EXPECT().PublishAccount(args.ctx, gomock.Any())
				testHelper.mockCacheRepository.EXPECT().Del(args.ctx, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - account number not found",
			args: args{
				ctx: context.Background(),
				req: mockReq,
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, models.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "error case - entity code not found",
			args: args{
				ctx: context.Background(),
				req: mockReq,
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error get entity",
			args: args{
				ctx: context.Background(),
				req: mockReq,
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(nil, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - account has a transaction",
			args: args{
				ctx: context.Background(),
				req: mockReq,
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetOneSplitAccount(args.ctx, args.req.AccountNumber).Return(true, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error get split account",
			args: args{
				ctx: context.Background(),
				req: mockReq,
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetOneSplitAccount(args.ctx, args.req.AccountNumber).Return(false, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - database error update account",
			args: args{
				ctx: context.Background(),
				req: mockReq,
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(&models.Entity{}, nil)
				testHelper.mockAcctRepository.EXPECT().GetOneSplitAccount(args.ctx, args.req.AccountNumber).Return(false, nil)
				testHelper.mockAccRepository.EXPECT().UpdateEntity(args.ctx, args.req).Return(models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.mockData)
			}
			_, err := testHelper.accountService.UpdateAccountEntity(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
func Test_account_GetOneByAccountNumber(t *testing.T) {
	testHelper := serviceTestHelper(t)

	account := models.GetAccountOut{
		AccountNumber: "22200100000001",
	}
	val, _ := json.Marshal(account)

	key := fmt.Sprintf("%s_%s", "pas_account_key", account.AccountNumber)
	type args struct {
		ctx           context.Context
		accountNumber string
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "success case - get from cache",
			args: args{
				ctx:           context.Background(),
				accountNumber: account.AccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, key).Return(string(val), nil)
			},
			wantErr: false,
		},
		{
			name: "error case - database error",
			args: args{
				ctx:           context.Background(),
				accountNumber: account.AccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, key).Return("", models.GetErrMap(models.ErrKeyAccountNumberNotFound))
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.accountNumber).Return(models.GetAccountOut{}, nil)
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, key, gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - failed set to cache",
			args: args{
				ctx:           context.Background(),
				accountNumber: account.AccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, key).Return("", models.GetErrMap(models.ErrKeyAccountNumberNotFound))
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.accountNumber).Return(models.GetAccountOut{}, models.GetErrMap(models.ErrKeyAccountNumberNotFound))
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, key, gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "success case - get from database",
			args: args{
				ctx:           context.Background(),
				accountNumber: account.AccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, key).Return("", nil)
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.accountNumber).Return(account, nil)
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, key, account, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}
			_, err := testHelper.accountService.GetOneByAccountNumber(tt.args.ctx, tt.args.accountNumber)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_account_GetOneByLegacyID(t *testing.T) {
	testHelper := serviceTestHelper(t)

	account := models.GetAccountOut{
		AccountNumber: "22200100000001",
	}
	val, _ := json.Marshal(account)

	key := fmt.Sprintf("%s_%s", "pas_account_legacy_key", account.AccountNumber)
	type args struct {
		ctx      context.Context
		legacyId string
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "success case - get from cache",
			args: args{
				ctx:      context.Background(),
				legacyId: account.AccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, key).Return(string(val), nil)
			},
			wantErr: false,
		},
		{
			name: "error case - legacy id not found",
			args: args{
				ctx:      context.Background(),
				legacyId: "0",
			},
			wantErr: true,
		},
		{
			name: "error case - database error",
			args: args{
				ctx:      context.Background(),
				legacyId: account.AccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, key).Return("", models.GetErrMap(models.ErrKeyAccountNumberNotFound))
				testHelper.mockAccRepository.EXPECT().GetOneByLegacyID(args.ctx, args.legacyId).Return(models.GetAccountOut{}, nil)
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, key, gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - failed set to cache",
			args: args{
				ctx:      context.Background(),
				legacyId: account.AccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, key).Return("", models.GetErrMap(models.ErrKeyAccountNumberNotFound))
				testHelper.mockAccRepository.EXPECT().GetOneByLegacyID(args.ctx, args.legacyId).Return(models.GetAccountOut{}, models.GetErrMap(models.ErrKeyAccountNumberNotFound))
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, key, gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "success case - get from database",
			args: args{
				ctx:      context.Background(),
				legacyId: account.AccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, key).Return("", nil)
				testHelper.mockAccRepository.EXPECT().GetOneByLegacyID(args.ctx, args.legacyId).Return(account, nil)
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, key, account, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}
			_, err := testHelper.accountService.GetOneByLegacyID(tt.args.ctx, tt.args.legacyId)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_account_GetAccountList(t *testing.T) {
	testHelper := serviceTestHelper(t)
	mockResponse := models.GetAccountOut{
		AccountNumber:   "11200100000001",
		AccountName:     "Kas BP Love You",
		CoaTypeCode:     "AST",
		CoaTypeName:     "Liability",
		CategoryCode:    "112",
		CategoryName:    "KAS BP",
		SubCategoryCode: "11201",
		SubCategoryName: "Kas BP",
		EntityCode:      "001",
		EntityName:      "AMF",
		ProductTypeCode: "1001",
		ProductTypeName: "Group Loan",
		AltID:           "5345345345553535235234234232",
		OwnerID:         "123456789",
		Status:          "active",
		CreatedAt:       atime.Now(),
		UpdatedAt:       atime.Now(),
	}
	type args struct {
		ctx context.Context
		req models.AccountFilterOptions
	}
	type mockData struct {
	}
	tests := []struct {
		name     string
		args     args
		mockData mockData
		doMock   func(args args, mockData mockData)
		wantErr  bool
	}{
		{
			name: "success case",
			args: args{
				ctx: context.Background(),
				req: models.AccountFilterOptions{
					CoaTypeCode:     "AST",
					EntityCode:      "001",
					CategoryCode:    "112",
					SubCategoryCode: "11201",
					ProductTypeCode: "1001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGuestModePayment.String()).Return(false)
				testHelper.mockAccRepository.EXPECT().GetAccountList(args.ctx, args.req).Return([]models.GetAccountOut{mockResponse}, nil)
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, gomock.Any()).Return("", assert.AnError)
				testHelper.mockAccRepository.EXPECT().GetAccountListCount(args.ctx, args.req).Return(1, nil)
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, gomock.Any(), 1, gomock.Any()).Return(assert.AnError)
			},
			wantErr: false,
		},
		{
			name: "success case - cache",
			args: args{
				ctx: context.Background(),
				req: models.AccountFilterOptions{
					CoaTypeCode:     "AST",
					EntityCode:      "001",
					CategoryCode:    "112",
					SubCategoryCode: "11201",
					ProductTypeCode: "1001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGuestModePayment.String()).Return(false)
				testHelper.mockAccRepository.EXPECT().GetAccountList(args.ctx, args.req).Return([]models.GetAccountOut{mockResponse}, nil)
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, gomock.Any()).Return("1", nil)
			},
			wantErr: false,
		},
		{
			name: "success case - without count total entries",
			args: args{
				ctx: context.Background(),
				req: models.AccountFilterOptions{
					CoaTypeCode:         "AST",
					EntityCode:          "001",
					CategoryCode:        "112",
					SubCategoryCode:     "11201",
					ProductTypeCode:     "1001",
					ExcludeTotalEntries: true,
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGuestModePayment.String()).Return(false)
				testHelper.mockAccRepository.EXPECT().GetAccountList(args.ctx, args.req).Return([]models.GetAccountOut{mockResponse}, nil)
			},
			wantErr: false,
		},
		{
			name: "error case - database error GetAccountList",
			args: args{
				ctx: context.Background(),
				req: models.AccountFilterOptions{
					CoaTypeCode:     "AST",
					EntityCode:      "001",
					CategoryCode:    "112",
					SubCategoryCode: "11201",
					ProductTypeCode: "1001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGuestModePayment.String()).Return(false)
				testHelper.mockAccRepository.EXPECT().GetAccountList(args.ctx, args.req).Return([]models.GetAccountOut{}, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - database error GetAccountListCount",
			args: args{
				ctx: context.Background(),
				req: models.AccountFilterOptions{
					CoaTypeCode:     "AST",
					EntityCode:      "001",
					CategoryCode:    "112",
					SubCategoryCode: "11201",
					ProductTypeCode: "1001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockFlag.EXPECT().IsEnabled(models.FlagGuestModePayment.String()).Return(false)
				testHelper.mockAccRepository.EXPECT().GetAccountList(args.ctx, args.req).Return([]models.GetAccountOut{mockResponse}, nil)
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, gomock.Any()).Return("", assert.AnError)
				testHelper.mockAccRepository.EXPECT().GetAccountListCount(args.ctx, args.req).Return(0, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.mockData)
			}

			_, _, err := testHelper.accountService.GetAccountList(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_account_DownloadCSVGetAccountList(t *testing.T) {
	testHelper := serviceTestHelper(t)
	now := atime.Now()
	req := []models.GetAccountOut{
		{
			AccountNumber:   "131001000005084",
			AccountName:     "SHINJI TAKERU",
			CoaTypeCode:     "AST",
			CoaTypeName:     "Asset",
			CategoryCode:    "131",
			CategoryName:    "Borrower Outstanding",
			SubCategoryCode: "13101",
			SubCategoryName: "Borrower Outstanding - Normal",
			EntityCode:      "001",
			EntityName:      "AMF",
			ProductTypeName: "1001",
			ProductTypeCode: "Group Loan",
			AltID:           "1467823",
			OwnerID:         "5000033192",
			Status:          "active",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}
	type args struct {
		in []models.GetAccountOut
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{

		{
			name: "error case - create csv write header",
			args: args{
				in: req,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write body",
			args: args{
				in: req,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).AnyTimes()
				testHelper.mockFile.EXPECT().CSVWriteBody(gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - create csv write process",
			args: args{
				in: req,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().CSVWriteBody(gomock.Any(), gomock.Any()).AnyTimes()
				testHelper.mockFile.EXPECT().CSVProcessWrite(gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "success case",
			args: args{
				in: req,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().NewCSVWriter(gomock.Any())
				testHelper.mockFile.EXPECT().CSVWriteHeader(gomock.Any(), gomock.Any()).AnyTimes()
				testHelper.mockFile.EXPECT().CSVWriteBody(gomock.Any(), gomock.Any()).AnyTimes()
				testHelper.mockFile.EXPECT().CSVProcessWrite(gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}
			_, _, err := testHelper.accountService.DownloadCSVGetAccountList(context.TODO(), tt.args.in)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_account_CheckAltIdIsExist(t *testing.T) {
	testHelper := serviceTestHelper(t)

	type args struct {
		ctx context.Context
		req models.AccountFilterOptions
	}
	type mockData struct {
	}
	tests := []struct {
		name     string
		args     args
		mockData mockData
		doMock   func(args args, mockData mockData)
		wantErr  bool
	}{
		{
			name: "alt id not exist",
			args: args{
				ctx: context.Background(),
				req: models.AccountFilterOptions{AltID: "22200100000001"},
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().CheckExistByParam(args.ctx, args.req).Return(false, nil)
			},
			wantErr: false,
		},
		{
			name: "alt id is exist",
			args: args{
				ctx: context.Background(),
				req: models.AccountFilterOptions{AltID: "22200100000001"},
			},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().CheckExistByParam(args.ctx, args.req).Return(true, nil)
			},
			wantErr: true,
		},
		{
			name: "test error database",
			args: args{
				ctx: context.Background(),
				req: models.AccountFilterOptions{AltID: "22200100000001"}},
			mockData: mockData{},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().CheckExistByParam(args.ctx, args.req).Return(false, assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args, tt.mockData)
			}
			err := testHelper.accountService.CheckAltIdIsExist(tt.args.ctx, tt.args.req.AltID)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_account_ProcessUploadAccounts(t *testing.T) {
	testHelper := serviceTestHelper(t)

	f := createMultipartFormData(t)
	o := mustOpen("../../storages/upload_accounts.csv")
	csvResp := [][]string{
		{"legacy_id", "name", "owner_id", "product_type_code", "entity_code", "category_code", "sub_category_code", "currency", "alt_id"},
		{"11100100000004", "KAS Teller Point 1", "1234567", "1001", "001", "111", "11101", "IDR", ""},
	}

	type args struct {
		ctx  context.Context
		file *multipart.FileHeader
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "success case",
			args: args{
				ctx:  context.Background(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().RemoveFile(gomock.Any()).AnyTimes()
				testHelper.mockFile.EXPECT().OpenFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CSVReadAll(gomock.Any()).Return(csvResp, nil)

				testHelper.mockPublisher.EXPECT().PublishSyncWithKeyAndLog(
					args.ctx,
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - no such file or directory",
			args: args{
				ctx: context.Background(),
				file: &multipart.FileHeader{
					Filename: "../../storages/upload_accounts.csv",
				},
			},
			wantErr: true,
		},
		{
			name: "error case - create file",
			args: args{
				ctx:  context.Background(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - copy file",
			args: args{
				ctx:  context.Background(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - open file",
			args: args{
				ctx:  context.Background(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().RemoveFile(gomock.Any()).AnyTimes()
				testHelper.mockFile.EXPECT().OpenFile(gomock.Any()).Return(o, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - csv read",
			args: args{
				ctx:  context.Background(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().RemoveFile(gomock.Any()).AnyTimes()
				testHelper.mockFile.EXPECT().OpenFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CSVReadAll(gomock.Any()).Return([][]string{}, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "error case - failed publish account",
			args: args{
				ctx:  context.Background(),
				file: f,
			},
			doMock: func(args args) {
				testHelper.mockFile.EXPECT().CreateFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CopyFile(gomock.Any(), gomock.Any()).Return(nil)
				testHelper.mockFile.EXPECT().RemoveFile(gomock.Any()).AnyTimes()
				testHelper.mockFile.EXPECT().OpenFile(gomock.Any()).Return(o, nil)
				testHelper.mockFile.EXPECT().CSVReadAll(gomock.Any()).Return(csvResp, nil)

				testHelper.mockPublisher.EXPECT().PublishSyncWithKeyAndLog(
					args.ctx,
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(models.GetErrMap(models.ErrCodeInternalServerError))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}
			err := testHelper.accountService.ProcessUploadAccounts(tt.args.ctx, tt.args.file)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_account_GetAllAccountNumbersByParam(t *testing.T) {
	testHelper := serviceTestHelper(t)
	ctx := context.Background()
	res := []models.GetAllAccountNumbersByParamOut{
		{
			AccountNumber: "12345678910",
			AccountType:   "LENDER_INSTITUSI_NON_RDL",
		},
	}
	b, _ := json.Marshal(res)

	tests := []struct {
		name    string
		req     models.GetAllAccountNumbersByParamIn
		doMock  func(ctx context.Context, req models.GetAllAccountNumbersByParamIn)
		wantErr bool
	}{
		{
			name: "success - get all accounts by param ownerId",
			req: models.GetAllAccountNumbersByParamIn{
				OwnerId: "12345678910",
			},
			doMock: func(ctx context.Context, req models.GetAllAccountNumbersByParamIn) {
				testHelper.mockAccRepository.EXPECT().
					GetAllAccountNumbersByParam(ctx, models.GetAllAccountNumbersByParamIn{
						OwnerId: req.OwnerId,
					}).Return(res, nil)
				testHelper.mockCacheRepository.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(string(b), nil)
			},
			wantErr: false,
		},
		{
			name: "success - get all accounts by param accountNumbers",
			req: models.GetAllAccountNumbersByParamIn{
				AccountNumbers: "131001000030128",
			},
			doMock: func(ctx context.Context, req models.GetAllAccountNumbersByParamIn) {
				testHelper.mockCacheRepository.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(string(b), nil)
			},
			wantErr: false,
		},
		{
			name: "success - get all accounts by param altId, accountType",
			req: models.GetAllAccountNumbersByParamIn{
				AltId:       "123456",
				AccountType: "LENDER_INSTITUSI_NON_RDL",
			},
			doMock: func(ctx context.Context, req models.GetAllAccountNumbersByParamIn) {
				testHelper.mockCacheRepository.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(string(b), nil)
			},
			wantErr: false,
		},
		{
			name: "error case - cache error",
			req: models.GetAllAccountNumbersByParamIn{
				AccountNumbers: "131001000030128",
			},
			doMock: func(ctx context.Context, req models.GetAllAccountNumbersByParamIn) {
				testHelper.mockAccRepository.EXPECT().GetAllAccountNumbersByParam(ctx, models.GetAllAccountNumbersByParamIn{
					AccountNumbers: req.AccountNumbers,
				}).Return(res, nil)
				testHelper.mockCacheRepository.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return("0", nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error",
			req: models.GetAllAccountNumbersByParamIn{
				SubCategoryCode: "JanganDipaksain",
			},
			doMock: func(ctx context.Context, req models.GetAllAccountNumbersByParamIn) {
				testHelper.mockCacheRepository.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return("0", models.ErrRedisNil)
				testHelper.mockAccRepository.EXPECT().
					GetAllAccountNumbersByParam(ctx, models.GetAllAccountNumbersByParamIn{
						SubCategoryCode: req.SubCategoryCode,
					}).Return([]models.GetAllAccountNumbersByParamOut{}, assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(ctx, tt.req)
			}
			_, err := testHelper.accountService.GetAllAccountNumbersByParam(ctx, tt.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_account_GetInvestedAccountNumberByCIHAccountNumber(t *testing.T) {
	testHelper := serviceTestHelper(t)

	account := models.AccountLender{
		CIHAccountNumber:         "211001000000420",
		InvestedAccountNumber:    "212001000000330",
		ReceivablesAccountNumber: "",
	}
	val, _ := json.Marshal(account)
	CIHAccountNumber := "211001000381110"
	type args struct {
		ctx           context.Context
		accountNumber string
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "success case - get from database",
			args: args{
				ctx:           context.Background(),
				accountNumber: CIHAccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, args.accountNumber).Return("", nil)
				testHelper.mockAccRepository.EXPECT().GetLenderAccountByCIHAccountNumber(args.ctx, args.accountNumber).Return(models.AccountLender{}, nil)
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, args.accountNumber, gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success case - get from cache",
			args: args{
				ctx:           context.Background(),
				accountNumber: CIHAccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, args.accountNumber).Return(string(val), nil)
			},
			wantErr: false,
		},
		{
			name: "error case - failed get to cache",
			args: args{
				ctx:           context.Background(),
				accountNumber: CIHAccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, args.accountNumber).Return("", models.GetErrMap(models.ErrKeyFailedSetToCache))
				testHelper.mockAccRepository.EXPECT().GetLenderAccountByCIHAccountNumber(args.ctx, args.accountNumber).Return(models.AccountLender{}, nil)
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, args.accountNumber, gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - database error",
			args: args{
				ctx:           context.Background(),
				accountNumber: CIHAccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, args.accountNumber).Return("", nil)
				testHelper.mockAccRepository.EXPECT().GetLenderAccountByCIHAccountNumber(args.ctx, args.accountNumber).Return(models.AccountLender{}, models.GetErrMap(models.ErrKeyDatabaseError))
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, args.accountNumber, gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error case - failed set to cache",
			args: args{
				ctx:           context.Background(),
				accountNumber: CIHAccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, args.accountNumber).Return("", nil)
				testHelper.mockAccRepository.EXPECT().GetLenderAccountByCIHAccountNumber(args.ctx, args.accountNumber).Return(models.AccountLender{}, nil)
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, args.accountNumber, gomock.Any(), gomock.Any()).Return(models.GetErrMap(models.ErrKeyFailedSetToCache))
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}
			_, err := testHelper.accountService.GetLenderAccountByCIHAccountNumber(tt.args.ctx, tt.args.accountNumber)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_account_GetLoanAdvanceAccountByLoanAccount(t *testing.T) {
	testHelper := serviceTestHelper(t)
	account := models.AccountLoan{
		LoanAccountNumber:               "211001000000001",
		LoanAdvancePaymentAccountNumber: "211001000000002",
	}
	val, _ := json.Marshal(account)
	cacheKey := fmt.Sprintf("%s_%s", "pas_loan_account_key", account.LoanAccountNumber)

	type args struct {
		ctx               context.Context
		loanAccountNumber string
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "success case - get from database",
			args: args{
				ctx:               context.Background(),
				loanAccountNumber: account.LoanAccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, cacheKey).Return("", nil)
				testHelper.mockAccRepository.EXPECT().GetLoanAdvanceAccountByLoanAccount(args.ctx, args.loanAccountNumber).Return(models.AccountLoan{}, nil)
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, cacheKey, gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success case - get from cache",
			args: args{
				ctx:               context.Background(),
				loanAccountNumber: account.LoanAccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, cacheKey).Return(string(val), nil)
			},
			wantErr: false,
		},
		{
			name: "error case - failed get to cache",
			args: args{
				ctx:               context.Background(),
				loanAccountNumber: account.LoanAccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, cacheKey).Return("", models.GetErrMap(models.ErrKeyFailedSetToCache))
				testHelper.mockAccRepository.EXPECT().GetLoanAdvanceAccountByLoanAccount(args.ctx, args.loanAccountNumber).Return(models.AccountLoan{}, nil)
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, cacheKey, gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - database error",
			args: args{
				ctx:               context.Background(),
				loanAccountNumber: account.LoanAccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, cacheKey).Return("", nil)
				testHelper.mockAccRepository.EXPECT().GetLoanAdvanceAccountByLoanAccount(args.ctx, args.loanAccountNumber).Return(models.AccountLoan{}, models.GetErrMap(models.ErrKeyDatabaseError))
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, cacheKey, gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error case - failed set to cache",
			args: args{
				ctx:               context.Background(),
				loanAccountNumber: account.LoanAccountNumber,
			},
			doMock: func(args args) {
				testHelper.mockCacheRepository.EXPECT().Get(args.ctx, cacheKey).Return("", nil)
				testHelper.mockAccRepository.EXPECT().GetLoanAdvanceAccountByLoanAccount(args.ctx, args.loanAccountNumber).Return(models.AccountLoan{}, nil)
				testHelper.mockCacheRepository.EXPECT().Set(args.ctx, cacheKey, gomock.Any(), gomock.Any()).Return(models.GetErrMap(models.ErrKeyFailedSetToCache))
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}
			_, err := testHelper.accountService.GetLoanAdvanceAccountByLoanAccount(tt.args.ctx, tt.args.loanAccountNumber)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_account_UpdateAccountByCustomerData(t *testing.T) {
	testHelper := serviceTestHelper(t)

	defaultResultAccounts := []models.GetAccountOut{
		{
			AccountNumber: "123456",
			OwnerID:       "666",
			AltID:         "123",
			AccountName:   "Obi Wan Kenobi",
		},
	}

	type args struct {
		ctx context.Context
		in  gocustomer.CustomerEventPayload
	}
	tests := []struct {
		name    string
		args    args
		doMock  func(args args)
		wantErr bool
	}{
		{
			name: "success case - update from customer data",
			args: args{
				ctx: context.Background(),
				in: gocustomer.CustomerEventPayload{
					CustomerNumber: "666",
					Name:           "Sheev Palpatine",
				},
			},
			doMock: func(args args) {
				testHelper.mockAccRepository.EXPECT().
					GetAccountList(args.ctx, models.AccountFilterOptions{
						Search:   args.in.CustomerNumber,
						SearchBy: "owner_id",
					}).
					Return(defaultResultAccounts, nil)

				testHelper.mockAccRepository.EXPECT().
					Update(args.ctx, gomock.Any()).
					Return(nil)

				testHelper.mockAcuanClient.
					EXPECT().
					PublishAccount(args.ctx, gomock.Any())
			},
			wantErr: false,
		},
		{
			name: "success case - name already same with data",
			args: args{
				ctx: context.Background(),
				in: gocustomer.CustomerEventPayload{
					CustomerNumber: "666",
					Name:           "Obi Wan Kenobi",
				},
			},
			doMock: func(args args) {
				testHelper.mockAccRepository.EXPECT().
					GetAccountList(args.ctx, models.AccountFilterOptions{
						Search:   args.in.CustomerNumber,
						SearchBy: "owner_id",
					}).
					Return(defaultResultAccounts, nil)
			},
			wantErr: false,
		},
		{
			name: "failed - unable get account list",
			args: args{
				ctx: context.Background(),
				in: gocustomer.CustomerEventPayload{
					CustomerNumber: "666",
					Name:           "Obi Wan Kenobi",
				},
			},
			doMock: func(args args) {
				testHelper.mockAccRepository.EXPECT().
					GetAccountList(args.ctx, models.AccountFilterOptions{
						Search:   args.in.CustomerNumber,
						SearchBy: "owner_id",
					}).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "failed - unable update data",
			args: args{
				ctx: context.Background(),
				in: gocustomer.CustomerEventPayload{
					CustomerNumber: "666",
					Name:           "Sheev Palpatine",
				},
			},
			doMock: func(args args) {
				testHelper.mockAccRepository.EXPECT().
					GetAccountList(args.ctx, models.AccountFilterOptions{
						Search:   args.in.CustomerNumber,
						SearchBy: "owner_id",
					}).
					Return(defaultResultAccounts, nil)

				testHelper.mockAccRepository.EXPECT().
					Update(args.ctx, gomock.Any()).
					Return(assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(tt.args)
			}
			err := testHelper.accountService.UpdateAccountByCustomerData(tt.args.ctx, tt.args.in)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
