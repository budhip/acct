package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type CursorPagination struct {
	Prev         string `json:"prev" example:"abc"`
	Next         string `json:"next" example:"cba"`
	TotalEntries int    `json:"totalEntries" example:"100"`
}

type PaginateableContent[ModelOut any] interface {
	GetCursor() string
	ToModelResponse() ModelOut
}

func isForward(c echo.Context) bool {
	return c.QueryParam("nextCursor") != ""
}

func isBackward(c echo.Context) bool {
	return !isForward(c) && c.QueryParam("prevCursor") != ""
}

func NewCursorPagination[ModelOut any, S ~[]E, E PaginateableContent[ModelOut]](c echo.Context, collections S, hasMorePages bool, totalEntries int) CursorPagination {
	var prevCursor, nextCursor string
	if len(collections) > 0 {
		if isBackward(c) || hasMorePages {
			nextCursor = collections[len(collections)-1].GetCursor()
		}

		if isForward(c) || (hasMorePages && isBackward(c)) {
			prevCursor = collections[0].GetCursor()
		}
	}

	return CursorPagination{
		Prev:         prevCursor,
		Next:         nextCursor,
		TotalEntries: totalEntries,
	}
}

func RestSuccessResponseCursorPagination[ModelResponse any, S ~[]E, E PaginateableContent[ModelResponse]](c echo.Context, data S, requestLimit, totalRows int, details any) error {
	// we use over-fetch to make sure nextPage exists or not
	hasMorePages := len(data) > (requestLimit - 1)

	if len(data) > 0 {
		if hasMorePages {
			data = data[:len(data)-1]
		}
		if isBackward(c) {
			// reverse data
			for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
				data[i], data[j] = data[j], data[i]
			}
		}
	}

	contents := make([]ModelResponse, 0)
	for _, datum := range data {
		res := datum.ToModelResponse()
		if &res != nil {
			contents = append(contents, res)
		}
	}

	pagination := NewCursorPagination[ModelResponse](c, data, hasMorePages, totalRows)

	return c.JSON(http.StatusOK, RestPaginationResponseModel[[]ModelResponse]{
		Kind:       "collection",
		Details:    details,
		Contents:   contents,
		Pagination: pagination,
	})
}
