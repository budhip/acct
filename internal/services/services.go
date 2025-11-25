package services

import (
	"context"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/config"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/acuanclient"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/dddnotification"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/file"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/godbledger"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/gofptransaction"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/goigate"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/kafka"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/queueunicorn"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/bq"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/cache"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/gcs"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/mysql"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/storage"

	flag "bitbucket.org/Amartha/go-feature-flag-sdk"
	xlog "bitbucket.org/Amartha/go-x/log"
)

type service struct {
	srv *Services
}

type Services struct {
	common  service
	conf    *config.Configuration
	flagger flag.IFlagger

	goDBLedger            godbledger.GoDBLedger
	mySqlRepo             mysql.SQLRepository
	cacheRepo             cache.CacheRepository
	storageRepo           storage.StorageRepository
	file                  file.IOFile
	acuanClient           acuanclient.AcuanClient
	igateClient           goigate.IgateClient
	dddNotificationClient dddnotification.DDDNotificationClient
	queueUnicornClient    queueunicorn.QueueUnicornClient
	publisher             kafka.Publisher
	fpTransactionClient   gofptransaction.FpTransactionClient
	cloudStorageRepo      gcs.CloudStorageRepository
	bigQueryRepo          bq.BigQueryRepository

	Account            *account
	Accounting         *accounting
	Entity             *entity
	Category           *category
	SubCategory        *subCategory
	Journal            *journalService
	COAType            *coaTypes
	ProductType        *productType
	LoanPartnerAccount *loanPartnerAccount
	DLQProcessor       *dlqProcessor
	RetryService       *retryService
	HealthService      *healthService
	Migration          *migrationService
	PublisherService   *publisherService
	TrialBalance       *trialBalance
}

func New(
	conf *config.Configuration,
	flagger flag.IFlagger,
	goDBLedger godbledger.GoDBLedger,
	mySqlRepo mysql.SQLRepository,
	cacheRepo cache.CacheRepository,
	storageRepo storage.StorageRepository,
	file file.IOFile,
	acuanClient acuanclient.AcuanClient,
	igateClient goigate.IgateClient,
	dddNotificationClient dddnotification.DDDNotificationClient,
	queueUnicornClient queueunicorn.QueueUnicornClient,
	publisher kafka.Publisher,
	fpTransactionClient gofptransaction.FpTransactionClient,
	cloudStorageRepo gcs.CloudStorageRepository,
	bigQueryRepo bq.BigQueryRepository,
) *Services {
	srv := &Services{
		conf:                  conf,
		flagger:               flagger,
		goDBLedger:            goDBLedger,
		mySqlRepo:             mySqlRepo,
		cacheRepo:             cacheRepo,
		storageRepo:           storageRepo,
		file:                  file,
		acuanClient:           acuanClient,
		igateClient:           igateClient,
		dddNotificationClient: dddNotificationClient,
		queueUnicornClient:    queueUnicornClient,
		publisher:             publisher,
		fpTransactionClient:   fpTransactionClient,
		cloudStorageRepo:      cloudStorageRepo,
		bigQueryRepo:          bigQueryRepo,
	}
	srv.common.srv = srv
	srv.Account = (*account)(&srv.common)
	srv.Accounting = (*accounting)(&srv.common)
	srv.Category = (*category)(&srv.common)
	srv.COAType = (*coaTypes)(&srv.common)
	srv.Entity = (*entity)(&srv.common)
	srv.Journal = (*journalService)(&srv.common)
	srv.ProductType = (*productType)(&srv.common)
	srv.SubCategory = (*subCategory)(&srv.common)
	srv.LoanPartnerAccount = (*loanPartnerAccount)(&srv.common)
	srv.DLQProcessor = (*dlqProcessor)(&srv.common)
	srv.RetryService = (*retryService)(&srv.common)
	srv.HealthService = (*healthService)(&srv.common)
	srv.Migration = (*migrationService)(&srv.common)
	srv.PublisherService = (*publisherService)(&srv.common)
	srv.TrialBalance = (*trialBalance)(&srv.common)

	return srv
}

const (
	logMessageService = "[SERVICE]"
)

func logService(ctx context.Context, err error) {
	if err != nil {
		logStatusError := xlog.String("status", "error")
		logError := xlog.Err(err)

		respErr, ok := models.IsErrMap(err)
		if ok {
			switch respErr.Code {
			case models.ErrCodeDatabaseError, models.ErrCodeCacheError, models.ErrCodeExternalServerError:
				xlog.Error(ctx, logMessageService, logStatusError, logError)
			default:
				xlog.Warn(ctx, logMessageService, logStatusError, logError)
			}
			return
		}
		xlog.Warn(ctx, logMessageService, logStatusError, logError)
	} else {
		xlog.Info(ctx, logMessageService, xlog.String("status", "success"))
	}
}

func logDuration(ctx context.Context, message string, now time.Time) {
	xlog.Info(ctx, message, xlog.Duration("elapsed-time", time.Since(now)))
}
