package dddnotification

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/config"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/metrics"
	xlog "bitbucket.org/Amartha/go-x/log"
	"bitbucket.org/Amartha/go-x/log/ctxdata"

	"github.com/go-resty/resty/v2"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type DDDNotificationClient interface {
	SendEmail(ctx context.Context, request RequestEmail) error
	SendMessageToSlack(ctx context.Context, message MessageData) error
}

type client struct {
	cfg        *config.Configuration
	httpClient *resty.Client
	metrics    metrics.Metrics
}

func New(cfg *config.Configuration, mtc metrics.Metrics) DDDNotificationClient {
	retryWaitTime := time.Duration(cfg.DDDNotification.RetryWaitTime) * time.Millisecond
	restyClient := resty.New().
		SetRetryCount(cfg.DDDNotification.RetryCount).
		SetRetryWaitTime(retryWaitTime)
	restyClient.SetTransport(newrelic.NewRoundTripper(restyClient.GetClient().Transport))

	return &client{cfg: cfg, httpClient: restyClient, metrics: mtc}
}

func (c *client) SendEmail(ctx context.Context, request RequestEmail) error {
	timeStart := time.Now()

	path := "/api/v1/email/mandrill"
	url := fmt.Sprintf("%s%s", c.cfg.DDDNotification.BaseURL, path)

	xlog.Infof(ctx, "send request to %s with body %v", url, request)

	resp, err := c.httpClient.R().SetContext(ctx).
		SetHeader("Accept", "application/json;  charset=utf-8").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("X-Correlation-Id", ctxdata.GetCorrelationId(ctx)).
		SetHeader("User-Agent", c.cfg.App.Name).
		SetBody(request).
		Post(url)
	if err != nil {
		return fmt.Errorf("error send request to %s: %w", url, err)
	}

	defer func() {
		if c.metrics != nil {
			c.metrics.GetHTTPClientPrometheus().Record(time.Since(timeStart), SERVICE_NAME, resp.Request.Method, url, resp.StatusCode())
		}
	}()

	var response *ResponseSendMessage
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		return fmt.Errorf("error unmarshal response from %s: %w", url, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("error response from %s: %s", url, response.Message)
	}

	return nil
}

func (c *client) SendMessageToSlack(ctx context.Context, message MessageData) error {
	timeStart := time.Now()

	path := "/api/v1/slack/send-message"
	url := fmt.Sprintf("%s%s", c.cfg.DDDNotification.BaseURL, path)

	payload := PayloadNotification{
		Title:        c.cfg.DDDNotification.TitleBot,
		Service:      c.cfg.App.Name,
		SlackChannel: c.cfg.DDDNotification.SlackChannel,
		Data:         message,
	}

	xlog.Infof(ctx, "send request to %s with body %v", url, payload)

	resp, err := c.httpClient.R().SetContext(ctx).
		SetHeader("Accept", "application/json;  charset=utf-8").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("X-Correlation-Id", ctxdata.GetCorrelationId(ctx)).
		SetHeader("User-Agent", c.cfg.App.Name).
		SetBody(payload).
		Post(url)
	if err != nil {
		return fmt.Errorf("error send request to %s: %w", url, err)
	}

	defer func() {
		if c.metrics != nil {
			c.metrics.GetHTTPClientPrometheus().Record(time.Since(timeStart), SERVICE_NAME, resp.Request.Method, url, resp.StatusCode())
		}
	}()

	var response *ResponseSendMessage
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		return fmt.Errorf("error unmarshal response from %s: %w", url, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("error response from %s: %s", url, response.Message)
	}

	return nil
}
