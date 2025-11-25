package models

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	ErrNoRowsAffected      = errors.New("no rows affected")
	ErrValidation          = errors.New("validation failed")
	ErrPositionInvalid     = errors.New("position invalid")
	ErrDataNotFound        = errors.New("data not found")
	ErrInternalServerError = errors.New("internal server error")
	ErrInvalidFormatDate   = errors.New("invalid format date")
	ErrDataTrxNotFound     = errors.New("data transaction not found")
	ErrDataTrxDuplicate    = errors.New("transaction duplicate")
	ErrIDEmpty             = errors.New("ID is empty")
	ErrDataExist           = errors.New("data is exist")
	ErrUnableToCreate      = errors.New("unable to create data")
	ErrTimeout             = errors.New("timeout")
	ErrNoRows              = sql.ErrNoRows
	ErrRedisClosed         = redis.ErrClosed
	ErrRedisNil            = redis.Nil
)

type (
	MapErrs     map[string]ErrorDetail
	ErrorDetail struct {
		Code         string `json:"code,omitempty"`
		Message      string `json:"message,omitempty"`
		ErrorMessage error  `json:"-"`
	}
)

func (e ErrorDetail) Error() string {
	return fmt.Sprintf("code: %s, message: %v", e.Code, e.ErrorMessage)
}

func GetErrMap(code string, args ...string) ErrorDetail {
	v, ok := MapErrors[code]
	if !ok {
		err := errors.New("unknown error mapping")
		return ErrorDetail{
			Code:         code,
			ErrorMessage: err,
			Message:      err.Error(),
		}
	}
	if len(args) > 0 {
		v.ErrorMessage = fmt.Errorf("%s caused by %s", v.ErrorMessage, args[0])
	}
	v.Message = v.ErrorMessage.Error()

	return v
}

func IsErrMap(err error) (ErrorDetail, bool) {
	respErr, ok := err.(ErrorDetail)
	return respErr, ok
}
