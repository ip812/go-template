package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Toast struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

func (t Toast) Error() string {
	return fmt.Sprintf("custom error: %s", t.Message)
}

func AddToast(w http.ResponseWriter, t Toast) {
	res, err := json.Marshal(struct {
		Toast Toast `json:"add-toast"`
	}{
		Toast: t,
	})
	if err != nil {
		return
	}
	w.Header().Set("HX-Trigger", string(res))
}

var (
	SuccEmailAddedToMailingList = fmt.Sprintf("your email was added to the mailing list")
)

func SuccessStatusOK(msg string) Toast {
	return Toast{
		Message:    msg,
		StatusCode: http.StatusOK,
	}
}

func SuccessStatusCreated(msg string) Toast {
	return Toast{
		Message:    msg,
		StatusCode: http.StatusCreated,
	}
}

func SuccessStatusNoContent(msg string) Toast {
	return Toast{
		Message:    msg,
		StatusCode: http.StatusNoContent,
	}
}

var (
	WarnEmailIsRequred     = fmt.Errorf("email is required")
	WarnInvalidEmailFormat = fmt.Errorf("provided email is not in valid format")
	WarnEmailAlreadyExists = fmt.Errorf("provided email already exists")
)

func WarningStatusBadRequest(err error) Toast {
	return Toast{
		Message:    err.Error(),
		StatusCode: http.StatusBadRequest,
	}
}

func WarningStatunUnauthorized(err error) Toast {
	return Toast{
		Message:    err.Error(),
		StatusCode: http.StatusUnauthorized,
	}
}

func WarningStatusForbidden(err error) Toast {
	return Toast{
		Message:    err.Error(),
		StatusCode: http.StatusForbidden,
	}
}

var (
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
