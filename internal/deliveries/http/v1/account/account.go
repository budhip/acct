package account

import (
	"errors"
	"net/http"
	"path"
	"strings"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/validation"
	"bitbucket.org/Amartha/go-accounting/internal/services"

	"github.com/hashicorp/go-multierror"
	"github.com/labstack/echo/v4"
)

type accountHandler struct {
	service services.AccountService
}

// New account handler will initialize the account/ resources endpoint
func New(app *echo.Group, accountSrv services.AccountService) {
	ah := accountHandler{
		service: accountSrv,
	}
	account := app.Group("/accounts")
	account.POST("", ah.createAccount)
	account.GET("", ah.getAccountList)
	account.GET("/:accountNumber", ah.getByAccountNumber)
	account.PATCH("/:accountNumber", ah.updateAccount)
	account.PUT("/:accountNumber/entity", ah.updateAccountEntity)
	account.GET("/download", ah.downloadCSVGetAccountList)
	account.GET("/alt-ids", ah.checkAltIdIsExist)
	account.POST("/upload", ah.uploadAccount)
	/*
		todo: remove GET /t24/:legacyId if no traffic again in the future
	*/
	account.GET("/t24/:legacyId", ah.getByLegactId)
	account.GET("/account-numbers", ah.getAllAccountNumbersByParam)

	account.POST("/loan-partner-accounts", ah.createLoanPartnerAccount)

	lenderAccount := app.Group("/lender-accounts")
	lenderAccount.GET("/:cihAccountNumber", ah.getLenderAccountByCIHAccountNumber)

	loanAccount := app.Group("/loan-accounts")
	loanAccount.GET("/advance-account/:loanAccountNumber", ah.getLoanAdvanceAccountByLoanAccount)
}

// @Summary 	Create Account
// @Description Create New Account
// @Tags 		Accounts
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	payload body models.DoCreateAccountRequest true "A JSON object containing create account payload"
// @Success 201 {object} models.DoCreateAccountResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while create account"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if there is an data not found while create account"
// @Failure 409 {object} commonhttp.RestErrorResponseModel "Data is exist. This can happen if altId data is exist"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while create account"
// @Router 	/v1/accounts [post]
func (ah accountHandler) createAccount(c echo.Context) error {
	req := new(models.DoCreateAccountRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	var errs *multierror.Error
	if err := validation.ValidateStruct(req); err != nil {
		errs = multierror.Append(errs, err)
	}
	if req.AccountType == "" {
		in := struct {
			EntityCode      string `json:"entityCode" validate:"required"`
			CategoryCode    string `json:"categoryCode" validate:"required"`
			SubCategoryCode string `json:"subCategoryCode" validate:"required"`
			Currency        string `json:"currency" validate:"required"`
		}{
			EntityCode:      req.EntityCode,
			CategoryCode:    req.CategoryCode,
			SubCategoryCode: req.SubCategoryCode,
			Currency:        req.Currency,
		}
		if err := validation.ValidateStruct(&in); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	if errs.ErrorOrNil() != nil {
		return commonhttp.RestErrorValidationResponse(c, errs)
	}

	res, err := ah.service.Create(c.Request().Context(), models.CreateAccount{
		OwnerID:         req.OwnerID,
		AccountType:     req.AccountType,
		ProductTypeCode: req.ProductTypeCode,
		CategoryCode:    req.CategoryCode,
		SubCategoryCode: req.SubCategoryCode,
		EntityCode:      req.EntityCode,
		Currency:        req.Currency,
		Name:            req.Name,
		AltId:           req.AltId,
		LegacyId:        req.LegacyId,
		Metadata:        req.Metadata,
	})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), models.ErrCodeDataNotFound) {
			statusCode = http.StatusNotFound
		}
		if errors.Is(err, models.GetErrMap(models.ErrKeyAltIdIsExist)) {
			statusCode = http.StatusConflict
		}
		return commonhttp.RestErrorResponse(c, statusCode, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusCreated, res.ToCreateAccountResponse())
}

