package consumer

import (
	"context"

	"bitbucket.org/Amartha/go-accounting/internal/contract"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/graceful"

	xlog "bitbucket.org/Amartha/go-x/log"
)

const (
	journalStream                = "journal_stream"
	journalStreamDLQ             = "journal_stream_dlq"
	notificationStream           = "notification_stream"
	pasAccountStream             = "pas_account_stream"
	accountRelationshipMigration = "account_relationships_migration"
	customerUpdatedStream        = "customer_updated_stream"
	journalEntryCreatedDLQ       = "journal_entry_created_dlq"
)

var (
	logProcessMessage = "[CONSUMER-HANDLER]"
	ListConsumerName  = []string{
		journalStream,
		journalStreamDLQ,
		notificationStream,
	}
)

type KafkaConsumer interface {
	Start() graceful.ProcessStarter
}

func NewKafkaConsumer(ctx context.Context, consumerName string, contract *contract.Contract) (KafkaConsumer, graceful.ProcessStopper, error) {
	switch consumerName {
	case journalStream:
		return newJournalStreamSubscriber(ctx, contract)
	case journalStreamDLQ:
		return newJournalStreamDLQSubscriber(ctx, contract)
	case notificationStream:
		return newNotificationStreamSubscriber(ctx, contract)
	case accountRelationshipMigration:
		return newAccountRelationshipMigrationSubscriber(ctx, contract)
	case pasAccountStream:
		return newPASAccountStreamSubscriber(ctx, contract)
	case customerUpdatedStream:
		return newCustomerUpdatedStreamSubscriber(ctx, contract)
	case journalEntryCreatedDLQ:
		return newJournalEntryCreatedDLQSubscriber(ctx, contract)
	default:
		xlog.Error(ctx, "invalid consumer instance name")
	}
	return nil, nil, nil
}

func logConsumerProcess(ctx context.Context, request any, err error) {
	if err != nil {
		xlog.Error(ctx, logProcessMessage, xlog.String("status", "failed"), xlog.Any("message", request), xlog.Err(err))
	} else {
		xlog.Info(ctx, logProcessMessage, xlog.String("status", "success"), xlog.Any("message", request))
	}
}
