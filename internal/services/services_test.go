package services_test

import (
	"os"
	"testing"

	"bitbucket.org/Amartha/go-accounting/internal/config"
	"bitbucket.org/Amartha/go-accounting/internal/services"

	mockAC "bitbucket.org/Amartha/go-accounting/internal/pkg/acuanclient/mock"
	mockDDDNotification "bitbucket.org/Amartha/go-accounting/internal/pkg/dddnotification/mock"
	mockFile "bitbucket.org/Amartha/go-accounting/internal/pkg/file/mock"
	mockFlag "bitbucket.org/Amartha/go-accounting/internal/pkg/flag/mock"
	mockGoDBLedger "bitbucket.org/Amartha/go-accounting/internal/pkg/godbledger/mock"
	mockFpTransaction "bitbucket.org/Amartha/go-accounting/internal/pkg/gofptransaction/mock"
	mockIgate "bitbucket.org/Amartha/go-accounting/internal/pkg/goigate/mock"
	mockPublisher "bitbucket.org/Amartha/go-accounting/internal/pkg/kafka/mock"
	mockQueueUnicorn "bitbucket.org/Amartha/go-accounting/internal/pkg/queueunicorn/mock"
	mockBigQuery "bitbucket.org/Amartha/go-accounting/internal/repositories/bq/mock"
	mockcache "bitbucket.org/Amartha/go-accounting/internal/repositories/cache/mock"
	mockgcs "bitbucket.org/Amartha/go-accounting/internal/repositories/gcs/mock"
	mockmysql "bitbucket.org/Amartha/go-accounting/internal/repositories/mysql/mock"
	mockStorage "bitbucket.org/Amartha/go-accounting/internal/repositories/storage/mock"
	xlog "bitbucket.org/Amartha/go-x/log"

	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	xlog.InitForTest()
	os.Exit(m.Run())
}

type testServiceHelper struct {
	mockCtrl *gomock.Controller
	config   config.Configuration
	mockFlag *mockFlag.MockFlaggerClient

	mockAccRepository                *mockmysql.MockAccountRepository
	mockAcctRepository               *mockmysql.MockAccountingRepository
	mockCategoryRepository           *mockmysql.MockCategoryRepository
	mockCOATypeRepository            *mockmysql.MockCOATypeRepository
	mockEntityRepository             *mockmysql.MockEntityRepository
	mockLoanPartnerAccountRepository *mockmysql.MockLoanPartnerAccountRepository
	mockMySQLRepository              *mockmysql.MockSQLRepository
	mockProductTypeRepository        *mockmysql.MockProductTypeRepository
	mockSubCategoryRepository        *mockmysql.MockSubCategoryRepository
	mockTrialBalanceRepository       *mockmysql.MockTrialBalanceRepository

	accountingService         services.AccountingService
	accountService            services.AccountService
	categoryService           services.CategoryService
	coaTypeService            services.COATypeService
	dlqProcessorService       services.DLQProcessorService
	entityService             services.EntityService
	healthService             services.HealthService
	journalService            services.JournalService
	loanPartnerAccountService services.LoanPartnerService
	migrationService          services.MigrationService
	productTypeService        services.ProductTypeService
	publisherService          services.PublisherService
	retryService              services.RetryService
	subCategoryService        services.SubCategoryService
	trialBalanceService       services.TrialBalanceService

	mockAcuanClient            *mockAC.MockAcuanClient
	mockCacheRepository        *mockcache.MockCacheRepository
	mockCloudStorageRepository *mockgcs.MockCloudStorageRepository
	mockDDDNotification        *mockDDDNotification.MockDDDNotificationClient
	mockFile                   *mockFile.MockIOFile
	mockGoDbLedger             *mockGoDBLedger.MockGoDBLedger
	mockGoFpTransaction        *mockFpTransaction.MockFpTransactionClient
	mockIgateClient            *mockIgate.MockIgateClient
	mockPublisher              *mockPublisher.MockPublisher
	mockQueueUnicorn           *mockQueueUnicorn.MockQueueUnicornClient
	mockStorageRepository      *mockStorage.MockStorageRepository
	mockBigQuery               *mockBigQuery.MockBigQueryRepository
}

