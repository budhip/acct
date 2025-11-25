export GOPRIVATE := bitbucket.org/Amartha
export GOOGLE_APPLICATION_CREDENTIALS := credentials.json
export USE_DB_MIGRATION := false

docker-start:
	docker-compose up -d

docker-stop:
	docker-compose down

run-api: tidy error-gen swag-gen 
	CGO_ENABLED=0 go run ./cmd/api/main.go
.PHONY: run-api

run-consumer-account-stream-t24: tidy
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=account_stream_t24
.PHONY: run-consumer-account-stream-t24

run-consumer-account-stream-t24-dlq:
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=account_stream_t24_dlq
.PHONY: run-consumer-account-stream-t24-dlq

run-consumer-journal:
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=journal_stream
.PHONY: run-consumer-journal

run-consumer-journal-mgr:
    CGO_ENABLED=0 USE_DB_MIGRATION="true" go run ./cmd/consumer/main.go run -n=journal_stream
.PHONY: run-consumer-journal-mgr

run-consumer-journal-dlq: 
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=journal_stream_dlq
.PHONY: run-consumer-journal-dlq

run-consumer-journal-entry-created-dlq:
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=journal_entry_created_dlq
.PHONY: run-consumer-journal-entry-created-dlq

run-consumer-notification:
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=notification_stream
.PHONY: run-consumer-notification

run-consumer-account-migration:
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=account_migration
.PHONY: run-consumer-account-migration

run-consumer-account-relationships-migration:
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=account_relationships_migration
.PHONY: run-consumer-account-relationships-migration

run-consumer-pas-account-stream:
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=pas_account_stream
.PHONY: run-consumer-pas-account-stream

run-consumer-pas-account-stream-mgr:
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=pas_account_stream_mgr
.PHONY: run-consumer-pas-account-stream

run-consumer-customer_updated_stream:
	CGO_ENABLED=0 go run ./cmd/consumer/main.go run -n=customer_updated_stream
.PHONY: run-consumer-customer_updated_stream

run-job-generate-account-balance-daily-transaction:
	CGO_ENABLED=0 go run cmd/job/main.go run -v=v1 -n=GenerateAccountBalanceDailyTransaction
.PHONY: run-job-generate-account-balance-daily-transaction

run-job-generate-account-trial-balance-daily:
	CGO_ENABLED=0 go run cmd/job/main.go run -v=v1 -n=GenerateAccountTrialBalanceDaily
.PHONY: run-job-generate-account-trial-balance-daily

error-gen: 
	CGO_ENABLED=0 go run ./cmd/errorgen/main.go

tidy:
	go mod tidy
	go mod download
.PHONY: tidy

prepare-release: swag-gen error-gen test

swag-prepare:
	go install github.com/swaggo/swag/cmd/swag@latest

swag-gen:
	swag init --parseDependency --parseInternal -g internal/deliveries/http/router.go --output docs/
.PHONY: swag

mock-prepare:
	go install go.uber.org/mock/mockgen@latest

mock-gen:
	@./scripts/generate-mock.sh repositories
	@./scripts/generate-mock.sh services
	@./scripts/generate-mock.sh acuanclient
	@./scripts/generate-mock.sh dddnotification
	@./scripts/generate-mock.sh file
	@./scripts/generate-mock.sh flag
	@./scripts/generate-mock.sh godbledger
	@./scripts/generate-mock.sh goigate
	@./scripts/generate-mock.sh kafka
	@./scripts/generate-mock.sh queueunicorn
	@./scripts/generate-mock.sh gofptransaction

test_files:= $(shell go list ./internal/... | grep -v /mock)

test: mock-gen
	CGO_ENABLED=1 GOPRIVATE=bitbucket.org/Amartha go test --count=1 -short -race -cover $(test_files)

test-cover: mock-gen
	CGO_ENABLED=1 GOPRIVATE=bitbucket.org/Amartha go test --count=1 -short -race -coverprofile=./cov.out $(test_files)

test-cover-display: mock-prepare test-cover
	go tool cover -html=cov.out

lint-prepare:
	$(eval GOPATH := $(shell go env GOPATH))
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.53.1

lint:
	golangci-lint run --out-format checkstyle > lint.xml

