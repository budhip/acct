package services

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/dddnotification"
)

type DLQProcessorService interface {
	SendNotificationJournalFailure(ctx context.Context, message models.JournalError) (err error)
	SendNotificationAccountT24Failure(ctx context.Context, message models.AccountError) (err error)
}

type dlqProcessor service

var _ DLQProcessorService = (*dlqProcessor)(nil)

func (dlq *dlqProcessor) SendNotificationJournalFailure(ctx context.Context, message models.JournalError) (err error) {
	return dlq.srv.dddNotificationClient.SendMessageToSlack(ctx, dddnotification.MessageData{
		Operation: "Process Journal",
		Message: fmt.Sprintf("<!channel> Failed with trx id: %s, trx type: %s, err: %v",
			message.TransactionId,
			message.Transactions[0].TransactionType,
			message.ErrCauser),
	})
}

func (dlq *dlqProcessor) SendNotificationAccountT24Failure(ctx context.Context, message models.AccountError) (err error) {
	return dlq.srv.dddNotificationClient.SendMessageToSlack(ctx, dddnotification.MessageData{
		Operation: "Process Account to T24",
		Message:   fmt.Sprintf("<!channel> Failed with acc number: %s. err: %v", message.AccountNumber, message.ErrCauser),
	})
}
