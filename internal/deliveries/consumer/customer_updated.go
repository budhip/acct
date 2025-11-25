package consumer

import (
	"context"
	"errors"

	"bitbucket.org/Amartha/go-accounting/internal/contract"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/gocustomer"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/graceful"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/kafka"

	"bitbucket.org/Amartha/go-payment-lib/messaging"
	"bitbucket.org/Amartha/go-payment-lib/messaging/codec"
	kafkaLib "bitbucket.org/Amartha/go-payment-lib/messaging/kafka"
)

type CustomerUpdatedSubscriber struct {
	ctx        context.Context
	contract   *contract.Contract
	subscriber messaging.Subscriber
}

func newCustomerUpdatedStreamSubscriber(ctx context.Context, contract *contract.Contract) (*CustomerUpdatedSubscriber, graceful.ProcessStopper, error) {
	sub, stopper, err := kafka.NewSubscriber(
		contract.Config,
		contract.NewRelic,
		contract.Config.Kafka.Consumers.CustomerUpdatedStream.ConsumerGroup,
		contract.Metrics,
	)
	if err != nil {
		return nil, stopper, err
	}

	b := &CustomerUpdatedSubscriber{
		ctx:        ctx,
		contract:   contract,
		subscriber: sub,
	}

	return b, stopper, nil
}

func (as *CustomerUpdatedSubscriber) Start() graceful.ProcessStarter {
	return func() error {
		return as.run()
	}
}

func (as *CustomerUpdatedSubscriber) run() error {
	err := as.subscriber.Subscribe(
		as.ctx,
		messaging.WithTopic(
			as.contract.Config.Kafka.Consumers.CustomerUpdatedStream.Topic,
			codec.NewJson("v1"),
			as.handlerConsumer,
		),
	)
	if err != nil {
		return models.GetErrMap(models.ErrKeyFailedSubscribingKafka, err.Error())
	}

	return nil
}

func (as *CustomerUpdatedSubscriber) handlerConsumer(message messaging.Message) messaging.Response {
	ctx := message.Context()
	var (
		request gocustomer.CustomerEventPayload
		err     error
	)

	// defer func() {
	// 	logConsumerProcess(ctx, request, err)
	// }()

	if err = message.Bind(&request); err != nil {
		err = models.GetErrMap(models.ErrKeyFailedBindingPayload, err.Error())
		return messaging.ExpectError(err, nil)
	}

	mc, ok := message.GetMessageClaim().(kafkaLib.MessageClaim)
	if !ok {
		return messaging.ExpectError(errors.New("unable get message claim"), nil)
	}

	// check header if customer updated
	isCustomerUpdate := hasHeader(mc, "type", "customer_updated")
	if !isCustomerUpdate {
		return messaging.Done(nil)
	}

	// currently we only care to update name from go-customer
	if request.Name == "" {
		return messaging.Done(nil)
	}

	err = as.contract.Service.Account.UpdateAccountByCustomerData(ctx, request)
	if err != nil {
		return messaging.ExpectError(err, nil)
	}

	return messaging.Done(request)
}

func hasHeader(mc kafkaLib.MessageClaim, key string, value string) bool {
	for _, header := range mc.Headers {
		if header != nil {
			if string(header.Key) == key && string(header.Value) == value {
				return true
			}
		}
	}

	return false
}
