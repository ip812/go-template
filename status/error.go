package status

import (
	"fmt"
	"net/http"
)

var (
	ErrQueriesNotInitialized         = fmt.Errorf("queries not initialized")
	ErrParsingFrom                   = fmt.Errorf("failed to parse a form")
	ErrDecodingForm                  = fmt.Errorf("failed to decode a form")
	ErrFailedtoValidateRequest       = fmt.Errorf("failed to validate a request")
	ErrFailedToAddEmailToMailingList = fmt.Errorf("failed to add email to mailing list")
)

func ErrorNotFound(err error) Toast {
	return Toast{
		Message:    err.Error(),
		StatusCode: http.StatusNotFound,
	}
}

func ErrorInternalServerError(err error) Toast {
	return Toast{
		Message:    err.Error(),
		StatusCode: http.StatusInternalServerError,
	}
}
