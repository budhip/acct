package http

import (
	"bytes"
	"fmt"
	"net/http"

	"bitbucket.org/Amartha/go-accounting/internal/models"

	"github.com/hashicorp/go-multierror"
	"github.com/labstack/echo/v4"
)

type (
	RestErrorResponseModel struct {
		Status  string      `json:"status" example:"error"`
		Code    interface{} `json:"code"`
		Message string      `json:"message" example:"error"`
	}

	RestTotalRowResponseModel struct {
		Kind      string      `json:"kind" example:"collection"`
		Contents  interface{} `json:"contents"`
		TotalRows int         `json:"total_rows" example:"100"`
	}

	RestSummaryResponseModel struct {
		Kind     string      `json:"kind" example:"collection"`
		Contents interface{} `json:"contents"`
		Summary  interface{} `json:"summary"`
	}

	RestPaginationResponseModel[T any] struct {
		Kind       string           `json:"kind" example:"collection"`
		Details    interface{}      `json:"details,omitempty"`
		Contents   T                `json:"contents"`
		Pagination CursorPagination `json:"pagination"`
	}

	RestErrorValidationResponseModel struct {
		Status  string      `json:"status" example:"error"`
		Message string      `json:"message" example:"validation error"`
		Errors  interface{} `json:"errors"`
	}
)

func RestSuccessResponse(c echo.Context, code int, in interface{}) error {
	return c.JSON(code, in)
}

func RestSuccessResponseCSV(c echo.Context, b *bytes.Buffer, filename string) error {
	c.Response().Writer.Header().Set(echo.HeaderContentType, "text/csv")
	c.Response().Writer.Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment;filename=%s", filename))
	c.Response().Write(b.Bytes())
	return nil
}

func RestSuccessResponseListWithTotalRows(c echo.Context, data interface{}, totalRows int) error {
	return c.JSON(http.StatusOK, RestTotalRowResponseModel{
		Kind:      "collection",
		Contents:  data,
		TotalRows: totalRows,
	})
}

func RestErrorResponse(c echo.Context, statusCode int, err error) error {
	res := RestErrorResponseModel{
		Status:  "error",
		Code:    statusCode,
		Message: err.Error(),
	}
	if data, ok := err.(models.ErrorDetail); ok {
		res.Code = data.Code
		res.Message = data.ErrorMessage.Error()
	}
	return c.JSON(statusCode, res)
}

func RestErrorValidationResponse(c echo.Context, errors interface{}) error {
	res := RestErrorValidationResponseModel{
		Status:  "error",
		Message: models.ErrValidation.Error(),
	}
	if data, ok := errors.(*multierror.Error); ok {
		res.Errors = data.Errors
	}

	return c.JSON(http.StatusUnprocessableEntity, res)
}

func RestSuccessResponseListWithSummary[
	ModelResponse any,
	S ~[]E,
	E interface{ ToModelResponse() ModelResponse },
](c echo.Context, data S, summary any) error {

	contents := make([]ModelResponse, 0, len(data))

	for _, d := range data {
		contents = append(contents, d.ToModelResponse())
	}

	return c.JSON(http.StatusOK, RestSummaryResponseModel{
		Kind:     "collection",
		Contents: contents,
		Summary:  summary,
	})
}
