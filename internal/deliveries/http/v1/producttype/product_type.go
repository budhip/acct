package producttype

import (
	"net/http"
	"strings"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/validation"
	"bitbucket.org/Amartha/go-accounting/internal/services"

	"github.com/labstack/echo/v4"
)

type productTypeHandler struct {
	productTypeSvc services.ProductTypeService
}

// New handler will initialize the /product-types resources endpoint
func New(app *echo.Group, productTypeSvc services.ProductTypeService) {
	handler := productTypeHandler{
		productTypeSvc: productTypeSvc,
	}
	api := app.Group("/product-types")
	api.GET("", handler.getAllProductType)
	api.POST("", handler.createProductType)
	api.PATCH("/:productTypeCode", handler.updateProductType)
}

// @Summary 	Get all data product type
// @Description Get all data product type
// @Tags 		Product Types
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Success 200 {object} commonhttp.RestTotalRowResponseModel{contents=[]models.ProductTypeOut{}} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while get product type"
// @Router 	/v1/product-types [get]
func (h productTypeHandler) getAllProductType(c echo.Context) error {
	res, err := h.productTypeSvc.GetAll(c.Request().Context())
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}

	var data []models.ProductTypeOut
	for _, v := range res {
		data = append(data, v.ToResponse())
	}

	return commonhttp.RestSuccessResponseListWithTotalRows(c, data, len(data))
}

// @Summary 	Create product type
// @Description Create product type
// @Tags 		Product Types
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param	payload body models.CreateProductTypeRequest true "A JSON object containing create product type payload"
// @Success 201 {object} models.ProductTypeOut "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while create product type"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create product type"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if there is an data not found while create  product type"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while create  product type"
// @Router 	/v1/product-types [post]
func (h productTypeHandler) createProductType(c echo.Context) error {
	req := new(models.CreateProductTypeRequest)

	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	res, err := h.productTypeSvc.Create(c.Request().Context(), *req)
	if err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), models.ErrCodeDataNotFound) {
			code = http.StatusNotFound
		} else if strings.Contains(err.Error(), models.ErrCodeDataIsExist) {
			code = http.StatusConflict
		}
		return commonhttp.RestErrorResponse(c, code, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusCreated, res.ToResponse())
}

// @Summary 	Update product type
// @Description Update product type
// @Tags 		Product Types
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param	payload body models.UpdateProductTypeRequest true "A JSON object containing update product type payload"
// @Success 201 {object} models.UpdateProductTypeResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while update product type"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while update product type"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if there is an data not found while update product type"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while update product type"
// @Router 	/v1/product-types/{productTypeCode} [patch]
func (h productTypeHandler) updateProductType(c echo.Context) error {
	req := new(models.UpdateProductTypeRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	out, err := h.productTypeSvc.Update(c.Request().Context(), models.UpdateProductType{
		Name:       req.Name,
		Status:     req.Status,
		Code:       req.Code,
		EntityCode: req.EntityCode,
	})
	if err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), models.ErrCodeDataNotFound) {
			code = http.StatusNotFound
		}
		return commonhttp.RestErrorResponse(c, code, err)
	}

	return commonhttp.RestSuccessResponse(c, http.StatusOK, out.ToResponse())
}
