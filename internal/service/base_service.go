package service

import (
	"multipass/pkg/apperror"
	"multipass/pkg/logging"
	"multipass/pkg/response"
)

type BaseService struct {
	Logger       logging.Logger
	Responder    response.Writer
	ErrorHandler apperror.ErrorHandler
}
