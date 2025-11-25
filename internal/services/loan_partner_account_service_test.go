package services_test

import (
	"context"
	"encoding/json"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestLoanPartnerService_Create(t *testing.T) {
	testHelper := serviceTestHelper(t)
	ctx := context.Background()
	req := models.LoanPartnerAccount{
		PartnerId:           "efishery",
		LoanKind:            "EFISHERY_LOAN",
		AccountNumber:       "22100100000001",
		AccountType:         "INTERNAL_ACCOUNTS_REVENUE_AMARTHA",
		EntityCode:          "001",
		LoanSubCategoryCode: "13101",
	}

	tests := []struct {
		name    string
		req     models.LoanPartnerAccount
		doMock  func(ctx context.Context, req models.LoanPartnerAccount)
		wantErr bool
	}{
		{
			name: "success case",
			req:  req,
			doMock: func(ctx context.Context, req models.LoanPartnerAccount) {
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, req.AccountNumber).
					Return(models.GetAccountOut{
						AccountNumber: req.AccountNumber,
						AccountType:   req.LoanSubCategoryCode,
						EntityCode:    req.EntityCode,
					}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						AccountNumber:       req.AccountNumber,
						LoanSubCategoryCode: req.LoanSubCategoryCode,
					}).Return([]models.LoanPartnerAccount{}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						PartnerId:           req.PartnerId,
						LoanKind:            req.LoanKind,
						AccountType:         req.AccountType,
						EntityCode:          req.EntityCode,
						LoanSubCategoryCode: req.LoanSubCategoryCode,
					}).Return([]models.LoanPartnerAccount{}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					Create(ctx, req).Return(nil)
				testHelper.mockCacheRepository.EXPECT().
					DeleteKeysWithPrefix(ctx, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - account number not found",
			req:  req,
			doMock: func(ctx context.Context, req models.LoanPartnerAccount) {
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, req.AccountNumber).
					Return(models.GetAccountOut{}, models.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "error case - database error get by param account number",
			req:  req,
			doMock: func(ctx context.Context, req models.LoanPartnerAccount) {
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, req.AccountNumber).
					Return(models.GetAccountOut{
						AccountNumber: req.AccountNumber,
						AccountType:   req.LoanSubCategoryCode,
						EntityCode:    req.EntityCode,
					}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						AccountNumber:       req.AccountNumber,
						LoanSubCategoryCode: req.LoanSubCategoryCode,
					}).
					Return([]models.LoanPartnerAccount{req}, models.GetErrMap(models.ErrKeyDatabaseError))
				testHelper.mockCacheRepository.EXPECT().
					DeleteKeysWithPrefix(ctx, gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error case - account is exist",
			req:  req,
			doMock: func(ctx context.Context, req models.LoanPartnerAccount) {
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, req.AccountNumber).
					Return(models.GetAccountOut{
						AccountNumber: req.AccountNumber,
						AccountType:   req.LoanSubCategoryCode,
						EntityCode:    req.EntityCode,
					}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						AccountNumber:       req.AccountNumber,
						LoanSubCategoryCode: req.LoanSubCategoryCode,
					}).
					Return([]models.LoanPartnerAccount{req}, nil)
				testHelper.mockCacheRepository.EXPECT().
					DeleteKeysWithPrefix(ctx, gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error case - data is exist",
			req:  req,
			doMock: func(ctx context.Context, req models.LoanPartnerAccount) {
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, req.AccountNumber).
					Return(models.GetAccountOut{
						AccountNumber: req.AccountNumber,
						AccountType:   req.LoanSubCategoryCode,
						EntityCode:    req.EntityCode,
					}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						AccountNumber:       req.AccountNumber,
						LoanSubCategoryCode: req.LoanSubCategoryCode,
					}).Return([]models.LoanPartnerAccount{}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						PartnerId:           req.PartnerId,
						LoanKind:            req.LoanKind,
						AccountType:         req.AccountType,
						EntityCode:          req.EntityCode,
						LoanSubCategoryCode: req.LoanSubCategoryCode,
					}).Return([]models.LoanPartnerAccount{req}, nil)
				testHelper.mockCacheRepository.EXPECT().
					DeleteKeysWithPrefix(ctx, gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error get by param",
			req:  req,
			doMock: func(ctx context.Context, req models.LoanPartnerAccount) {
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, req.AccountNumber).
					Return(models.GetAccountOut{
						AccountNumber: req.AccountNumber,
						AccountType:   req.LoanSubCategoryCode,
						EntityCode:    req.EntityCode,
					}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						AccountNumber:       req.AccountNumber,
						LoanSubCategoryCode: req.LoanSubCategoryCode,
					}).Return([]models.LoanPartnerAccount{}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						PartnerId:           req.PartnerId,
						LoanKind:            req.LoanKind,
						AccountType:         req.AccountType,
						EntityCode:          req.EntityCode,
						LoanSubCategoryCode: req.LoanSubCategoryCode,
					}).
					Return([]models.LoanPartnerAccount{}, models.GetErrMap(models.ErrKeyDatabaseError))
				testHelper.mockCacheRepository.EXPECT().
					DeleteKeysWithPrefix(ctx, gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error insert loan partner account",
			req:  req,
			doMock: func(ctx context.Context, req models.LoanPartnerAccount) {
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, req.AccountNumber).
					Return(models.GetAccountOut{
						AccountNumber: req.AccountNumber,
						AccountType:   req.LoanSubCategoryCode,
						EntityCode:    req.EntityCode,
					}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						AccountNumber:       req.AccountNumber,
						LoanSubCategoryCode: req.LoanSubCategoryCode,
					}).Return([]models.LoanPartnerAccount{}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						PartnerId:           req.PartnerId,
						LoanKind:            req.LoanKind,
						AccountType:         req.AccountType,
						EntityCode:          req.EntityCode,
						LoanSubCategoryCode: req.LoanSubCategoryCode,
					}).Return([]models.LoanPartnerAccount{}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					Create(ctx, req).Return(models.GetErrMap(models.ErrKeyDatabaseError))
				testHelper.mockCacheRepository.EXPECT().
					DeleteKeysWithPrefix(ctx, gomock.Any()).Return(nil)
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
			_, err := testHelper.loanPartnerAccountService.Create(ctx, tt.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestLoanPartnerService_Update(t *testing.T) {
	testHelper := serviceTestHelper(t)
	ctx := context.Background()
	req := models.UpdateLoanPartnerAccount{
		PartnerId:     "efishery",
		LoanKind:      "EFISHERY_LOAN",
		AccountNumber: "22100100000001",
		AccountType:   "INTERNAL_ACCOUNTS_REVENUE_AMARTHA",
	}

	tests := []struct {
		name    string
		req     models.UpdateLoanPartnerAccount
		doMock  func(ctx context.Context, req models.UpdateLoanPartnerAccount)
		wantErr bool
	}{
		{
			name: "success case",
			req:  req,
			doMock: func(ctx context.Context, req models.UpdateLoanPartnerAccount) {
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, req.AccountNumber).
					Return(models.GetAccountOut{}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						AccountNumber: req.AccountNumber,
					}).
					Return([]models.LoanPartnerAccount{{}}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						PartnerId:           req.PartnerId,
						LoanKind:            req.LoanKind,
						AccountType:         req.AccountType,
						LoanSubCategoryCode: req.LoanSubCategoryCode,
					}).
					Return([]models.LoanPartnerAccount{}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					Update(ctx, req).Return(nil)
				testHelper.mockCacheRepository.EXPECT().
					DeleteKeysWithPrefix(ctx, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case - account number not found in acct_account",
			req:  req,
			doMock: func(ctx context.Context, req models.UpdateLoanPartnerAccount) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(ctx, req.AccountNumber).Return(models.GetAccountOut{}, models.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "error case - account number not found in acct_loan_partner_account",
			req:  req,
			doMock: func(ctx context.Context, req models.UpdateLoanPartnerAccount) {
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, req.AccountNumber).
					Return(models.GetAccountOut{}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						AccountNumber: req.AccountNumber,
					}).
					Return([]models.LoanPartnerAccount{}, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - loan partner account exist",
			req:  req,
			doMock: func(ctx context.Context, req models.UpdateLoanPartnerAccount) {
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, req.AccountNumber).
					Return(models.GetAccountOut{}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						AccountNumber: req.AccountNumber,
					}).
					Return([]models.LoanPartnerAccount{{}}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						PartnerId:           req.PartnerId,
						LoanKind:            req.LoanKind,
						AccountType:         req.AccountType,
						LoanSubCategoryCode: req.LoanSubCategoryCode,
					}).
					Return([]models.LoanPartnerAccount{{}}, nil)
			},
			wantErr: true,
		},
		{
			name: "error case - database error when update",
			req:  req,
			doMock: func(ctx context.Context, req models.UpdateLoanPartnerAccount) {
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, req.AccountNumber).
					Return(models.GetAccountOut{}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						AccountNumber: req.AccountNumber,
					}).
					Return([]models.LoanPartnerAccount{{}}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						PartnerId:           req.PartnerId,
						LoanKind:            req.LoanKind,
						AccountType:         req.AccountType,
						LoanSubCategoryCode: req.LoanSubCategoryCode,
					}).
					Return([]models.LoanPartnerAccount{}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					Update(ctx, req).Return(models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case - redis error closed",
			req:  req,
			doMock: func(ctx context.Context, req models.UpdateLoanPartnerAccount) {
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, req.AccountNumber).
					Return(models.GetAccountOut{}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						AccountNumber: req.AccountNumber,
					}).
					Return([]models.LoanPartnerAccount{{}}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, models.GetLoanPartnerAccountByParamsIn{
						PartnerId:           req.PartnerId,
						LoanKind:            req.LoanKind,
						AccountType:         req.AccountType,
						LoanSubCategoryCode: req.LoanSubCategoryCode,
					}).
					Return([]models.LoanPartnerAccount{}, nil)
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					Update(ctx, req).Return(nil)
				testHelper.mockCacheRepository.EXPECT().
					DeleteKeysWithPrefix(ctx, gomock.Any()).Return(models.ErrRedisClosed)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock(ctx, tt.req)
			}
			_, err := testHelper.loanPartnerAccountService.Update(ctx, tt.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestLoanPartnerService_GetByParams(t *testing.T) {
	testHelper := serviceTestHelper(t)
	res := []models.LoanPartnerAccount{
		{
			PartnerId:           "efishery",
			LoanKind:            "EFISHERY_LOAN",
			AccountNumber:       "22100100000001",
			AccountType:         "INTERNAL_ACCOUNTS_REVENUE_AMARTHA",
			EntityCode:          "001",
			LoanSubCategoryCode: "13101",
		},
	}
	b, _ := json.Marshal([]models.LoanPartnerAccount{})

	type args struct {
		ctx context.Context
		req models.GetLoanPartnerAccountByParamsIn
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
			name: "success case - empty params",
			args: args{
				ctx: context.Background(),
				req: models.GetLoanPartnerAccountByParamsIn{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockLoanPartnerAccountRepository.EXPECT().GetByParams(args.ctx, args.req).Return(res, nil)
			},
			wantErr: false,
		},
		{
			name: "error case - empty params",
			args: args{
				ctx: context.Background(),
				req: models.GetLoanPartnerAccountByParamsIn{},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockLoanPartnerAccountRepository.EXPECT().GetByParams(args.ctx, args.req).Return([]models.LoanPartnerAccount{}, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "success case with params - get from cache",
			args: args{
				ctx: context.Background(),
				req: models.GetLoanPartnerAccountByParamsIn{
					PartnerId: "efishery",
				},
			},
			doMock: func(args args, mockData mockData) {
				b, _ := json.Marshal(res)
				testHelper.mockCacheRepository.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(string(b), nil)
			},
			wantErr: false,
		},
		{
			name: "success case with params - get from cache and database",
			args: args{
				ctx: context.Background(),
				req: models.GetLoanPartnerAccountByParamsIn{
					PartnerId:           "efishery",
					LoanKind:            "EFISHERY_LOAN",
					AccountNumber:       "22100100000001",
					AccountType:         "INTERNAL_ACCOUNTS_REVENUE_AMARTHA",
					EntityCode:          "001",
					LoanSubCategoryCode: "13101",
					LoanAccountNumber:   "12100100000001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{
					AccountNumber: args.req.AccountNumber,
					EntityCode:    "001",
				}, nil)
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.LoanAccountNumber).Return(models.GetAccountOut{
					AccountNumber:   args.req.LoanAccountNumber,
					EntityCode:      "001",
					SubCategoryCode: "13101",
				}, nil)
				testHelper.mockCacheRepository.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return("", models.GetErrMap(models.ErrKeyDataNotFound))
				testHelper.mockLoanPartnerAccountRepository.EXPECT().GetByParams(args.ctx, args.req).Return(res, nil)
				testHelper.mockCacheRepository.EXPECT().
					Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error case with params - account number not found",
			args: args{
				ctx: context.Background(),
				req: models.GetLoanPartnerAccountByParamsIn{
					AccountNumber: "22100100000001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.AccountNumber).Return(models.GetAccountOut{}, models.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "error case with params - loan account number not found",
			args: args{
				ctx: context.Background(),
				req: models.GetLoanPartnerAccountByParamsIn{
					LoanAccountNumber: "22100100000001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockAccRepository.EXPECT().GetOneByAccountNumber(args.ctx, args.req.LoanAccountNumber).Return(models.GetAccountOut{}, models.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "error case with params - entity code not found",
			args: args{
				ctx: context.Background(),
				req: models.GetLoanPartnerAccountByParamsIn{
					EntityCode: "001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error case with params - database error get entity",
			args: args{
				ctx: context.Background(),
				req: models.GetLoanPartnerAccountByParamsIn{
					EntityCode: "001",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockEntityRepository.EXPECT().GetByCode(args.ctx, args.req.EntityCode).Return(nil, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error case with params - cache error",
			args: args{
				ctx: context.Background(),
				req: models.GetLoanPartnerAccountByParamsIn{
					PartnerId: "efishery",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCacheRepository.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return("0", nil)
			},
			wantErr: true,
		},
		{
			name: "error case with params - no result",
			args: args{
				ctx: context.Background(),
				req: models.GetLoanPartnerAccountByParamsIn{
					PartnerId: "efishery",
				},
			},
			doMock: func(args args, mockData mockData) {
				testHelper.mockCacheRepository.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(string(b), nil)
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
			_, err := testHelper.loanPartnerAccountService.GetByParams(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
