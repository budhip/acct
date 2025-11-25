package services_test

import (
	"context"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/godbledger"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/mysql"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_account_Create(t *testing.T) {
	testHelper := serviceTestHelper(t)

	ctx := context.Background()
	reqWithoutAccountType := models.CreateAccount{
		OwnerID:         "12345",
		ProductTypeCode: "1013",
		EntityCode:      "001",
		CategoryCode:    "131",
		SubCategoryCode: "13112",
		Currency:        models.CurrencyIDR,
		Status:          models.AccountStatusActive,
		Name:            "Testing Name",
		AltId:           "1234567890",
		Metadata:        &models.Metadata{},
	}

	reqWithAccountType := reqWithoutAccountType
	reqWithAccountType.AccountType = "LOAN_ACCOUNT_PAYLATER_TELKOMSEL"

	reqWithAccountTypeLoanPartner := reqWithoutAccountType
	reqWithAccountTypeLoanPartner.AccountType = "PARTNERSHIP_LOAN"
	reqWithAccountTypeLoanPartner.Metadata = &models.Metadata{
		"partnerId": "123456789",
	}

	tests := []struct {
		name    string
		req     models.CreateAccount
		doMock  func(ctx context.Context, req models.CreateAccount)
		wantErr bool
	}{
		{
			name: "error: invalid account type (with account type)",
			req:  reqWithAccountType,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByAccountType(ctx, req.AccountType).
					Return(&models.SubCategory{}, models.GetErrMap(models.ErrKeyAccountTypeNotValid))
			},
			wantErr: true,
		},
		{
			name: "error: database error get product type (with account type)",
			req:  reqWithAccountType,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByAccountType(ctx, req.AccountType).
					Return(&models.SubCategory{
						ProductTypeCode: req.ProductTypeCode,
					}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(nil, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error: product type not found (with account type)",
			req:  reqWithAccountType,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByAccountType(ctx, req.AccountType).
					Return(&models.SubCategory{
						ProductTypeCode: req.ProductTypeCode,
					}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error: metadata has invalid type for partnerId (with account type loan partner)",
			req: func() models.CreateAccount {
				r := reqWithAccountTypeLoanPartner
				r.Metadata = &models.Metadata{
					"partnerId": make(chan int), // invalid type for JSON
				}
				return r
			}(),
			doMock:  nil,
			wantErr: true,
		},
		{
			name: "error: partnerId empty (with account type loan partner)",
			req: func() models.CreateAccount {
				r := reqWithAccountTypeLoanPartner
				r.Metadata = &models.Metadata{
					"partnerId": "",
				}
				return r
			}(),
			doMock:  nil,
			wantErr: true,
		},
		{
			name: "error: database error get loan partner account (with account type loan partner)",
			req:  reqWithAccountTypeLoanPartner,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, gomock.Any()).
					Return([]models.LoanPartnerAccount{}, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error: loan partner account not found (with account type loan partner)",
			req:  reqWithAccountTypeLoanPartner,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, gomock.Any()).
					Return([]models.LoanPartnerAccount{}, nil)
			},
			wantErr: true,
		},
		{
			name: "error: database error get sub category (with account type loan partner)",
			req:  reqWithAccountTypeLoanPartner,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, gomock.Any()).
					Return([]models.LoanPartnerAccount{{LoanSubCategoryCode: req.SubCategoryCode}}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(ctx, req.SubCategoryCode).
					Return(nil, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error: sub category not found (with account type loan partner)",
			req:  reqWithAccountTypeLoanPartner,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, gomock.Any()).
					Return([]models.LoanPartnerAccount{{LoanSubCategoryCode: req.SubCategoryCode}}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(ctx, req.SubCategoryCode).
					Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error: database error get product type (with account type loan partner)",
			req:  reqWithAccountTypeLoanPartner,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, gomock.Any()).
					Return([]models.LoanPartnerAccount{{LoanSubCategoryCode: req.SubCategoryCode}}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(ctx, req.SubCategoryCode).
					Return(&models.SubCategory{ProductTypeCode: req.ProductTypeCode}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(nil, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error: product type not found (with account type loan partner)",
			req:  reqWithAccountTypeLoanPartner,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, gomock.Any()).
					Return([]models.LoanPartnerAccount{{LoanSubCategoryCode: req.SubCategoryCode}}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(ctx, req.SubCategoryCode).
					Return(&models.SubCategory{ProductTypeCode: req.ProductTypeCode}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error: category code not found (without account type)",
			req:  reqWithoutAccountType,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockCategoryRepository.EXPECT().
					CheckCategoryByCode(ctx, req.CategoryCode).
					Return(models.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "error: sub category code not found (without account type)",
			req:  reqWithoutAccountType,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockCategoryRepository.EXPECT().
					CheckCategoryByCode(ctx, req.CategoryCode).
					Return(nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					CheckSubCategoryByCodeAndCategoryCode(ctx, req.SubCategoryCode, req.CategoryCode).
					Return(nil, models.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "error: database error get product type (without account type)",
			req:  reqWithoutAccountType,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockCategoryRepository.EXPECT().
					CheckCategoryByCode(ctx, req.CategoryCode).
					Return(nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					CheckSubCategoryByCodeAndCategoryCode(ctx, req.SubCategoryCode, req.CategoryCode).
					Return(&models.SubCategory{AccountType: req.AccountType}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(nil, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error: product type not found (without account type)",
			req:  reqWithoutAccountType,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockCategoryRepository.EXPECT().
					CheckCategoryByCode(ctx, req.CategoryCode).
					Return(nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					CheckSubCategoryByCodeAndCategoryCode(ctx, req.SubCategoryCode, req.CategoryCode).
					Return(&models.SubCategory{AccountType: req.AccountType}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "error: entity not found (without account type)",
			req:  reqWithoutAccountType,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockCategoryRepository.EXPECT().
					CheckCategoryByCode(ctx, req.CategoryCode).
					Return(nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					CheckSubCategoryByCodeAndCategoryCode(ctx, req.SubCategoryCode, req.CategoryCode).
					Return(&models.SubCategory{AccountType: req.AccountType}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(&models.ProductType{}, nil)
				testHelper.mockEntityRepository.EXPECT().
					CheckEntityByCode(ctx, req.EntityCode, models.StatusActive).
					Return(models.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "error: currency not found (without account type)",
			req:  reqWithoutAccountType,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockCategoryRepository.EXPECT().
					CheckCategoryByCode(ctx, req.CategoryCode).
					Return(nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					CheckSubCategoryByCodeAndCategoryCode(ctx, req.SubCategoryCode, req.CategoryCode).
					Return(&models.SubCategory{AccountType: req.AccountType}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(&models.ProductType{}, nil)
				testHelper.mockEntityRepository.EXPECT().
					CheckEntityByCode(ctx, req.EntityCode, models.StatusActive).
					Return(nil)
				testHelper.mockGoDbLedger.EXPECT().
					GetCurrency(ctx, req.Currency).Return(nil, models.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "error: metadata has invalid type for t24AccountNumber (without account type)",
			req: func() models.CreateAccount {
				r := reqWithoutAccountType
				r.LegacyId = &models.AccountLegacyId{
					"t24AccountNumber": make(chan int), // invalid type for JSON
				}
				return r
			}(),
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockCategoryRepository.EXPECT().
					CheckCategoryByCode(ctx, req.CategoryCode).
					Return(nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					CheckSubCategoryByCodeAndCategoryCode(ctx, req.SubCategoryCode, req.CategoryCode).
					Return(&models.SubCategory{AccountType: req.AccountType}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(&models.ProductType{}, nil)
				testHelper.mockEntityRepository.EXPECT().
					CheckEntityByCode(ctx, req.EntityCode, models.StatusActive).
					Return(nil)
				testHelper.mockGoDbLedger.EXPECT().
					GetCurrency(ctx, req.Currency).Return(godbledger.CurrencyIDR, nil)
			},
			wantErr: true,
		},
		{
			name: "error: failed get sequence (without account type)",
			req: func() models.CreateAccount {
				r := reqWithoutAccountType
				r.LegacyId = &models.AccountLegacyId{
					"t24AccountNumber": "",
				}
				return r
			}(),
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockCategoryRepository.EXPECT().
					CheckCategoryByCode(ctx, req.CategoryCode).
					Return(nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					CheckSubCategoryByCodeAndCategoryCode(ctx, req.SubCategoryCode, req.CategoryCode).
					Return(&models.SubCategory{AccountType: req.AccountType}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(&models.ProductType{}, nil)
				testHelper.mockEntityRepository.EXPECT().
					CheckEntityByCode(ctx, req.EntityCode, models.StatusActive).
					Return(nil)
				testHelper.mockGoDbLedger.EXPECT().
					GetCurrency(ctx, req.Currency).Return(godbledger.CurrencyIDR, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(ctx, gomock.Any()).
					Return(int64(1), models.ErrRedisClosed)
			},
			wantErr: true,
		},
		{
			name: "error: database error get account (without account type)",
			req: func() models.CreateAccount {
				r := reqWithoutAccountType
				r.LegacyId = &models.AccountLegacyId{
					"t24AccountNumber": "",
				}
				return r
			}(),
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockCategoryRepository.EXPECT().
					CheckCategoryByCode(ctx, req.CategoryCode).
					Return(nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					CheckSubCategoryByCodeAndCategoryCode(ctx, req.SubCategoryCode, req.CategoryCode).
					Return(&models.SubCategory{AccountType: req.AccountType}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(&models.ProductType{}, nil)
				testHelper.mockEntityRepository.EXPECT().
					CheckEntityByCode(ctx, req.EntityCode, models.StatusActive).
					Return(nil)
				testHelper.mockGoDbLedger.EXPECT().
					GetCurrency(ctx, req.Currency).Return(godbledger.CurrencyIDR, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(ctx, gomock.Any()).
					Return(int64(1), nil)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, gomock.Any()).Return(models.GetAccountOut{}, models.GetErrMap(models.ErrKeyDatabaseError))
			},
			wantErr: true,
		},
		{
			name: "error: generate other account with subcategory 21102 (with account type)",
			req: func() models.CreateAccount {
				r := reqWithAccountType
				r.LegacyId = &models.AccountLegacyId{
					"t24AccountNumber": "",
				}
				r.CategoryCode = "211"
				r.SubCategoryCode = "21102"
				return r
			}(),
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByAccountType(ctx, req.AccountType).
					Return(&models.SubCategory{
						Code:            req.SubCategoryCode,
						ProductTypeCode: req.ProductTypeCode,
					}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(&models.ProductType{}, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(ctx, gomock.Any()).
					Return(int64(1), nil).MaxTimes(2)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, gomock.Any()).
					Return(models.GetAccountOut{}, models.ErrNoRows)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(ctx, gomock.Any()).
					Return(&models.SubCategory{}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(ctx, gomock.Any()).
					Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "success: the account already create (without account type)",
			req: func() models.CreateAccount {
				r := reqWithoutAccountType
				r.LegacyId = &models.AccountLegacyId{
					"t24AccountNumber": "",
				}
				return r
			}(),
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockCategoryRepository.EXPECT().
					CheckCategoryByCode(ctx, req.CategoryCode).
					Return(nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					CheckSubCategoryByCodeAndCategoryCode(ctx, req.SubCategoryCode, req.CategoryCode).
					Return(&models.SubCategory{AccountType: req.AccountType}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(&models.ProductType{}, nil)
				testHelper.mockEntityRepository.EXPECT().
					CheckEntityByCode(ctx, req.EntityCode, models.StatusActive).
					Return(nil)
				testHelper.mockGoDbLedger.EXPECT().
					GetCurrency(ctx, req.Currency).Return(godbledger.CurrencyIDR, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(ctx, gomock.Any()).
					Return(int64(1), nil)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, gomock.Any()).
					Return(models.GetAccountOut{}, nil)
				testHelper.mockAccRepository.EXPECT().
					UpdateAltId(ctx, gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success: create with subcategory 13112 (with account type loan partner)",
			req:  reqWithAccountTypeLoanPartner,
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockLoanPartnerAccountRepository.EXPECT().
					GetByParams(ctx, gomock.Any()).
					Return([]models.LoanPartnerAccount{{LoanSubCategoryCode: req.SubCategoryCode}}, nil)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(ctx, req.SubCategoryCode).
					Return(&models.SubCategory{ProductTypeCode: req.ProductTypeCode}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(&models.ProductType{EntityCode: req.EntityCode}, nil)

				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(ctx, gomock.Any()).
					Return(int64(1), nil)
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, gomock.Any()).
					Return(models.GetAccountOut{}, models.ErrNoRows)

				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockAccRepository.EXPECT().BulkInsertAcctAccount(ctx, gomock.Any()).Return(nil)
						testHelper.mockAccRepository.EXPECT().BulkInsertAccount(ctx, gomock.Any()).Return(nil)
						testHelper.mockAccRepository.EXPECT().CreateLoanAccount(ctx, gomock.Any()).Return(nil)
						return steps(ctx, testHelper.mockMySQLRepository)
					})
				testHelper.mockPublisher.EXPECT().PublishSyncWithKeyAndLog(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				testHelper.mockAcuanClient.EXPECT().PublishAccount(ctx, gomock.Any()).AnyTimes()
				testHelper.mockCacheRepository.EXPECT().Del(ctx, gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error: database error with subcategory 13101 (with account type)",
			req: func() models.CreateAccount {
				r := reqWithAccountType
				r.LegacyId = &models.AccountLegacyId{
					"t24AccountNumber": "",
				}
				r.CategoryCode = "131"
				r.SubCategoryCode = "13101"
				return r
			}(),
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByAccountType(ctx, req.AccountType).
					Return(&models.SubCategory{
						Code:            req.SubCategoryCode,
						ProductTypeCode: req.ProductTypeCode,
					}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(&models.ProductType{}, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(ctx, gomock.Any()).
					Return(int64(1), nil).AnyTimes()
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, gomock.Any()).
					Return(models.GetAccountOut{}, models.ErrNoRows)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(ctx, gomock.Any()).
					Return(&models.SubCategory{}, nil).AnyTimes()

				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockAccRepository.EXPECT().BulkInsertAcctAccount(ctx, gomock.Any()).Return(nil)
						testHelper.mockAccRepository.EXPECT().BulkInsertAccount(ctx, gomock.Any()).Return(nil)
						testHelper.mockAccRepository.EXPECT().CreateLoanAccount(ctx, gomock.Any()).Return(models.ErrNoRowsAffected)
						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(models.ErrNoRowsAffected)
				testHelper.mockCacheRepository.EXPECT().Del(ctx, gomock.Any()).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "error: database error with subcategory 13101 (with account type)",
			req: func() models.CreateAccount {
				r := reqWithAccountType
				r.LegacyId = &models.AccountLegacyId{
					"t24AccountNumber": "",
				}
				r.CategoryCode = "131"
				r.SubCategoryCode = "13101"
				return r
			}(),
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByAccountType(ctx, req.AccountType).
					Return(&models.SubCategory{
						Code:            req.SubCategoryCode,
						ProductTypeCode: req.ProductTypeCode,
					}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(&models.ProductType{}, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(ctx, gomock.Any()).
					Return(int64(1), nil).AnyTimes()
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, gomock.Any()).
					Return(models.GetAccountOut{}, models.ErrNoRows)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(ctx, gomock.Any()).
					Return(&models.SubCategory{}, nil).AnyTimes()

				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockAccRepository.EXPECT().BulkInsertAcctAccount(ctx, gomock.Any()).Return(nil)
						testHelper.mockAccRepository.EXPECT().BulkInsertAccount(ctx, gomock.Any()).Return(nil)
						testHelper.mockAccRepository.EXPECT().CreateLoanAccount(ctx, gomock.Any()).Return(models.ErrNoRowsAffected)
						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(models.ErrNoRowsAffected)
			},
			wantErr: true,
		},
		{
			name: "error: database error with subcategory 21102 (with account type)",
			req: func() models.CreateAccount {
				r := reqWithAccountType
				r.LegacyId = &models.AccountLegacyId{
					"t24AccountNumber": "",
				}
				r.CategoryCode = "211"
				r.SubCategoryCode = "21102"
				return r
			}(),
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByAccountType(ctx, req.AccountType).
					Return(&models.SubCategory{
						Code:            req.SubCategoryCode,
						ProductTypeCode: req.ProductTypeCode,
					}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(&models.ProductType{}, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(ctx, gomock.Any()).
					Return(int64(1), nil).AnyTimes()
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, gomock.Any()).
					Return(models.GetAccountOut{}, models.ErrNoRows)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(ctx, gomock.Any()).
					Return(&models.SubCategory{}, nil).AnyTimes()

				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockAccRepository.EXPECT().BulkInsertAcctAccount(ctx, gomock.Any()).Return(models.ErrNoRowsAffected)
						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(models.ErrNoRowsAffected)
			},
			wantErr: true,
		},
		{
			name: "error: database error with subcategory 21102 (with account type)",
			req: func() models.CreateAccount {
				r := reqWithAccountType
				r.LegacyId = &models.AccountLegacyId{
					"t24AccountNumber": "",
				}
				r.CategoryCode = "211"
				r.SubCategoryCode = "21102"
				return r
			}(),
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByAccountType(ctx, req.AccountType).
					Return(&models.SubCategory{
						Code:            req.SubCategoryCode,
						ProductTypeCode: req.ProductTypeCode,
					}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(&models.ProductType{}, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(ctx, gomock.Any()).
					Return(int64(1), nil).AnyTimes()
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, gomock.Any()).
					Return(models.GetAccountOut{}, models.ErrNoRows)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(ctx, gomock.Any()).
					Return(&models.SubCategory{}, nil).AnyTimes()

				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockAccRepository.EXPECT().BulkInsertAcctAccount(ctx, gomock.Any()).Return(nil)
						testHelper.mockAccRepository.EXPECT().BulkInsertAccount(ctx, gomock.Any()).Return(models.ErrNoRowsAffected)
						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(models.ErrNoRowsAffected)
			},
			wantErr: true,
		},
		{
			name: "error: database error with subcategory 21102 (with account type)",
			req: func() models.CreateAccount {
				r := reqWithAccountType
				r.LegacyId = &models.AccountLegacyId{
					"t24AccountNumber": "",
				}
				r.CategoryCode = "211"
				r.SubCategoryCode = "21102"
				return r
			}(),
			doMock: func(ctx context.Context, req models.CreateAccount) {
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByAccountType(ctx, req.AccountType).
					Return(&models.SubCategory{
						Code:            req.SubCategoryCode,
						ProductTypeCode: req.ProductTypeCode,
					}, nil)
				testHelper.mockProductTypeRepository.EXPECT().
					GetByCode(ctx, req.ProductTypeCode).
					Return(&models.ProductType{}, nil)
				testHelper.mockCacheRepository.EXPECT().
					GetIncrement(ctx, gomock.Any()).
					Return(int64(1), nil).AnyTimes()
				testHelper.mockAccRepository.EXPECT().
					GetOneByAccountNumber(ctx, gomock.Any()).
					Return(models.GetAccountOut{}, models.ErrNoRows)
				testHelper.mockSubCategoryRepository.EXPECT().
					GetByCode(ctx, gomock.Any()).
					Return(&models.SubCategory{}, nil).AnyTimes()

				testHelper.mockMySQLRepository.EXPECT().Atomic(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, steps func(ctx context.Context, r mysql.SQLRepository) error) error {
						testHelper.mockAccRepository.EXPECT().BulkInsertAcctAccount(ctx, gomock.Any()).Return(nil)
						testHelper.mockAccRepository.EXPECT().BulkInsertAccount(ctx, gomock.Any()).Return(nil)
						testHelper.mockAccRepository.EXPECT().CreateLenderAccount(ctx, gomock.Any()).Return(models.ErrNoRowsAffected)
						return steps(ctx, testHelper.mockMySQLRepository)
					}).Return(models.ErrNoRowsAffected)
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
			_, err := testHelper.accountService.Create(ctx, tt.req)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
