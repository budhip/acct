package contract

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/config"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/acuanclient"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/dbutil"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/dddnotification"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/file"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/flag"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/godbledger"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/gofptransaction"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/goigate"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/graceful"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/kafka"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/metrics"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/queueunicorn"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/bq"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/cache"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/gcs"
	"bitbucket.org/Amartha/go-accounting/internal/repositories/mysql"
	"bitbucket.org/Amartha/go-accounting/internal/services"
	"bitbucket.org/Amartha/go-x/environment"

	storageR "bitbucket.org/Amartha/go-accounting/internal/repositories/storage"
	xlog "bitbucket.org/Amartha/go-x/log"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/storage"
	"github.com/darcys22/godbledger/godbledger/db/mysqldb"
	"github.com/newrelic/go-agent/v3/integrations/nrzap"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/redis/go-redis/v9"
)

const (
	// defaultMaxOpen is default value for max open connection
	defaultMaxOpen = 20
	// defaultMaxIdle is default value for max idle connection
	defaultMaxIdle = 5
	// defaultMaxLifetime is default value for max connection lifetime in minutes
	defaultMaxLifetime = 5 * time.Minute
)

type Contract struct {
	Config   *config.Configuration
	NewRelic *newrelic.Application
	DB       dbutil.DB
	Cache    *redis.Client
	Flagger  flag.FlaggerClient
	Service  *services.Services
	Metrics  metrics.Metrics
}

func New(ctx context.Context) (c *Contract, stopper graceful.ProcessStopper, err error) {
	var stoppers []graceful.ProcessStopper
	stopper = func(ctx context.Context) error {
		for _, st := range stoppers {
			err := st(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	}

	cfg := &config.Configuration{}
	err = config.Load(ctx, cfg)
	if err != nil {
		return c, stopper, fmt.Errorf("failed load config: %v", err)
	}

	level := xlog.DebugLogLevel()
	if environment.ToEnvironment(cfg.App.Env) == environment.PROD_ENV {
		level = xlog.InfoLogLevel()
	}
	xlog.Init(cfg.App.Name,
		xlog.WithLogToOption(cfg.App.LogOption),
		xlog.WithLogEnvOption(cfg.App.Env),
		xlog.WithCaller(true),
		xlog.AddCallerSkip(2),
		level,
	)
	stoppers = append(stoppers, func(ctx context.Context) error {
		xlog.Sync()
		return nil
	})

	projectID := cfg.GcloudProjectID
	if projectID == "" {
		projectID, _ = metadata.ProjectID()
		cfg.GcloudProjectID = projectID
		xlog.Warn(ctx, "can not determine google cloud project, for local use set the gcloud_project_id in config yaml")
	}

	if os.Getenv("APP_TYPE") == "job" {
		cfg.MySQL = cfg.MySQLJob
	}

	newRelic := setupNR(ctx, cfg)

	// connect to mysql db
	db, err := setupDB(cfg)
	if err != nil {
		err = fmt.Errorf("failed setup mysql: %v", err)
		return
	}
	if err = db.PingContext(ctx); err != nil {
		err = fmt.Errorf("failed connect to mysql: %v", err)
		return
	}
	stoppers = append(stoppers, func(ctx context.Context) error {
		return db.Close()
	})

	// connect to redis
	redis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Db,
	})
	if _, err = redis.Ping(ctx).Result(); err != nil {
		err = fmt.Errorf("failed connect to redis: %v", err)
		return
	}
	stoppers = append(stoppers, func(ctx context.Context) error {
		return redis.Close()
	})

	// register metrics
	mtc := metrics.New()
	if mtc != nil {
		mtc.RegisterDB(db.Primary(), cfg.App.Name, "amartha", cfg.MySQL.DbName)
		for _, dbReplica := range db.Replicas() {
			mtc.RegisterDB(dbReplica, cfg.App.Name, "amartha", cfg.MySQL.DbName)
		}
		mtc.RegisterRedis(redis, cfg.App.Name, "amartha")
	}

	flagger, err := flag.New(cfg)
	if err != nil {
		return c, stopper, fmt.Errorf("failed initializing flagger %v", err)
	}
	stoppers = append(stoppers, func(ctx context.Context) error {
		return flagger.Close()
	})

	// connect to gcp storage
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		err = fmt.Errorf("failed connect to storage: %v", err)
		return
	}
	stoppers = append(stoppers, func(ctx context.Context) error {
		return storageClient.Close()
	})

	publisher, pubStopper, err := kafka.NewPublisher(cfg, mtc)
	if err != nil {
		err = fmt.Errorf("error initiate kafka publisher: %w", err)
		return
	}
	stoppers = append(stoppers, pubStopper)

	// register repository
	mySqlRepo := mysql.NewMySQLRepository(db)
	goDBLedger := godbledger.New(&mysqldb.Database{
		DB: db.Primary(),
	})
	cacheRepo := cache.NewCacheRepository(redis)
	storageRepo := storageR.NewStorageRepository(storageClient)
	cloudStorageRepo, err := gcs.NewCloudStorageRepository(cfg)
	if err != nil {
		return
	}
	stoppers = append(stoppers, func(ctx context.Context) error {
		return cloudStorageRepo.Close()
	})

	bqRepo, err := bq.NewCloudStorageRepository(cfg)
	if err != nil {
		return
	}
	stoppers = append(stoppers, func(ctx context.Context) error {
		return bqRepo.Close()
	})

	acuanCilent, err := acuanclient.New(cfg)
	if err != nil {
		return
	}

	queueUnicornClient, err := queueunicorn.New(cfg)
	if err != nil {
		return
	}

	ioFile := file.New()
	igateClient := goigate.New(cfg, newRelic)
	dddNotificationClient := dddnotification.New(cfg, mtc)
	fptransactionClient := gofptransaction.New(cfg.GoFpTransaction, mtc)

	// register service
	service := services.New(cfg, flagger,
		goDBLedger,
		mySqlRepo,
		cacheRepo,
		storageRepo,
		ioFile,
		acuanCilent,
		igateClient,
		dddNotificationClient,
		queueUnicornClient,
		publisher,
		fptransactionClient,
		cloudStorageRepo,
		bqRepo,
	)

	return &Contract{
		Config:   cfg,
		NewRelic: newRelic,
		DB:       db,
		Cache:    redis,
		Flagger:  flagger,
		Service:  service,
		Metrics:  mtc,
	}, stopper, nil
}

