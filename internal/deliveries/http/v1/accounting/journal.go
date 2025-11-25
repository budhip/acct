package accounting

import (
	"context"
	"errors"
	"net/http"
	"path"
	"strings"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/validation"

	"github.com/labstack/echo/v4"
)

// @Summary 	Create Journal Transaction
// @Description Create Journal Transaction
// @Tags 		Accounting
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param	payload body models.JournalRequest true "A JSON object containing create journal payload"
// @Success 201 {object} models.JournalResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while create journal"
// @Failure 409 {object} commonhttp.RestErrorResponseModel "Data is exist. This can happen if transactionId is exist"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while create journal"
// @Router 	/v1/journals [post]
func (ah accountingHandler) create(c echo.Context) error {
	req := new(models.JournalRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	if _, err := ah.JournalService.InsertJournalTransaction(c.Request().Context(), *req); err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), models.ErrCodeDataNotFound) {
			code = http.StatusNotFound
		} else if strings.Contains(err.Error(), models.ErrCodeDataIsExist) {
			code = http.StatusConflict
		}
		return commonhttp.RestErrorResponse(c, code, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusCreated, req.ToResponse())
}

// @Summary 	Publish Journal Transaction
// @Description Publish Journal Transaction
// @Tags 		Accounting
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param	payload body models.JournalRequest true "A JSON object containing publish journal payload"
// @Success 201 {object} models.JournalResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while publish journal"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if there is an data not found while publish journal"
// @Failure 409 {object} commonhttp.RestErrorResponseModel "Data is exist. This can happen if transactionId is exist"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while publish journal"
// @Router 	/v1/journals/publish [post]
func (ah accountingHandler) publish(c echo.Context) error {
	req := new(models.JournalRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	if err := ah.JournalService.PublishJournalTransaction(c.Request().Context(), *req); err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), models.ErrCodeDataNotFound) {
			code = http.StatusNotFound
		} else if strings.Contains(err.Error(), models.ErrCodeDataIsExist) {
			code = http.StatusConflict
		}
		return commonhttp.RestErrorResponse(c, code, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusCreated, req.ToResponse())
}

// @Summary 	Upload Journal Transaction
// @Description Upload Journal Transaction
// @Tags 		Accounting
// @Accept		multipart/form-data
// @Produce		text/csv
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param	file formData file true "csv file"
// @Success 200 string success "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while upload journal"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while upload journal"
// @Router 	/v1/journals/upload [post]
func (ah accountingHandler) uploadJournal(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if path.Ext(file.Filename) != ".csv" {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, errors.New("file not csv"))
	}

	ctx := context.WithoutCancel(c.Request().Context())
	go ah.JournalService.ProcessUploadJournal(ctx, file)

	return commonhttp.RestSuccessResponse(c, http.StatusOK, "processing")
}

// @Summary 	Get journal Detail by transaction id
// @Description Get journal Detail by transaction id
// @Tags 		Accounting
// @Accept		json
// @Produce		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	transactionId path string true "account identifier"
// @Success 200 {object} commonhttp.RestTotalRowResponseModel{contents=[]models.DoGetJournalDetailResponse{}} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if data not found while get journal detail"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get journal detail"
// @Router 	/v1/journals/:transactionId [get]
func (ah accountingHandler) getByTransactionId(c echo.Context) error {
	result, err := ah.JournalService.GetJournalByTransactionId(c.Request().Context(), c.Param("transactionId"))
	if err != nil {
		if errors.Is(err, models.GetErrMap(models.ErrKeyTransactionIdNotFound)) {
			return commonhttp.RestErrorResponse(c, http.StatusNotFound, err)
		}
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	var data []models.DoGetJournalDetailResponse
	for _, v := range result {
		data = append(data, v.ToResponse())
	}

	return commonhttp.RestSuccessResponseListWithTotalRows(c, data, len(data))
}
