package account

import (
	"errors"
	"net/http"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/validation"
	"github.com/labstack/echo/v4"
)

// @Summary 	Create Loan Partner Account
// @Description Create New Loan Partner Account
// @Tags 		Accounts
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	payload body models.DoCreateAccountLoanPartnerRequest true "A JSON object containing create account payload"
// @Success 201 {object} models.DoCreateAccountLoanPartnerResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while create account"
// @Failure 409 {object} commonhttp.RestErrorResponseModel "Data is exist. This can happen if altId data is exist"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while create account"
// @Router 	/v1/accounts/loan-partner-accounts [post]
func (ah accountHandler) createLoanPartnerAccount(c echo.Context) error {
	req := new(models.DoCreateAccountLoanPartnerRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	res, err := ah.service.CreateLoanPartnerAccount(c.Request().Context(), models.CreateAccountLoanPartner{
		PartnerName: req.PartnerName,
		LoanKind:    req.LoanKind,
		PartnerId:   req.PartnerId,
		Metadata:    req.Metadata,
	})
	if err != nil {
		if errors.Is(err, models.GetErrMap(models.ErrKeyDataIsExist)) {
			return commonhttp.RestErrorResponse(c, http.StatusConflict, models.GetErrMap(models.ErrKeyDataIsExist))
		}
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusCreated, res.ToResponse())
}
