package cache

import (
	"errors"
	"net/http"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/pkg/validation"
	"bitbucket.org/Amartha/go-accounting/internal/services"

	"github.com/labstack/echo/v4"
)

type cacheHandler struct {
	accountSrv services.AccountService
}

// New cache type handler will initialize the cache/ resources endpoint
func New(app *echo.Group, accountSrv services.AccountService) {
	h := cacheHandler{
		accountSrv: accountSrv,
	}
	api := app.Group("/cache")
	accCatSeq := api.Group("/accounts/category-codes-seq")
	accCatSeq.GET("", h.getAllCategoryCodeSeq)
	accCatSeq.PUT("", h.updateCategoryCodeSeq)
	accCatSeq.POST("", h.createCategoryCodeSeq)
}

// @Summary 	Get all category code sequence
// @Description Get all category code sequence
// @Tags 		Cache
// @Accept  	json
// @Produce 	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Success 200 {object} []models.DoGetAllCategoryCodeSeqResponse "Response indicates that the request succeeded and the resources has been updated"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while category code sequence"
// @Router 	/v1/cache/accounts/category-codes-seq [get]
func (h cacheHandler) getAllCategoryCodeSeq(c echo.Context) error {
	res, err := h.accountSrv.GetAllCategoryCodeSeq(c.Request().Context())
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusInternalServerError, err)
	}
	return commonhttp.RestSuccessResponseListWithTotalRows(c, res, len(res))
}

// @Summary 	Update value category code sequence
// @Description Update value category code sequence
// @Tags 		Cache
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	payload body models.DoUpdateCategoryCodeSeqRequest true "A JSON object containing update value category code sequence payload"
// @Success 200 {object} models.DoUpdateCategoryCodeSeqResponse "Response indicates that the request succeeded and the resources has been updated"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while update value category code sequence"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while update value category code sequence"
// @Failure 404 {object} commonhttp.RestErrorResponseModel "Data not found. This can happen if there is an data not found while update value category code sequence"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while update value category code sequence"
// @Router 	/v1/cache/accounts/category-codes-seq [put]
func (h cacheHandler) updateCategoryCodeSeq(c echo.Context) error {
	req := new(models.DoUpdateCategoryCodeSeqRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	if err := h.accountSrv.UpdateCategoryCodeSeq(c.Request().Context(), *req); err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, models.GetErrMap(models.ErrKeyDataNotFound)) {
			statusCode = http.StatusNotFound
		}
		return commonhttp.RestErrorResponse(c, statusCode, err)
	}
	return commonhttp.RestSuccessResponse(c, http.StatusOK, req.ToResponse())
}

// @Summary 	Create value category code sequence
// @Description Create value category code sequence
// @Tags 		Cache
// @Accept  	json
// @Produce  	json
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param 	payload body models.DoCreateCategoryCodeSeqRequest true "A JSON object containing create value category code sequence payload"
// @Success 200 {object} models.DoCreateCategoryCodeSeqResponse "Response indicates that the request succeeded and the resources has been created"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while create value category code sequence"
// @Failure 409 {object} commonhttp.RestErrorResponseModel "Data is exist error. This can happen if there is an error while create value category code sequence"
// @Failure 422 {object} commonhttp.RestErrorValidationResponseModel{errors=[]validation.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create value category code sequence"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while create value category code sequence"
// @Router 	/v1/cache/accounts/category-codes-seq [post]
func (h cacheHandler) createCategoryCodeSeq(c echo.Context) error {
	req := new(models.DoCreateCategoryCodeSeqRequest)
	if err := c.Bind(req); err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if err := validation.ValidateStruct(req); err != nil {
		return commonhttp.RestErrorValidationResponse(c, err)
	}

	if err := h.accountSrv.CreateCategoryCodeSeq(c.Request().Context(), *req); err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, models.GetErrMap(models.ErrKeyDataIsExist)) {
			statusCode = http.StatusConflict
		}
		return commonhttp.RestErrorResponse(c, statusCode, err)
	}
	return commonhttp.RestSuccessResponse(c, http.StatusOK, req.ToResponse())
}
