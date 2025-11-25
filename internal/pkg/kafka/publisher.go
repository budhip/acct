package kafka

import (
	"context"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/config"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/graceful"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/metrics"

	"bitbucket.org/Amartha/go-payment-lib/messaging"
	"bitbucket.org/Amartha/go-payment-lib/messaging/kafka"
	xlog "bitbucket.org/Amartha/go-x/log"
)

type publisher struct {
	messaging.PublisherWithKeyAndHeader
}

type Publisher interface {
	messaging.PublisherWithKeyAndHeader
	PublishSyncWithKeyAndLog(ctx context.Context, description, topic, key string, data interface{}) error
}

func NewPublisher(cfg *config.Configuration, mtc metrics.Metrics) (Publisher, graceful.ProcessStopper, error) {
	stopper := func(context.Context) error { return nil }

	defaultOpts := []kafka.PublisherOption{kafka.WithOrigin(cfg.App.Name)}
	if mtc != nil {
		defaultOpts = append(defaultOpts, kafka.WithPublisherGenericPromMetrics(mtc.PrometheusRegisterer(), cfg.App.Name, 1*time.Second))
	}

	pub, err := kafka.NewPublisher(cfg.Kafka.Brokers, defaultOpts...)
	if err != nil {
		return nil, stopper, err
	}

	stopper = func(ctx context.Context) error {
		return pub.ClosePublisher()
	}

	p := &publisher{pub}

	return p, stopper, nil
}

func (p *publisher) PublishSyncWithKeyAndLog(ctx context.Context, description, topic, key string, data interface{}) error {
	logMessage := "[PUBLISH-MESSAGE]"
	xlog.Info(ctx, logMessage, xlog.String("key", key), xlog.String("topic", topic), xlog.String("description", description), xlog.Any("data", data))

	if err := p.PublishSyncWithKey(ctx, topic, key, data); err != nil {
		xlog.Error(ctx, logMessage, xlog.String("key", key), xlog.String("topic", topic), xlog.String("description", description), xlog.Any("data", data), xlog.Err(err))
		return err
	}

	return nil
}
