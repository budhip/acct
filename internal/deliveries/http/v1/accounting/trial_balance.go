package accounting

import (
	"context"
	"errors"
	"net/http"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/validation"

	"github.com/labstack/echo/v4"
)

// @Summary 	Get Trial Balance
// @Description Get Trial Balance
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.DoGetTrialBalanceRequest true "Get trial balance query parameters"
// @Success 200 {object} commonhttp.RestTotalRowResponseModel[contents=[]models.GetTrialBalanceResponse] "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while get trial balance"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Not found error. This can happen if there is an error while get trial balance"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation get trial balance"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get trial balance"
// @Router 	/v1/trial-balances [get]
func (ah accountingHandler) getTrialBalance(c echo.Context) error {
	queryFilter := new(models.DoGetTrialBalanceRequest)
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

	res, err := ah.GetTrialBalance(c.Request().Context(), *opts)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, models.GetErrMap(models.ErrKeyEntityCodeNotFound)) {
			statusCode = http.StatusNotFound
		}
		return commonhttp.RestErrorResponse(c, statusCode, err)
	}

	return commonhttp.RestSuccessResponseListWithTotalRows(c, res, len(res.COATypes))
}

// @Summary 	Download Trial Balance Send to Email
// @Description Download Trial Balance Send to Email
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.DoGetTrialBalanceRequest true "Get trial balance query parameters"
// @Success 200 {object} commonhttp.RestTotalRowResponseModel[contents=[]models.GetTrialBalanceResponse] "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while download trial balance"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation download trial balance"
// @Router 	/v1/trial-balances/download [get]
func (ah accountingHandler) downloadCSVgetTrialBalance(c echo.Context) error {
	queryFilter := new(models.DoGetTrialBalanceRequest)
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

	ctx := context.WithoutCancel(c.Request().Context())
	go ah.DownloadTrialBalance(ctx, *opts)

	return commonhttp.RestSuccessResponse(c, http.StatusOK, "processing")
}

// @Summary 	Get Trial Balance Details
// @Description Get Trial Balance Details
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.DoGetTrialBalanceDetailsRequest true "Get list trial balance detail query parameters"
// @Success 200 {object} commonhttp.RestPaginationResponseModel[[]models.GetTrialBalanceDetailOut] "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while get account list"
// @Failure 422 {object} commonhttp.RestErrorResponseModel "Validation error. This can happen if there is an error while get account list"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get account list"
// @Router 	/v1/trial-balances/details [get]
func (ah accountingHandler) getTrialBalanceDetails(c echo.Context) error {
	queryFilter := new(models.DoGetTrialBalanceDetailsRequest)
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

	tba, total, err := ah.GetTrialBalanceDetails(c.Request().Context(), *opts)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, models.GetErrMap(models.ErrKeySubCategoryCodeNotFound)) ||
			errors.Is(err, models.GetErrMap(models.ErrKeyEntityCodeNotFound)) {
			statusCode = http.StatusNotFound
		}
		return commonhttp.RestErrorResponse(c, statusCode, err)
	}

	return commonhttp.RestSuccessResponseCursorPagination[models.GetTrialBalanceDetailOut](c, tba, opts.Limit, total, nil)
}

// @Summary 	Get Trial Balance by sub category code
// @Description Get Trial Balance by sub category code
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.DoGetTrialBalanceBySubCategoryRequest true "Get trial balance query parameters"
// @Success 200 {object} models.GetTrialBalanceBySubCategoryOut "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while get trial balance"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Not found error. This can happen if there is an error while get trial balance"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation get trial balance"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get trial balance"
// @Router 	/v1/trial-balances/sub-categories/:subCategoryCode [get]
func (ah accountingHandler) getTrialBalanceBySubCategoryCode(c echo.Context) error {
	queryFilter := new(models.DoGetTrialBalanceBySubCategoryRequest)
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

	res, err := ah.GetTrialBalanceBySubCategoryCode(c.Request().Context(), *opts)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, models.GetErrMap(models.ErrKeySubCategoryCodeNotFound)) {
			statusCode = http.StatusNotFound
		}
		return commonhttp.RestErrorResponse(c, statusCode, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, res.ToResponse())
}

// @Summary 	Send to Email Trial Balance Details
// @Description  Send to Email Trial Balance Details
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.DownloadTrialBalanceDetailsRequest true "Get download trial balance details query parameters"
// @Success 200 string success "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while download trial balance"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation download trial balance"
// @Router 	/v1/trial-balances/details/download [get]
func (ah accountingHandler) sendToEmailCSVgetTrialBalanceDetails(c echo.Context) error {
	queryFilter := new(models.DownloadTrialBalanceDetailsRequest)
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

	ctx := context.WithoutCancel(c.Request().Context())
	go ah.SendToEmailGetTrialBalanceDetails(ctx, *opts)

	return commonhttp.RestSuccessResponse(c, http.StatusOK, "processing")
}

// @Summary 	Close Trial Balance Period
// @Description  Close Trial Balance Period
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param		X-Secret-Key header string true "X-Secret-Key"
// @Param		period path string true "Trial Balance Period" example("2023-10")
// @Param		params body models.CloseTrialBalanceRequest true "Close trial balance period request body"
// @Success 	200 {object} models.CloseTrialBalanceResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 	400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while close trial balance"
// @Failure 	422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation close trial balance"
// @Failure 	404 {object} commonhttp.RestErrorResponseModel "Not found error. This can happen if there is an error while close trial balance"
// @Failure 	500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while close trial balance"
// @Router 		/v1/trial-balances/{period}/close [post]
func (ah accountingHandler) closeTrialBalance(c echo.Context) error {
	req := new(models.CloseTrialBalanceRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}
	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	req.Period = c.Param("period")
	out, err := ah.CloseTrialBalance(c.Request().Context(), *req)
	if err != nil {
		if errors.Is(err, models.GetErrMap(models.ErrKeyClosedPeriodNotFound)) {
			return commonhttp.RestErrorResponse(c, http.StatusNotFound, err)
		}
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, out.ToCloseTrialBalanceResponse())
}

// @Summary 	Trigger Adjustment Trial Balance
// @Description Trigger Adjustment Trial Balance
// @Tags 		Accounting
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param	payload body models.AdjustmentTrialBalanceRequest true "A JSON object containing trigger adjustment trial balance payload"
// @Success 200 string success "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while trigger adjustment trial balance"
// @Failure 422 {object} commonhttp.RestErrorResponseModel "Validation error. This can happen if there is an error while trigger adjustment trial balance"
// @Router 	/v1/trial-balances/adjustment [post]
func (ah accountingHandler) adjustmentTrialBalance(c echo.Context) error {
	req := new(models.AdjustmentTrialBalanceRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	opts, err := req.ToFilterOpts()
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err = ah.AccountingService.GenerateAdjustmentTrialBalanceBigQuery(c.Request().Context(), *opts); err != nil {
		code := http.StatusInternalServerError
		return commonhttp.RestErrorResponse(c, code, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, "processing")
}
