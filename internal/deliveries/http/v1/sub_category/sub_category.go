package subcategory

import (
	"net/http"
	"strings"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/validation"
	"bitbucket.org/Amartha/go-accounting/internal/services"

	"github.com/labstack/echo/v4"
)

type subCategoryHandler struct {
	subCatSvc services.SubCategoryService
}

// New transaction handler will initialize the sub-categories/ resources endpoint
func New(app *echo.Group, subCatSvc services.SubCategoryService) {
	handler := subCategoryHandler{
		subCatSvc: subCatSvc,
	}
	api := app.Group("/sub-categories")
	api.POST("", handler.createSubCategory)
	api.GET("", handler.getAllSubCategory)
	api.PATCH("/:subCategoryCode", handler.updateSubCategory)
}

// @Summary 	Create data sub category
// @Description Create data sub category
// @Tags 		Sub Categories
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	payload body models.CreateSubCategoryRequest true "A JSON object containing create sub category payload"
// @Success 201 {object} models.CreateSubCategoryResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while create sub category"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create sub category"
// @Failure 409 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if there is an data not found while create sub category"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while create sub category"
// @Router 	/v1/sub-categories [post]
func (h subCategoryHandler) createSubCategory(c echo.Context) error {
	req := new(models.CreateSubCategoryRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	res, err := h.subCatSvc.Create(c.Request().Context(), models.CreateSubCategory(*req))
	if err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), models.ErrCodeDataNotFound) {
			code = http.StatusNotFound
		} else if strings.Contains(err.Error(), models.ErrCodeDataIsExist) {
			code = http.StatusConflict
		}
		return commonhttp.RestErrorResponse(c, code, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusCreated, res.ToCreateResponse())
}

// @Summary 	Get all data sub category
// @Description Get all data sub category
// @Tags 		Sub Categories
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Success 200 {object} commonhttp.RestTotalRowResponseModel{contents=[]models.SubCategoryOut{}} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get sub category"
// @Router 	/v1/sub-categories [get]
func (h subCategoryHandler) getAllSubCategory(c echo.Context) error {
	categoryCode := c.Request().URL.Query().Get("categoryCode")
	param := models.GetAllSubCategoryParam{}
	if categoryCode != "" {
		param.CategoryCode = categoryCode
	}
	res, err := h.subCatSvc.GetAll(c.Request().Context(), param)
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	data := []models.SubCategoryOut{}
	for _, v := range *res {
		data = append(data, *v.ToResponse())
	}

	return commonhttp.RestSuccessResponseListWithTotalRows(c, data, len(data))
}

// @Summary 	Update data sub category
// @Description Update data sub category
// @Tags 		Sub Categories
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	payload body models.UpdateSubCategoryRequest true "A JSON object containing update sub category payload"
// @Success 200 {object} models.UpdateSubCategoryResponse "Response indicates that the request succeeded and the resources has been updated"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while update sub category"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while update sub category"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if there is an data not found while update sub category"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while update sub category"
// @Router 	/v1/sub-categories/:subCategoryCode [patch]
func (h subCategoryHandler) updateSubCategory(c echo.Context) error {
	req := new(models.UpdateSubCategoryRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	out, err := h.subCatSvc.Update(c.Request().Context(), models.UpdateSubCategory{
		Name:            req.Name,
		Description:     req.Description,
		Status:          req.Status,
		ProductTypeCode: req.ProductTypeCode,
		Code:            req.Code,
		Currency:        req.Currency,
	})
	if err != nil {
		if strings.Contains(err.Error(), models.ErrCodeDataNotFound) {
			return commonhttp.RestErrorResponse(c, http.StatusNotFound, err)
		}
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}
	return commonhttp.RestSuccessResponse(c, http.StatusOK, out.ToResponse())
}
