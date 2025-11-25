package loanpartneraccount

import (
	"errors"
	"net/http"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/validation"
	"bitbucket.org/Amartha/go-accounting/internal/services"

	"github.com/labstack/echo/v4"
)

type loanPartnerAccountHandler struct {
	loanPartnerAccountSvc services.LoanPartnerService
}

// New entity handler will initialize the /loan-partner-accounts resources endpoint
func New(app *echo.Group, loanPartnerAccountSvc services.LoanPartnerService) {
	handler := loanPartnerAccountHandler{
		loanPartnerAccountSvc: loanPartnerAccountSvc,
	}
	api := app.Group("/loan-partner-accounts")
	api.POST("", handler.create)
	api.PATCH("/:accountNumber", handler.update)
	api.GET("", handler.getByParams)
}

// @Summary 	Create loan partner account
// @Description Create loan partner account
// @Tags 		Loan Partner Accounts
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	payload body models.DoCreateLoanPartnerAccountRequest true "A JSON object containing loan partner account payload"
// @Success 201 {object} models.DoCreateLoanPartnerAccountResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while create loan partner account"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Account not found. This can happen if account not found while create loan partner account"
// @Failure 409 {object} commonhttp.RestErrorResponseModel "Data is exist. This can happen if there is an data is exist while create loan partner account"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create loan partner account"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while create loan partner account"
// @Router 	/v1/loan-partner-accounts [post]
func (h loanPartnerAccountHandler) create(c echo.Context) error {
	req := new(models.DoCreateLoanPartnerAccountRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	res, err := h.loanPartnerAccountSvc.Create(c.Request().Context(), models.LoanPartnerAccount{
		PartnerId:           req.PartnerId,
		LoanKind:            req.LoanKind,
		AccountNumber:       req.AccountNumber,
		AccountType:         req.AccountType,
		LoanSubCategoryCode: req.LoanSubCategoryCode,
	})
	if err != nil {
		httpStatus := http.StatusInternalServerError
		if errors.Is(err, models.GetErrMap(models.ErrKeyAccountNumberNotFound)) {
			httpStatus = http.StatusNotFound
		}
		if errors.Is(err, models.GetErrMap(models.ErrKeyDataIsExist)) ||
			errors.Is(err, models.GetErrMap(models.ErrKeyAccountNumberIsExist)) {
			httpStatus = http.StatusConflict
		}
		return commonhttp.RestErrorResponse(c, httpStatus, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusCreated, res.ToCreateResponse())
}

// @Summary 	Update Acct Loan Partner Account
// @Description Update Acct Loan Partner Account by Account ID
// @Tags 		Loan Partner Accounts
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param	accountNumber path string true "account identifier"
// @Param 	payload body models.DoUpdateLoanPartnerAccountRequest true "A JSON object containing update loan partner account payload"
// @Success 200 {object} models.DoUpdateLoanPartnerAccountResponse "Response indicates that the request succeeded and the resources has been retransmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while update loan partner accountt"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Account not found. This can happen if account not found while update loan partner account"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while update loan partner account"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while update loan partner account"
// @Router 	/v1/loan-partner-accounts/:accountNumber [patch]
func (h loanPartnerAccountHandler) update(c echo.Context) error {
	req := new(models.DoUpdateLoanPartnerAccountRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	res, err := h.loanPartnerAccountSvc.Update(c.Request().Context(), models.UpdateLoanPartnerAccount{
		PartnerId:           req.PartnerId,
		LoanKind:            req.LoanKind,
		AccountNumber:       req.AccountNumber,
		AccountType:         req.AccountType,
		LoanSubCategoryCode: req.LoanSubCategoryCode,
	})
	if err != nil {
		if errors.Is(err, models.GetErrMap(models.ErrKeyAccountNumberNotFound)) {
			return commonhttp.RestErrorResponse(c, http.StatusNotFound, err)
		}
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, res.ToUpdateResponse())
}

// @Summary 	Get loan partner account
// @Description Get loan partner account
// @Tags 		Loan Partner Accounts
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.DoGetLoanPartnerAccountByParamsRequest true "Get loan partner account query parameters"
// @Success 200 {object} commonhttp.RestPaginationResponseModel[[]models.DoGetLoanPartnerAccountByParamsResponse] "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while get loan partner account"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Account not found. This can happen if account not found while get loan partner account"
// @Failure 422 {object} commonhttp.RestErrorResponseModel "Validation error. This can happen if there is an error while get loan partner account"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get loan partner account"
// @Router 	/v1/loan-partner-accounts [get]
func (h loanPartnerAccountHandler) getByParams(c echo.Context) error {
	req := new(models.DoGetLoanPartnerAccountByParamsRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	res, err := h.loanPartnerAccountSvc.GetByParams(c.Request().Context(), models.GetLoanPartnerAccountByParamsIn{
		PartnerId:           req.PartnerId,
		LoanKind:            req.LoanKind,
		AccountNumber:       req.AccountNumber,
		AccountType:         req.AccountType,
		EntityCode:          req.EntityCode,
		LoanSubCategoryCode: req.LoanSubCategoryCode,
		LoanAccountNumber:   req.LoanAccountNumber,
	})
	if err != nil {
		if errors.Is(err, models.GetErrMap(models.ErrKeyAccountNumberNotFound)) ||
			errors.Is(err, models.GetErrMap(models.ErrKeyEntityCodeNotFound)) ||
			errors.Is(err, models.GetErrMap(models.ErrKeyDataNotFound)) {
			return commonhttp.RestErrorResponse(c, http.StatusNotFound, err)
		}
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	var data []models.DoGetLoanPartnerAccountByParamsResponse
	for _, v := range res {
		data = append(data, v.ToGetResponse())
	}

	return commonhttp.RestSuccessResponseListWithTotalRows(c, data, len(data))
}
