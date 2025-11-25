package models

import "mime/multipart"

type PublishRequest struct {
	Topic   string                `json:"topic" example:"journal_stream"`
	Message *multipart.FileHeader `json:"message"`
}
