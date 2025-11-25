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

type JournalStreamDLQSubscriber struct {
	ctx        context.Context
	contract   *contract.Contract
	subscriber messaging.Subscriber
}

func newJournalStreamDLQSubscriber(ctx context.Context, contract *contract.Contract) (*JournalStreamDLQSubscriber, graceful.ProcessStopper, error) {
	sub, stopper, err := kafka.NewSubscriber(
		contract.Config,
		contract.NewRelic,
		contract.Config.Kafka.Consumers.JournalStreamDLQ.ConsumerGroup,
		contract.Metrics,
	)
	if err != nil {
		return nil, stopper, err
	}

	b := &JournalStreamDLQSubscriber{
		ctx:        ctx,
		contract:   contract,
		subscriber: sub,
	}

	return b, stopper, nil
}

func (js *JournalStreamDLQSubscriber) Start() graceful.ProcessStarter {
	return func() error {
		return js.run()
	}
}

func (js *JournalStreamDLQSubscriber) run() error {
	if err := js.subscriber.Subscribe(js.ctx,
		messaging.WithTopic(
			js.contract.Config.Kafka.Consumers.JournalStreamDLQ.Topic,
			codec.NewJson("v1"),
			js.handlerConsumer),
	); err != nil {
		return models.GetErrMap(models.ErrKeyFailedSubscribingKafka, err.Error())
	}

	return nil
}

func (js *JournalStreamDLQSubscriber) handlerConsumer(message messaging.Message) messaging.Response {
	ctx := message.Context()
	var (
		request models.JournalError
		err     error
	)

	// defer func() {
	// 	logConsumerProcess(ctx, request, err)
	// }()

	if err = message.Bind(&request); err != nil {
		err = models.GetErrMap(models.ErrKeyFailedBindingPayload, err.Error())
		return messaging.ExpectError(err, nil)
	}

	if err = js.contract.Service.RetryService.RetryInsertJournalTransaction(ctx, request); err != nil {
		return messaging.ReportError(err, request)
	}

	return messaging.Done(request)
}
