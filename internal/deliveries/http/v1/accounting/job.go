package accounting

import (
	"context"
	"net/http"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/validation"

	"github.com/labstack/echo/v4"
)

// @Summary 	Run Job Trial Balance
// @Description Run Job Trial Balance
// @Tags 		Accounting
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param	payload body models.RunTrialBalanceRequest true "A JSON object containing run job trial balance payload"
// @Success 200 string success "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while run job trial balance"
// @Failure 422 {object} commonhttp.RestErrorResponseModel "Validation error. This can happen if there is an error while run job trial balance"
// @Router 	/v1/jobs/trial-balance [post]
func (ah accountingHandler) runJobTrialBalance(c echo.Context) error {
	req := new(models.RunTrialBalanceRequest)
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

	newCtx := context.WithoutCancel(c.Request().Context())
	go ah.AccountingService.GenerateTrialBalanceBigQuery(newCtx, opts.Date, opts.IsAdjustment)

	return commonhttp.RestSuccessResponse(c, http.StatusOK, "processing")
}
