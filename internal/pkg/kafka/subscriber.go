package kafka

import (
	"context"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/config"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/graceful"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/metrics"

	"bitbucket.org/Amartha/go-payment-lib/messaging"
	"bitbucket.org/Amartha/go-payment-lib/messaging/kafka"
	"bitbucket.org/Amartha/go-payment-lib/messaging/kafka/middleware"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/newrelic/go-agent/v3/newrelic"
)

var logField = "[KAFKA-CONSUMER]"

type Subscriber interface {
	messaging.Subscriber
}

func NewSubscriber(cfg *config.Configuration, nr *newrelic.Application, consumerGroup string, mtc metrics.Metrics) (Subscriber, graceful.ProcessStopper, error) {
	stopper := func(context.Context) error { return nil }

	defaultOpts := []kafka.SubscriberOption{
		kafka.WithSubscriberKafkaVersion("3.3.1"),
		kafka.WithMiddleware(
			middleware.Context,
			middlewareMessageClaim,
			// middleware.Newrelic(consumerGroup, nr),
		),
		kafka.WithSubscriberPreStartCallback(preStart),
		kafka.WithSubscriberEndCallback(end),
		kafka.WithSubscriberErrorCallback(errcb),
		kafka.WithSubscriberSessionTimeout(5 * time.Minute),
	}
	if mtc != nil {
		defaultOpts = append(
			defaultOpts,
			kafka.WithSubscriberGenericPromMetrics(mtc.PrometheusRegisterer(), cfg.App.Name, "subscriber", 1*time.Second),
		)
	}

	sub, err := kafka.NewSubscriber(
		cfg.Kafka.Brokers,
		consumerGroup,
		// kafka.WithSubscriberInitialOffset(kafka.OffsetNewest),
		defaultOpts...,
	)
	if err != nil {
		return nil, stopper, err
	}

	stopper = func(ctx context.Context) error {
		return sub.CloseSubscriber()
	}

	return sub, stopper, nil
}

func GetMessageClaim(message messaging.Message) kafka.MessageClaim {
	v, ok := message.GetMessageClaim().(kafka.MessageClaim)
	if !ok {
		return nil
	}
	return v
}

func middlewareSetContext(next messaging.SubscriptionHandler) messaging.SubscriptionHandler {
	return middleware.Context(next)
}

func middlewareSetNewrelic(name string, newRelicApp *newrelic.Application) messaging.MiddlewareFunc {
	return middleware.Newrelic(name, newRelicApp)
}

func middlewareMessageClaim(next messaging.SubscriptionHandler) messaging.SubscriptionHandler {
	return func(message messaging.Message) messaging.Response {
		v := GetMessageClaim(message)
		if v != nil {
			logField := []xlog.Field{
				xlog.Time("timestamp", v.Timestamp),
				xlog.Any("block-timestamp", v.BlockTimestamp),
				xlog.String("topic", v.Topic),
				xlog.Int32("partition", v.Partition),
				xlog.Int64("offset", v.Offset),
				xlog.Any("header", v.Headers),
				xlog.String("key", string(v.Key)),
				xlog.String("message-claimed", string(v.Value)),
			}
			xlog.Info(message.Context(), "[MESSAGE-CLAIM]", logField...)
		}
		return next(message)
	}
}

func preStart(ctx context.Context, claims map[string][]int32) error {
	xlog.Info(ctx, logField, xlog.String("status", "start"), xlog.Any("claims", claims))
	return nil
}

func end(ctx context.Context, claims map[string][]int32) error {
	xlog.Info(ctx, logField, xlog.String("status", "end"), xlog.Any("claims", claims))
	return nil
}

func errcb(err error) {
	xlog.Info(context.Background(), logField, xlog.Err(err))
}
