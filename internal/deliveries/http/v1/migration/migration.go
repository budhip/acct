package migration

import (
	"context"
	"net/http"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/validation"
	"bitbucket.org/Amartha/go-accounting/internal/services"

	"github.com/labstack/echo/v4"
)

type handler struct {
	migrationSvc services.MigrationService
}

func New(app *echo.Group, migrationSvc services.MigrationService) {
	handler := handler{migrationSvc}
	api := app.Group("/migrations")

	api.POST("/buckets/journal-load", handler.loanInFileBucket)
}

// @Summary 	Load migration journal from buckets
// @Description Load migration journal from buckets
// @Tags 		Migration
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	payload body models.MigrationBucketsJournalLoadRequest true "A JSON object containing create entity payload"
// @Success 201 {object} models.MigrationLoadResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while create entity"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create entity"
// @Failure 409 {object} commonhttp.RestErrorResponseModel "Data is exist. This can happen if there is an data is exist while create entity"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while create entity"
// @Router 	/v1/migrations/buckets/journal-load [post]
func (h handler) loanInFileBucket(c echo.Context) error {
	req := new(models.MigrationBucketsJournalLoadRequest)

	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	// TODO: Use proper background task executor
	go h.migrationSvc.BucketsJournalLoad(context.Background(), *req)

	return commonhttp.RestSuccessResponse(c, http.StatusCreated, models.NewMigrationLoadResponse("Migration process started"))
}