func serviceTestHelper(t *testing.T) testServiceHelper {
	t.Helper()
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockAcuanClient := mockAC.NewMockAcuanClient(mockCtrl)
	mockCacheRepository := mockcache.NewMockCacheRepository(mockCtrl)
	mockCloudStorageRepository := mockgcs.NewMockCloudStorageRepository(mockCtrl)
	mockDDDNotification := mockDDDNotification.NewMockDDDNotificationClient(mockCtrl)
	mockFile := mockFile.NewMockIOFile(mockCtrl)
	mockFlag := mockFlag.NewMockFlaggerClient(mockCtrl)
	mockGoDBLedger := mockGoDBLedger.NewMockGoDBLedger(mockCtrl)
	mockGoFpTransaction := mockFpTransaction.NewMockFpTransactionClient(mockCtrl)
	mockIgateClient := mockIgate.NewMockIgateClient(mockCtrl)
	mockPublisher := mockPublisher.NewMockPublisher(mockCtrl)
	mockQueueUnicorn := mockQueueUnicorn.NewMockQueueUnicornClient(mockCtrl)
	mockStorageRepository := mockStorage.NewMockStorageRepository(mockCtrl)
	mockBigQueryRepository := mockBigQuery.NewMockBigQueryRepository(mockCtrl)

	mockAccountingRepository := mockmysql.NewMockAccountingRepository(mockCtrl)
	mockAccountRepository := mockmysql.NewMockAccountRepository(mockCtrl)
	mockCategoryRepository := mockmysql.NewMockCategoryRepository(mockCtrl)
	mockCOATypeRepository := mockmysql.NewMockCOATypeRepository(mockCtrl)
	mockEntityRepository := mockmysql.NewMockEntityRepository(mockCtrl)
	mockLoanPartnerAccountRepository := mockmysql.NewMockLoanPartnerAccountRepository(mockCtrl)
	mockMySQLRepository := mockmysql.NewMockSQLRepository(mockCtrl)
	mockProductTypeRepository := mockmysql.NewMockProductTypeRepository(mockCtrl)
	mockSubCategoryRepository := mockmysql.NewMockSubCategoryRepository(mockCtrl)
	mockTrialBalanceRepository := mockmysql.NewMockTrialBalanceRepository(mockCtrl)

	mockMySQLRepository.EXPECT().GetAccountingRepository().Return(mockAccountingRepository).AnyTimes()
	mockMySQLRepository.EXPECT().GetAccountRepository().Return(mockAccountRepository).AnyTimes()
	mockMySQLRepository.EXPECT().GetCategoryRepository().Return(mockCategoryRepository).AnyTimes()
	mockMySQLRepository.EXPECT().GetCOATypeRepository().Return(mockCOATypeRepository).AnyTimes()
	mockMySQLRepository.EXPECT().GetEntityRepository().Return(mockEntityRepository).AnyTimes()
	mockMySQLRepository.EXPECT().GetLoanPartnerAccountRepository().Return(mockLoanPartnerAccountRepository).AnyTimes()
	mockMySQLRepository.EXPECT().GetProductTypeRepository().Return(mockProductTypeRepository).AnyTimes()
	mockMySQLRepository.EXPECT().GetSubCategoryRepository().Return(mockSubCategoryRepository).AnyTimes()
	mockMySQLRepository.EXPECT().GetTrialBalanceRepository().Return(mockTrialBalanceRepository).AnyTimes()

	conf := config.Configuration{
		AccountConfig: config.AccountConfig{
			AccountNumberPadWidth:    8,
			IsCreateAccountT24:       true,
			LimitAccountSubLedger:    5,
			LimitAccountTrialBalance: 10,
			InvestedAccountNumber:    make(map[string]string),
			ReceivablesAccountNumber: make(map[string]string),
			MultiLoanAccount:         make(map[string]string),
			ChunkSizeAccountBalance:  10,
			LoanPartnerAccountConfig: map[string]config.LoanPartnerAccountConfig{
				"wht23_26": {
					Name:            "WHT 23/26 -",
					CategoryCode:    "121",
					SubCategoryCode: "12101",
					Currency:        "IDR",
					AccountType:     "INTERNAL_ACCOUNTS_PPH_AMARTHA",
				},
			},
			LoanPartnerAccountEntities: []string{
				"001", "005",
			},
			AllowedEntitiesAccountDailyBalance: []string{
				"001",
			},
		},
		SecretKey: "",
		SQLTransaction: config.SQLTransactionConfiguration{
			BulkLimit: 1,
		},
	}
	conf.AccountConfig.InvestedAccountNumber["21101"] = "21201"
	conf.AccountConfig.InvestedAccountNumber["21102"] = "21202"
	conf.AccountConfig.ReceivablesAccountNumber["21102"] = "14201"
	conf.AccountConfig.MultiLoanAccount["13101"] = "21303"

	serv := services.New(
		&conf,
		mockFlag,
		mockGoDBLedger,
		mockMySQLRepository,
		mockCacheRepository,
		mockStorageRepository,
		mockFile,
		mockAcuanClient,
		mockIgateClient,
		mockDDDNotification,
		mockQueueUnicorn,
		mockPublisher,
		mockGoFpTransaction,
		mockCloudStorageRepository,
		mockBigQueryRepository,
	)

	return testServiceHelper{
		mockCtrl: mockCtrl,
		config:   conf,
		mockFlag: mockFlag,

		mockAccRepository:                mockAccountRepository,
		mockAcctRepository:               mockAccountingRepository,
		mockCategoryRepository:           mockCategoryRepository,
		mockCloudStorageRepository:       mockCloudStorageRepository,
		mockCOATypeRepository:            mockCOATypeRepository,
		mockEntityRepository:             mockEntityRepository,
		mockLoanPartnerAccountRepository: mockLoanPartnerAccountRepository,
		mockMySQLRepository:              mockMySQLRepository,
		mockProductTypeRepository:        mockProductTypeRepository,
		mockSubCategoryRepository:        mockSubCategoryRepository,
		mockTrialBalanceRepository:       mockTrialBalanceRepository,

		mockAcuanClient:       mockAcuanClient,
		mockCacheRepository:   mockCacheRepository,
		mockDDDNotification:   mockDDDNotification,
		mockFile:              mockFile,
		mockGoDbLedger:        mockGoDBLedger,
		mockGoFpTransaction:   mockGoFpTransaction,
		mockIgateClient:       mockIgateClient,
		mockPublisher:         mockPublisher,
		mockQueueUnicorn:      mockQueueUnicorn,
		mockStorageRepository: mockStorageRepository,
		mockBigQuery:          mockBigQueryRepository,

		accountingService:         serv.Accounting,
		accountService:            serv.Account,
		categoryService:           serv.Category,
		coaTypeService:            serv.COAType,
		dlqProcessorService:       serv.DLQProcessor,
		entityService:             serv.Entity,
		healthService:             serv.HealthService,
		journalService:            serv.Journal,
		loanPartnerAccountService: serv.LoanPartnerAccount,
		migrationService:          serv.Migration,
		productTypeService:        serv.ProductType,
		publisherService:          serv.PublisherService,
		retryService:              serv.RetryService,
		subCategoryService:        serv.SubCategory,
		trialBalanceService:       serv.TrialBalance,
	}
}
