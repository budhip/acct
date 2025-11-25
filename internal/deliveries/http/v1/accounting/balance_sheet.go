package accounting

import (
	"errors"
	"net/http"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/validation"

	"github.com/labstack/echo/v4"
)

// @Summary 	Get Balance Sheet
// @Description Get Balance Sheet
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.GetBalanceSheetRequest true "Get balance sheet query parameters"
// @Success 200 {object} models.GetBalanceSheetResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while get balance sheet"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Not found error. This can happen if there is an error while get balance sheet"
// @Failure 422 {object} commonhttp.RestErrorResponseModel "Validation error. This can happen if there is an error while get balance sheet"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get balance sheet"
// @Router 	/v1/balance-sheets [get]
func (ah accountingHandler) getBalanceSheet(c echo.Context) error {
	queryFilter := new(models.GetBalanceSheetRequest)
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

	res, err := ah.GetBalanceSheet(c.Request().Context(), *opts)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, models.GetErrMap(models.ErrKeyEntityCodeNotFound)) {
			statusCode = http.StatusNotFound
		}
		return commonhttp.RestErrorResponse(c, statusCode, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, res)
}

// @Summary 	Download CSV Balance Sheet
// @Description Download CSV Balance Sheet
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.GetBalanceSheetRequest true "Download balance sheet query parameters"
// @Success 200 {object} models.GetBalanceSheetResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while download balance sheet"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Not found error. This can happen if there is an error while download balance sheet"
// @Failure 422 {object} commonhttp.RestErrorResponseModel "Validation error. This can happen if there is an error while download balance sheet"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while download balance sheet"
// @Router 	/v1/balance-sheets/download [get]
func (ah accountingHandler) downloadCSVBalanceSheet(c echo.Context) error {
	queryFilter := new(models.GetBalanceSheetRequest)
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

	res, err := ah.GetBalanceSheet(c.Request().Context(), *opts)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, models.GetErrMap(models.ErrKeyEntityCodeNotFound)) {
			statusCode = http.StatusNotFound
		}
		return commonhttp.RestErrorResponse(c, statusCode, err)
	}

	b, filename, err := ah.DownloadCSVGetBalanceSheet(c.Request().Context(), *opts, res)
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	return commonhttp.RestSuccessResponseCSV(c, b, filename)
}
