package api

import (
	"multipass/pkg/apperror"
	"multipass/pkg/logging"
	"multipass/pkg/response"
)

type BaseHandler struct {
	Logger       logging.Logger
	Responder    response.Writer
	ErrorHandler apperror.ErrorHandler
}
