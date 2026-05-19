package httperror

import (
	"errors"
	"net/http"

	"github.com/glennprays/golang-clean-arch-starter/domain"
)

type APIError struct {
	Status  int
	Message string
}

func FromError(err error) APIError {
	var apiError APIError
	var domainError domain.Error

	if errors.As(err, &domainError) {
		apiError.Message = domainError.Error()
		switch domainError.ServiceError() {
		case domain.ErrBadRequest:
			apiError.Status = http.StatusBadRequest
		case domain.ErrNotFound:
			apiError.Status = http.StatusNotFound
		case domain.ErrUnauthorized:
			apiError.Status = http.StatusUnauthorized
		case domain.ErrForbidden:
			apiError.Status = http.StatusForbidden
		case domain.ErrConflict:
			apiError.Status = http.StatusConflict
		default:
			apiError.Status = http.StatusInternalServerError
		}
	}

	return apiError
}
