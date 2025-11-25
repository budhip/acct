package services

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/google/uuid"
)

type PublisherService interface {
	PublishMessage(ctx context.Context, in models.PublishRequest) (err error)
}

type publisherService service

var _ PublisherService = (*publisherService)(nil)

func (as *publisherService) PublishMessage(ctx context.Context, in models.PublishRequest) (err error) {
	defer func() {
		logService(ctx, err)
	}()

	records, err := as.readData(in.Message)
	if err != nil {
		return
	}

	if len(records) == 0 {
		err = fmt.Errorf("data is empty")
		return err
	}

	for _, item := range records {
		if err := as.publish(ctx, in.Topic, item); err != nil {
			return err
		}
	}

	return
}

func (as *publisherService) publish(
	ctx context.Context,
	topic string,
	message any,
) error {
	desc := fmt.Sprintf("publish message to %s", topic)
	if err := as.srv.publisher.PublishSyncWithKeyAndLog(
		ctx,
		desc,
		topic,
		uuid.New().String(),
		message,
	); err != nil {
		xlog.Error(ctx, desc, xlog.String("topic", topic), xlog.Any("message", message), xlog.Err(err))
		return err
	}
	return nil
}

func (as *publisherService) readData(file *multipart.FileHeader) (records []map[string]interface{}, err error) {
	src, err := file.Open()
	if err != nil {
		return
	}
	defer src.Close()

	data, err := as.srv.file.ReadAll(src)
	if err != nil {
		return
	}

	if err = json.Unmarshal(data, &records); err != nil {
		return
	}

	return
}
