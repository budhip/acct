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

func (ah accountingHandler) getTrialBalanceDetails(c echo.Context) error {

	queryFilter := new(models.DoGetTrialBalanceDetailsByPeriodRequest)
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

	tbd, summary, err := ah.GetTrialBalanceFromGCS(c.Request().Context(), *opts)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, models.GetErrMap(models.ErrKeySubCategoryCodeNotFound)) ||
			errors.Is(err, models.GetErrMap(models.ErrKeyEntityCodeNotFound)) {
			statusCode = http.StatusNotFound
		}
		return commonhttp.RestErrorResponse(c, statusCode, err)
	}

	//service

	return commonhttp.RestSuccessResponseListWithSummary(c, tbd, summary.ToResponse())
}

func (ah accountingHandler) sendTrialBalanceDetailsToEmail(c echo.Context) error {

	queryFilter := new(models.DownloadTrialBalanceDetailsByPeriodRequest)
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
	go ah.SendEmailTrialBalanceDetails(ctx, *opts)

	return commonhttp.RestSuccessResponse(c, http.StatusOK, "processing")
}

func (ah accountingHandler) sendTrialBalanceSummaryToEmail(c echo.Context) error {

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
	go ah.SendEmailTrialBalanceSummary(ctx, *opts)

	return commonhttp.RestSuccessResponse(c, http.StatusOK, "processing")
}
