package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/dddnotification"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/queueunicorn"
	xlog "bitbucket.org/Amartha/go-x/log"
)

type RetryService interface {
	RetryInsertJournalTransaction(ctx context.Context, req models.JournalError) (err error)
	RetryUpdateLegacyId(ctx context.Context, req models.AccountError) (err error)
}

type retryService service

var _ RetryService = (*retryService)(nil)

func (rs *retryService) RetryInsertJournalTransaction(ctx context.Context, req models.JournalError) (err error) {
	defer func() {
		logService(ctx, err)
		// send notif to slack
		if err != nil {
			rs.sendMessageToSlack(ctx, "Start process send job retry journal transaction", err.Error())
		}
	}()

	respErr, err := rs.getErrMap(req.ErrCauser)
	if err != nil {
		return err
	}

	switch respErr.Code {
	case models.ErrCodeCacheError, models.ErrCodeDatabaseError, models.ErrCodeInternalServerError:
		{
			if err = rs.sendJob(ctx, "v1/journals", req.TransactionId, req.JournalRequest, req.ErrCauser); err != nil {
				return err
			}
		}
	default:
		rs.logRetry(ctx, "exclude retryable error", req.TransactionId, false, req.JournalRequest, req.ErrCauser)
	}

	return nil
}

func (rs *retryService) RetryUpdateLegacyId(ctx context.Context, req models.AccountError) (err error) {
	defer func() {
		logService(ctx, err)
		// send notif to slack
		if err != nil {
			rs.sendMessageToSlack(ctx, "Start process send job retry update legacy id", err.Error())
		}
	}()

	respErr, err := rs.getErrMap(req.ErrCauser)
	if err != nil {
		return err
	}

	switch respErr.Code {
	case models.ErrCodeDatabaseError, models.ErrCodeExternalServerError, models.ErrCodeInternalServerError:
		{
			if err = rs.sendJob(ctx, "v1/accounts/t24", req.AccountNumber, req.CreateAccount, req.ErrCauser); err != nil {
				return err
			}
		}
	default:
		rs.logRetry(ctx, "exclude retryable error", req.AccountNumber, false, req.CreateAccount, req.ErrCauser)
	}

	return nil
}

func (rs *retryService) sendJob(ctx context.Context, url, requestId string, request, errCauser any) error {
	rs.logRetry(ctx, "retryable error", requestId, true, request, errCauser)
	req := queueunicorn.RequestJobHTTP{
		Name: queueunicorn.HttpRequestJobKey,
		Payload: queueunicorn.PayloadJob{
			Host:    fmt.Sprintf("%s/%s", rs.srv.conf.HostGoAccounting, url),
			Method:  http.MethodPost,
			Body:    request,
			Headers: queueunicorn.RequestHeaderJob(rs.srv.conf.SecretKey, requestId),
		},
		Options: queueunicorn.Options{
			ProcessIn: 60,
			MaxRetry:  5,
			// Timeout:   10,
			// Deadline:  1000,
		},
	}
	if err := rs.srv.queueUnicornClient.SendJobHTTP(ctx, req); err != nil {
		return err
	}

	return nil
}

func (rs *retryService) sendMessageToSlack(ctx context.Context, operation, message string) {
	if err := rs.srv.dddNotificationClient.SendMessageToSlack(ctx, dddnotification.MessageData{
		Operation: operation,
		Message:   message,
	}); err != nil {
		xlog.Error(ctx, "[PROCESS-RETRY]", xlog.String("description", "failed to send slack message"), xlog.Err(err))
	}
}

func (rs *retryService) logRetry(ctx context.Context, desc string, id string, isRetry bool, request, err any) {
	xlog.Info(ctx, "[PROCESS-RETRY]",
		xlog.String("request-id", id),
		xlog.String("description", desc),
		xlog.Any("request", request),
		xlog.Bool("is-retry", isRetry),
		xlog.Any("error-causer", err))
}

func (rs *retryService) getErrMap(errMap interface{}) (respErr models.ErrorDetail, err error) {
	byErr, err := json.Marshal(errMap)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyFailedMarshal, err.Error())
		return
	}

	if err = json.Unmarshal(byErr, &respErr); err != nil {
		err = models.GetErrMap(models.ErrKeyFailedUnmarshal, err.Error())
		return
	}
	return
}
