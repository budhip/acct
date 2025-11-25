package account

import (
	"errors"
	"net/http"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/labstack/echo/v4"
)

// @Summary 	Get lender account by cih account number
// @Description Get lender account by cih account number
// @Tags 		Accounts
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	cihAccountNumber path string true "account number identifier"
// @Success 200 {object} models.DoGetInvestedAccountResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if data not found while get lender account by cih account number"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get lender account by cih account number"
// @Router 	/v1/lender-accounts/{cihAccountNumber} [get]
func (ah accountHandler) getLenderAccountByCIHAccountNumber(c echo.Context) error {
	res, err := ah.service.GetLenderAccountByCIHAccountNumber(c.Request().Context(), c.Param("cihAccountNumber"))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, models.GetErrMap(models.ErrKeyAccountNumberNotFound)) {
			statusCode = http.StatusNotFound
		}
		return commonhttp.RestErrorResponse(c, statusCode, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, res.ToResponse())
}