// @Summary 	Update Account
// @Description Update Existing Account
// @Tags 		Accounts
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param	accountNumber path string true "account identifier"
// @Param 	payload body models.DoUpdateAccountRequest true "A JSON object containing update account payload"
// @Success 201 {object} models.DoUpdateAccountResponse "Response indicates that the request succeeded and the resources has been retransmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while update account"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if there is an data not found while update account"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while update account"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while update account"
// @Router 	/v1/accounts/:accountNumber [patch]
func (ah accountHandler) updateAccount(c echo.Context) error {
	req := new(models.DoUpdateAccountRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	res, err := ah.service.Update(c.Request().Context(), models.UpdateAccount{
		AccountNumber: req.AccountNumber,
		Name:          req.Name,
		OwnerID:       req.OwnerID,
		AltID:         req.AltID,
		LegacyId:      req.LegacyID,
	})
	if err != nil {
		if errors.Is(err, models.GetErrMap(models.ErrKeyAccountNumberNotFound)) {
			return commonhttp.RestErrorResponse(c, http.StatusNotFound, err)
		}
		if errors.Is(err, models.GetErrMap(models.ErrKeyLegacyIdalreadyExists)) {
			return commonhttp.RestErrorResponse(c, http.StatusConflict, err)
		}
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, res.ToUpdateAccountResponse())
}

// @Summary 	Update Account Entity
// @Description Update Account Entity
// @Tags 		Accounts
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param	accountNumber path string true "account identifier"
// @Param 	payload body models.DoUpdateAccountEntityRequest true "A JSON object containing update account entity payload"
// @Success 200 {object} models.DoUpdateAccountEntityResponse "Response indicates that the request succeeded and the resources has been retransmitted in the message body"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if there is an data not found while update account entity"
// @Failure 409 {object} commonhttp.RestErrorResponseModel "Data is exist. This can happen if there is an data is exist in journal while update account entity"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while update account entity"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while update account entity"
// @Router 	/v1/accounts/:accountNumber/entity [put]
func (ah accountHandler) updateAccountEntity(c echo.Context) error {
	req := new(models.DoUpdateAccountEntityRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	res, err := ah.service.UpdateAccountEntity(c.Request().Context(), models.UpdateAccountEntity{
		AccountNumber: req.AccountNumber,
		EntityCode:    req.EntityCode,
	})
	if err != nil {
		if errors.Is(err, models.GetErrMap(models.ErrKeyAccountNumberNotFound)) ||
			errors.Is(err, models.GetErrMap(models.ErrKeyEntityCodeNotFound)) {
			return commonhttp.RestErrorResponse(c, http.StatusNotFound, err)
		}
		if errors.Is(err, models.GetErrMap(models.ErrKeyJournalAccountIsExist)) {
			return commonhttp.RestErrorResponse(c, http.StatusConflict, err)
		}
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, res.ToResponse())
}

// @Summary 	Get one account by account number
// @Description Get one account detail by account number
// @Tags 		Accounts
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	accountNumber path string true "account identifier"
// @Success 200 {object} models.DoGetAccountResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while get account"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if data not found while get account"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get account"
// @Router 	/v1/accounts/:accountNumber [get]
func (ah accountHandler) getByAccountNumber(c echo.Context) error {
	result, err := ah.service.GetOneByAccountNumber(c.Request().Context(), c.Param("accountNumber"))
	if err != nil {
		if errors.Is(err, models.GetErrMap(models.ErrKeyAccountNumberNotFound)) {
			return commonhttp.RestErrorResponse(c, http.StatusNotFound, err)
		}
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}
	return commonhttp.RestSuccessResponse(c, http.StatusOK, result.ToGetAccountResponse())
}

// @Summary 	Get one account by legacy id
// @Description Get one account detail by legacy id
// @Tags 		Accounts
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	legacyId path string true "account identifier"
// @Success 200 {object} models.DoGetAccountResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while get account"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if data not found while get account"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get account"
// @Router 	/v1/accounts/t24/:legacyId [get]
func (ah accountHandler) getByLegactId(c echo.Context) error {
	result, err := ah.service.GetOneByLegacyID(c.Request().Context(), c.Param("legacyId"))
	if err != nil {
		if errors.Is(err, models.GetErrMap(models.ErrKeyLegacyIdNotFound)) {
			return commonhttp.RestErrorResponse(c, http.StatusNotFound, err)
		}
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}
	return commonhttp.RestSuccessResponse(c, http.StatusOK, result.ToGetAccountResponse())
}

// @Summary 	Get Account List
// @Description Get Account List
// @Tags 		Accounts
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.DoGetListAccountRequest true "Get account list query parameters"
// @Success 200 {object} commonhttp.RestPaginationResponseModel[[]models.GetAccountOut] "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while get account list"
// @Failure 422 {object} commonhttp.RestErrorResponseModel "Validation error. This can happen if there is an error while get account list"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get account list"
// @Router 	/v1/accounts [get]
func (ah accountHandler) getAccountList(c echo.Context) error {
	queryFilter := new(models.DoGetListAccountRequest)
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

	accounts, total, err := ah.service.GetAccountList(c.Request().Context(), *opts)
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	return commonhttp.RestSuccessResponseCursorPagination[models.DoGetListAccountResponse](c, accounts, opts.Limit, total, nil)
}

// @Summary 	Download account List
// @Description Download account List
// @Tags 		Accounts
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.DoDownloadGetListAccountRequest true "Download account list query parameters"
// @Success 200 {object} commonhttp.RestTotalRowResponseModel[contents=[]models.GetAccountOut] "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while download account list"
// @Failure 422 {object} commonhttp.RestErrorResponseModel "Validation error. This can happen if there is an error while download account list"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while download account list"
// @Router 	/v1/accounts/download [get]
func (ah accountHandler) downloadCSVGetAccountList(c echo.Context) error {
	query := new(models.DoDownloadGetListAccountRequest)
	if err := c.Bind(query); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(query); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	queryFilter := models.DoGetListAccountRequest{
		Search:          query.Search,
		SearchBy:        query.SearchBy,
		CoaTypeCode:     query.CoaTypeCode,
		EntityCode:      query.EntityCode,
		ProductTypeCode: query.ProductTypeCode,
		CategoryCode:    query.CategoryCode,
		SubCategoryCode: query.SubCategoryCode,
		Limit:           query.Limit,
	}
	opts, err := queryFilter.ToFilterOpts()
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}
	opts.Limit = 0

	accounts, _, err := ah.service.GetAccountList(c.Request().Context(), *opts)
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	b, filename, err := ah.service.DownloadCSVGetAccountList(c.Request().Context(), accounts)
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	return commonhttp.RestSuccessResponseCSV(c, b, filename)
}

// @Summary 	Check alt id is exist
// @Description Check alt id is exist
// @Tags 		Accounts
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	altId query string true "alt id identifier"
// @Success 200 string success "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Success 200 {object} models.DoCheckAltIdResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 409 {object} commonhttp.RestErrorResponseModel "Data is exist. This can happen if there is an data is exist"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while check alt id"
// @Router 	/v1/accounts/alt-ids [get]
func (ah accountHandler) checkAltIdIsExist(c echo.Context) error {
	in := models.DoCheckAltIdRequest{AltId: c.QueryParam("altId")}

	if err := validation.ValidateStruct(in); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	if err := ah.service.CheckAltIdIsExist(c.Request().Context(), in.AltId); err != nil {
		if errors.Is(err, models.GetErrMap(models.ErrKeyAltIdIsExist)) {
			return commonhttp.RestErrorResponse(c, http.StatusConflict, err)
		}
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	res := models.DoCheckAltIdResponse{AltId: in.AltId}
	return commonhttp.RestSuccessResponse(c, http.StatusOK, res.ToResponse())
}

// @Summary 	Upload account
// @Description Upload account
// @Tags 		Accounts
// @Accept		multipart/form-data
// @Produce		text/csv
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param	file formData file true "csv file"
// @Success 200 string success "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while upload account"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while upload account"
// @Router 	/v1/accounts/upload [post]
func (ah accountHandler) uploadAccount(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if path.Ext(file.Filename) != ".csv" {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, errors.New("file not csv"))
	}

	if err = ah.service.ProcessUploadAccounts(c.Request().Context(), file); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, "success")
}

// @Summary 	Get all account by params
// @Description Get all account by params
// @Tags 		Accounts
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param   params query models.DoGetAllAccountNumbersByParamRequest true "get accounts query parameters"
// @Success 200 {object} commonhttp.RestTotalRowResponseModel{contents=[]models.DoGetAllAccountNumbersByParamResponse{}} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get account by params"
// @Router 	/v1/accounts/account-numbers [get]
func (ah accountHandler) getAllAccountNumbersByParam(c echo.Context) error {
	params := new(models.DoGetAllAccountNumbersByParamRequest)

	if err := c.Bind(params); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(params); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	result, err := ah.service.GetAllAccountNumbersByParam(c.Request().Context(), models.GetAllAccountNumbersByParamIn{
		OwnerId:        params.OwnerId,
		AltId:          params.AltId,
		AccountNumbers: params.AccountNumbers,
		AccountType:    params.AccountType,
		Limit:          params.Limit,
	})
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	var data []models.DoGetAllAccountNumbersByParamResponse
	for _, v := range result {
		data = append(data, *v.ToResponse())
	}

	return commonhttp.RestSuccessResponseListWithTotalRows(c, data, len(data))
}
