package goigate

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httputil"

	"bitbucket.org/Amartha/go-accounting/internal/config"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-igate"
	"bitbucket.org/Amartha/go-x/environment"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type IgateClient interface {
	CreateLenderAccountNonRDL(ctx context.Context, in igate.LenderAccountRequest) (resp *igate.LenderAccount, err error)
	CreateCorporateLenderAccount(ctx context.Context, in igate.CorporateLenderAccountRequest) (resp *igate.CorporateLenderAccount, err error)
	GetLender(ctx context.Context, opts igate.CustomerGetLenderOptions) (resp *igate.Lender, err error)
	GetAccountIA(ctx context.Context, opts igate.GetAccountIAOptions) (resp []*igate.CoreResGetAccountIA, err error)
	GetLoanAccount(ctx context.Context, opts igate.LoanAccountGetOptions) (resp *igate.LoanAccount, err error)
	RegisterSavingAccount(ctx context.Context, in igate.GeneralSavingAccountRequest) (resp *igate.GeneralSavingAccount, err error)
}

type client struct {
	*igate.Client
}

func New(conf *config.Configuration, nrapp *newrelic.Application) IgateClient {
	opts := igate.ClientOptions{
		Source:       conf.App.Name,
		RateLimit:    conf.IGateClient.RequestPerSec,
		RetryMax:     conf.IGateClient.MaxRetry,
		NewrelicApp:  nrapp,
		MockResponse: conf.IGateClient.IsMockResponse,
		Debug: func(conf *config.Configuration) bool {
			if env := environment.ToEnvironment(conf.App.Env); env == environment.PROD_ENV {
				return false
			}
			return true
		}(conf),
		LogFormat: &logrus.JSONFormatter{},
		RequestLogHook: func(debug bool, req *http.Request, retry int) (logMap []interface{}) {
			logMap = append(logMap, "request-hashCode", req.Header.Get("hashCode"))
			if debug {
				out, _ := httputil.DumpRequestOut(req, true)
				logMap = append(logMap, "request-dump", base64.StdEncoding.EncodeToString(out))
				if req.Body != nil {
					body, _ := io.ReadAll(req.Body)
					req.Body = io.NopCloser(bytes.NewBuffer(body))
					logMap = append(logMap, "request-body", base64.StdEncoding.EncodeToString(body))
				}
			}
			return logMap
		},
		ResponseLogHook: func(debug bool, resp *http.Response) (logMap []interface{}) {
			logMap = append(logMap, "response-http-code", resp.StatusCode)
			if debug {
				out, _ := httputil.DumpResponse(resp, true)
				logMap = append(logMap, "response-dump", base64.StdEncoding.EncodeToString(out))
				if resp.Body != nil {
					body, _ := io.ReadAll(resp.Body)
					resp.Body = io.NopCloser(bytes.NewBuffer(body))
					logMap = append(logMap, "response-body", base64.StdEncoding.EncodeToString(body))
				}
			}
			return logMap
		},
	}
	igate := igate.NewClient(conf.IGateClient.BaseURL, conf.IGateClient.SecretKey, &opts)

	return &client{igate}
}

func (c *client) CreateLenderAccountNonRDL(ctx context.Context, in igate.LenderAccountRequest) (resp *igate.LenderAccount, err error) {
	defer func() {
		logIgate(ctx, "create lender account non rdl", in, err)
	}()

	resp, _, err = c.Accounts.CreateLenderAccountNonRDL(ctx, &in)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyFailedFromExternalClient, err.Error())
		return nil, err
	}

	return resp, nil
}

func (c *client) CreateCorporateLenderAccount(ctx context.Context, in igate.CorporateLenderAccountRequest) (resp *igate.CorporateLenderAccount, err error) {
	defer func() {
		logIgate(ctx, "create lender corporate account", in, err)
	}()

	resp, _, err = c.Accounts.CreateCorporateLenderAccount(ctx, &in)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyFailedFromExternalClient, err.Error())
		return nil, err
	}

	return resp, nil
}

func (c *client) GetLender(ctx context.Context, opts igate.CustomerGetLenderOptions) (resp *igate.Lender, err error) {
	defer func() {
		logIgate(ctx, "get customer lender", opts, err)
	}()

	resp, _, err = c.Customers.GetLender(ctx, &opts)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyFailedFromExternalClient, err.Error())
		return nil, err
	}

	return resp, nil
}

func (c *client) GetAccountIA(ctx context.Context, opts igate.GetAccountIAOptions) (resp []*igate.CoreResGetAccountIA, err error) {
	defer func() {
		logIgate(ctx, "get account ia", opts, err)
	}()

	resp, _, err = c.Accounts.GetAccountIA(ctx, &opts)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyFailedFromExternalClient, err.Error())
		return nil, err
	}

	return resp, nil
}

func (c *client) GetLoanAccount(ctx context.Context, opts igate.LoanAccountGetOptions) (resp *igate.LoanAccount, err error) {
	defer func() {
		logIgate(ctx, "get account loan", opts, err)
	}()

	resp, _, err = c.Accounts.GetLoanAccount(ctx, &opts)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyFailedFromExternalClient, err.Error())
		return nil, err
	}

	return resp, nil
}

func (c *client) RegisterSavingAccount(ctx context.Context, in igate.GeneralSavingAccountRequest) (resp *igate.GeneralSavingAccount, err error) {
	defer func() {
		logIgate(ctx, "register saving account", in, err)
	}()

	resp, _, err = c.Accounts.RegisterGeneralSavingAccount(ctx, &in)
	if err != nil {
		err = models.GetErrMap(models.ErrKeyFailedFromExternalClient, err.Error())
		return nil, err
	}

	return resp, nil
}

func logIgate(ctx context.Context, desc string, req any, err error) {
	if err != nil {
		xlog.Error(ctx, "[GO-IGATE]", xlog.String("status", "error"), xlog.String("description", desc), xlog.Any("request", req), xlog.Err(err))
	} else {
		xlog.Info(ctx, "[GO-IGATE]", xlog.String("status", "success"), xlog.String("description", desc), xlog.Any("request", req))
	}
}
