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

type PASAccountStreamSubscriber struct {
	ctx        context.Context
	contract   *contract.Contract
	subscriber messaging.Subscriber
}

func newPASAccountStreamSubscriber(ctx context.Context, contract *contract.Contract) (*PASAccountStreamSubscriber, graceful.ProcessStopper, error) {
	sub, stopper, err := kafka.NewSubscriber(
		contract.Config,
		contract.NewRelic,
		contract.Config.Kafka.Consumers.PASAccountStream.ConsumerGroup,
		contract.Metrics,
	)
	if err != nil {
		return nil, stopper, err
	}

	b := &PASAccountStreamSubscriber{
		ctx:        ctx,
		contract:   contract,
		subscriber: sub,
	}

	return b, stopper, nil
}

func (s *PASAccountStreamSubscriber) Start() graceful.ProcessStarter {
	return func() error {
		return s.run()
	}
}

func (s *PASAccountStreamSubscriber) run() error {
	if err := s.subscriber.Subscribe(s.ctx,
		messaging.WithTopic(
			s.contract.Config.Kafka.Consumers.PASAccountStream.Topic,
			codec.NewJson("v1"),
			s.handlerConsumer),
	); err != nil {
		return models.GetErrMap(models.ErrKeyFailedSubscribingKafka, err.Error())
	}

	return nil
}

func (s *PASAccountStreamSubscriber) handlerConsumer(message messaging.Message) messaging.Response {
	ctx := message.Context()
	var (
		request models.CreateAccount
		err     error
	)

	// defer func() {
	// 	logConsumerProcess(ctx, request, err)
	// }()

	if err = message.Bind(&request); err != nil {
		err = models.GetErrMap(models.ErrKeyFailedBindingPayload, err.Error())
		return messaging.ExpectError(err, nil)
	}

	if err = s.contract.Service.Account.ConsumerAccountStream(ctx, request); err != nil {
		return messaging.ReportError(err, request)
	}

	return messaging.Done(request)
}
