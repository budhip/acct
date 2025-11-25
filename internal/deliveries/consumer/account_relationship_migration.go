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

type AccountRelationshipMigrationSubscriber struct {
	ctx        context.Context
	contract   *contract.Contract
	subscriber messaging.Subscriber
}

func newAccountRelationshipMigrationSubscriber(ctx context.Context, contract *contract.Contract) (*AccountRelationshipMigrationSubscriber, graceful.ProcessStopper, error) {
	sub, stopper, err := kafka.NewSubscriber(
		contract.Config,
		contract.NewRelic,
		contract.Config.Kafka.Consumers.AccountRelationshipMigration.ConsumerGroup,
		contract.Metrics,
	)
	if err != nil {
		return nil, stopper, err
	}

	b := &AccountRelationshipMigrationSubscriber{
		ctx:        ctx,
		contract:   contract,
		subscriber: sub,
	}

	return b, stopper, nil
}

func (as *AccountRelationshipMigrationSubscriber) Start() graceful.ProcessStarter {
	return func() error {
		return as.run()
	}
}

func (as *AccountRelationshipMigrationSubscriber) run() error {
	if err := as.subscriber.Subscribe(as.ctx,
		messaging.WithTopic(
			as.contract.Config.Kafka.Consumers.AccountRelationshipMigration.Topic,
			codec.NewJson("v1"),
			as.handlerConsumer),
	); err != nil {
		return models.GetErrMap(models.ErrKeyFailedSubscribingKafka, err.Error())
	}

	return nil
}

func (as *AccountRelationshipMigrationSubscriber) handlerConsumer(message messaging.Message) messaging.Response {
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

	if err = as.contract.Service.Account.ConsumerCreateAccountRelationship(ctx, request); err != nil {
		return messaging.ReportError(err, request)
	}

	return messaging.Done(request)
}
