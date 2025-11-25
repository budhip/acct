package config

import (
	"context"
	"time"

	confloader "bitbucket.org/Amartha/go-config-loader-library"
)

type (
	Configuration struct {
		App                  App                         `json:"app"`
		MySQL                MySQL                       `json:"mysql"`
		MySQLJob             MySQL                       `json:"mysql_job"`
		MySQLMigration       MySQL                       `json:"mysql_migration"` // TODO: remove after testing for migration
		Redis                Redis                       `json:"redis"`
		Kafka                KafkaConfiguration          `json:"kafka"`
		AccountConfig        AccountConfig               `json:"account_config"`
		CacheTTL             CacheTTL                    `json:"cache_ttl"`
		FeatureFlag          FeatureFlag                 `json:"feature_flag"`
		FeatureFlagMigration FeatureFlag                 `json:"feature_flag_migration"` // TODO: remove after testing for migration
		JournalConfig        JournalConfig               `json:"journal_config"`
		IGateClient          IGateClient                 `json:"igate_client"`
		DDDNotification      DDDNotification             `json:"ddd_notification"`
		GoQueueUnicorn       HTTPConfiguration           `json:"go_queue_unicorn"`
		GoFpTransaction      HTTPConfiguration           `json:"go_fp_transaction"`
		CloudStorageConfig   CloudStorageConfig          `json:"cloud_storage_config"`
		Migration            MigrationConfiguration      `json:"migration"`
		SQLTransaction       SQLTransactionConfiguration `json:"sql_transaction"`

		GcloudProjectID    string `json:"gcloud_project_id"`
		BigQueryDataset    string `json:"big_query_dataset"`
		NewRelicLicenseKey string `json:"new_relic_license_key"`
		HostGoAccounting   string `json:"host_go_accounting"`
		SecretKey          string `json:"secret_key"`

		AcuanLibConfig  AcuanLibConfig `json:"go_acuan_lib"`
		GQUTrialBalance GQUConfig      `json:"gqu_trial_balance"`
	}

	App struct {
		Env             string        `json:"env"`
		HTTPPort        int           `json:"http_port"`
		HTTPTimeout     time.Duration `json:"http_timeout"`
		GracefulTimeout time.Duration `json:"graceful_timeout"`
		Name            string        `json:"name"`
		Type            string        `json:"type"`
		LogOption       string        `json:"log_option"`
		LogLevel        string        `json:"log_level"`
	}

	MySQL struct {
		DbHost            string        `json:"db_host"`
		DbHostReplication string        `json:"db_host_replication"`
		DbPort            string        `json:"db_port"`
		DbUser            string        `json:"db_user"`
		DbPass            string        `json:"db_pass"`
		DbName            string        `json:"db_name"`
		MaxOpenConnection int           `json:"max_open_connection"`
		MaxIdleConnection int           `json:"max_idle_connection"`
		MaxLifetime       time.Duration `json:"max_lifetime"`
	}

	Redis struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Password string `json:"password"`
		Db       int    `json:"db"`
	}

	AccountConfig struct {
		AccountNumberPadWidth                  int64                               `json:"account_number_pad_width"`
		ChunkSizeAccountBalance                int                                 `json:"chunk_size_account_balance"`
		IsCreateAccountT24                     bool                                `json:"is_create_account_t24"`
		IsT24CreateAccountPAS                  map[string]bool                     `json:"is_t24_create_account_pas"`
		LimitAccountSubLedger                  int                                 `json:"limit_account_sub_ledger"`
		LimitAccountTrialBalance               int                                 `json:"limit_account_trial_balance"`
		InvestedAccountNumber                  map[string]string                   `json:"invested_account_number"`
		ReceivablesAccountNumber               map[string]string                   `json:"receivables_account_number"`
		MultiLoanAccount                       map[string]string                   `json:"multi_loan_account"`
		LenderInstiReceivablesAccount          map[string]string                   `json:"lender_insti_receivables_account"`
		LoanPartnerAccountConfig               map[string]LoanPartnerAccountConfig `json:"loan_partner_account_config"`
		LoanPartnerAccountEntities             []string                            `json:"loan_partner_account_entities"`
		CashInTransitRepaymentEntity           map[string]string                   `json:"cash_in_transit_repayment_entity"`
		AllowedEntitiesAccountDailyBalance     []string                            `json:"allowed_entities_account_daily_balance"`
		IsSkipAllZeroAmountAccountDailyBalance bool                                `json:"is_skip_all_zero_amount_account_daily_balance"`
		IsSequentialAccountDailyBalance        bool                                `json:"is_sequential_account_daily_balance"`
		MaxConcurrentAccountBalance            int                                 `json:"max_concurrent_account_balance"`
	}

	LoanPartnerAccountConfig struct {
		Name            string `json:"name"`
		CategoryCode    string `json:"category_code"`
		SubCategoryCode string `json:"sub_category_code"`
		AccountType     string `json:"account_type"`
		Currency        string `json:"currency"`
	}

	JournalConfig struct {
		SplitIdPadWidth int64 `json:"split_id_pad_width"`
	}

	IGateClient struct {
		BaseURL        string `json:"base_url"`
		RequestPerSec  int    `json:"request_per_sec"`
		MaxRetry       int    `json:"max_retry"`
		IsMockResponse bool   `json:"is_mock_response"`
		SecretKey      string `json:"secret_key"`
	}

	DDDNotification struct {
		BaseURL       string `json:"base_url"`
		RetryCount    int    `json:"retry_count"`
		RetryWaitTime int    `json:"retry_wait_time"`

		SlackChannel                    string `json:"slack_channel"`
		TitleBot                        string `json:"title_bot"`
		EmailTemplateSubLedger          string `json:"email_template_sub_ledger"`
		EmailTemplateTrialBalance       string `json:"email_template_trial_balance"`
		EmailTemplateTrialBalanceDetail string `json:"email_template_trial_balance_detail"`
	}

	HTTPConfiguration struct {
		BaseURL       string        `json:"base_url"`
		SecretKey     string        `json:"secret_key"`
		RetryCount    int           `json:"retry_count"`
		RetryWaitTime int           `json:"retry_wait_time"`
		Timeout       time.Duration `json:"timeout"`
	}

	CacheTTL struct {
		GetOneByAccountNumber              time.Duration `json:"get_one_by_account_number"`
		GetAccounts                        time.Duration `json:"get_accounts"`
		GetLenderAccountByCIHAccountNumber time.Duration `json:"get_lender_account_by_cih_account_number"`
		GetLoanAdvanceAccountByLoanAccount time.Duration `json:"get_loan_advance_account_by_loan_account"`
		GetSubLedgerAccountsCount          time.Duration `json:"get_sub_ledger_accounts_count"`
		GetAccountListCount                time.Duration `json:"get_account_list_count"`
		GetTrialBalanceDetailsCount        time.Duration `json:"get_trial_balance_detail_count"`
		GetLoanPartnerAccount              time.Duration `json:"get_loan_partner_account"`
		GetEntity                          time.Duration `json:"get_entity"`
	}

	FeatureFlag struct {
		URL             string        `json:"url"`
		Token           string        `json:"token"`
		Env             string        `json:"env"`
		RefreshInterval time.Duration `json:"refresh_interval"`
	}

	KafkaConfiguration struct {
		HealthCheckPort int             `json:"health_check_port"`
		Brokers         []string        `json:"brokers"`
		Publishers      KafkaPublishers `json:"publishers"`
		Consumers       KafkaConsumers  `json:"consumers"`
	}

	KafkaPublishers struct {
		JournalStream             KafkaPublisherConfiguration `json:"journal_stream"`
		JournalStreamMigration    KafkaPublisherConfiguration `json:"journal_stream_migration"`
		JournalStreamDLQ          KafkaPublisherConfiguration `json:"journal_stream_dlq"`
		JournalStreamDLQMigration KafkaPublisherConfiguration `json:"journal_stream_dlq_migration"`
		AccountStreamT24          KafkaPublisherConfiguration `json:"account_stream_t24"`
		AccountStreamT24DLQ       KafkaPublisherConfiguration `json:"account_stream_t24_dlq"`
		PASAccountStream          KafkaPublisherConfiguration `json:"pas_account_stream"`
		PASAccountStreamDLQ       KafkaPublisherConfiguration `json:"pas_account_stream_dlq"`
		JournalEntryCreated       KafkaPublisherConfiguration `json:"journal_entry_created"`
		JournalEntryCreatedDLQ    KafkaPublisherConfiguration `json:"journal_entry_created_dlq"`

		PASAccountStreamMigration    KafkaPublisherConfiguration `json:"pas_account_stream_migration"`
		PASAccountStreamMigrationDLQ KafkaPublisherConfiguration `json:"pas_account_stream_migration_dlq"`
		AccountMigrationStreamDLQ    KafkaPublisherConfiguration `json:"account_migration_stream_dlq"`
	}
	KafkaPublisherConfiguration struct {
		Topic string `json:"topic"`
	}

	KafkaConsumers struct {
		JournalStream          KafkaConsumerConfiguration `json:"journal_stream"`
		JournalStreamMigration KafkaConsumerConfiguration `json:"journal_stream_migration"`
		JournalStreamDLQ       KafkaConsumerConfiguration `json:"journal_stream_dlq"`
		AccountStreamT24       KafkaConsumerConfiguration `json:"account_stream_t24"`
		AccountStreamT24DLQ    KafkaConsumerConfiguration `json:"account_stream_t24_dlq"`
		NotificationStream     KafkaConsumerConfiguration `json:"notification_stream"`
		PASAccountStream       KafkaConsumerConfiguration `json:"pas_account_stream"`
		JournalEntryCreatedDLQ KafkaConsumerConfiguration `json:"journal_entry_created_dlq"`

		PASAccountStreamMigration    KafkaConsumerConfiguration `json:"pas_account_stream_migration"`
		AccountMigrationStream       KafkaConsumerConfiguration `json:"account_migration_stream"`
		AccountRelationshipMigration KafkaConsumerConfiguration `json:"account_relationships_migration"`
		CustomerUpdatedStream        KafkaConsumerConfiguration `json:"customer_updated_stream"`
	}

	KafkaConsumerConfiguration struct {
		Topic         string `json:"topic"`
		ConsumerGroup string `json:"consumer_group"`
	}

	CloudStorageConfig struct {
		BaseURL                       string        `json:"base_url"`
		BucketName                    string        `json:"bucket_name"`
		TrialBalanceDetailURLDuration time.Duration `json:"trial_balance_detail_url_duration"`
		SubLedgerURLDuration          time.Duration `json:"sub_ledger_url_duration"`
	}

	MigrationConfiguration struct {
		Buckets string `json:"buckets"`
	}

	SQLTransactionConfiguration struct {
		BulkLimit int `json:"bulk_limit"`
	}

	AcuanLibConfig struct {
		Kafka                 AcuanLibKafkaConfig `json:"kafka"`
		SourceSystem          string              `json:"source_system"`
		Topic                 string              `json:"topic"`
		TopicAccounting       string              `json:"topic_accounting"`
		TopUpKey              string              `json:"topup_key"`
		InvestmentKey         string              `json:"investment_key"`
		CashoutKey            string              `json:"cashout_key"`
		DisbursementKey       string              `json:"disbursement_key"`
		DisbursementFailedKey string              `json:"disbursement_failed_key"`
		RepaymentKey          string              `json:"repayment_key"`
		RefundKey             string              `json:"refund_key"`
	}

	AcuanLibKafkaConfig struct {
		BrokerList        string `json:"broker_list"`
		PartitionStrategy string `json:"partition_strategy"`
	}
	GQUConfig struct {
		ProcessIn int `json:"process_in"`
		MaxRetry  int `json:"max_retry"`
	}
)

func Load(ctx context.Context, cfg *Configuration) error {
	loader := confloader.New("", "", "",
		confloader.WithConfigFileSearchPaths(
			"./config",
			"./../config",
			"./../../config"),
	)
	return loader.Load(cfg)
}
