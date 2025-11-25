package category

import (
	"net/http"
	"strings"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/validation"
	"bitbucket.org/Amartha/go-accounting/internal/services"

	"github.com/labstack/echo/v4"
)

type categoryHandler struct {
	categorySvc services.CategoryService
}

// New transaction handler will initialize the categories/ resources endpoint
func New(app *echo.Group, categorySvc services.CategoryService) {
	handler := categoryHandler{
		categorySvc: categorySvc,
	}
	api := app.Group("/categories")
	api.POST("", handler.createCategory)
	api.GET("", handler.getAllCategory)
	api.PATCH("/:categoryCode", handler.updateCategory)
}

// @Summary 		Create data category
// @Description 	Create data category
// @Tags 			Categories
// @Accept  		json
// @Produce  		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	payload body models.DoCreateCategoryRequest true "A JSON object containing create category payload"
// @Success 201 {object} models.CategoryOut "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while create category"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create category"
// @Failure 409 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if there is an data not found while create category"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while create category"
// @Router 	/v1/categories [post]
func (h categoryHandler) createCategory(c echo.Context) error {
	req := new(models.DoCreateCategoryRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	res, err := h.categorySvc.Create(c.Request().Context(), models.CreateCategoryIn(*req))
	if err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), models.ErrCodeDataNotFound) {
			code = http.StatusNotFound
		} else if strings.Contains(err.Error(), models.ErrCodeDataIsExist) {
			code = http.StatusConflict
		}
		return commonhttp.RestErrorResponse(c, code, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusCreated, res.ConvertToCategoryOut())
}

// getAllCategory 	API get all category
// @Summary 		Get all data category
// @Description 	Get all data category
// @Tags			Categories
// @Accept			json
// @Produce			json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Success 200 {object} commonhttp.RestTotalRowResponseModel{contents=[]models.COATypeCategory{}} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get entity"
// @Router	/v1/categories [get]
func (h categoryHandler) getAllCategory(c echo.Context) error {
	res, err := h.categorySvc.GetAll(c.Request().Context())
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	data := []models.CategoryOut{}
	for _, v := range *res {
		data = append(data, *v.ConvertToCategoryOut())
	}

	return commonhttp.RestSuccessResponseListWithTotalRows(c, data, len(data))
}

// @Summary 		Update data category
// @Description 	Create data category
// @Tags 			Categories
// @Accept  		json
// @Produce  		json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	payload body models.DoUpdateCategoryRequest true "A JSON object containing update category payload"
// @Success 200 {object} models.DoUpdateCategoryResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while update category"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while update category"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if there is an data not found while update category"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while update category"
// @Router 	/v1/categories/{categoryCode} [patch]
func (h categoryHandler) updateCategory(c echo.Context) error {
	req := new(models.DoUpdateCategoryRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	res, err := h.categorySvc.Update(c.Request().Context(), models.UpdateCategoryIn{
		Name:        req.Name,
		Description: req.Description,
		CoaTypeCode: req.CoaTypeCode,
		Code:        req.Code,
	})
	if err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), models.ErrCodeDataNotFound) {
			code = http.StatusNotFound
		}
		return commonhttp.RestErrorResponse(c, code, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, res.ToResponse())
}
