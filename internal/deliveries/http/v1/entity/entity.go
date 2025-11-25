package entity

import (
	"errors"
	"net/http"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/validation"
	"bitbucket.org/Amartha/go-accounting/internal/services"

	"github.com/labstack/echo/v4"
)

type entityHandler struct {
	entitySvc services.EntityService
}

// New entity handler will initialize the /entities resources endpoint
func New(app *echo.Group, entitySvc services.EntityService) {
	handler := entityHandler{
		entitySvc: entitySvc,
	}
	api := app.Group("/entities")
	api.POST("", handler.createEntity)
	api.GET("", handler.getAllEntity)
	api.GET("/search", handler.getEntityByParam)
	api.PATCH("/:entityCode", handler.updateEntity)
}

// @Summary 	Create data entity
// @Description Create data entity
// @Tags 		Entities
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	payload body models.CreateEntityRequest true "A JSON object containing create entity payload"
// @Success 201 {object} models.EntityOut "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while create entity"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create entity"
// @Failure 409 {object} commonhttp.RestErrorResponseModel "Data is exist. This can happen if there is an data is exist while create entity"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while create entity"
// @Router 	/v1/entities [post]
func (h entityHandler) createEntity(c echo.Context) error {
	req := new(models.CreateEntityRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	res, err := h.entitySvc.Create(c.Request().Context(), models.CreateEntityIn(*req))
	if err != nil {
		if errors.Is(err, models.ErrDataExist) {
			return commonhttp.RestErrorResponse(c, http.StatusConflict, models.GetErrMap(models.ErrKeyDataIsExist))
		}
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusCreated, res.ToResponse())
}

// @Summary 	Get all data entity
// @Description Get all data entity
// @Tags 		Entities
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Success 200 {object} commonhttp.RestTotalRowResponseModel{contents=[]models.EntityOut{}} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get entity"
// @Router 	/v1/entities [get]
func (h entityHandler) getAllEntity(c echo.Context) error {
	res, err := h.entitySvc.GetAll(c.Request().Context())
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	var data []models.EntityOut
	for _, v := range *res {
		data = append(data, *v.ToResponse())
	}

	return commonhttp.RestSuccessResponseListWithTotalRows(c, data, len(data))
}

// @Summary 	Update data entity
// @Description Update data entity
// @Tags 		Entities
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	payload body models.UpdateEntityRequest true "A JSON object containing update entity payload"
// @Success 200 {object} models.UpdateEntityResponse "Response indicates that the request succeeded and the resources has been updated"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while update entity"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while update entity"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if there is an data not found while update entity"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while update entity"
// @Router 	/v1/entities/:entityCode [patch]
func (h entityHandler) updateEntity(c echo.Context) error {
	req := new(models.UpdateEntityRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	out, err := h.entitySvc.Update(c.Request().Context(), models.UpdateEntity{
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Code:        req.Code,
	})
	if err != nil {
		if errors.Is(err, models.GetErrMap(models.ErrKeyEntityCodeNotFound)) {
			return commonhttp.RestErrorResponse(c, http.StatusNotFound, err)
		}
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, out.ToResponse())
}

// @Summary 	Get data entity by param
// @Description Get data entity by param
// @Tags 		Entities
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	payload body models.GetEntityRequest true "A JSON object containing update entity payload"
// @Success 200 {object} models.GetEntityResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while update entity"
// @Router 	/v1/entities/search [get]
func (h entityHandler) getEntityByParam(c echo.Context) error {
	req := new(models.GetEntityRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	out, err := h.entitySvc.GetByParam(c.Request().Context(), models.GetEntity{
		Name: req.Name,
		Code: req.EntityCode,
	})
	if err != nil {
		if errors.Is(err, models.GetErrMap(models.ErrKeyDataNotFound)) {
			return commonhttp.RestErrorResponse(c, http.StatusNotFound, err)
		}
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, out.ToResponseGetEntity())
}
