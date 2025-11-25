package coatype

import (
	"errors"
	"net/http"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/validation"
	"bitbucket.org/Amartha/go-accounting/internal/services"

	"github.com/labstack/echo/v4"
)

type coaTypeHandler struct {
	coaTypeSvc services.COATypeService
}

// New coa type handler will initialize the coa-types/ resources endpoint
func New(app *echo.Group, coaTypeSvc services.COATypeService) {
	handler := coaTypeHandler{
		coaTypeSvc: coaTypeSvc,
	}
	api := app.Group("/coa-types")
	api.POST("", handler.createCOAType)
	api.GET("", handler.getAllCOATypes)
	api.PATCH("/:coaTypeCode", handler.updateCOAType)
}

// @Summary 	Create data coa type
// @Description Create data coa type
// @Tags 		Coa Types
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	payload body models.CreateCOATypeRequest true "A JSON object containing create coa type payload"
// @Success 201 {object} models.COATypeOut "Response indicates that the request succeeded and the resources has been created into the system, and then retransmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while create coa type"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create coa type"
// @Failure 409 {object} commonhttp.RestErrorResponseModel "Data conflicted. This can happen if there is existing data found while create coa type"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while create coa type"
// @Router 	/v1/coa-types [post]
func (h coaTypeHandler) createCOAType(c echo.Context) error {
	req := new(models.CreateCOATypeRequest)

	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	res, err := h.coaTypeSvc.Create(c.Request().Context(), models.CreateCOATypeIn(*req))
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, models.GetErrMap(models.ErrKeyDataIsExist)) {
			code = http.StatusConflict
		}
		return commonhttp.RestErrorResponse(c, code, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusCreated, res.ToResponse())
}

// @Summary 	Get all data coa type
// @Description Get all data coa type
// @Tags 		Coa Types
// @Accept  	json
// @Produce 	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Success 200 {object} commonhttp.RestTotalRowResponseModel{contents=[]models.COATypeCategory{}} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get all data coa type"
// @Router 	/v1/coa-types [get]
func (h coaTypeHandler) getAllCOATypes(c echo.Context) error {
	res, err := h.coaTypeSvc.GetAll(c.Request().Context())
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}
	return commonhttp.RestSuccessResponseListWithTotalRows(c, res, len(res))
}

// @Summary 	Update data COA type
// @Description Update data COA type
// @Tags 		Coa Types
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	payload body models.UpdateCOATypeRequest true "A JSON object containing update coa type payload"
// @Success 200 {object} models.UpdateCOATypeResponse "Response indicates that the request succeeded and the resources has been updated"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while update coa type"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while update coa type"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if there is an data not found while update coa type"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while update coa type"
// @Router 	/v1/coa-types/:coaTypeCode [patch]
func (h coaTypeHandler) updateCOAType(c echo.Context) error {
	req := new(models.UpdateCOATypeRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	out, err := h.coaTypeSvc.Update(c.Request().Context(), models.UpdateCOAType{
		Name:          req.Name,
		NormalBalance: req.NormalBalance,
		Status:        req.Status,
		Code:          req.Code,
	})
	if err != nil {
		if errors.Is(err, models.GetErrMap(models.ErrKeyCoaTypeNotFound)) {
			return commonhttp.RestErrorResponse(c, http.StatusNotFound, err)
		}
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, out.ToResponse())
}
