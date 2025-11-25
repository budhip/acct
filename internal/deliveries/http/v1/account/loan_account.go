package account

import (
	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
)

// @Summary 	Get loan advance account by loan account
// @Description Get loan advance account by loan account
// @Tags 		Accounts
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	loanAccountNumber path string true "loan account number identifier"
// @Success 200 {object} models.DoGetLoanAccountResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if data not found while get loan advance account by loan account"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get loan advance account by loan account"
// @Router 	/v1/loan-accounts/advance-account/{loanAccountNumber} [get]
func (ah accountHandler) getLoanAdvanceAccountByLoanAccount(c echo.Context) error {
	res, err := ah.service.GetLoanAdvanceAccountByLoanAccount(c.Request().Context(), c.Param("loanAccountNumber"))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, models.GetErrMap(models.ErrKeyAccountNumberNotFound)) {
			statusCode = http.StatusNotFound
		}
		return commonhttp.RestErrorResponse(c, statusCode, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, res.ToResponse())
}
