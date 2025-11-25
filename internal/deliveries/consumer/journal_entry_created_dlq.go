package consumer

import (
	"context"

	"bitbucket.org/Amartha/go-accounting/internal/contract"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/graceful"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/kafka"
	"bitbucket.org/Amartha/go-payment-lib/messaging"
	"bitbucket.org/Amartha/go-payment-lib/messaging/codec"
)

type JournalEntryCreatedDLQSubscriber struct {
	ctx        context.Context
	contract   *contract.Contract
	subscriber messaging.Subscriber
}

func newJournalEntryCreatedDLQSubscriber(ctx context.Context, contract *contract.Contract) (*JournalEntryCreatedDLQSubscriber, graceful.ProcessStopper, error) {
	sub, stopper, err := kafka.NewSubscriber(
		contract.Config,
		contract.NewRelic,
		contract.Config.Kafka.Consumers.JournalEntryCreatedDLQ.ConsumerGroup,
		contract.Metrics,
	)
	if err != nil {
		return nil, stopper, err
	}

	b := &JournalEntryCreatedDLQSubscriber{
		ctx:        ctx,
		contract:   contract,
		subscriber: sub,
	}

	return b, stopper, nil
}

func (j *JournalEntryCreatedDLQSubscriber) Start() graceful.ProcessStarter {
	return func() error {
		return j.run()
	}
}

func (j *JournalEntryCreatedDLQSubscriber) run() error {
	if err := j.subscriber.Subscribe(j.ctx,
		messaging.WithTopic(
			j.contract.Config.Kafka.Consumers.JournalEntryCreatedDLQ.Topic,
			codec.NewJson("v1"),
			j.handlerConsumer),
	); err != nil {
		return models.GetErrMap(models.ErrKeyFailedSubscribingKafka, err.Error())
	}
	return nil
}

func (j *JournalEntryCreatedDLQSubscriber) handlerConsumer(message messaging.Message) messaging.Response {
	ctx := message.Context()
	var (
		request models.JournalEntryCreatedRequest
		err     error
	)

	if err = message.Bind(&request); err != nil {
		err = models.GetErrMap(models.ErrKeyFailedBindingPayload, err.Error())
		return messaging.ExpectError(err, nil)
	}

	if err = j.contract.Service.Journal.RetryPublishToJournalEntryCreated(ctx, request); err != nil {
		return messaging.ReportError(err, request)
	}

	return messaging.Done(request)
}
