package httperror

import (
	"errors"
	"net/http"

	"github.com/glennprays/golang-clean-arch-starter/internal/app/model"
)

type APIError struct {
	Status  int
	Message string
}

func FromError(err error) APIError {
	var apiError APIError
	var svcError model.Error

	if errors.As(err, &svcError) {
		apiError.Message = svcError.AppError().Error()
		svcErr := svcError.ServiceError()
		switch svcErr {
		case model.ErrBadRequest:
			apiError.Status = http.StatusBadRequest
		case model.ErrInternalFailure:
			apiError.Status = http.StatusInternalServerError
		case model.ErrNotFound:
			apiError.Status = http.StatusNotFound
		case model.ErrUnauthorized:
			apiError.Status = http.StatusUnauthorized
		case model.ErrForbidden:
			apiError.Status = http.StatusForbidden
		case model.ErrConflict:
			apiError.Status = http.StatusConflict
		}
	}

	return apiError
}
