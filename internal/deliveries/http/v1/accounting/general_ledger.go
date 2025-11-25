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

// @Summary 	Get General Ledger
// @Description Get General Ledger
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.GetSubLedgerRequest true "Get general ledger query parameters"
// @Success 200 {object} commonhttp.RestTotalRowResponseModel[contents=[]models.GetSubLedgerResponse,details=models.SubLedgerAccountResponse] "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while get general ledger"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Not found error. This can happen if there is an error while get general ledger"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get general ledger"
// @Router 	/v1/general-ledgers [get]
func (ah accountingHandler) generalLedger(c echo.Context) error {
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

	account, ledgers, total, err := ah.GetGeneralLedger(c.Request().Context(), *opts)
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

// @Summary 	Download CSV General Ledger
// @Description Download CSV General Ledger
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.DownloadCSVToEmailSubLedgerRequest true "Get general ledger query parameters"
// @Success 200 {object} commonhttp.RestTotalRowResponseModel[contents=[]models.GetSubLedgerResponse,details=models.SubLedgerAccountResponse] "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while get general ledger"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Not found error. This can happen if there is an error while get general ledger"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get general ledger"
// @Router 	/v1/general-ledgers/download [get]
func (ah accountingHandler) download(c echo.Context) error {
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

	go ah.SendSubLedgerCSVToEmail(context.Background(), *opts)

	return commonhttp.RestSuccessResponse(c, http.StatusOK, "success")
}
