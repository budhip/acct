package accounting

import (
	"context"
	"net/http"
	"strings"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/validation"

	"github.com/labstack/echo/v4"
)

// @Summary 	Get Sub Ledger
// @Description Get Sub Ledger
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.GetSubLedgerRequest true "Get sub ledger query parameters"
// @Success 200 {object} commonhttp.RestPaginationResponseModel[[]models.GetSubLedgerOut] "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while get sub ledger"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Not found error. This can happen if there is an error while get sub ledger"
// @Failure 422 {object} commonhttp.RestErrorResponseModel "Validation error. This can happen if there is an error while get sub ledger"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get sub ledger"
// @Router 	/v1/sub-ledgers [get]
func (ah accountingHandler) getSubLedger(c echo.Context) error {
	queryFilter := new(models.GetSubLedgerRequest)
	if err := c.Bind(queryFilter); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(queryFilter); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	opts, err := queryFilter.ToFilterOpts()
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	account, ledgers, total, err := ah.GetSubLedger(c.Request().Context(), *opts)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), models.ErrCodeDataNotFound) {
			statusCode = http.StatusNotFound
		}
		if strings.Contains(err.Error(), models.ErrCodeDataIsExceedsLimit) {
			statusCode = http.StatusOK
		}
		return commonhttp.RestErrorResponse(c, statusCode, err)
	}

	return commonhttp.RestSuccessResponseCursorPagination[models.GetSubLedgerResponse](c, ledgers, opts.Limit, total, account)
}

// @Summary 	Get Sub Ledger Count
// @Description Get Sub Ledger Count
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.GetSubLedgerRequest true "Get sub ledger count query parameters"
// @Success 200 {object} commonhttp.RestPaginationResponseModel[[]models.GetSubLedgerOut] "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while get sub ledger count"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Not found error. This can happen if there is an error while get sub ledger count"
// @Failure 422 {object} commonhttp.RestErrorResponseModel "Validation error. This can happen if there is an error while get sub ledger count"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get sub ledger count"
// @Router 	/v1/sub-ledgers/count [get]
func (ah accountingHandler) getSubLedgerCount(c echo.Context) error {
	queryFilter := new(models.GetSubLedgerRequest)
	if err := c.Bind(queryFilter); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(queryFilter); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	opts, err := queryFilter.ToFilterOpts()
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	res, err := ah.GetSubLedgerCount(c.Request().Context(), *opts)
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, res.ToModelResponse())
}

// @Summary 	Get Sub Ledger Accounts
// @Description Get Sub Ledger Accounts
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.DoGetSubLedgerAccountsRequest true "Get sub ledger accounts query parameters"
// @Success 200 {object} commonhttp.RestPaginationResponseModel[[]models.GetSubLedgerAccountsOut] "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while get sub ledger accounts"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Not found error. This can happen if there is an error while get sub ledger accounts"
// @Failure 422 {object} commonhttp.RestErrorResponseModel "Validation error. This can happen if there is an error while get sub ledger accounts"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get sub ledger accounts"
// @Router 	/v1/sub-ledgers/accounts [get]
func (ah accountingHandler) getSubLedgerAccounts(c echo.Context) error {
	queryFilter := new(models.DoGetSubLedgerAccountsRequest)
	if err := c.Bind(queryFilter); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(queryFilter); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	opts, err := queryFilter.ToFilterOpts()
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	subLedgerAccounts, total, err := ah.GetSubLedgerAccounts(c.Request().Context(), *opts)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), models.ErrCodeDataNotFound) {
			statusCode = http.StatusNotFound
		}
		return commonhttp.RestErrorResponse(c, statusCode, err)
	}

	return commonhttp.RestSuccessResponseCursorPagination[models.DoGetSubLedgerAccountsResponse](c, subLedgerAccounts, opts.Limit, total, nil)
}

// @Summary 	Send Sub Ledger CSV To Email
// @Description Send Sub Ledger CSV To Email
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.DownloadCSVToEmailSubLedgerRequest true "Send sub ledger csv to email query parameters"
// @Success 200 string success "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while send sub ledger csv to email"
// @Failure 422 {object} commonhttp.RestErrorResponseModel "Validation error. This can happen if there is an error while send sub ledger csv to email"
// @Router 	/v1/sub-ledgers/send-email [get]
func (ah accountingHandler) sendSubLedgerCSVToEmail(c echo.Context) error {
	queryParam := new(models.DownloadCSVToEmailSubLedgerRequest)

	if err := c.Bind(queryParam); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(queryParam); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	opts, err := queryParam.ToFilterOpts()
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	ctx := context.WithoutCancel(c.Request().Context())
	go ah.SendSubLedgerCSVToEmail(ctx, *opts)

	return commonhttp.RestSuccessResponse(c, http.StatusOK, "processing")
}

// @Summary 	Download CSV Sub Ledger
// @Description Download CSV Sub Ledger
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.DownloadCSVSubLedgerRequest true "Download sub ledger csv query parameters"
// @Success 200 string success "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while download sub ledger csv"
// @Failure 422 {object} commonhttp.RestErrorResponseModel "Validation error. This can happen if there is an error while download sub ledger csv"
// @Router 	/v1/sub-ledgers/download [get]
func (ah accountingHandler) downloadSubLedgerCSV(c echo.Context) error {
	queryParam := new(models.DownloadCSVSubLedgerRequest)

	if err := c.Bind(queryParam); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(queryParam); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	opts, err := queryParam.ToFilterOpts()
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	b, filename, err := ah.DownloadSubLedgerCSV(c.Request().Context(), *opts)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), models.ErrCodeDataNotFound) {
			statusCode = http.StatusNotFound
		}
		if strings.Contains(err.Error(), models.ErrCodeDataIsExceedsLimit) {
			statusCode = http.StatusOK
		}
		return commonhttp.RestErrorResponse(c, statusCode, err)
	}

	return commonhttp.RestSuccessResponseCSV(c, b, filename)
}