func setupDB(conf *config.Configuration) (*dbutil.DbConn, error) {
	createConnectionString := func(host string) string {
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?allowAllFiles=true",
			conf.MySQL.DbUser,
			conf.MySQL.DbPass,
			host,
			conf.MySQL.DbPort,
			conf.MySQL.DbName,
		)
	}

	dbPrimary, err := mysqldb.NewDB(createConnectionString(conf.MySQL.DbHost))
	if err != nil {
		return nil, err
	}

	var dbReplicas []*sql.DB
	if conf.MySQL.DbHostReplication != "" {
		connectionReplica := createConnectionString(conf.MySQL.DbHostReplication)
		dbReplica, errInitReplica := mysqldb.NewDB(connectionReplica)
		if errInitReplica != nil {
			return nil, errInitReplica
		}
		dbReplicas = append(dbReplicas, dbReplica.DB)
	}

	db := dbutil.New(dbPrimary.DB, dbReplicas...)

	db.SetMaxOpenConns(defaultMaxOpen)
	db.SetMaxIdleConns(defaultMaxIdle)
	db.SetConnMaxLifetime(defaultMaxLifetime)

	c := conf.MySQL
	if c.MaxOpenConnection > 0 {
		db.SetMaxOpenConns(conf.MySQL.MaxOpenConnection)
	}
	if c.MaxIdleConnection > 0 {
		db.SetMaxIdleConns(conf.MySQL.MaxIdleConnection)
	}
	if c.MaxLifetime > 0 {
		db.SetConnMaxLifetime(conf.MySQL.MaxLifetime)
	}

	return db, nil
}

func setupNR(ctx context.Context, cfg *config.Configuration) *newrelic.Application {
	if os.Getenv("APP_TYPE") == "job" {
		return nil
	}

	if env := environment.ToEnvironment(cfg.App.Env); env == environment.PROD_ENV {
		logger, ok := xlog.Loggers.Load(xlog.DefaultLogger)
		if !ok {
			return nil
		}
		app, err := newrelic.NewApplication(
			newrelic.ConfigAppName(cfg.App.Name),
			newrelic.ConfigLicense(cfg.NewRelicLicenseKey),
			func(config *newrelic.Config) {
				config.Logger = nrzap.Transform(logger)
				config.ErrorCollector.ExpectStatusCodes = []int{
					http.StatusBadRequest,
					http.StatusUnauthorized,
					http.StatusForbidden,
					http.StatusNotFound,
					http.StatusMethodNotAllowed,
					http.StatusConflict,
					http.StatusRequestEntityTooLarge,
					http.StatusUnprocessableEntity,
				}
			},
			newrelic.ConfigDistributedTracerEnabled(true),
		)
		if err != nil {
			xlog.Errorf(ctx, "setupNR.NewApplication - %v", err)
		}
		if err = app.WaitForConnection(15 * time.Second); nil != err {
			xlog.Errorf(ctx, "setupNR.WaitForConnection - %v", err)
		}
		return app
	}
	return nil
}
