package gofptransaction

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
)

var logMessage = "[FP-TRANSACTION-CLIENT]"

type FpTransactionClient interface {
	UpdateAccountBySubCategory(ctx context.Context, in UpdateBySubCategory) (err error)
	GetAllOrderTypes(ctx context.Context, orderTypeCode, orderTypeName string) (GetAllOrderTypesOut, error)
}

type client struct {
	baseURL    string
	secretKey  string
	httpClient *resty.Client
	metrics    metrics.Metrics
}

func New(configuration config.HTTPConfiguration, metrics metrics.Metrics) FpTransactionClient {
	retryWaitTime := time.Duration(configuration.RetryWaitTime) * time.Millisecond

	restyClient := resty.New().
		SetRetryCount(configuration.RetryCount).
		SetRetryWaitTime(retryWaitTime)

	return client{
		baseURL:    configuration.BaseURL,
		secretKey:  configuration.SecretKey,
		httpClient: restyClient,
		metrics:    metrics,
	}
}

func (c client) UpdateAccountBySubCategory(ctx context.Context, in UpdateBySubCategory) (err error) {
	timeStart := time.Now()
	url := fmt.Sprintf("%s/api/v1/accounts/sub-category/%s", c.baseURL, in.Code)

	logFields := []xlog.Field{
		xlog.String("url", url),
		xlog.Any("request", in),
	}

	defer func() {
		if err != nil {
			xlog.Warn(ctx, logMessage, append(logFields, xlog.Err(err))...)
		}
	}()

	xlog.Info(ctx, logMessage, append(logFields, xlog.String("message", "send request to go_fp_transaction"))...)

	body := map[string]interface{}{}
	if in.ProductTypeName != nil {
		body["productTypeName"] = *in.ProductTypeName
	}
	if in.Currency != nil {
		body["currency"] = *in.Currency
	}

	httpRes, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json;  charset=utf-8").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("X-Correlation-Id", ctxdata.GetCorrelationId(ctx)).
		SetHeader("X-Secret-Key", c.secretKey).
		SetBody(body).
		Patch(url)
	if err != nil {
		return fmt.Errorf("failed send request: %w", err)
	}

	defer func() {
		if c.metrics != nil {
			c.metrics.GetHTTPClientPrometheus().Record(time.Since(timeStart), SERVICE_NAME, httpRes.Request.Method, url, httpRes.StatusCode())
		}
	}()

	logFields = append(logFields,
		xlog.String("httpStatusCode", httpRes.Status()),
		xlog.Any("httpResponse", httpRes.Body()))

	if httpRes.StatusCode() != http.StatusOK {
		return fmt.Errorf("invalid response http code: got %d", httpRes.StatusCode())
	}

	return nil
}

func (c client) GetAllOrderTypes(ctx context.Context, orderTypeCode, orderTypeName string) (out GetAllOrderTypesOut, err error) {
	timeStart := time.Now()
	url := fmt.Sprintf("%s/api/v1/order-types", c.baseURL)

	logFields := []xlog.Field{
		xlog.String("url", url),
		xlog.String("orderTypeCode", orderTypeCode),
		xlog.String("orderTypeName", orderTypeName),
	}

	defer func() {
		if err != nil {
			xlog.Warn(ctx, logMessage, append(logFields, xlog.Err(err))...)
		}
	}()

	xlog.Info(ctx, logMessage, append(logFields, xlog.String("message", "send request to go_fp_transaction"))...)

	httpRes, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json;  charset=utf-8").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("X-Correlation-Id", ctxdata.GetCorrelationId(ctx)).
		SetHeader("X-Secret-Key", c.secretKey).
		SetQueryParam("code", orderTypeCode).
		SetQueryParam("name", orderTypeName).
		Get(url)
	if err != nil {
		return out, fmt.Errorf("failed send request: %w", err)
	}

	defer func() {
		if c.metrics != nil {
			c.metrics.GetHTTPClientPrometheus().Record(time.Since(timeStart), SERVICE_NAME, httpRes.Request.Method, url, httpRes.StatusCode())
		}
	}()

	logFields = append(logFields,
		xlog.String("httpStatusCode", httpRes.Status()),
		xlog.Any("httpResponse", httpRes.Body()))

	if httpRes.StatusCode() != http.StatusOK {
		return out, fmt.Errorf("invalid response http code: got %d", httpRes.StatusCode())
	}

	err = json.Unmarshal(httpRes.Body(), &out)
	if err != nil {
		return out, fmt.Errorf("failed unmarshal response: %w", err)
	}

	return
}
