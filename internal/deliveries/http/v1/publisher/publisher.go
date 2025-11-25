package account

import (
	"context"
	"errors"
	"net/http"
	"path"

	commonhttp "bitbucket.org/Amartha/go-accounting/internal/deliveries/http/common"
	"bitbucket.org/Amartha/go-accounting/internal/models"
	"bitbucket.org/Amartha/go-accounting/internal/services"

	"github.com/labstack/echo/v4"
)

type publisherHandler struct {
	service services.PublisherService
}

func New(app *echo.Group, srv services.PublisherService) {
	ah := publisherHandler{
		service: srv,
	}
	publisher := app.Group("/topics")
	publisher.POST("/:topic", ah.publish)
}

// @Summary 	Publish Message
// @Description Publish Message
// @Tags 		Publisher
// @Accept		multipart/form-data
// @Produce		text/csv
// @Param	X-Secret-Key header string true "X-Secret-Key"
// @Param	file formData file true "csv file"
// @Success 200 string processing "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} commonhttp.RestErrorResponseModel "Bad request error. This can happen if there is an error while publish message"
// @Failure 422 {object} commonhttp.RestErrorResponseModel "Validation error. This can happen if there is an error while publish message"
// @Failure 500 {object} commonhttp.RestErrorResponseModel "Internal server error. This can happen if there is an error while publish message"
// @Router 	/v1/topics/:topic [post]
func (ah publisherHandler) publish(c echo.Context) error {
	req := new(models.PublishRequest)

	req.Topic = c.Param("topic")
	file, err := c.FormFile("file")
	if err != nil {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, err)
	}

	if path.Ext(file.Filename) != ".json" {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, errors.New("file not csv"))
	}

	if file.Size == 0 {
		return commonhttp.RestErrorResponse(c, http.StatusBadRequest, errors.New("file is empty"))
	}

	req.Message = file

	ctx := context.WithoutCancel(c.Request().Context())
	go ah.service.PublishMessage(ctx, *req)

	return commonhttp.RestSuccessResponse(c, http.StatusOK, "processing")
}
