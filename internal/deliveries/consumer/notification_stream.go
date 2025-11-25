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

type NotificationStreamSubscriber struct {
	ctx        context.Context
	contract   *contract.Contract
	subscriber messaging.Subscriber
}

func newNotificationStreamSubscriber(ctx context.Context, contract *contract.Contract) (*NotificationStreamSubscriber, graceful.ProcessStopper, error) {
	sub, stopper, err := kafka.NewSubscriber(
		contract.Config,
		contract.NewRelic,
		contract.Config.Kafka.Consumers.NotificationStream.ConsumerGroup,
		contract.Metrics,
	)
	if err != nil {
		return nil, stopper, err
	}

	b := &NotificationStreamSubscriber{
		ctx:        ctx,
		contract:   contract,
		subscriber: sub,
	}

	return b, stopper, nil
}

func (ns *NotificationStreamSubscriber) Start() graceful.ProcessStarter {
	return func() error {
		return ns.run()
	}
}

func (ns *NotificationStreamSubscriber) run() error {
	if err := ns.subscriber.Subscribe(ns.ctx,
		messaging.WithTopic(
			ns.contract.Config.Kafka.Consumers.JournalStreamDLQ.Topic,
			codec.NewJson("v1"),
			ns.handlerJournalStreamDLQConsumer),
		messaging.WithTopic(
			ns.contract.Config.Kafka.Consumers.AccountStreamT24DLQ.Topic,
			codec.NewJson("v1"),
			ns.handlerAccountStreamT24DLQConsumer),
	); err != nil {
		return models.GetErrMap(models.ErrKeyFailedSubscribingKafka, err.Error())
	}
	return nil
}

func (ns *NotificationStreamSubscriber) handlerJournalStreamDLQConsumer(message messaging.Message) messaging.Response {
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
		return messaging.ReportError(err, nil)
	}

	if err = ns.contract.Service.DLQProcessor.SendNotificationJournalFailure(ctx, request); err != nil {
		return messaging.ReportError(err, request)
	}

	return messaging.Done(request)
}

func (ns *NotificationStreamSubscriber) handlerAccountStreamT24DLQConsumer(message messaging.Message) messaging.Response {
	ctx := message.Context()
	var (
		request models.AccountError
		err     error
	)

	// defer func() {
	// 	logConsumerProcess(ctx, request, err)
	// }()

	if err = message.Bind(&request); err != nil {
		err = models.GetErrMap(models.ErrKeyFailedBindingPayload, err.Error())
		return messaging.ExpectError(err, nil)
	}

	if err = ns.contract.Service.DLQProcessor.SendNotificationAccountT24Failure(ctx, request); err != nil {
		return messaging.ReportError(err, request)
	}

	return messaging.Done(request)
}
